package world

import (
	"fmt"
	"github.com/mmcdole/lunar"
)

type ItemId uint64

type Item struct {
	// A unique ID for this item instance.
	// When a spell or item targets another item, they use this ID to refer to their target.
	Id ItemId
	*ItemType
	Amount int32
	Durability int32
}

type ItemLocation int
const (
	ITEM_LOCATION_INVENTORY ItemLocation = iota
	ITEM_LOCATION_EQUIPMENT
)

type ItemKind int
const (
	ITEM_KIND_CONSUMABLE = iota
	ITEM_KIND_MISC
	ITEM_KIND_EQUIPMENT_ONE_HANDED_SWORD
	ITEM_KIND_EQUIPMENT_ONE_HANDED_AXE
	ITEM_KIND_EQUIPMENT_ONE_HANDED_BLUNT
	ITEM_KIND_EQUIPMENT_ONE_HANDED_DAGGER
	ITEM_KIND_EQUIPMENT_TWO_HANDED_SWORD
	ITEM_KIND_EQUIPMENT_TWO_HANDED_AXE
	ITEM_KIND_EQUIPMENT_TWO_HANDED_BLUNT
	ITEM_KIND_EQUIPMENT_TWO_HANDED_BOW
	ITEM_KIND_EQUIPMENT_SHIELD
	ITEM_KIND_EQUIPMENT_OUTFIT_CLOTHING
	ITEM_KIND_EQUIPMENT_OUTFIT_LIGHT_ARMOR
	ITEM_KIND_EQUIPMENT_OUTFIT_HEAVY_ARMOR
	ITEM_KIND_EQUIPMENT_ACCESSORY
	ITEM_KIND_COUNT
)

func (location ItemLocation) String() string {
	switch location {
		case ITEM_LOCATION_INVENTORY:
			return "Inventory"
		case ITEM_LOCATION_EQUIPMENT:
			return "Equipment"
		default:
			panic(fmt.Sprintf("Item location %d not handled.", location))
	}
}

func (kind ItemKind) String() string {
	switch kind {
	case ITEM_KIND_EQUIPMENT_ONE_HANDED_SWORD:
		return "One-Handed Sword"
	case ITEM_KIND_EQUIPMENT_ONE_HANDED_AXE:
		return "One-Handed Axe"
	case ITEM_KIND_EQUIPMENT_ONE_HANDED_BLUNT:
		return "One-Handed Blunt"
	case ITEM_KIND_EQUIPMENT_ONE_HANDED_DAGGER:
		return "Dagger"
	case ITEM_KIND_EQUIPMENT_TWO_HANDED_SWORD:
		return "Two-Handed Sword"
	case ITEM_KIND_EQUIPMENT_TWO_HANDED_AXE:
		return "Two-Handed Axe"
	case ITEM_KIND_EQUIPMENT_TWO_HANDED_BLUNT:
		return "Two-Handed Blunt"
	case ITEM_KIND_EQUIPMENT_TWO_HANDED_BOW:
		return "Bow"
	case ITEM_KIND_EQUIPMENT_SHIELD:
		return "Shield"
	case ITEM_KIND_EQUIPMENT_OUTFIT_CLOTHING:
		return "Clothing"
	case ITEM_KIND_EQUIPMENT_OUTFIT_LIGHT_ARMOR:
		return "Light Armor"
	case ITEM_KIND_EQUIPMENT_OUTFIT_HEAVY_ARMOR:
		return "Heavy Armor"
	case ITEM_KIND_EQUIPMENT_ACCESSORY:
		return "Accessory"
	case ITEM_KIND_CONSUMABLE:
		return "Consumable"
	case ITEM_KIND_MISC:
		return "Misc"
	default:
		panic(fmt.Sprintf("Item kind %d not handled", kind))
	}
}

func (kind ItemKind) toScriptEnum() string {
	switch kind {
	case ITEM_KIND_EQUIPMENT_ONE_HANDED_SWORD:
		return "ONE_HANDED_SWORD"
	case ITEM_KIND_EQUIPMENT_ONE_HANDED_AXE:
		return "ONE_HANDED_AXE"
	case ITEM_KIND_EQUIPMENT_ONE_HANDED_BLUNT:
		return "ONE_HANDED_BLUNT"
	case ITEM_KIND_EQUIPMENT_ONE_HANDED_DAGGER:
		return "DAGGER"
	case ITEM_KIND_EQUIPMENT_TWO_HANDED_SWORD:
		return "TWO_HANDED_SWORD"
	case ITEM_KIND_EQUIPMENT_TWO_HANDED_AXE:
		return "TWO_HANDED_AXE"
	case ITEM_KIND_EQUIPMENT_TWO_HANDED_BLUNT:
		return "TWO_HANDED_BLUNT"
	case ITEM_KIND_EQUIPMENT_TWO_HANDED_BOW:
		return "BOW"
	case ITEM_KIND_EQUIPMENT_SHIELD:
		return "SHIELD"
	case ITEM_KIND_EQUIPMENT_OUTFIT_CLOTHING:
		return "CLOTHING"
	case ITEM_KIND_EQUIPMENT_OUTFIT_LIGHT_ARMOR:
		return "LIGHT_ARMOR"
	case ITEM_KIND_EQUIPMENT_OUTFIT_HEAVY_ARMOR:
		return "HEAVY_ARMOR"
	case ITEM_KIND_EQUIPMENT_ACCESSORY:
		return "ACCESSORY"
	case ITEM_KIND_CONSUMABLE:
		return "CONSUMABLE"
	case ITEM_KIND_MISC:
		return "MISC"
	default:
		panic(fmt.Sprintf("Item kind %d not handled", kind))
	}
}

