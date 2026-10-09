package world

import (
	"log"
	"fmt"
	"github.com/mmcdole/lunar"
)

type NpcId int

// The types an NPC script can declare for its behavior params. Each unique NPC in the
// world JSON stores a value of the declared type for each param.
type NpcBehaviorParamType int
const (
	NPC_BEHAVIOR_PARAM_TYPE_STRING = iota
	NPC_BEHAVIOR_PARAM_TYPE_NUMBER
	NPC_BEHAVIOR_PARAM_TYPE_BOOLEAN
	NPC_BEHAVIOR_PARAM_TYPE_ITEM
	NPC_BEHAVIOR_PARAM_TYPE_DIRECTION
	NPC_BEHAVIOR_PARAM_TYPE_COUNT
)

type NpcData struct {
	// Key is an internal name used for saving the NPC in JSON data
	// Name is the actual display name of the NPC
	// This allows us to have a Goblin whose name is still "Goblin" but who has
	// some kind of special behavior and thus has a different Key
	Key string
	Name string
	description string
	experienceWorth int32
	experienceWorthScaling int32
	stats StatBlock
	scaling StatBlock
	equipment Equipment
	dropTable DropTable

	startingDisposition NpcDisposition
	MovementType NpcMovementType

	// Maps each behavior param name to its type
	BehaviorParams map[string]NpcBehaviorParamType `ts_type:"Record<string, NpcBehaviorParamType>"`

	init *lua.Function
	update *lua.Function
	onAttacked *lua.Function
	onPlayerEntered *lua.Function
	onItemGiven *lua.Function
	getStatusDescription *lua.Function
}

func (paramType NpcBehaviorParamType) String() string {
	switch paramType {
		case NPC_BEHAVIOR_PARAM_TYPE_STRING:
			return "String"
		case NPC_BEHAVIOR_PARAM_TYPE_NUMBER:
			return "Number"
		case NPC_BEHAVIOR_PARAM_TYPE_BOOLEAN:
			return "Boolean"
		case NPC_BEHAVIOR_PARAM_TYPE_ITEM:
			return "Item"
		case NPC_BEHAVIOR_PARAM_TYPE_DIRECTION:
			return "Direction"
		default:
			return ""
	}
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
		_, duplicateNpcId := NPC_KEY_TO_ID[npcData.Key]
		if duplicateNpcId {
			log.Fatalf("NPC %s has id '%s' which is a duplicate of another NPC.", path, npcData.Key)
		}

		// Store item in NPC_DATA
		NPC_KEY_TO_ID[npcData.Key] = NpcId(len(NPC_DATA))
		NPC_DATA = append(NPC_DATA, npcData)
		log.Printf("Loaded NPC '%s'.", npcData.Key)
	}

	log.Printf("All NPC data has been loaded.")
}

func (parser *ScriptParser) parseNpc(table *lua.Table) *NpcData {
	npcData := &NpcData{}

	npcData.Key = parser.getString(table, "key")
	npcData.Name = parser.getString(table, "name")
	npcData.description = parser.getString(table, "description")

	npcData.experienceWorth = parser.getInt32(table, "experience_worth")
	parser.checkInt32NonNegative(npcData.experienceWorth, "experience_worth")

	npcData.experienceWorthScaling = parser.getInt32(table, "experience_worth_scaling")
	parser.checkInt32NonNegative(npcData.experienceWorthScaling, "experience_worth_scaling")

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
	npcData.MovementType, ok = EnumFromString(movementTypeString, NpcMovementType(NPC_MOVEMENT_TYPE_COUNT))
	if !ok {
		parser.addProblem(fmt.Errorf("movement_type value '%s' is not a valid NPC movement type.", movementTypeString))
	}

	npcData.BehaviorParams = parser.getNpcBehaviorParams(table, "behavior_params")

	// Behavior hooks

	npcData.init = parser.getOptionalFunction(table, "init")
	npcData.update = parser.getOptionalFunction(table, "update")
	npcData.onAttacked = parser.getOptionalFunction(table, "on_attacked")
	npcData.onPlayerEntered = parser.getOptionalFunction(table, "on_player_entered")
	npcData.onItemGiven = parser.getOptionalFunction(table, "on_item_given")
	npcData.getStatusDescription = parser.getOptionalFunction(table, "get_status_description")

	if parser.getError() != nil {
		return nil
	}

	return npcData
}

// Behavior params are optional. If present, they are a list of { name, type } tables
func (parser *ScriptParser) getNpcBehaviorParams(table *lua.Table, key string) map[string]NpcBehaviorParamType {
	params := map[string]NpcBehaviorParamType{}

	value := table.RawGetString(key)
	if value.IsNil() {
		return params
	}

	paramsTable, ok := value.AsTable()
	if !ok {
		parser.addProblem(fmt.Errorf("invalid type for '%s'. expected table, got %s", key, value.Kind().String()))
		return params
	}

	paramCount := paramsTable.RawLen()
	for index := 1; index <= paramCount; index++ {
		paramKey := fmt.Sprintf("%s[%d]", key, index)
		paramTable, ok := paramsTable.RawGetInt(index).AsTable()
		if !ok {
			parser.addProblem(fmt.Errorf("field '%s' must be a table", paramKey))
			continue
		}

		name := parser.getString(paramTable, "name")
		if name == "" {
			parser.addProblem(fmt.Errorf("field '%s.name' must not be empty", paramKey))
			continue
		}

		typeString := parser.getString(paramTable, "type")
		paramType, ok := EnumFromString(typeString, NpcBehaviorParamType(NPC_BEHAVIOR_PARAM_TYPE_COUNT))
		if !ok {
			parser.addProblem(fmt.Errorf("field '%s.type' value '%s' is not a valid behavior param type.", paramKey, typeString))
			continue
		}

		_, isDuplicate := params[name]
		if isDuplicate {
			parser.addProblem(fmt.Errorf("behavior param '%s' is declared more than once.", name))
			continue
		}

		params[name] = paramType
	}

	return params
}
