package world

import (
	"fmt"
	"math"
	"slices"
	"mud/bitset"
)

type Character struct {
	PlayerKey string

	Race *RaceData
	Class *ClassData

	Level int32
	Experience int32

	// Calculated on load and level up. Not saved
	Attributes AttributeBlock
	Skills SkillFlags

	Equipment Equipment
	Inventory Inventory

	SpellsKnown []*Spell
	SpellsPrepared []*Spell

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
	character.calculateAttributes()
	character.Skills = character.Race.Skills | character.Class.Skills

	character.RoomsDiscovered = bitset.New(len(world.Rooms))

	// Set the character's room to the start room
	character.Room = world.StartRoom
	bitset.Set(character.RoomsDiscovered, character.Room, true)

	character.Equipment = NewEquipment()
	character.SpellsKnown = make([]*Spell, 0, 1)
	character.SpellsPrepared = make([]*Spell, 0, 1)
	character.calculateStats()

	// Set health and mana
	character.Health = character.Stats.Values[STAT_MAX_HEALTH]
	character.Mana = character.Stats.Values[STAT_MAX_MANA]

	return character
}

func (character *Character) Attribute(attr Attribute) int32 {
	return character.Attributes.Values[attr]
}

// MOB FUNCTIONS
// Called through the Mob functions of the same name

func (character *Character) DescribeToPlayer(world *World, playerKey string) {
	world.MessagePlayer(playerKey, fmt.Sprintf("<y>%s</y> is a level %d %s %s.",
		character.Name, character.Level, character.Race.Name, character.Class.Name))
}

func (character *Character) InventoryCapacity() int32 {
	return INVENTORY_CAPACITY_PLAYER
}

func (character *Character) GetAttacks() []MobAttack {
	attacks := make([]MobAttack, 0, 2)

	// Main hand attack
	mainItem := character.Equipment.Get(EQUIPMENT_SLOT_MAIN_HAND)
	if mainItem != nil && mainItem.IsWeapon() {
		weaponData := mainItem.ItemType.Data.(ItemDataEquipment)
		attacks = append(attacks, MobAttack {
			damage: weaponData.damage + (character.Attribute(ATTR_STR) / 2),
			penetration: weaponData.penetration,
			hasWeaponSlot: true,
			weaponSlot: EQUIPMENT_SLOT_MAIN_HAND,
		})
	}

	// Off-hand attack
	offItem := character.Equipment.Get(EQUIPMENT_SLOT_OFF_HAND)
	mainHandIsTwoHanded := mainItem != nil && mainItem.IsTwoHandedWeapon()
	if !mainHandIsTwoHanded && offItem != nil && offItem.IsWeapon() {
		weaponData := offItem.ItemType.Data.(ItemDataEquipment)
		attacks = append(attacks, MobAttack {
			damage: weaponData.damage + (character.Attribute(ATTR_STR) / 4),
			penetration: weaponData.penetration,
			hasWeaponSlot: true,
			weaponSlot: EQUIPMENT_SLOT_OFF_HAND,
		})
	}

	// Unarmed attack, added only if main and off hand both did not produce an attack and character has a free hand
	if len(attacks) == 0 && (mainItem == nil || offItem == nil) {
		attacks = append(attacks, MobAttack {
			damage: character.Attribute(ATTR_STR) / 2,
			penetration: 0,
		})
	}

	return attacks
}

// CHARACTER FUNCTIONS

func (character *Character) SpellSlots() int {
	return int(character.Attribute(ATTR_INT)) / 4
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
	world.MessagePlayer(character.PlayerKey, fmt.Sprintf("You gained %d experience.", amount))

	for character.Experience > ExpToNextLevel(character.Level) {
		character.Level++
		character.calculateAttributes()
		character.calculateStats()
		world.MessagePlayer(character.PlayerKey, fmt.Sprintf("Level up! You are now level %d.", character.Level))
	}
}

func (character *Character) calculateAttributes() {
	baseAttributes := character.Race.Attributes.Add(&character.Class.Attributes)
	character.Attributes = baseAttributes.ScaledToLevel(character.Level)
}

