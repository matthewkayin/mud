package world

import (
	"fmt"
)

type ItemLocation int
const (
	ITEM_LOCATION_INVENTORY ItemLocation = iota
	ITEM_LOCATION_EQUIPMENT
)

// Identifies an item in a mob's inventory or equipment by its position.
// Items move around (removing a stack moves the last stack into its index, equipping copies the item into a slot),
// so a handle only stays valid until the mob's inventory or equipment changes.
type ItemHandle struct {
	Mob MobHandle
	Location ItemLocation
	// Inventory index, or EquipmentSlot when Location is ITEM_LOCATION_EQUIPMENT
	Index int
	// The ID of the item
	Id ItemId
}

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

// Returns the item the handle refers to, or false if the mob no longer exists
// or the item at the handle's position is not the expected item
func (world *World) getItem(handle ItemHandle) (*Item, bool) {
	mob, exists := world.Mobs.GetIfExists(handle.Mob)
	if !exists {
		return nil, false
	}

	var item *Item
	switch handle.Location {
		case ITEM_LOCATION_INVENTORY: {
			if handle.Index < 0 || handle.Index >= len(mob.Inventory.Items) {
				return nil, false
			}
			item = &mob.Inventory.Items[handle.Index]
		}
		case ITEM_LOCATION_EQUIPMENT: {
			// if handle.Index < 0 || handle.Index >= EQUIPMENT_SLOT_COUNT {
				// return nil, false
			// }
			// item = mob.Data.Equipment.Get(EquipmentSlot(handle.Index))
			// if item == nil {
				// return nil, false
			// }
			return nil, false
		}
		default:
			return nil, false
	}

	if item.Id != handle.Id {
		return nil, false
	}

	return item, true
}

// Returns a handle to the expected item, following an inventory item to its new index if it has moved
func (world *World) relocateItemHandle(handle ItemHandle) (ItemHandle, bool) {
	if handle.Location != ITEM_LOCATION_INVENTORY {
		_, exists := world.getItem(handle)
		return handle, exists
	}

	mob, exists := world.Mobs.GetIfExists(handle.Mob)
	if !exists {
		return handle, false
	}

	index, found := mob.Inventory.FindExpectedItem(handle.Id, handle.Index)
	if !found {
		return handle, false
	}

	handle.Index = index
	return handle, true
}
