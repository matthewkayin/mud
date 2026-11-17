package world

import (
	"fmt"
	"strings"
	"github.com/mmcdole/lunar"
)

type RaceData struct {
	Name string
	Description string
	Stats StatBlock
	Skills SkillFlags
}

const RACE_DATA_FOLDER = "races"

func (world *World) loadRaceData() []ScriptProblem {
	files, problems := world.scriptLoadFolder(RACE_DATA_FOLDER)

	world.RaceData = make(map[string]*RaceData)

	for _, file := range files {
		// Parse race data
		parser := ScriptParser{ path: file.path }
		raceData := parser.parseRace(file.table)
		if raceData == nil {
			problems = append(problems, parser.problems...)
			continue
		}

		// Check for duplicates
		_, duplicateName := world.RaceData[raceData.Name]
		if duplicateName {
			parser.addProblem("name", fmt.Errorf("Race has name '%s' which is a duplicate of another race.", raceData.Name))
			problems = append(problems, parser.problems...)
			continue
		}

		world.RaceData[strings.ToLower(raceData.Name)] = raceData
	}

	return problems
}

func (parser *ScriptParser) parseRace(table *lua.Table) *RaceData {
	raceData := &RaceData{}

	raceData.Name = parser.getString(table, "name")
	raceData.Description = parser.getString(table, "description")
	raceData.Stats = parser.getStatBlock(table, "stats")
	raceData.Skills = parser.getSkillFlags(table, "skills")

	if len(parser.problems) != 0 {
		return nil
	}

	return raceData
}
