package world

const MOB_MAX_LEVEL int32 = 3
const MOB_EXP_PER_LEVEL int32 = 300

// Making Cast K higher increases how effective intelligence is at reducing spell learn-time
// Spell Casts to learn = Base + (1 - (INT / 50.0) * 0.75)
// 					    => Base * (1 - INT * (0.75 / 50.0))
// Where Base is the spell's base casts to learn
// Higher intelligence therefore reduces the number of casts it takes to learn the spell
const MOB_CASTS_TO_LEARN_K float32 = 0.75 / 50.0

type MobData struct {
	Name string
	Room int

	Level int32
	Experience int32
	ExperienceToNextLevel int32

	Stats StatBlock
	Abilities uint32

	Health int32
	Mana int32

	Spells []SpellId
	Inventory Inventory
	Equipment Equipment
}

func (mobData *MobData) GetExpToNextLevel() int32 {
	if mobData.Level == MOB_MAX_LEVEL {
		return 0
	}

	return mobData.Level * MOB_EXP_PER_LEVEL
}

func (mobData *MobData) MaxHealth() int32 {
	return mobData.Vitality() * 5
}

func (mobData *MobData) MaxMana() int32 {
	return mobData.Intelligence() * 5
}

func (mobData *MobData) Armor() int32 {
	outfit := mobData.Equipment.Get(EQUIPMENT_SLOT_OUTFIT)
	if outfit == nil {
		return 0
	}

	itemData := ITEM_DATA[outfit.Id]
	outfitData := itemData.Data.(*ItemDataOutfit)
	return outfitData.Armor
}

func (mobData *MobData) Vitality() int32 {
	return mobData.Stats.Values[STAT_VIT] + mobData.Equipment.StatBonuses.Values[STAT_VIT]
}

func (mobData *MobData) Strength() int32 {
	return mobData.Stats.Values[STAT_STR] + mobData.Equipment.StatBonuses.Values[STAT_STR]
}

func (mobData *MobData) Agility() int32 {
	return mobData.Stats.Values[STAT_AGI] + mobData.Equipment.StatBonuses.Values[STAT_AGI]
}

func (mobData *MobData) Intelligence() int32 {
	return mobData.Stats.Values[STAT_INT] + mobData.Equipment.StatBonuses.Values[STAT_INT]
}

func (mobData *MobData) Faith() int32 {
	return mobData.Stats.Values[STAT_FTH] + mobData.Equipment.StatBonuses.Values[STAT_FTH]
}

func (mobData *MobData) SpellSlots() int32 {
	// Spell slots is based on base INT, not INT bonus,
	// otherwise they could equip INT attributes to increase
	// their spell slots, prepare the spells, and then unequip the items
	//
	// There is a case to be made that spell slots should be class determined
	// and separate from the int stat entirely

	return int32(float32(mobData.Stats.Values[STAT_INT]) / 3.0)
}

func (mobData *MobData) CastsToLearn(spell SpellId) int32 {
	spellData := SPELL_DATA[spell]
	spellCastsToLearn := float32(spellData.CastsToLearn)
	mobInt := float32(mobData.Intelligence())

	return int32(spellCastsToLearn * (1.0 - (mobInt * MOB_CASTS_TO_LEARN_K)))
}

func (mobData *MobData) RemoveSpell(toRemove SpellId) {
	for index, spell := range mobData.Spells {
		if spell == toRemove {
			lastIndex := len(mobData.Spells) - 1
			mobData.Spells[index] = mobData.Spells[lastIndex]
			mobData.Spells = mobData.Spells[:lastIndex]
			return
		}
	}
}

func (mobData *MobData) HasAbility(ability MobAbility) bool {
	var abilityFlag uint32 = 1 << ability
	return (mobData.Abilities & abilityFlag) == abilityFlag
}

func (mobData *MobData) SetHasAbility(ability MobAbility, value bool) {
	var abilityFlag uint32 = 1 << ability
	if value {
		mobData.Abilities |= abilityFlag
	} else {
		mobData.Abilities &= ^abilityFlag
	}
}

func (mobData *MobData) GetAbilityList() []MobAbility {
	list := make([]MobAbility, 0, 1)
	for abilityIndex := range MOB_ABILITY_COUNT {
		ability := MobAbility(abilityIndex)
		if mobData.HasAbility(ability) {
			list = append(list, ability)
		}
	}

	return list
}
