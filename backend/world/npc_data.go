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

	stats StatBlock

	// ATTACKS

	damage int32
	// armor penetration score
	penetration int32
	attacksPerRound int32

	// LOOT

	experienceWorth int32
	loot LootTable
	wares []*ItemType

	// BEHAVIOR

	startingDisposition NpcDisposition
	init *lua.Function
	update *lua.Function
}

func (world *World) loadNpcData() []ScriptProblem {
	files, problems := world.scriptLoadFolder(NPC_DATA_FOLDER)

	world.NpcData = make(map[string]*NpcData)

	for _, file := range files {
		// Parse NPC data
		parser := ScriptParser{ path: file.path }
		npcData := parser.parserNpc(world, file.table)
		if npcData == nil {
			problems = append(problems, parser.problems...)
			continue
		}

		// Check for duplicate IDs
		_, duplicateNpcId := world.NpcData[npcData.id]
		if duplicateNpcId {
			parser.addProblem("id", fmt.Errorf("NPC has ID '%s' which is a duplicate of another NPC.", npcData.id))
			problems = append(problems, parser.problems...)
			continue
		}

		world.NpcData[npcData.id] = npcData
	}

	return problems
}

func (parser *ScriptParser) parserNpc(world *World, table *lua.Table) *NpcData {
	npcData := &NpcData{}

	npcData. id = parser.getString(table, "id")
	npcData.name = parser.getString(table, "name")
	npcData.description = parser.getString(table, "description")

	// Stats

	// A stat of 0 means it was not provided
	npcData.stats = parser.getOptionalStatBlock(table, "stats")
	if npcData.stats.Values[STAT_MAX_HEALTH] == 0 {
		npcData.stats.Values[STAT_MAX_HEALTH] = 1
	}
	if npcData.stats.Values[STAT_DEFLECTION_AMOUNT] == 0 {
		npcData.stats.Values[STAT_DEFLECTION_AMOUNT] = 40
	}
	for index := range STAT_COUNT {
		value := npcData.stats.Values[index]
		switch index {
			case STAT_MAX_HEALTH:
				if value <= 0 {
					parser.addProblem("stats", fmt.Errorf("stat '%s' must be greater than 0, got %d", index.String(), value))
				}
			case STAT_DEFLECTION_AMOUNT:
				if value <= 0 || value > 100 {
					parser.addProblem("stats", fmt.Errorf("stat '%s' must be a percent value from 1 to 100, got %d", index.String(), value))
				}
			default:
				if value < 0 {
					parser.addProblem("stats", fmt.Errorf("stat '%s' must be greater than or equal to 0, got %d", index.String(), value))
				}
		}
	}

	// Attacks

	npcData.damage = parser.getOptionalInt32(table, "damage", 1)
	parser.checkInt32NonNegative(npcData.damage, "damage")

	npcData.penetration = parser.getOptionalInt32(table, "penetration", 0)
	parser.checkInt32NonNegative(npcData.penetration, "penetration")

	npcData.attacksPerRound = parser.getOptionalInt32(table, "attacks_per_round", 1)
	parser.checkInt32Positive(npcData.attacksPerRound, "attacks_per_round")

	// Loot

	npcData.experienceWorth = parser.getOptionalInt32(table, "experience_worth", 0)
	parser.checkInt32NonNegative(npcData.experienceWorth, "experience_worth")

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
				parser.addProblem("wares", fmt.Errorf("wares[%d] must be an item name", index))
				continue
			}

			itemType, ok := world.ItemTypes[itemName]
			if !ok {
				parser.addProblem("wares", fmt.Errorf("wares[%d] has value %s which is not a valid item", index, itemName))
				continue
			}

			npcData.wares = append(npcData.wares, itemType)
		}
	}

	// Behavior
	dispositionString := parser.getString(table, "starting_disposition")
	npcData.startingDisposition, ok = util.EnumFromString(dispositionString, NPC_DISPOSITION_COUNT)
	if !ok {
		parser.addProblem("starting_disposition", fmt.Errorf("disposition value '%s' is not a valid NPC disposition", dispositionString))
	}

	npcData.init = parser.getOptionalFunction(table, "init")
	npcData.update = parser.getOptionalFunction(table, "update")

	if len(parser.problems) != 0 {
		return nil
	}

	return npcData
}