type ItemType struct {
	Name string
	Description string

	Kind ItemKind
	price int32
	size int32
	Data any
}

type ItemDataConsumable struct {
	TargetType TargetType
	ValidateTarget *lua.Function
	onUse *lua.Function
}

type CreateItemParams struct {
	itemType *ItemType
	amount int32
	durabilityPercent int32
	durability int32
}

type ArmorData struct {
	maxDurability int32
	reinforcedPrefix string

	// Only parsed for armor kinds which have a required skill
	skillRequirement int32

	deflectChance int32
	deflectBonus int32
	damageResistance int32

	sneakModifier int32
	evadeModifier int32
	initiativeModifier int32
}

type WeaponData struct {
	maxDurability int32
	reinforcedPrefix string
	skillRequirement int32

	damage int32
	penetration int32
}

// Creates a new item and assigns it a unique ID. All Item instances should be created this way
// When specifying the item's durability, you should use the durability field or the durabilityPercent field, not both
func (world *World) createItem(params CreateItemParams) Item {
	itemId := world.NextItemId
	world.NextItemId++

	durability := params.durability
	if durability == 0 && params.durabilityPercent != 0 {
		durability = (params.durabilityPercent * params.itemType.MaxDurability()) / 100
	}

	return Item {
		Id: itemId,
		ItemType: params.itemType,
		Amount: params.amount,
		Durability: durability,
	}
}

func (item *Item) NameWithCondition() string {
	maxDurability := item.MaxDurability()
	if maxDurability == 0 {
		return item.Name
	}

	if item.Durability < maxDurability / 2 {
		return "Damaged " + item.Name
	}

	reinforcedPrefix, hasReinforcedPrefix := item.GetReinforcedPrefix()
	if item.Durability > maxDurability && hasReinforcedPrefix {
		return reinforcedPrefix + " " + item.Name
	}

	return item.Name
}

func (item *Item) NameWithAmount() string {
	if item.Amount == 1 {
		return item.NameWithCondition()
	}
	return fmt.Sprintf("%d %s", item.Amount, item.NameWithCondition())
}

func (item *Item) Size() int32 {
	return item.size * item.Amount
}

func (itemType *ItemType) CanStack() bool {
	switch itemType.Kind {
		case ITEM_KIND_CONSUMABLE:
			return true
		default:
			return false
	}
}

func (itemType *ItemType) HasWeaponData() bool {
switch itemType.Kind {
	case ITEM_KIND_EQUIPMENT_ONE_HANDED_SWORD,
	ITEM_KIND_EQUIPMENT_ONE_HANDED_AXE,
	ITEM_KIND_EQUIPMENT_ONE_HANDED_BLUNT,
	ITEM_KIND_EQUIPMENT_ONE_HANDED_DAGGER,
	ITEM_KIND_EQUIPMENT_TWO_HANDED_SWORD,
	ITEM_KIND_EQUIPMENT_TWO_HANDED_AXE,
	ITEM_KIND_EQUIPMENT_TWO_HANDED_BLUNT,
	ITEM_KIND_EQUIPMENT_TWO_HANDED_BOW:
		return true
	default:
		return false
	}
}

func (itemType *ItemType) HasArmorData() bool {
	switch itemType.Kind {
		case ITEM_KIND_EQUIPMENT_OUTFIT_CLOTHING,
   		 ITEM_KIND_EQUIPMENT_OUTFIT_LIGHT_ARMOR,
		 ITEM_KIND_EQUIPMENT_OUTFIT_HEAVY_ARMOR,
		 ITEM_KIND_EQUIPMENT_SHIELD,
		 ITEM_KIND_EQUIPMENT_ACCESSORY:
		 return true
	default:
		 return false
	}
}

func (item *Item) IsOneHandedWeapon() bool {
	switch item.ItemType.Kind {
	 	case ITEM_KIND_EQUIPMENT_ONE_HANDED_SWORD,
	   		 ITEM_KIND_EQUIPMENT_ONE_HANDED_AXE,
			 ITEM_KIND_EQUIPMENT_ONE_HANDED_BLUNT,
			 ITEM_KIND_EQUIPMENT_ONE_HANDED_DAGGER:
			 return true
		default:
			 return false
	}
}

