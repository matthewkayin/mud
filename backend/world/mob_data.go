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

	Health int32
	Mana int32

	Spells []Spell
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
	outfitData := itemData.data.(*ItemDataOutfit)
	return outfitData.armor
}

func (mobData *MobData) Vitality() int32 {
	return mobData.Stats.Vitality + mobData.Equipment.statBonuses.Vitality
}

func (mobData *MobData) Strength() int32 {
	return mobData.Stats.Strength + mobData.Equipment.statBonuses.Strength
}

func (mobData *MobData) Agility() int32 {
	return mobData.Stats.Agility + mobData.Equipment.statBonuses.Agility
}

func (mobData *MobData) Intelligence() int32 {
	return mobData.Stats.Intelligence + mobData.Equipment.statBonuses.Intelligence
}

func (mobData *MobData) Faith() int32 {
	return mobData.Stats.Faith + mobData.Equipment.statBonuses.Faith
}

func (mobData *MobData) SpellSlots() int32 {
	// Spell slots is based on base INT, not INT bonus,
	// otherwise they could equip INT attributes to increase
	// their spell slots, prepare the spells, and then unequip the items
	//
	// There is a case to be made that spell slots should be class determined
	// and separate from the int stat entirely

	return int32(float32(mobData.Stats.Intelligence) / 3.0)
}

func (mobData *MobData) CastsToLearn(spell Spell) int32 {
	spellData := SPELL_DATA[spell]
	spellCastsToLearn := float32(spellData.castsToLearn)
	mobInt := float32(mobData.Intelligence())

	return int32(spellCastsToLearn * (1.0 - (mobInt * MOB_CASTS_TO_LEARN_K)))
}

func (mobData *MobData) RemoveSpell(toRemove Spell) {
	for index, spell := range mobData.Spells {
		if spell == toRemove {
			lastIndex := len(mobData.Spells) - 1
			mobData.Spells[index] = mobData.Spells[lastIndex]
			mobData.Spells = mobData.Spells[:lastIndex]
			return
		}
	}
}
