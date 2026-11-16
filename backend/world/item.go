package world

import (
	"fmt"
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

func (kind ItemKind) String() string {
	switch kind {
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

	kind ItemKind
	price int32
	size int32
	maxDurability int32
	requirements int32
	data any
}

type CreateItemParams struct {
	itemType *ItemType
	amount int32
	durabilityPercent int32
	durability int32
}

type ArmorData struct {
	prefix string

	deflectChance int32
	deflectBonus int32
	damnageResistance int32

	sneakModified int32
	evadeModifier int32
	initiativeModifier int32
}

type WeaponData struct {
	prefix string

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
		durability = (params.durabilityPercent * params.itemType.maxDurability) / 100
	}

	return Item {
		Id: itemId,
		ItemType: params.itemType,
		Amount: params.amount,
		Durability: durability,
	}
}

func (item *Item) NameWithCondition() string {
	maxDurability := item.ItemType.maxDurability
	if maxDurability == 0 {
		return item.Name
	}

	var prefix string
	hasPrefix := false

	if item.IsOneHandedWeapon() || item.IsTwoHandedWeapon() {
		prefix = item.ItemType.data.(WeaponData).prefix
		hasPrefix = true
	}

	if item.IsOutfit() || item.IsShield() {
		prefix = item.ItemType.data.(ArmorData).prefix
		hasPrefix = true
	}

	if item.Durability < maxDurability / 2 {
		return "Damaged " + item.Name
	} else if item.Durability > maxDurability && hasPrefix {
		return prefix + " " + item.Name
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
	switch itemType.kind {
		case ITEM_KIND_CONSUMABLE:
			return true
		default:
			return false
	}
}

func (item *Item) IsOneHandedWeapon() bool {
	switch item.ItemType.kind {
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
	switch item.ItemType.kind {
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
	switch item.ItemType.kind {
	 	case ITEM_KIND_EQUIPMENT_OUTFIT_CLOTHING,
	   		 ITEM_KIND_EQUIPMENT_OUTFIT_LIGHT_ARMOR,
			 ITEM_KIND_EQUIPMENT_OUTFIT_HEAVY_ARMOR:
			 return true
		default:
			 return false
	}
}

func (item *Item) IsShield() bool {
	switch item.ItemType.kind {
		case ITEM_KIND_EQUIPMENT_SHIELD:
			return true
		default:
			return false
	}
}

func (item *Item) IsAccessory() bool {
	switch item.ItemType.kind {
		case ITEM_KIND_EQUIPMENT_ACCESSORY:
			return true
		default:
			return false
	}
}

func (item *Item) getRequirements() (bool, SkillBlock) {
	var hasRequirements bool
	var requirements SkillBlock

	switch item.ItemType.kind {
		case ITEM_KIND_EQUIPMENT_ONE_HANDED_SWORD,
			 ITEM_KIND_EQUIPMENT_TWO_HANDED_SWORD:
			hasRequirements = true
		 	requirements.Values[SKILL_SWORDS] = item.ItemType.requirements
		case ITEM_KIND_EQUIPMENT_ONE_HANDED_AXE,
			 ITEM_KIND_EQUIPMENT_TWO_HANDED_AXE:
			hasRequirements = true
		 	requirements.Values[SKILL_AXES] = item.ItemType.requirements
		case ITEM_KIND_EQUIPMENT_ONE_HANDED_BLUNT,
			 ITEM_KIND_EQUIPMENT_TWO_HANDED_BLUNT:
			hasRequirements = true
		 	requirements.Values[SKILL_BLUDGEONING] = item.ItemType.requirements
		case ITEM_KIND_EQUIPMENT_ONE_HANDED_DAGGER:
			hasRequirements = true
		 	requirements.Values[SKILL_DAGGERS] = item.ItemType.requirements
		case ITEM_KIND_EQUIPMENT_TWO_HANDED_BOW:
			hasRequirements = true
			requirements.Values[SKILL_MARKSMANSHIP] = item.ItemType.requirements
		case ITEM_KIND_EQUIPMENT_OUTFIT_LIGHT_ARMOR:
			hasRequirements = true
			requirements.Values[SKILL_LIGHT_ARMOR] = item.ItemType.requirements
		case ITEM_KIND_EQUIPMENT_OUTFIT_HEAVY_ARMOR:
			hasRequirements = true
			requirements.Values[SKILL_HEAVY_ARMOR] = item.ItemType.requirements
		case ITEM_KIND_EQUIPMENT_SHIELD:
			hasRequirements = true
			requirements.Values[SKILL_SHIELDS] = item.ItemType.requirements
		default:
			hasRequirements = false
	}

	return hasRequirements, requirements
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