func (item *Item) IsTwoHandedWeapon() bool {
	switch item.ItemType.Kind {
	 	case ITEM_KIND_EQUIPMENT_TWO_HANDED_SWORD,
	   		 ITEM_KIND_EQUIPMENT_TWO_HANDED_AXE,
			 ITEM_KIND_EQUIPMENT_TWO_HANDED_BLUNT,
			 ITEM_KIND_EQUIPMENT_TWO_HANDED_BOW:
			 return true
		default:
			 return false
	}
}

func (item *Item) IsOutfit() bool {
	switch item.ItemType.Kind {
	 	case ITEM_KIND_EQUIPMENT_OUTFIT_CLOTHING,
	   		 ITEM_KIND_EQUIPMENT_OUTFIT_LIGHT_ARMOR,
			 ITEM_KIND_EQUIPMENT_OUTFIT_HEAVY_ARMOR:
			 return true
		default:
			 return false
	}
}

func (item *Item) IsShield() bool {
	switch item.ItemType.Kind {
		case ITEM_KIND_EQUIPMENT_SHIELD:
			return true
		default:
			return false
	}
}

func (item *Item) IsAccessory() bool {
	switch item.ItemType.Kind {
		case ITEM_KIND_EQUIPMENT_ACCESSORY:
			return true
		default:
			return false
	}
}

// Returns the skill which items of this kind require, and whether the kind has a skill requirement
func (kind ItemKind) requiredSkill() (SkillName, bool) {
	switch kind {
		case ITEM_KIND_EQUIPMENT_ONE_HANDED_SWORD,
			 ITEM_KIND_EQUIPMENT_TWO_HANDED_SWORD:
			return SKILL_SWORDS, true
		case ITEM_KIND_EQUIPMENT_ONE_HANDED_AXE,
			 ITEM_KIND_EQUIPMENT_TWO_HANDED_AXE:
			return SKILL_AXES, true
		case ITEM_KIND_EQUIPMENT_ONE_HANDED_BLUNT,
			 ITEM_KIND_EQUIPMENT_TWO_HANDED_BLUNT:
			return SKILL_BLUDGEONING, true
		case ITEM_KIND_EQUIPMENT_ONE_HANDED_DAGGER:
			return SKILL_DAGGERS, true
		case ITEM_KIND_EQUIPMENT_TWO_HANDED_BOW:
			return SKILL_MARKSMANSHIP, true
		case ITEM_KIND_EQUIPMENT_OUTFIT_LIGHT_ARMOR:
			return SKILL_LIGHT_ARMOR, true
		case ITEM_KIND_EQUIPMENT_OUTFIT_HEAVY_ARMOR:
			return SKILL_HEAVY_ARMOR, true
		case ITEM_KIND_EQUIPMENT_SHIELD:
			return SKILL_SHIELDS, true
		default:
			return 0, false
	}
}

// Returns the skill and rank needed to equip items of this type, and whether the type has a skill requirement
func (itemType *ItemType) getSkillRequirement() (SkillName, int32, bool) {
	skill, hasSkillRequirement := itemType.Kind.requiredSkill()
	if !hasSkillRequirement {
		return 0, 0, false
	}

	if itemType.HasWeaponData() {
		return skill, itemType.Data.(WeaponData).skillRequirement, true
	}
	if itemType.HasArmorData() {
		return skill, itemType.Data.(ArmorData).skillRequirement, true
	}

	return 0, 0, false
}

func (item *Item) EquipmentSlotForItem() (EquipmentSlot, bool) {
	if item.IsOneHandedWeapon() || item.IsTwoHandedWeapon() {
			return EQUIPMENT_SLOT_MAIN_HAND, true
	} else if item.IsShield() {
			return EQUIPMENT_SLOT_OFF_HAND, true
	} else if item.IsOutfit() {
			return EQUIPMENT_SLOT_OUTFIT, true
	} else if item.IsAccessory(){
		return EQUIPMENT_SLOT_ACCESSORY, true
	} else {
		return EQUIPMENT_SLOT_COUNT, false
	}
}

// Returns 0 for item types which have no durability
func (itemType *ItemType) MaxDurability() int32 {
	if itemType.HasWeaponData() {
		return itemType.Data.(WeaponData).maxDurability
	}
	if itemType.HasArmorData() {
		return itemType.Data.(ArmorData).maxDurability
	}

	return 0
}

// Returns the prefix shown on an item whose durability is above its max durability
func (itemType *ItemType) GetReinforcedPrefix() (string, bool) {
	if itemType.HasWeaponData() {
		return itemType.Data.(WeaponData).reinforcedPrefix, true
	}
	if itemType.HasArmorData() {
		return itemType.Data.(ArmorData).reinforcedPrefix, true
	}

	return "", false
}
