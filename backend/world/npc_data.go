package world

import (
	"fmt"
	"mud/util"

	"github.com/mmcdole/lunar"
)

const NPC_DATA_FOLDER = "npcs"

type NpcDisposition int
const (
	NPC_DISPOSITION_NEUTRAL NpcDisposition = iota
	NPC_DISPOSITION_FRIENDLY
	NPC_DISPOSITION_HOSTILE
	NPC_DISPOSITION_COUNT
)

func (disposition NpcDisposition) String() string {
	switch disposition {
		case NPC_DISPOSITION_NEUTRAL:
			return "Neutral"
		case NPC_DISPOSITION_FRIENDLY:
			return "Friendly"
		case NPC_DISPOSITION_HOSTILE:
			return "Hostile"
		default:
			return ""
	}
}

type NpcData struct {
	id string
	name string
	description string

	// STATS

	maxHealth int32
	maxMana int32

	damage int32
	// armor penetration score
	penetration int32
	// flat armor resistance
	resistance int32
	// deflection score
	deflection int32
	// amount of damage to deflect on a successful deflect
	deflectionAmount int32

	// LOOT

	experienceWorth int32
	loot LootTable
	wares []*ItemType

	// BEHAVIOR

	startingDisposition NpcDisposition
	init *lua.Function
	update *lua.Function
}

func (world *World) loadNpcData() error {
	// Read NPC folder
	paths, err := scriptGetFilesFrom(NPC_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening NPC data folder: %s", err.Error())
	}

	world.NpcData = make([]*NpcData, 0, len(paths))
	world.NpcIdToData = make(map[string]*NpcData)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse NPC data
		parser := ScriptParser{}
		npcData := parser.parserNpc(world, table)
		if npcData == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		world.NpcData = append(world.NpcData, npcData)
	}

	// Set the NPC ID to type mapping
	for index := range len(world.NpcData) {
		npcData := world.NpcData[index]

		// Check for duplicate IDs
		_, duplicateNpcId := world.NpcIdToData[npcData.id]
		if duplicateNpcId {
			return fmt.Errorf("NPC has ID '%s' which is a duplicate of another NPC.", npcData.id)
		}

		world.NpcIdToData[npcData.id] = npcData
	}

	return nil
}

func (parser *ScriptParser) parserNpc(world *World, table *lua.Table) *NpcData {
	npcData := &NpcData{}

	npcData. id = parser.getString(table, "id")
	npcData.name = parser.getString(table, "name")
	npcData.description = parser.getString(table, "description")

	// Stats

	npcData.maxHealth = parser.getOptionalInt32(table, "max_health", 1)
	parser.checkInt32Positive(npcData.maxHealth, "max_health")

	npcData.maxMana = parser.getOptionalInt32(table, "max_mana", 0)
	parser.checkInt32Positive(npcData.maxMana, "max_mana")

	npcData.damage = parser.getOptionalInt32(table, "damage", 0)
	parser.checkInt32Positive(npcData.damage, "damage")

	npcData.penetration = parser.getOptionalInt32(table, "penetration", 0)
	parser.checkInt32Positive(npcData.penetration, "penetration")

	npcData.resistance = parser.getOptionalInt32(table, "resistance", 0)
	parser.checkInt32Positive(npcData.resistance, "resistance")

	npcData.deflection = parser.getOptionalInt32(table, "deflection", 0)
	parser.checkInt32Positive(npcData.deflection, "deflection")

	npcData.deflectionAmount = parser.getOptionalInt32(table, "deflection_amount", 60)
	parser.checkInt32IsPositivePercent(npcData.deflectionAmount, "deflection_amount")

	// Loot

	npcData.experienceWorth = parser.getInt32(table, "experience_worth")
	parser.checkInt32Positive(npcData.experienceWorth, "experience_worth")

	npcData.loot = parser.getOptionalLootTable(world, table, "loot")

	// Wares
	npcData.wares = make([]*ItemType, 0, 1)
	waresTable, ok := table.RawGetString("wares").AsTable()
	if ok {
		wareCount := waresTable.RawLen()
		npcData.wares = make([]*ItemType, 0, wareCount)

		for index := 1; index <= wareCount; index++ {
			itemName, ok := waresTable.RawGetInt(index).AsString()
			if !ok {
				parser.addProblem(fmt.Errorf("wares[%d] must be an item name", index))
				continue
			}

			itemType, ok := world.ItemNameToType[itemName]
			if !ok {
				parser.addProblem(fmt.Errorf("wares[%d] has value %s which is not a valid item", index, itemName))
				continue
			}

			npcData.wares = append(npcData.wares, itemType)
		}
	}

	// Behavior
	dispositionString := parser.getString(table, "starting_disposition")
	npcData.startingDisposition, ok = util.EnumFromString(dispositionString, NPC_DISPOSITION_COUNT)
	if !ok {
		parser.addProblem(fmt.Errorf("disposition value '%s' is not a valid NPC disposition", dispositionString))
	}

	npcData.init = parser.getOptionalFunction(table, "init")
	npcData.update = parser.getOptionalFunction(table, "update")

	if len(parser.problems) != 0 {
		return nil
	}

	return npcData
}
