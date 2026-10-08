package world

import (
	"fmt"
	"log"

	"github.com/mmcdole/lunar"
)

var SPELL_CAST_TIME_INSTANT int32 = 0

type SpellId int32

type SpellData struct {
	Name string
	Description string
	CastsToLearn int32

	ManaCost int32
	CastTime int32
	CanTargetPlayers bool

	OnHit *lua.Function
}

var SPELL_DATA []*SpellData
var SPELL_NAME_TO_ID map[string]SpellId

const SPELL_DATA_FOLDER = WORLD_DATA_FOLDER + "/spells"

func (world *World) loadSpellData() {
	// Read spell folder
	log.Printf("Loading spell data...")
	paths, err := scriptGetFilesFrom(SPELL_DATA_FOLDER)
	if err != nil {
		log.Fatalf("Error opening spell data folder: %s", err.Error())
	}

	SPELL_DATA = make([]*SpellData, 0, len(paths))
	SPELL_NAME_TO_ID = make(map[string]SpellId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			log.Fatalf("%s: %s", path, err.Error())
		}

		// Parse spell data
		parser := ScriptParser{}
		spellData := parser.parseSpell(table)
		if spellData == nil {
			log.Fatalf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateSpellName := SPELL_NAME_TO_ID[spellData.Name]
		if duplicateSpellName {
			log.Fatalf("Spell %s has name '%s' which is a duplicate of another spell.", path, spellData.Name)
		}

		// Store spell in SPELL_DATA
		SPELL_NAME_TO_ID[spellData.Name] = SpellId(len(SPELL_DATA))
		SPELL_DATA = append(SPELL_DATA, spellData)
		log.Printf("Loaded spell '%s'.", spellData.Name)
	}

	log.Printf("All spell data has been loaded.")
}

func (parser *ScriptParser) parseSpell(table *lua.Table) *SpellData {
	spellData := &SpellData{}

	spellData.Name = parser.getString(table, "name")
	spellData.Description = parser.getString(table, "description")

	spellData.CastsToLearn = parser.getInt32(table, "casts_to_learn")
	if spellData.CastsToLearn <= 0 {
		parser.addProblem(fmt.Errorf("field 'CastsToLearn' must be greater than 0, got %d", spellData.CastsToLearn))
	}

	spellData.ManaCost = parser.getInt32(table, "mana_cost")
	if spellData.ManaCost < 0 {
		parser.addProblem(fmt.Errorf("field 'ManaCost' must not be negative, got %d", spellData.ManaCost))
	}

	spellData.CastTime = parser.getInt32(table, "cast_time")
	if spellData.CastTime < 0 {
		parser.addProblem(fmt.Errorf("field 'CastTime' must not be negative, got %d", spellData.CastTime))
	}

	spellData.CanTargetPlayers = parser.getBool(table, "can_target_players")
	spellData.OnHit = parser.getFunction(table, "on_hit")

	if len(parser.problems) != 0 {
		return nil
	}

	return spellData
}
