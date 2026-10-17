package world

import (
	"log"
	"slices"
	"encoding/json"
)

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
	Conditions []Condition

	Spells []SpellId
	Inventory Inventory
	Equipment Equipment
}

type MobDataJson struct {
	Name string
	Room int

	Level int32
	Experience int32
	ExperienceToNextLevel int32

	Stats StatBlock
	Abilities []string

	Health int32
	Mana int32
	Conditions []Condition

	Spells []string
	Inventory Inventory
	Equipment Equipment
}

// Returns a copy of the mob data that shares no slices with the original
func (mobData *MobData) Clone() MobData {
	clone := *mobData
	clone.Conditions = slices.Clone(mobData.Conditions)
	clone.Spells = slices.Clone(mobData.Spells)
	clone.Inventory = Inventory {
		Items: slices.Clone(mobData.Inventory.Items),
	}
	clone.Equipment.IsSlotInUse = slices.Clone(mobData.Equipment.IsSlotInUse)
	clone.Equipment.SlotItem = slices.Clone(mobData.Equipment.SlotItem)

	return clone
}

func (mobData *MobData) MarshalJSON() ([]byte, error) {
	mobDataJson := MobDataJson {
		Name: mobData.Name,
		Room: mobData.Room,

		Level: mobData.Level,
		Experience: mobData.Experience,
		ExperienceToNextLevel: mobData.ExperienceToNextLevel,

		Stats: mobData.Stats,
		Abilities: mobData.abilitiesToStringArray(),

		Health: mobData.Health,
		Mana: mobData.Mana,
		Conditions: mobData.Conditions,

		Spells: spellsToStringArray(mobData.Spells),
		Inventory: mobData.Inventory,
		Equipment: mobData.Equipment,
	}
	return json.Marshal(&mobDataJson)
}

func (mobData *MobData) UnmarshalJSON(data []byte) error {
	var mobDataJson MobDataJson
	err := json.Unmarshal(data, &mobDataJson)
	if err != nil {
		return err
	}

	mobData.Name = mobDataJson.Name
	mobData.Room = mobDataJson.Room

	mobData.Level = mobDataJson.Level
	mobData.Experience = mobDataJson.Experience
	mobData.ExperienceToNextLevel = mobDataJson.ExperienceToNextLevel

	mobData.Stats = mobDataJson.Stats
	mobData.setAbilitiesFromStringArray(mobDataJson.Abilities)

	mobData.Health = mobDataJson.Health
	mobData.Mana = mobDataJson.Mana
	mobData.Conditions = mobDataJson.Conditions

	mobData.Spells = spellsFromStringArray(mobDataJson.Spells)
	mobData.Inventory = mobDataJson.Inventory
	mobData.Equipment = mobDataJson.Equipment

	return nil
}

func (mobData *MobData) abilitiesToStringArray() []string {
	results := make([]string, 0, 1)

	for index := range MOB_ABILITY_COUNT {
		if !mobData.HasAbility(MobAbility(index)) {
			continue
		}

		results = append(results, MOB_ABILITY_DATA[index].Name)
	}

	return results
}

func (mobData *MobData) setAbilitiesFromStringArray(abilities []string) {
	for _, abilityName := range abilities {
		var abilityIndex int = -1
		for index := range MOB_ABILITY_COUNT {
			if abilityName == MOB_ABILITY_DATA[index].Name {
				abilityIndex = int(index)
				break
			}
		}

		if abilityIndex == -1 {
			log.Fatalf("No mob ability exists with name '%s'.", abilityName)
		}

		mobData.SetHasAbility(MobAbility(abilityIndex), true)
	}
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
	shield := mobData.Equipment.Get(EQUIPMENT_SLOT_OFF_HAND)
	var armorVal int32 = 0

	if outfit != nil {
		itemData := ITEM_DATA[outfit.Id]
		armorVal += itemData.Data.(*ItemDataArmor).Armor
	}

	if shield != nil {
		itemData := ITEM_DATA[shield.Id]
		armorVal += itemData.Data.(*ItemDataArmor).Armor
	}

	armorVal = mobData.GetConditionStatChange(armorVal, CONDITION_EFFECT_TYPE_STAT_CHANGE_ARMOR)

	return armorVal
}

func (mobData *MobData) GetStat(index int) int32 {
	return mobData.Stats.Values[index] + mobData.Equipment.StatBonuses.Values[index]
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

func (mobData *MobData) HasCondition(conditionId ConditionId) bool {
	for index := range len(mobData.Conditions) {
		if mobData.Conditions[index].Id == conditionId {
			return true
		}
	}

	return false
}

func (mobData *MobData) AddCondition(condition Condition) {
	for index := range len(mobData.Conditions) {
		if mobData.Conditions[index].Id == condition.Id {
			incomingIsLonger :=
				condition.TimeToLive == CONDITION_DURATION_INFINITE ||
				(mobData.Conditions[index].TimeToLive < condition.TimeToLive &&
					mobData.Conditions[index].TimeToLive != CONDITION_DURATION_INFINITE)
			if incomingIsLonger {
				mobData.Conditions[index].TimeToLive = condition.TimeToLive
			}

			return
		}
	}

	mobData.Conditions = append(mobData.Conditions, condition)
}

func (mobData *MobData) RemoveCondition(conditionId ConditionId) {
	for index := range len(mobData.Conditions) {
		if mobData.Conditions[index].Id == conditionId {
			lastIndex := len(mobData.Conditions) - 1
			mobData.Conditions[index] = mobData.Conditions[lastIndex]
			mobData.Conditions = mobData.Conditions[:lastIndex]
			return
		}
	}
}

func (mobData *MobData) GetConditionStatChange(base int32, conditionEffectType ConditionEffectType) int32 {
	percent := float32(1.0)

	for _, condition := range mobData.Conditions {
		conditionData := CONDITION_DATA[condition.Id]

		for index := range len(conditionData.Effects) {
			effect := &conditionData.Effects[index]
			if effect.Type == conditionEffectType {
				statChange := effect.Data.(ConditionStatChange)
				switch statChange.Type {
					case CONDITION_STAT_CHANGE_FLAT:
						base += statChange.Value
					case CONDITION_STAT_CHANGE_PERCENT:
						percent += float32(statChange.Value / 100.0)
				}
			}
		}
	}

	return int32(float32(base) * percent)
}

func (mobData *MobData) GetConditionChanceModifier(base float32, conditionEffectType ConditionEffectType) float32 {
	var chanceModifier int32 = 0.0

	for _, condition := range mobData.Conditions {
		conditionData := CONDITION_DATA[condition.Id]
		for index := range len(conditionData.Effects) {
			effect := &conditionData.Effects[index]
			if effect.Type == conditionEffectType {
				chanceModifier += effect.Data.(int32)
			}
		}
	}

	return float32(chanceModifier) / 100
}
