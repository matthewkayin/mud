package world

import (
	"fmt"
)

type QuestId int

// Characters store completed quests in a uint32 bitfield
const QUEST_MAX = 32
const QUEST_DATA_FILE = "quests.lua"

var QUEST_NAMES []string
var QUEST_NAME_TO_ID map[string]QuestId

func (world *World) loadQuestData() error {
	table, err := world.scriptLoadTable(QUEST_DATA_FILE)
	if err != nil {
		return err
	}

	questCount := table.RawLen()
	if questCount > QUEST_MAX {
		return fmt.Errorf("%s: there are %d quests, but the maximum is %d.", QUEST_DATA_FILE, questCount, QUEST_MAX)
	}

	QUEST_NAMES = make([]string, 0, questCount)
	QUEST_NAME_TO_ID = make(map[string]QuestId)

	for index := 1; index <= questCount; index++ {
		questName, ok := table.RawGetInt(index).AsString()
		if !ok {
			return fmt.Errorf("%s: quests[%d] must be a string.", QUEST_DATA_FILE, index)
		}

		// Check for duplicates
		_, duplicateQuestName := QUEST_NAME_TO_ID[questName]
		if duplicateQuestName {
			return fmt.Errorf("%s: quest '%s' is a duplicate of another quest.", QUEST_DATA_FILE, questName)
		}

		QUEST_NAME_TO_ID[questName] = QuestId(len(QUEST_NAMES))
		QUEST_NAMES = append(QUEST_NAMES, questName)
	}

	return nil
}

func (character *Character) IsQuestComplete(quest QuestId) bool {
	return character.QuestsComplete & (1 << quest) != 0
}

func (character *Character) MarkQuestComplete(quest QuestId) {
	character.QuestsComplete |= 1 << quest
}

func questsToStringArray(quests uint32) []string {
	result := make([]string, 0, 1)
	for quest := range len(QUEST_NAMES) {
		if quests & (1 << quest) != 0 {
			result = append(result, QUEST_NAMES[quest])
		}
	}

	return result
}

func questsFromStringArray(questNames []string) (uint32, error) {
	var result uint32 = 0
	for _, questName := range questNames {
		quest, exists := QUEST_NAME_TO_ID[questName]
		if !exists {
			return 0, fmt.Errorf("No quest ID matches quest '%s'.", questName)
		}
		result |= 1 << quest
	}

	return result, nil
}
