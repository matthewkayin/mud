package world

import (
	"log"
)

// Where an item target is held
type ItemTargetHolder int
const (
	// In the inventory or equipment of ItemTarget.Mob
	ITEM_TARGET_HOLDER_MOB ItemTargetHolder = iota
	// On the floor of ItemTarget.Room
	ITEM_TARGET_HOLDER_ROOM
	// In one of the chests in ItemTarget.Room
	ITEM_TARGET_HOLDER_CHEST
)

// An item that a spell or item targets, found by its instance ID
type ItemTarget struct {
	Holder ItemTargetHolder
	// Set when Holder is ITEM_TARGET_HOLDER_MOB
	Mob MobHandle
	// Set when Holder is ITEM_TARGET_HOLDER_ROOM or ITEM_TARGET_HOLDER_CHEST
	Room int
	Id ItemId
}

// Returns the item that the target refers to, if it still exists and is within the mob's reach
func (mob *Mob) findTargetItem(world *World, target ItemTarget) (*Item, bool) {
	switch target.Holder {
		case ITEM_TARGET_HOLDER_MOB: {
			holderMob, exists := world.Mobs.GetIfExists(target.Mob)
			if !exists {
				return nil, false
			}
			if holderMob.Handle != mob.Handle && holderMob.Room != mob.Room {
				return nil, false
			}
			if !holderMob.IsPlayer() {
				log.Printf("warn: in findTargetItem(), the item-holding mob %s is an NPC", holderMob.Name)
				return nil, false
			}

			return holderMob.Character.FindItem(target.Id)
		}

		case ITEM_TARGET_HOLDER_ROOM: {
			if target.Room != mob.Room {
				return nil, false
			}

			room := &world.Rooms[target.Room]
			index, found := room.Inventory.FindItemById(target.Id)
			if !found {
				return nil, false
			}

			return &room.Inventory.Items[index], true
		}

		case ITEM_TARGET_HOLDER_CHEST: {
			if target.Room != mob.Room {
				return nil, false
			}

			// Chests are searched by item ID rather than by index, since rotted corpses are removed and chest indices shift.
			// The unexported inventory is read directly so that looking up the item never rolls a chest's loot.
			room := &world.Rooms[target.Room]
			for chestIndex := range len(room.Chests) {
				chest := &room.Chests[chestIndex]
				index, found := chest.inventory.FindItemById(target.Id)
				if found {
					return &chest.inventory.Items[index], true
				}
			}

			return nil, false
		}

		default: {
			log.Printf("warn: item target holder %d not handled in findTargetItem()", target.Holder)
			return nil, false
		}
	}
}
