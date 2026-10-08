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
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 8,
				STAT_STR: 10,
				STAT_AGI: 6,
				STAT_INT: 6,
				STAT_FTH: 8,
			},
		},
		Scaling: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 8,
				STAT_STR: 10,
				STAT_AGI: 6,
				STAT_INT: 6,
				STAT_FTH: 8,
			},
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
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 8,
				STAT_STR: 8,
				STAT_AGI: 10,
				STAT_INT: 6,
				STAT_FTH: 8,
			},
		},
		Scaling: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 8,
				STAT_STR: 8,
				STAT_AGI: 10,
				STAT_INT: 6,
				STAT_FTH: 8,
			},
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
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 6,
				STAT_STR: 6,
				STAT_AGI: 8,
				STAT_INT: 10,
				STAT_FTH: 8,
			},
		},
		Scaling: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 6,
				STAT_STR: 6,
				STAT_AGI: 8,
				STAT_INT: 10,
				STAT_FTH: 8,
			},
		},
		UnlocksAtLevel: [MOB_MAX_LEVEL][]ClassUnlock {},
	},

	CLASS_PRIEST: {
		Name: "Priest",
		Stats: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 6,
				STAT_STR: 6,
				STAT_AGI: 8,
				STAT_INT: 8,
				STAT_FTH: 10,
			},
		},
		Scaling: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 6,
				STAT_STR: 6,
				STAT_AGI: 8,
				STAT_INT: 8,
				STAT_FTH: 10,
			},
		},
		UnlocksAtLevel: [MOB_MAX_LEVEL][]ClassUnlock {
			1: {
			},
		},
	},
}
