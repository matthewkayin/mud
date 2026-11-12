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

func (world *World) loadBackgroundData() error {
	// Read class folder
	paths, err := scriptGetFilesFrom(BACKGROUND_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening background data folder: %s", err.Error())
	}

	world.BackgroundData = make([]*BackgroundData, 0, len(paths))
	world.BackgroundNameToId = make(map[string]BackgroundId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse class data
		parser := ScriptParser{}
		backgroundData := parser.parseBackground(table)
		if backgroundData == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateBackgroundName := world.BackgroundNameToId[backgroundData.Name]
		if duplicateBackgroundName {
			return fmt.Errorf("Background %s has name '%s' which is a duplicate of another background.", path, backgroundData.Name)
		}

		world.BackgroundNameToId[backgroundData.Name] = BackgroundId(len(world.BackgroundData))
		world.BackgroundData = append(world.BackgroundData, backgroundData)
	}

	return nil
}

func (parser *ScriptParser) parseBackground(table *lua.Table) *BackgroundData {
	backgroundData := &BackgroundData{}

	backgroundData.Name = parser.getString(table, "name")
	backgroundData.Description = parser.getString(table, "description")
	backgroundData.Attributes = parser.getAttributeBlock(table, "attributes")
	backgroundData.Skills = parser.getSkillBlock(table, "skills")

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
