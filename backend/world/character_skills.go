package world

import "fmt"

type SkillName int

const (
	// VIT
	SKILL_FORTITUDE = iota
	SKILL_HEAVY_ARMOR
	SKILL_SURVIVAL
	SKILL_TAUNT

	// STR
	SKILL_SWORDS
	SKILL_AXES
	SKILL_BLUDGEONING
	SKILL_SHIELDS
	SKILL_BLACKSMITHING
	SKILL_POWER_ATTACK

	// AGI
	SKILL_MARKSMANSHIP
	SKILL_DAGGERS
	SKILL_DUAL_WIELD
	SKILL_SNEAK
	SKILL_EVADE
	SKILL_LIGHT_ARMOR
	SKILL_SECURITY

	// INT
	SKILL_KEEN
	SKILL_SPELLCRAFT
	SKILL_SCRIBE
	SKILL_ALCHEMY
	SKILL_ENCHANTING
	SKILL_PSIONICS
	SKILL_CANTRIPS

	// FTH
	SKILL_MEDITATION
	SKILL_WILLPOWER
	SKILL_EVOCATION
	SKILL_DRUIDRY
	SKILL_CONJURATION

	SKILL_COUNT
)

type SkillBlock struct {
	Values [SKILL_COUNT]int32
}

type SkillData struct {
	Name        string
	Attribute   AttributeName
	Description string
}

var SKILL_DATA []*SkillData = []*SkillData{
	// VIT
	SKILL_FORTITUDE: {
		Name:        "Fortitude",
		Attribute:   ATTRIBUTE_VIT,
		Description: "",
	},
	SKILL_HEAVY_ARMOR: {
		Name:        "Heavy Armor",
		Attribute:   ATTRIBUTE_VIT,
		Description: "",
	},
	SKILL_SURVIVAL: {
		Name:        "Survival",
		Attribute:   ATTRIBUTE_VIT,
		Description: "",
	},
	SKILL_TAUNT: {
		Name:        "Taunt",
		Attribute:   ATTRIBUTE_VIT,
		Description: "",
	},

	// STR
	SKILL_SWORDS: {
		Name:        "Swords",
		Attribute:   ATTRIBUTE_STR,
		Description: "",
	},
	SKILL_AXES: {
		Name:        "Axes",
		Attribute:   ATTRIBUTE_STR,
		Description: "",
	},
	SKILL_BLUDGEONING: {
		Name:        "Bludgeoning",
		Attribute:   ATTRIBUTE_STR,
		Description: "",
	},
	SKILL_SHIELDS: {
		Name:        "Shields",
		Attribute:   ATTRIBUTE_STR,
		Description: "",
	},
	SKILL_BLACKSMITHING: {
		Name:        "Blacksmithing",
		Attribute:   ATTRIBUTE_STR,
		Description: "",
	},
	SKILL_POWER_ATTACK: {
		Name:        "Power Attack",
		Attribute:   ATTRIBUTE_STR,
		Description: "",
	},

	// AGI
	SKILL_MARKSMANSHIP: {
		Name:        "Marksmanship",
		Attribute:   ATTRIBUTE_AGI,
		Description: "",
	},
	SKILL_DAGGERS: {
		Name:        "Daggers",
		Attribute:   ATTRIBUTE_AGI,
		Description: "",
	},
	SKILL_DUAL_WIELD: {
		Name:        "Dual Wield",
		Attribute:   ATTRIBUTE_AGI,
		Description: "",
	},
	SKILL_SNEAK: {
		Name:        "Sneak",
		Attribute:   ATTRIBUTE_AGI,
		Description: "",
	},
	SKILL_EVADE: {
		Name:        "Evade",
		Attribute:   ATTRIBUTE_AGI,
		Description: "",
	},
	SKILL_LIGHT_ARMOR: {
		Name:        "Light Armor",
		Attribute:   ATTRIBUTE_AGI,
		Description: "",
	},
	SKILL_SECURITY: {
		Name:        "Security",
		Attribute:   ATTRIBUTE_AGI,
		Description: "",
	},

	// INT
	SKILL_KEEN: {
		Name:        "Keen",
		Attribute:   ATTRIBUTE_INT,
		Description: "",
	},
	SKILL_SPELLCRAFT: {
		Name:        "Spellcraft",
		Attribute:   ATTRIBUTE_INT,
		Description: "",
	},
	SKILL_SCRIBE: {
		Name:        "Scribe",
		Attribute:   ATTRIBUTE_INT,
		Description: "",
	},
	SKILL_ALCHEMY: {
		Name:        "Alchemy",
		Attribute:   ATTRIBUTE_INT,
		Description: "",
	},
	SKILL_ENCHANTING: {
		Name:        "Enchanting",
		Attribute:   ATTRIBUTE_INT,
		Description: "",
	},
	SKILL_PSIONICS: {
		Name:        "Psionics",
		Attribute:   ATTRIBUTE_INT,
		Description: "",
	},
	SKILL_CANTRIPS: {
		Name:        "Cantrips",
		Attribute:   ATTRIBUTE_INT,
		Description: "",
	},

	// FTH
	SKILL_MEDITATION: {
		Name:        "Meditation",
		Attribute:   ATTRIBUTE_FTH,
		Description: "",
	},
	SKILL_WILLPOWER: {
		Name:        "Willpower",
		Attribute:   ATTRIBUTE_FTH,
		Description: "",
	},
	SKILL_EVOCATION: {
		Name:        "Evocation",
		Attribute:   ATTRIBUTE_FTH,
		Description: "",
	},
	SKILL_DRUIDRY: {
		Name:        "Druidry",
		Attribute:   ATTRIBUTE_FTH,
		Description: "",
	},
	SKILL_CONJURATION: {
		Name:        "Conjuration",
		Attribute:   ATTRIBUTE_FTH,
		Description: "",
	},
}

func skillNameToEnum(name string) (SkillName, error) {
	for index := range SKILL_COUNT {
		if name == SKILL_DATA[index].Name {
			return SkillName(index), nil
		}
	}

	return 0, fmt.Errorf("%s is not a valid skill name.", name)
}

func (skills *SkillBlock) Add(other *SkillBlock) SkillBlock {
	result := *skills
	for index := range SKILL_COUNT {
		result.Values[index] += other.Values[index]
	}

	return result
}

// Returns true if skills >= other
// Use this function for checking guild requirements
func (skills *SkillBlock) Meets(requirements *SkillBlock) bool {
	for index := range SKILL_COUNT {
		if skills.Values[index] < requirements.Values[index] {
			return false
		}
	}

	return true
}
