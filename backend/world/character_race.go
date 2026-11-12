package world

import (
	"fmt"
	"strings"

	"github.com/mmcdole/lunar"
)

type RaceId int

type RaceData struct {
	Name string
	Description string
	Attributes AttributeBlock
	Skills SkillBlock
}

const RACE_DATA_FOLDER = "races"

func (world *World) loadRaceData() error {
	// Read race folder
	paths, err := scriptGetFilesFrom(RACE_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening race data folder: %s", err.Error())
	}

	world.RaceData = make([]*RaceData, 0, len(paths))
	world.RaceNameToId = make(map[string]RaceId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse class data
		parser := ScriptParser{}
		raceData := parser.parseRace(table)
		if raceData == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateClassName := world.BackgroundNameToId[raceData.Name]
		if duplicateClassName {
			return fmt.Errorf("Class %s has name '%s' which is a duplicate of another class.", path, raceData.Name)
		}

		world.RaceNameToId[raceData.Name] = RaceId(len(world.RaceData))
		world.RaceData = append(world.RaceData, raceData)
	}

	return nil
}

func (parser *ScriptParser) parseRace(table *lua.Table) *RaceData {
	raceData := &RaceData{}

	raceData.Name = parser.getString(table, "name")
	raceData.Description = parser.getString(table, "description")
	raceData.Attributes = parser.getAttributeBlock(table, "attributes")
	raceData.Skills = parser.getSkillBlock(table, "skills")

	if len(parser.problems) != 0 {
		return nil
	}

	return raceData
}

func (world *World) RaceIdFromString(name string) (RaceId, error) {
	for raceId, raceData := range world.RaceData {
		if strings.EqualFold(name, raceData.Name) {
			return RaceId(raceId), nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid race.", name)
}
