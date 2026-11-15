package world

import (
	"fmt"
	"strings"

	"github.com/mmcdole/lunar"
)

type BackgroundId int

type BackgroundData struct {
	Name string
	Description string
	Attributes AttributeBlock
	Skills SkillBlock
}

const BACKGROUND_DATA_FOLDER = "backgrounds"

func (world *World) loadBackgroundData() []ScriptProblem {
	files, problems := world.scriptLoadFolder(BACKGROUND_DATA_FOLDER)

	world.BackgroundData = make([]*BackgroundData, 0, len(files))
	world.BackgroundNameToId = make(map[string]BackgroundId)

	for _, file := range files {
		// Parse background data
		parser := ScriptParser{ path: file.path }
		backgroundData := parser.parseBackground(file.table)
		if backgroundData == nil {
			problems = append(problems, parser.problems...)
			continue
		}

		// Check for duplicates
		_, duplicateName := world.BackgroundNameToId[backgroundData.Name]
		if duplicateName {
			parser.addProblem([]string{ "name" }, fmt.Errorf("Background has name '%s' which is a duplicate of another background.", backgroundData.Name))
			problems = append(problems, parser.problems...)
			continue
		}

		world.BackgroundNameToId[backgroundData.Name] = BackgroundId(len(world.BackgroundData))
		world.BackgroundData = append(world.BackgroundData, backgroundData)
	}

	return problems
}

func (parser *ScriptParser) parseBackground(luaTable *lua.Table) *BackgroundData {
	backgroundData := &BackgroundData{}
	table := newScriptTable(luaTable)

	backgroundData.Name = parser.getString(table, "name")
	backgroundData.Description = parser.getString(table, "description")
	backgroundData.Attributes = parser.getAttributeBlock(table, "attributes")
	backgroundData.Skills = parser.getSkillBlock(table, "skills")

	parser.checkNoUnknownFields(table)

	if len(parser.problems) != 0 {
		return nil
	}

	return backgroundData
}

func (world *World) BackgroundIdFromString(name string) (BackgroundId, error) {
	for backgroundId, backgroundData := range world.BackgroundData {
		if strings.EqualFold(name, backgroundData.Name) {
			return BackgroundId(backgroundId), nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid background.", name)
}
