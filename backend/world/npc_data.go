package world

import (
	"log"
	"fmt"
	"github.com/mmcdole/lunar"
)

type NpcId int

type NpcData struct {
	// key is an internal name used for saving the NPC in JSON data
	// name is the actual display name of the NPC
	// This allows us to have a Goblin whose name is still "Goblin" but who has
	// some kind of special behavior and thus has a different key
	key string
	name string
	description string
	experienceWorth int32
	experienceWorthScaling int32
	stats StatBlock
	scaling StatBlock
	equipment Equipment
	dropTable DropTable

	startingDisposition NpcDisposition
	movementType NpcMovementType

	init *lua.Function
	update *lua.Function
	onEvent *lua.Function
	getDescription *lua.Function
}

// LOAD

var NPC_DATA []*NpcData
var NPC_KEY_TO_ID map[string]NpcId

const NPC_DATA_FOLDER = "npcs"

func (world *World) loadNpcData() {
	// Read npc folder
	log.Printf("Loading NPC data...")
	paths, err := scriptGetFilesFrom(world.dataFolder, NPC_DATA_FOLDER)
	if err != nil {
		log.Fatalf("Error opening NPC data folder: %s", err.Error())
	}

	NPC_DATA = make([]*NpcData, 0, len(paths))
	NPC_KEY_TO_ID = make(map[string]NpcId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			log.Fatalf("%s: %s", path, err.Error())
		}

		// Parse item data
		parser := ScriptParser{}
		npcData := parser.parseNpc(table)
		if npcData == nil {
			log.Fatalf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateNpcId := NPC_KEY_TO_ID[npcData.key]
		if duplicateNpcId {
			log.Fatalf("NPC %s has id '%s' which is a duplicate of another NPC.", path, npcData.key)
		}

		// Store item in NPC_DATA
		NPC_KEY_TO_ID[npcData.key] = NpcId(len(NPC_DATA))
		NPC_DATA = append(NPC_DATA, npcData)
		log.Printf("Loaded NPC '%s'.", npcData.key)
	}

	log.Printf("All NPC data has been loaded.")
}

func (parser *ScriptParser) parseNpc(table *lua.Table) *NpcData {
	npcData := &NpcData{}

	npcData.key = parser.getString(table, "key")
	npcData.name = parser.getString(table, "name")

	// Get description, which could be a string or a function
	descriptionValue := table.RawGetString("description")
	if descriptionValue.IsNil() {
		parser.addProblem(fmt.Errorf("Missing required field 'description'"))
	} else if descriptionValue.Kind() == lua.StringKind {
		description, _ := descriptionValue.AsString()
		npcData.description = description
		npcData.getDescription = nil
	} else if descriptionValue.Kind() == lua.FunctionKind {
		descriptionFn, _ := descriptionValue.AsFunction()
		npcData.description = ""
		npcData.getDescription = descriptionFn
	} else {
		parser.addProblem(fmt.Errorf("invalid type for 'description'. expected string or function, got %s", descriptionValue.Kind().String()))
	}

	npcData.experienceWorth = parser.getInt32(table, "experienceWorth")
	parser.checkInt32NonNegative(npcData.experienceWorth, "experienceWorth")

	npcData.experienceWorthScaling = parser.getInt32(table, "experienceWorthScaling")
	parser.checkInt32NonNegative(npcData.experienceWorth, "experienceWorthScaling")

	npcData.stats = parser.getStatBlock(table, "stats", false)
	npcData.scaling = parser.getStatBlock(table, "scaling", false)
	npcData.equipment = parser.getEquipment(table, "equipment")
	npcData.dropTable = parser.getDropTable(table, "drop_table")

	var ok bool
	startingDispositionString := parser.getString(table, "starting_disposition")
	npcData.startingDisposition, ok = EnumFromString(startingDispositionString, NpcDisposition(NPC_DISPOSITION_COUNT))
	if !ok {
		parser.addProblem(fmt.Errorf("starting_disposition value '%s' is not a valid NPC disposition.", startingDispositionString))
	}

	movementTypeString := parser.getString(table, "movement_type")
	npcData.movementType, ok = EnumFromString(movementTypeString, NpcMovementType(NPC_MOVEMENT_TYPE_COUNT))
	if !ok {
		parser.addProblem(fmt.Errorf("movement_type value '%s' is not a valid NPC movement type.", movementTypeString))
	}

	// Behavior hooks

	npcData.init = parser.getFunction(table, "init")
	npcData.update = parser.getFunction(table, "update")
	npcData.onEvent = parser.getFunction(table, "on_event")

	return npcData
}
