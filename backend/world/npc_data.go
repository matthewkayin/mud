package world

import (
	"log"
	"fmt"
	"github.com/mmcdole/lunar"
)

type NpcId int

type NpcData struct {
	// Id is an internal name used for saving the NPC in JSON data
	// Name is the actual display name of the NPC
	// This allows us to have a Goblin whose name is still "Goblin" but who has
	// some kind of special behavior and thus has a different Id
	Id string
	Name string
	description string
	experienceWorth int32
	experienceWorthScaling int32
	respawnDuration int32

	stats StatBlock
	scaling StatBlock
	baseDamage int32
	baseArmorPenetration int32
	baseDeflection int32
	baseDamageResistance int32

	equipment Equipment
	infusion ItemInfusion
	dropTable DropTable
	wares []ItemId

	startingDisposition NpcDisposition
	MovementType NpcMovementType

	// Maps each behavior param name to its type
	BehaviorParams map[string]BehaviorParamType `ts_type:"Record<string, BehaviorParamType>"`

	init *lua.Function
	update *lua.Function
	onAttacked *lua.Function
	onPlayerEntered *lua.Function
	onItemGiven *lua.Function
	getStatusDescription *lua.Function
	onDialog *lua.Function
}

// LOAD

var NPC_DATA []*NpcData
var NPC_ID_STR_TO_ID map[string]NpcId

const NPC_DATA_FOLDER = "npcs"

func (world *World) loadNpcData() {
	// Read npc folder
	log.Printf("Loading NPC data...")
	paths, err := scriptGetFilesFrom(world.dataFolder, NPC_DATA_FOLDER)
	if err != nil {
		log.Fatalf("Error opening NPC data folder: %s", err.Error())
	}

	NPC_DATA = make([]*NpcData, 0, len(paths))
	NPC_ID_STR_TO_ID = make(map[string]NpcId)

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
		_, duplicateNpcId := NPC_ID_STR_TO_ID[npcData.Id]
		if duplicateNpcId {
			log.Fatalf("NPC %s has id '%s' which is a duplicate of another NPC.", path, npcData.Id)
		}

		// Store npc in NPC_DATA
		NPC_ID_STR_TO_ID[npcData.Id] = NpcId(len(NPC_DATA))
		NPC_DATA = append(NPC_DATA, npcData)
		log.Printf("Loaded NPC '%s'.", npcData.Id)
	}

	log.Printf("All NPC data has been loaded.")
}

func (parser *ScriptParser) parseNpc(table *lua.Table) *NpcData {
	npcData := &NpcData{}

	npcData.Id = parser.getString(table, "id")
	npcData.Name = parser.getString(table, "name")
	npcData.description = parser.getString(table, "description")

	npcData.experienceWorth = parser.getInt32(table, "experience_worth")
	parser.checkInt32NonNegative(npcData.experienceWorth, "experience_worth")

	npcData.experienceWorthScaling = parser.getInt32(table, "experience_worth_scaling")
	parser.checkInt32NonNegative(npcData.experienceWorthScaling, "experience_worth_scaling")

	npcData.respawnDuration = parser.getInt32(table, "respawn_duration")
	parser.checkInt32Positive(npcData.respawnDuration, "respawn_duration")

	npcData.stats = parser.getStatBlock(table, "stats", false)
	npcData.scaling = parser.getStatBlock(table, "scaling", false)
	npcData.baseDamage = parser.getInt32(table, "damage")
	npcData.baseArmorPenetration = parser.getInt32(table, "armor_penetration")
	npcData.baseDeflection = parser.getInt32(table, "deflection")
	npcData.baseDamageResistance = parser.getInt32(table, "damage_resistance")

	npcData.equipment = parser.getEquipment(table, "equipment")

	infusion := parser.getOptionalItemInfusion(table, "infusion")
	if infusion == nil {
		npcData.infusion = ItemInfusion{}
	} else {
		npcData.infusion = *infusion
	}

	npcData.dropTable = parser.getDropTable(table, "drop_table")

	var ok bool
	startingDispositionString := parser.getString(table, "starting_disposition")
	npcData.startingDisposition, ok = EnumFromString(startingDispositionString, NpcDisposition(NPC_DISPOSITION_COUNT))
	if !ok {
		parser.addProblem(fmt.Errorf("starting_disposition value '%s' is not a valid NPC disposition.", startingDispositionString))
	}

	movementTypeString := parser.getString(table, "movement_type")
	npcData.MovementType, ok = EnumFromString(movementTypeString, NpcMovementType(NPC_MOVEMENT_TYPE_COUNT))
	if !ok {
		parser.addProblem(fmt.Errorf("movement_type value '%s' is not a valid NPC movement type.", movementTypeString))
	}

	npcData.BehaviorParams = parser.getBehaviorParams(table, "behavior_params")

	// Wares
	npcData.wares = make([]ItemId, 0, 1)
	waresTableValue := table.RawGetString("wares")
	if waresTableValue.Kind() == lua.TableKind {
		waresTable, _ := waresTableValue.AsTable()
		wareCount := waresTable.RawLen()
		npcData.wares = make([]ItemId, 0, wareCount)
		for index := 1; index <= wareCount; index++ {
			wareKey := fmt.Sprintf("wares[%d]", index)
			wareItemName, ok := waresTable.RawGetInt(index).AsString()
			if !ok {
				parser.addProblem(fmt.Errorf("field '%s' must be a string", wareKey))
				continue
			}

			itemId, ok := ITEM_NAME_TO_ID[wareItemName]
			if !ok {
				parser.addProblem(fmt.Errorf("field '%s' (value %s) is not a valid item", wareKey, wareItemName))
				continue
			}

			npcData.wares = append(npcData.wares, itemId)
		}
	} else if !waresTableValue.IsNil() {
		parser.addProblem(fmt.Errorf("invalid type for 'wares'. expected table or nil, got %s", waresTableValue.Kind().String()))
	}

	// Behavior hooks

	npcData.init = parser.getOptionalFunction(table, "init")
	npcData.update = parser.getOptionalFunction(table, "update")
	npcData.onAttacked = parser.getOptionalFunction(table, "on_attacked")
	npcData.onPlayerEntered = parser.getOptionalFunction(table, "on_player_entered")
	npcData.onItemGiven = parser.getOptionalFunction(table, "on_item_given")
	npcData.getStatusDescription = parser.getOptionalFunction(table, "get_status_description")
	npcData.onDialog = parser.getOptionalFunction(table, "on_dialog")

	if parser.getError() != nil {
		return nil
	}

	return npcData
}
