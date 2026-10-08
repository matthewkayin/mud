package world

import (
	"fmt"
	"log"
	"strings"

	"github.com/mmcdole/lunar"
)

type ClassId int

type ClassUnlockType int
const (
	CLASS_UNLOCK_TYPE_ABILITY = iota
	CLASS_UNLOCK_TYPE_SPELL
)

type ClassUnlock struct {
	Type ClassUnlockType
	Data any
}

type ClassData struct {
	Name string
	Stats StatBlock
	Scaling StatBlock
	UnlocksAtLevel [MOB_MAX_LEVEL][]ClassUnlock
}

var CLASS_DATA []*ClassData
var CLASS_NAME_TO_ID map[string]ClassId

const CLASS_DATA_FOLDER = "classes"

func ClassIdFromString(name string) (ClassId, error) {
	for classId, classData := range CLASS_DATA {
		if strings.EqualFold(name, classData.Name) {
			return ClassId(classId), nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid class.", name)
}

// LOAD

// Must be called after loadSpellData and loadItemData, since classes reference spells and items by name
func (world *World) loadClassData() {
	// Read class folder
	log.Printf("Loading class data...")
	paths, err := scriptGetFilesFrom(world.dataFolder, CLASS_DATA_FOLDER)
	if err != nil {
		log.Fatalf("Error opening class data folder: %s", err.Error())
	}

	CLASS_DATA = make([]*ClassData, 0, len(paths))
	CLASS_NAME_TO_ID = make(map[string]ClassId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			log.Fatalf("%s: %s", path, err.Error())
		}

		// Parse class data
		parser := ScriptParser{}
		classData := parser.parseClass(table)
		if classData == nil {
			log.Fatalf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateClassName := CLASS_NAME_TO_ID[classData.Name]
		if duplicateClassName {
			log.Fatalf("Class %s has name '%s' which is a duplicate of another class.", path, classData.Name)
		}

		// Store class in CLASS_DATA
		CLASS_NAME_TO_ID[classData.Name] = ClassId(len(CLASS_DATA))
		CLASS_DATA = append(CLASS_DATA, classData)
		log.Printf("Loaded class '%s'.", classData.Name)
	}

	log.Printf("All class data has been loaded.")
}

func (parser *ScriptParser) parseClass(table *lua.Table) *ClassData {
	classData := &ClassData{}

	classData.Name = parser.getString(table, "name")
	classData.Stats = parser.getStatBlock(table, "stats", false)
	classData.Scaling = parser.getStatBlock(table, "scaling", false)

	// Unlocks
	unlocksTable := parser.getTable(table, "unlocks")
	if unlocksTable != nil {
		unlockCount := unlocksTable.RawLen()
		for index := 1; index <= unlockCount; index++ {
			key := fmt.Sprintf("unlocks[%d]", index)
			unlockTable, ok := unlocksTable.RawGetInt(index).AsTable()
			if !ok {
				parser.addProblem(fmt.Errorf("field '%s' must be a table", key))
				continue
			}

			level, unlock, ok := parser.parseClassUnlock(unlockTable, key)
			if !ok {
				continue
			}
			classData.UnlocksAtLevel[level] = append(classData.UnlocksAtLevel[level], unlock)
		}
	}

	if len(parser.problems) != 0 {
		return nil
	}

	return classData
}

// Parses a table of the form { level = N, ability = "<ability name>" } or { level = N, spell = "<spell name>" }
func (parser *ScriptParser) parseClassUnlock(table *lua.Table, key string) (int32, ClassUnlock, bool) {
	level := parser.getInt32(table, "level")
	if level < 1 || level > MOB_MAX_LEVEL {
		parser.addProblem(fmt.Errorf("field '%s' level must be between 1 and %d, got %d", key, MOB_MAX_LEVEL - 1, level))
		return 0, ClassUnlock{}, false
	}

	abilityValue := table.RawGetString("ability")
	spellValue := table.RawGetString("spell")
	if abilityValue.IsNil() == spellValue.IsNil() {
		parser.addProblem(fmt.Errorf("field '%s' must set exactly one of 'ability' or 'spell'", key))
		return 0, ClassUnlock{}, false
	}

	if !abilityValue.IsNil() {
		abilityName := parser.getString(table, "ability")
		ability, exists := MobAbilityFromName(abilityName)
		if !exists {
			parser.addProblem(fmt.Errorf("field '%s' references ability '%s' which does not exist", key, abilityName))
			return 0, ClassUnlock{}, false
		}

		return level, ClassUnlock { Type: CLASS_UNLOCK_TYPE_ABILITY, Data: ability }, true
	}

	spellName := parser.getString(table, "spell")
	spell, exists := SPELL_NAME_TO_ID[spellName]
	if !exists {
		parser.addProblem(fmt.Errorf("field '%s' references spell '%s' which does not exist", key, spellName))
		return 0, ClassUnlock{}, false
	}

	return level, ClassUnlock { Type: CLASS_UNLOCK_TYPE_SPELL, Data: spell }, true
}
