package world

import (
	"fmt"
	"math"
	"mud/bitset"
)

type Character struct {
	PlayerKey string

	Race *RaceData
	Class *ClassData

	Level int32
	Experience int32

	Stats StatBlock
	Skills SkillFlags

	Equipment Equipment

	RoomsDiscovered []byte

	MobData
}

type CharacterSheet struct {
	Name string

	Race *RaceData
	Class *ClassData
}

func (world *World) NewCharacter(playerKey string, characterSheet *CharacterSheet) *Character {
	character := &Character{}

	character.PlayerKey = playerKey
	character.Name = characterSheet.Name
	character.Race = characterSheet.Race
	character.Class = characterSheet.Class

	// Set attributes and skills
	character.Level = 1
	character.Stats = character.Race.Stats.Add(&character.Class.Stats)
	character.Skills = character.Race.Skills | character.Class.Skills

	character.RoomsDiscovered = bitset.New(len(world.Rooms))

	// Set the character's room to the start room
	character.Room = world.StartRoom
	bitset.Set(character.RoomsDiscovered, character.Room, true)

	// Set health and mana
	character.Health = character.MaxHealth()
	character.Mana = character.MaxMana()

	character.Equipment = NewEquipment()

	return character
}

// MOB OWNER FUNCTIONS

func (character *Character) IsPlayer() bool {
	return true
}

func (character *Character) IsNpc() bool {
	return false
}

func (character *Character) GetPlayerKey() string {
	return character.PlayerKey
}

func (character *Character) getMobData() *MobData {
	return &character.MobData
}

func (character *Character) GetFaction() MobFaction {
	return MOB_FACTION_ALLIANCE
}

func (character *Character) DescribeToPlayer(world *World, playerKey string) {
	world.MessagePlayer(playerKey, fmt.Sprintf("<y>%s</y> is a level %d %s %s.",
		character.Name, character.Level, character.Race.Name, character.Class.Name))
}

func (character *Character) MaxHealth() int32 {
	return 10 * character.Stats.Values[STAT_VIT]
}

func (character *Character) MaxMana() int32 {
	return 10 * character.Stats.Values[STAT_INT]
}

func (character *Character) InventoryCapacity() int32 {
	return INVENTORY_CAPACITY_PLAYER
}

func (character *Character) Message(world *World, message string) {
	world.MessagePlayer(character.PlayerKey, message)
}

func (character *Character) update(world *World) {
	// npc-only function, does nothing
}

// CHARACTER FUNCTIONS

func (character *Character) MaxSpellSlots() int32 {
	return int32(character.Stats.Values[STAT_INT] / 4)
}

func ExpToNextLevel(level int32) int32 {
	return int32(756.0 * math.Pow(1.26, float64(level - 1)))
}

func (character *Character) ExpTowardNextLevel() int32 {
	if character.Level == 1 {
		return character.Experience
	}
	return character.Experience - ExpToNextLevel(character.Level - 1)
}

func LevelFromExpTotal(expTotal int32) int32 {
	level := int32(1)
	for expTotal > ExpToNextLevel(level) {
		level++
	}
	return level
}

func (character *Character) grantExperience(world *World, amount int32) {
	character.Experience += amount
	for character.Experience > ExpToNextLevel(character.Level) {
		character.Level++
		character.calculateStats()
		world.MessagePlayer(character.PlayerKey, fmt.Sprintf("Level up! You are now level %d.", character.Level))
	}
}

func (character *Character) calculateStats() {
	baseStats := character.Race.Stats.Add(&character.Class.Stats)
	character.Stats = baseStats.ScaledToLevel(character.Level)
}

func (character *Character) subtractDurabilityFromEquipment(world *World, slot EquipmentSlot) {
	item := character.Equipment.Get(slot)
	if item == nil {
		return
	}

	maxDurability := item.MaxDurability()
	itemWasDamaged := item.Durability < maxDurability / 2
	item.Durability--

	if item.Durability <= 0 {
		world.MessageRoom(character.Room, fmt.Sprintf("<y>%s's</y> <g>%s</g> broke!", character.Name, item.Name))
		character.Equipment.Unequip(slot)
		return
	}

	// If item has become damaged, tell the user
	itemIsDamaged := item.Durability < maxDurability / 2
	if itemIsDamaged && !itemWasDamaged {
		character.Message(world, fmt.Sprintf("Your <g>%s</g> is now damaged.", item.Name))
		return
	}

	// If an item has lost its sharpness, tell the user
	itemLostSharpness := item.Durability == maxDurability
	if itemLostSharpness {
		itemIsWeapon :=
			item.IsOneHandedWeapon() ||
			item.IsTwoHandedWeapon()
		if itemIsWeapon {
			character.Message(world, fmt.Sprintf("Your <g>%s</g> has lost its sharpness.", item.Name))
		} else {
			character.Message(world, fmt.Sprintf("Your <g>%s</g> has lost its fortification.", item.Name)) }
	}
}

func (character *Character) EquipFromInventory(world *World, itemIndex int, slot EquipmentSlot) (Item, []string, error) {
	item := character.Inventory.Items[itemIndex]
	item.Amount = 1

	// Check item type against equipment slot
	if !ItemMatchesEquipmentSlot(item, slot) {
		return item, nil, fmt.Errorf("%s cannot be equipped to slot %s.", item.NameWithCondition(), slot.String())
	}

	// Check skill requirement
	skill, hasRequirement := item.Kind.requiredSkill()
	if hasRequirement && !character.Skills.Has(skill) {
		return item, nil, fmt.Errorf("You need the skill %s to equip %s.", skill.String(), item.NameWithCondition())
	}

	// Check that the displaced items will fit in the inventory once this item is removed from it
	var sizeDelta int32 = -item.Size()
	for _, displacedItem := range character.Equipment.ItemsDisplacedBy(slot, item) {
		sizeDelta += displacedItem.Size()
	}
	if !character.Inventory.HasSpaceFor(sizeDelta, character.InventoryCapacity()) {
		return item, nil, fmt.Errorf("You don't have enough space in your inventory to equip %s.", item.NameWithCondition())
	}

	// Move the item from the inventory into the slot
	character.Inventory.RemoveItem(world, itemIndex)
	unequippedItems, _ := character.Equipment.Equip(slot, item)

	// Move the displaced items into the inventory
	messages := []string{}
	for _, unequippedItem := range unequippedItems {
		character.Inventory.AddItem(unequippedItem)
		messages = append(messages, fmt.Sprintf("<g>%s</g> was unequipped and added to your inventory.", unequippedItem.NameWithCondition()))
	}

	return item, messages, nil
}

func (character *Character) UnequipToInventory(slot EquipmentSlot) (Item, error) {
	if slot == EQUIPMENT_SLOT_OFF_HAND && character.Equipment.IsTwoHandedWeaponEquipped() {
		slot = EQUIPMENT_SLOT_MAIN_HAND
	}

	equippedItem := character.Equipment.Get(slot)
	if equippedItem == nil {
		return Item{}, fmt.Errorf("You have nothing equipped in your %s slot.", slot.String())
	}

	if !character.Inventory.HasSpaceFor(equippedItem.Size(), character.InventoryCapacity()) {
		return *equippedItem, fmt.Errorf("You don't have enough space in your inventory to unequip %s.", equippedItem.NameWithCondition())
	}

	item, _ := character.Equipment.Unequip(slot)
	character.Inventory.AddItem(item)

	return item, nil
}
