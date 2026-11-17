package world

import (
	"fmt"
	"strings"

	"github.com/mmcdole/lunar"
)

type ClassData struct {
	Name string
	Description string
	Stats StatBlock
	Skills SkillFlags
}

const CLASS_DATA_FOLDER = "classes"

func (world *World) loadClassData() []ScriptProblem {
	files, problems := world.scriptLoadFolder(CLASS_DATA_FOLDER)

	world.ClassData = make(map[string]*ClassData)

	for _, file := range files {
		// Parse class data
		parser := ScriptParser{ path: file.path }
		classData := parser.parseClass(file.table)
		if classData == nil {
			problems = append(problems, parser.problems...)
			continue
		}

		// Check for duplicates
		_, duplicateName := world.ClassData[classData.Name]
		if duplicateName {
			parser.addProblem("name", fmt.Errorf("Class has name '%s' which is a duplicate of another class.", classData.Name))
			problems = append(problems, parser.problems...)
			continue
		}

		world.ClassData[strings.ToLower(classData.Name)] = classData
	}

	return problems
}

func (parser *ScriptParser) parseClass(table *lua.Table) *ClassData {
	classData := &ClassData{}

	classData.Name = parser.getString(table, "name")
	classData.Description = parser.getString(table, "description")
	classData.Stats = parser.getStatBlock(table, "stats")
	classData.Skills = parser.getSkillFlags(table, "skills")

	if len(parser.problems) != 0 {
		return nil
	}

	return classData
}
