package world

import (
	"strings"
	"fmt"

	"github.com/mmcdole/lunar"
)

type RaceId int

type RaceData struct {
	Name string
	Description string
	Stats StatBlock
}

var RACE_DATA []*RaceData
var RACE_NAME_TO_ID map[string]RaceId

const RACE_DATA_FOLDER = "races"

func RaceIdFromString(raceName string) (RaceId, error) {
	for raceId, raceData := range RACE_DATA {
		if strings.EqualFold(raceName, raceData.Name) {
			return RaceId(raceId), nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid race.", raceName)
}

// LOAD

func (world *World) loadRaceData() error {
	// Read race folder
	paths, err := scriptGetFilesFrom(world.dataFolder, RACE_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening race data folder: %s", err.Error())
	}

	RACE_DATA = make([]*RaceData, 0, len(paths))
	RACE_NAME_TO_ID = make(map[string]RaceId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse race data
		parser := ScriptParser{}
		raceData := parser.parseRace(table)
		if raceData == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateRaceName := RACE_NAME_TO_ID[raceData.Name]
		if duplicateRaceName {
			return fmt.Errorf("Race %s has name '%s' which is a duplicate of another race.", path, raceData.Name)
		}

		// Store race in RACE_DATA
		RACE_NAME_TO_ID[raceData.Name] = RaceId(len(RACE_DATA))
		RACE_DATA = append(RACE_DATA, raceData)
	}

	return nil
}

func (parser *ScriptParser) parseRace(table *lua.Table) *RaceData {
	raceData := &RaceData{}

	raceData.Name = parser.getString(table, "name")
	raceData.Description = parser.getString(table, "description")
	// Race stats are modifiers, so they are allowed to be negative
	raceData.Stats = parser.getStatBlock(table, "stats", true)

	if len(parser.problems) != 0 {
		return nil
	}

	return raceData
}