// Must be called whenever the character's attributes or equipment change
func (character *Character) calculateStats() {
	attributes := &character.Attributes
	stats := StatBlock{}

	stats.Values[STAT_MAX_HEALTH] = 10 * attributes.Values[ATTR_VIT]
	stats.Values[STAT_MAX_MANA] = 10 * attributes.Values[ATTR_INT]
	stats.Values[STAT_INITIATIVE] = attributes.Values[ATTR_AGI]
	stats.Values[STAT_ACCURACY] = attributes.Values[ATTR_AGI]
	stats.Values[STAT_EVASION] = attributes.Values[ATTR_AGI]
	stats.Values[STAT_CRIT_CHANCE] = attributes.Values[ATTR_AGI] / 2
	stats.Values[STAT_DEFLECTION] = 0
	stats.Values[STAT_DEFLECTION_AMOUNT] = 40
	stats.Values[STAT_RESISTANCE] = 0
	stats.Values[STAT_MAGIC_DAMAGE] = attributes.Values[ATTR_FTH] / 2
	stats.Values[STAT_MAGIC_RESISTANCE] = attributes.Values[ATTR_FTH] / 2
	stats.Values[STAT_FOCUS] = attributes.Values[ATTR_INT] / 2

	// Add armor from each equipped item that has armor data (outfit, shield)
	for index := range EQUIPMENT_SLOT_COUNT {
		item := character.Equipment.Get(EquipmentSlot(index))
		if item == nil || !item.IsEquipment() {
			continue
		}

		equipmentStats := item.ItemType.Data.(ItemDataEquipment).stats
		stats = stats.Add(&equipmentStats)
	}

	character.Stats = stats
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
		world.MessagePlayer(character.PlayerKey, fmt.Sprintf("Your <g>%s</g> broke!", item.Name))
		world.PlayerInventoryChanged(character.PlayerKey)
		character.Equipment.Unequip(slot)
		character.calculateStats()
		return
	}

	// If item has become damaged, tell the user
	itemIsDamaged := item.Durability < maxDurability / 2
	if itemIsDamaged && !itemWasDamaged {
		world.MessagePlayer(character.PlayerKey, fmt.Sprintf("Your <g>%s</g> is now damaged.", item.Name))
		world.PlayerInventoryChanged(character.PlayerKey)
		return
	}

	// If an item has lost its sharpness, tell the user
	itemLostSharpness := item.Durability == maxDurability
	if itemLostSharpness {
		prefix, _ := item.GetReinforcedPrefix()
		world.MessagePlayer(character.PlayerKey, fmt.Sprintf("Your <g>%s</g> is no longer %s.", item.Name, prefix))
		world.PlayerInventoryChanged(character.PlayerKey)
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
	character.calculateStats()

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
	character.calculateStats()

	return item, nil
}

func (character *Character) FindItem(id ItemId) (*Item, bool) {
	// Check inventory
	index, found := character.Inventory.FindItemById(id)
	if found {
		return &character.Inventory.Items[index], true
	}

	// Check equipment
	for index := range EQUIPMENT_SLOT_COUNT {
		slot := EquipmentSlot(index)
		item := character.Equipment.Get(slot)
		if item != nil && item.Id == id {
			return item, true
		}
	}

	return nil, false
}

// Called on death, performs random durability damage to equipment and dumps the equipment into the inventory
func (character *Character) dumpEquipment(world *World) {
	for index := range EQUIPMENT_SLOT_COUNT {
		slot := EquipmentSlot(index)
		item, exists := character.Equipment.Unequip(slot)
		if !exists {
			continue
		}

		halfMaxDurability := item.MaxDurability() / 2
		item.Durability -= halfMaxDurability + world.Random.Int32N(halfMaxDurability)
		if item.Durability <= 0 {
			continue
		}

		character.Inventory.AddItem(item)
	}

	character.calculateStats()
}

func (character *Character) HasKnownSpell(spell *Spell) bool {
	return slices.Index(character.SpellsKnown, spell) != -1
}

func (character *Character) HasPreparedSpell(spell *Spell) bool {
	return slices.Index(character.SpellsPrepared, spell) != -1
}
