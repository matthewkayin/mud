package world

import (
	"log"
	"os"
	"strings"
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

func (world *World) loadSpells() {
	// Spells
	log.Printf("Loading spell data...")
	files, err := os.ReadDir(WORLD_SPELLS_FOLDER)
	if err != nil {
		log.Fatalf("Error opening spells folder: %s", err.Error())
	}

	SPELL_DATA = make([]*SpellData, 0, len(files))
	SPELL_NAME_TO_ID = make(map[string]SpellId)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".lua") {
			log.Printf("Skipping non-lua file %s in spells folder.", file.Name())
			continue
		}

		path := WORLD_SPELLS_FOLDER + "/" + file.Name()
		path = path[len(WORLD_DATA_FOLDER) + 1:]
		spellData, err := world.scriptLoadSpell(path)
		if err != nil {
			log.Fatal(err.Error())
		}

		_, duplicateSpellName := SPELL_NAME_TO_ID[spellData.Name]
		if duplicateSpellName {
			log.Fatalf("Spell %s has name '%s' which is a duplicate of another spell.", path, spellData.Name)
		}

		SPELL_NAME_TO_ID[spellData.Name] = SpellId(len(SPELL_DATA))
		SPELL_DATA = append(SPELL_DATA, spellData)
		log.Printf("Loaded spell %s.", file.Name())
	}
	log.Printf("All spell data has been loaded.\n")
}
