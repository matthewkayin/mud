package world

import (
	"fmt"
	"strconv"
	"github.com/mmcdole/lunar"
)

const SPELL_CAST_TIME_INSTANT int32 = 0

type Spell struct {
	Name string
	Description string
	RequiredSkill Skill

	ManaCost int32
	CastTime int32
	TargetType TargetType

	ValidateTarget *lua.Function
	onCast *lua.Function
}

func (spell *Spell) DescribeToPlayer(world *World, playerKey string) {
	world.MessagePlayer(playerKey, fmt.Sprintf("%s (%s)", spell.Name, spell.RequiredSkill.String()))
	world.MessagePlayer(playerKey, fmt.Sprintf("Cost: %d | Cast Time: %s", spell.ManaCost, spellCastTimeString(spell.CastTime)))
	world.MessagePlayer(playerKey, spell.Description)
}

func spellCastTimeString(castTime int32) string {
	if castTime == SPELL_CAST_TIME_INSTANT {
		return "Instant"
	}

	return strconv.Itoa(int(castTime))
}

func (world *World) SpellListToStringArray(spells []*Spell) []string {
	if len(spells) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(spells))
	for index := range len(spells) {
		result = append(result, spells[index].Name)
	}

	return result
}

func (world *World) spellListFromStringArray(spells []string) ([]*Spell, error) {
	if len(spells) == 0 {
		return []*Spell{}, nil
	}

	result := make([]*Spell, 0, len(spells))
	for index := range len(spells) {
		spell, exists := world.Spells[spells[index]]
		if !exists {
			return result, fmt.Errorf("Unrecognized spell '%s'.", spells[index])
		}
		result = append(result, spell)
	}

	return result, nil
}

const SPELL_DATA_FOLDER = "spells"

func (world *World) loadSpells() []ScriptProblem {
	files, problems := world.scriptLoadFolder(SPELL_DATA_FOLDER)

	world.Spells = make(map[string]*Spell)

	for _, file := range files {
		// Parse spell
		parser := ScriptParser{ path: file.path }
		spell := parser.parseSpell(file.table)
		if spell == nil {
			problems = append(problems, parser.problems...)
			continue
		}

		// Check for duplicates
		_, duplicateSpell := world.Spells[spell.Name]
		if duplicateSpell {
			parser.addProblem("name", fmt.Errorf("Spell has name '%s' which is a duplicate of another spell.", spell.Name))
			problems = append(problems, parser.problems...)
			continue
		}

		world.Spells[spell.Name] = spell
	}

	return problems
}

func (parser *ScriptParser) parseSpell(table *lua.Table) *Spell {
	spell := &Spell{}

	spell.Name = parser.getString(table, "name")
	spell.Description = parser.getString(table, "description")
	spell.RequiredSkill = parser.getSkill(table, "required_skill")

	spell.ManaCost = parser.getInt32(table, "mana_cost")
	parser.checkInt32NonNegative(spell.ManaCost, "mana_cost")

	spell.CastTime = parser.getInt32(table, "cast_time")
	parser.checkInt32NonNegative(spell.CastTime, "cast_time")

	spell.TargetType = parser.getTargetType(table, "target_type")
	if spell.TargetType == TARGET_TYPE_ITEM || spell.TargetType == TARGET_TYPE_ARGUMENT {
		spell.ValidateTarget = parser.getOptionalFunction(table, "validate_target")
		if spell.ValidateTarget == nil {
			parser.addProblem("validate_target", fmt.Errorf("'validate_target' is a required field when target_type is 'Item' or 'Argument'."))
		}
	}

	spell.onCast = parser.getFunction(table, "on_cast")

	if len(parser.problems) != 0 {
		return nil
	}

	return spell
}
