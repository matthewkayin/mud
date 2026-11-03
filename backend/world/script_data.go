package world

import (
	"sync"
)

// The script data globals (ITEM_DATA, ITEM_NAME_TO_ID, ...) as they were when the snapshot was taken.
// Loading a world overwrites the globals, so WorldLoad takes a snapshot first and restores it if the load fails.
//
// Snapshots only copy the slice headers and map pointers, so a loader must always assign a new slice or map
// to a global (e.g. ITEM_DATA = make(...)) and never clear or truncate the existing one in place,
// or it would change the data held by the snapshot too.
type scriptDataSnapshot struct {
	world *World

	questNames []string
	questNameToId map[string]QuestId
	raceData []*RaceData
	raceNameToId map[string]RaceId
	jobData []*JobData
	jobNameToId map[string]JobId
	jobEnchanter JobId
	conditionData []*ConditionData
	conditionNameToId map[string]ConditionId
	spellData []*SpellData
	spellNameToId map[string]SpellId
	recipeData []*RecipeData
	recipeNameToId map[string]RecipeId
	itemData []*ItemData
	itemNameToId map[string]ItemId
	itemGold ItemId
	classData []*ClassData
	classNameToId map[string]ClassId
	npcData []*NpcData
	npcIdStrToId map[string]NpcId
	fixtureData []*FixtureData
	fixtureNameToId map[string]FixtureId
}

// Guards the script data globals against readers outside the game loop (the editor's HTTP handlers).
// Only the game loop writes the globals, and it holds the write lock while it does,
// so code running on the game loop can read them without locking.
var scriptDataLock sync.RWMutex

// Must be held by any goroutine other than the game loop while it reads script data
func LockScriptDataForReading() {
	scriptDataLock.RLock()
}

func UnlockScriptDataForReading() {
	scriptDataLock.RUnlock()
}

func captureScriptData() scriptDataSnapshot {
	return scriptDataSnapshot {
		world: __world,

		questNames: QUEST_NAMES,
		questNameToId: QUEST_NAME_TO_ID,
		raceData: RACE_DATA,
		raceNameToId: RACE_NAME_TO_ID,
		jobData: JOB_DATA,
		jobNameToId: JOB_NAME_TO_ID,
		jobEnchanter: JOB_ENCHANTER,
		conditionData: CONDITION_DATA,
		conditionNameToId: CONDITION_NAME_TO_ID,
		spellData: SPELL_DATA,
		spellNameToId: SPELL_NAME_TO_ID,
		recipeData: RECIPE_DATA,
		recipeNameToId: RECIPE_NAME_TO_ID,
		itemData: ITEM_DATA,
		itemNameToId: ITEM_NAME_TO_ID,
		itemGold: ITEM_GOLD,
		classData: CLASS_DATA,
		classNameToId: CLASS_NAME_TO_ID,
		npcData: NPC_DATA,
		npcIdStrToId: NPC_ID_STR_TO_ID,
		fixtureData: FIXTURE_DATA,
		fixtureNameToId: FIXTURE_NAME_TO_STRING,
	}
}

func (snapshot *scriptDataSnapshot) restore() {
	__world = snapshot.world

	QUEST_NAMES = snapshot.questNames
	QUEST_NAME_TO_ID = snapshot.questNameToId
	RACE_DATA = snapshot.raceData
	RACE_NAME_TO_ID = snapshot.raceNameToId
	JOB_DATA = snapshot.jobData
	JOB_NAME_TO_ID = snapshot.jobNameToId
	JOB_ENCHANTER = snapshot.jobEnchanter
	CONDITION_DATA = snapshot.conditionData
	CONDITION_NAME_TO_ID = snapshot.conditionNameToId
	SPELL_DATA = snapshot.spellData
	SPELL_NAME_TO_ID = snapshot.spellNameToId
	RECIPE_DATA = snapshot.recipeData
	RECIPE_NAME_TO_ID = snapshot.recipeNameToId
	ITEM_DATA = snapshot.itemData
	ITEM_NAME_TO_ID = snapshot.itemNameToId
	ITEM_GOLD = snapshot.itemGold
	CLASS_DATA = snapshot.classData
	CLASS_NAME_TO_ID = snapshot.classNameToId
	NPC_DATA = snapshot.npcData
	NPC_ID_STR_TO_ID = snapshot.npcIdStrToId
	FIXTURE_DATA = snapshot.fixtureData
	FIXTURE_NAME_TO_STRING = snapshot.fixtureNameToId
}
