package world

import (
	"fmt"
	"log"
	"mud/util"
)

type SkillFlags uint32

type Skill int
const (
	SKILL_HEAVY_ARMOR Skill = iota
	SKILL_LIGHT_ARMOR
	SKILL_TAUNT
	SKILL_SWORDS
	SKILL_AXES
	SKILL_SHIELDS
	SKILL_BLUNT
	SKILL_BOWS
	SKILL_DAGGERS
	SKILL_SNEAK
	SKILL_SECURITY
	SKILL_METAMAGIC
	SKILL_CANTRIPS
	SKILL_PSIONICS
	SKILL_EVOCATION
	SKILL_DRUIDRY
	SKILL_CONJURATION
	SKILL_NECROMANCY
	SKILL_PRAYER
	SKILL_COUNT
)

func (skill Skill) String() string {
	switch skill {
		case SKILL_HEAVY_ARMOR:
			return "Heavy Armor"
		case SKILL_LIGHT_ARMOR:
			return "Light Armor"
		case SKILL_TAUNT:
			return "Taunt"
		case SKILL_SWORDS:
			return "Swords"
		case SKILL_AXES:
			return "Axes"
		case SKILL_SHIELDS:
			return "Shields"
		case SKILL_BLUNT:
			return "Blunt"
		case SKILL_BOWS:
			return "Bows"
		case SKILL_DAGGERS:
			return "Daggers"
		case SKILL_SNEAK:
			return "Sneak"
		case SKILL_SECURITY:
			return "Security"
		case SKILL_METAMAGIC:
			return "Metamagic"
		case SKILL_CANTRIPS:
			return "Cantrips"
		case SKILL_PSIONICS:
			return "Psionics"
		case SKILL_EVOCATION:
			return "Evocation"
		case SKILL_DRUIDRY:
			return "Druidry"
		case SKILL_CONJURATION:
			return "Conjuration"
		case SKILL_NECROMANCY:
			return "Necromancy"
		case SKILL_PRAYER:
			return "Prayer"
		default:
			log.Printf("warn: skill %d not handled in String()", skill)
			return ""
	}
}

func (flags SkillFlags) Has(skill Skill) bool {
	flag := SkillFlags(1) << skill
	return (flags & flag) == flag
}

func (flags *SkillFlags) Grant(skill Skill) {
	*flags |= SkillFlags(1) << skill
}

func (flags SkillFlags) ToStringArray() []string {
	skills := make([]string, 0, 1)
	for index := range SKILL_COUNT {
		skill := Skill(index)
		if flags.Has(skill) {
			skills = append(skills, skill.String())
		}
	}

	return skills
}

func skillFlagsFromStringArray(skills []string) (SkillFlags, error) {
	flags := SkillFlags(0)
	for _, skillName := range skills {
		skill, ok := util.EnumFromString[Skill](skillName, SKILL_COUNT)
		if !ok {
			return flags, fmt.Errorf("Unrecognized skill '%s'.", skillName)
		}

		flags.Grant(skill)
	}

	return flags, nil
}
