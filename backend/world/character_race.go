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

func (world *World) loadRaceData() []ScriptProblem {
	files, problems := world.scriptLoadFolder(RACE_DATA_FOLDER)

	world.RaceData = make([]*RaceData, 0, len(files))
	world.RaceNameToId = make(map[string]RaceId)

	for _, file := range files {
		// Parse race data
		parser := ScriptParser{ path: file.path }
		raceData := parser.parseRace(file.table)
		if raceData == nil {
			problems = append(problems, parser.problems...)
			continue
		}

		// Check for duplicates
		_, duplicateName := world.RaceNameToId[raceData.Name]
		if duplicateName {
			parser.addProblem("name", fmt.Errorf("Race has name '%s' which is a duplicate of another race.", raceData.Name))
			problems = append(problems, parser.problems...)
			continue
		}

		world.RaceNameToId[raceData.Name] = RaceId(len(world.RaceData))
		world.RaceData = append(world.RaceData, raceData)
	}

	return problems
}

func (parser *ScriptParser) parseRace(luaTable *lua.Table) *RaceData {
	raceData := &RaceData{}
	table := newScriptTable(luaTable, "")

	raceData.Name = parser.getString(table, "name")
	raceData.Description = parser.getString(table, "description")
	raceData.Attributes = parser.getAttributeBlock(table, "attributes")
	raceData.Skills = parser.getSkillBlock(table, "skills")

	parser.checkNoUnknownFields(table)

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
