package world

import (
	"fmt"
	"strings"

	"github.com/mmcdole/lunar"
)

type ClassId int

type ClassData struct {
	Name string
	Description string
}

const CLASS_DATA_FOLDER = "classes"

func (world *World) loadClassData() error {
	// Read class folder
	paths, err := scriptGetFilesFrom(CLASS_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening class data folder: %s", err.Error())
	}

	world.ClassData = make([]*ClassData, 0, len(paths))
	world.ClassNameToId = make(map[string]ClassId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse class data
		parser := ScriptParser{}
		classData := parser.parseClass(table)
		if classData == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateClassName := world.ClassNameToId[classData.Name]
		if duplicateClassName {
			return fmt.Errorf("Class %s has name '%s' which is a duplicate of another class.", path, classData.Name)
		}

		world.ClassNameToId[classData.Name] = ClassId(len(world.ClassData))
		world.ClassData = append(world.ClassData, classData)
	}

	return nil
}

func (parser *ScriptParser) parseClass(table *lua.Table) *ClassData {
	classData := &ClassData{}

	classData.Name = parser.getString(table, "name")
	classData.Description = parser.getString(table, "description")

	if len(parser.problems) != 0 {
		return nil
	}

	return classData
}

func (world *World) classIdFromString(name string) (ClassId, error) {
	for classId, classData := range world.ClassData {
		if strings.EqualFold(name, classData.Name) {
			return ClassId(classId), nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid class.", name)
}
