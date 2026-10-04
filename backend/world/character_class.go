package world

import (
	"fmt"
	"strings"
)

type ClassId int
const (
	CLASS_WARRIOR = iota
	CLASS_THIEF
	CLASS_WIZARD
	CLASS_PRIEST
)

type ClassUnlockType int
const (
	CLASS_UNLOCK_TYPE_ABILITY = iota
	CLASS_UNLOCK_TYPE_SPELL
)

type ClassUnlock struct {
	Type ClassUnlockType
	Data any
}

type ClassData struct {
	Name string
	Stats StatBlock
	Scaling StatBlock
	UnlocksAtLevel [MOB_MAX_LEVEL][]ClassUnlock
}

func ClassIdFromString(name string) (ClassId, error) {
	for classId, classData := range CLASS_DATA {
		if strings.EqualFold(name, classData.Name) {
			return ClassId(classId), nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid class.", name)
}

var CLASS_DATA []*ClassData = []*ClassData {
	CLASS_WARRIOR: {
		Name: "Warrior",
		Stats: StatBlock {
			Vitality: 8,
			Strength: 10,
			Agility: 6,
			Intelligence: 6,
			Faith: 8,
		},
		Scaling: StatBlock {
			Vitality: 8,
			Strength: 10,
			Agility: 6,
			Intelligence: 6,
			Faith: 8,
		},
		UnlocksAtLevel: [MOB_MAX_LEVEL][]ClassUnlock {
			2: {
				{ Type: CLASS_UNLOCK_TYPE_ABILITY,  Data: MOB_ABILITY_TAUNT, },
			},
		},
	},

	CLASS_THIEF: {
		Name: "Thief",
		Stats: StatBlock {
			Vitality: 8,
			Strength: 8,
			Agility: 10,
			Intelligence: 6,
			Faith: 6,
		},
		Scaling: StatBlock {
			Vitality: 8,
			Strength: 8,
			Agility: 10,
			Intelligence: 6,
			Faith: 6,
		},
		UnlocksAtLevel: [MOB_MAX_LEVEL][]ClassUnlock {
			1: {
				{ Type: CLASS_UNLOCK_TYPE_ABILITY,  Data: MOB_ABILITY_SNEAK, },
			},
		},
	},

	CLASS_WIZARD: {
		Name: "Wizard",
		Stats: StatBlock {
			Vitality: 6,
			Strength: 6,
			Agility: 8,
			Intelligence: 10,
			Faith: 8,
		},
		Scaling: StatBlock {
			Vitality: 6,
			Strength: 6,
			Agility: 8,
			Intelligence: 10,
			Faith: 8,
		},
		UnlocksAtLevel: [MOB_MAX_LEVEL][]ClassUnlock {},
	},

	CLASS_PRIEST: {
		Name: "Priest",
		Stats: StatBlock {
			Vitality: 6,
			Strength: 6,
			Agility: 8,
			Intelligence: 8,
			Faith: 10,
		},
		Scaling: StatBlock {
			Vitality: 6,
			Strength: 6,
			Agility: 8,
			Intelligence: 8,
			Faith: 10,
		},
		UnlocksAtLevel: [MOB_MAX_LEVEL][]ClassUnlock {
			1: {
				{ Type: CLASS_UNLOCK_TYPE_SPELL, Data: SPELL_CURE, },
			},
		},
	},
}
