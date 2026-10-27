package world

import (
	"testing"
)

func primeTestQuests() {
	QUEST_NAMES = []string { "Find the Lost Ring", "Clear the Bridge", "Fetch Water" }
	QUEST_NAME_TO_ID = make(map[string]QuestId)
	for index, questName := range QUEST_NAMES {
		QUEST_NAME_TO_ID[questName] = QuestId(index)
	}
}

func TestQuestMarkComplete(t *testing.T) {
	primeTestQuests()

	character := Character {}
	character.MarkQuestComplete(1)

	for quest := range QuestId(len(QUEST_NAMES)) {
		expected := quest == 1
		if character.IsQuestComplete(quest) != expected {
			t.Errorf("IsQuestComplete(%d) = %t, expected %t", quest, !expected, expected)
		}
	}

	// Marking a quest twice keeps it complete
	character.MarkQuestComplete(1)
	if character.QuestsComplete != 1 << 1 {
		t.Errorf("QuestsComplete = %b, expected %b", character.QuestsComplete, 1 << 1)
	}
}

func TestQuestStringArrayRoundTrip(t *testing.T) {
	primeTestQuests()

	var quests uint32 = (1 << 0) | (1 << 2)
	questNames := questsToStringArray(quests)
	if len(questNames) != 2 || questNames[0] != "Find the Lost Ring" || questNames[1] != "Fetch Water" {
		t.Errorf("questsToStringArray = %v", questNames)
	}

	result := questsFromStringArray(questNames)
	if result != quests {
		t.Errorf("questsFromStringArray = %b, expected %b", result, quests)
	}
}
