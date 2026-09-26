package world

import (
	"fmt"
	"log"
)

const BEHAVIOR_TROLL_ANGRY_DURATION = (5 * 60) / WORLD_SECONDS_PER_UPDATE
const BEHAVIOR_TROLL_PEACE_DURATION = (5 * 60) / WORLD_SECONDS_PER_UPDATE

type BehaviorTroll struct {
	ExitToBlock Direction
	Toll Item
	angryTimer int32
	peaceTimer int32
}

func (behavior *BehaviorTroll) init(npc *Npc, world *World) {
	behavior.angryTimer = 0
	behavior.peaceTimer = 0
	behavior.setIsBlockingDoor(npc, world, true)
}

func (behavior *BehaviorTroll) update(npc *Npc, world *World) {
	if behavior.angryTimer > 0 {
		behavior.angryTimer--
		if behavior.angryTimer <= 0 {
			npc.disposition = NPC_DISPOSITION_NEUTRAL
		}
	}

	if behavior.peaceTimer > 0 {
		behavior.peaceTimer--
		if behavior.peaceTimer <= 0 {
			behavior.setIsBlockingDoor(npc, world, true)

			// Remove toll from inventory
			npcMob := world.Mobs.Get(npc.mobHandle)
			index, found := npcMob.Data.Inventory.FindItem(behavior.Toll.Id)
			if found {
				npcMob.Data.Inventory.RemoveItems(index, npcMob.Data.Inventory.Items[index].Amount)
			}
		}
	}
}

func (behavior *BehaviorTroll) onEvent(npc *Npc, world *World, event BehaviorEvent) bool {
	switch event.Type {
		case BEHAVIOR_EVENT_TYPE_ATTACKED: {
			behavior.angryTimer = BEHAVIOR_TROLL_ANGRY_DURATION
			behavior.peaceTimer = 0
			behavior.setIsBlockingDoor(npc, world, true)
			npc.disposition = NPC_DISPOSITION_HOSTILE
			return true
		}

		case BEHAVIOR_EVENT_TYPE_PLAYER_ENTERED: {
			eventData := event.Data.(BehaviorEventPlayerEntered)

			playerMob := world.Mobs.Get(eventData.PlayerHandle)
			if playerMob.PlayerCharacter == nil {
				log.Printf("Warn - Behavior troll was given player entered event with a non-player mob.")
				return false
			}

			if !behavior.isBlockingDoor(npc, world) {
				return false
			}

			if behavior.angryTimer > 0 {
				return false
			}

			npcMob := world.Mobs.Get(npc.mobHandle)
			message := fmt.Sprintf("%s: 'Stop, %s. You want pass? Must pay toll. Give me %s.'",
				npcMob.Data.Name, RACE_DATA[playerMob.PlayerCharacter.Race].Name, behavior.Toll.GetNameWithAmount())
			world.messagePlayer(playerMob.PlayerCharacter.PlayerId, message)

			return true
		}

		case BEHAVIOR_EVENT_TYPE_ITEM_GIVEN: {
			eventData := event.Data.(BehaviorEventItemGiven)

			if behavior.angryTimer > 0 || behavior.peaceTimer > 0 {
				return false
			}

			npcMob := world.Mobs.Get(npc.mobHandle)

			// If player gives a non-toll item, drop it on the floor
			if npcMob.Data.Inventory.Items[eventData.AddedToIndex].Id != behavior.Toll.Id {
				item := npcMob.Data.Inventory.RemoveItems(eventData.AddedToIndex, eventData.Amount)
				world.Rooms[npcMob.Data.Room].Inventory.AddItem(item)

				world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s: 'This not toll! Give me %s!'",
					npcMob.Data.Name, behavior.Toll.GetNameWithAmount()))
				world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s dropped %s on the floor.",
					npcMob.Data.Name, item.GetNameWithAmount()))
				return true
			}

			// If the player gave the toll but not the right amount, then the troll complains
			if npcMob.Data.Inventory.AmountOf(behavior.Toll.Id) < behavior.Toll.Amount {
				world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s: 'You think me no count? Me need %s!'",
					npcMob.Data.Name, behavior.Toll.GetNameWithAmount()))
				return true
			}

			// Otherwise, the toll is paid, unblock the exit
			behavior.setIsBlockingDoor(npc, world, false)
			behavior.peaceTimer = BEHAVIOR_TROLL_PEACE_DURATION

			world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s grins and pockets the toll.",
				npcMob.Data.Name))
			world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s: 'Me see you are smart one. Me step aside now. Go on.'",
				npcMob.Data.Name))
			world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s stepped aside. The way %s is clear.",
				npcMob.Data.Name, DirectionToString(behavior.ExitToBlock)))

			return true
		}
	}

	return false
}

func (behavior *BehaviorTroll) getDescription(npc *Npc, world *World) (string, bool) {
	npcMob := world.Mobs.Get(npc.mobHandle)

	if behavior.isBlockingDoor(npc, world) {
		return fmt.Sprintf("%s is blocking the way %s.", npcMob.Data.Name, DirectionToString(behavior.ExitToBlock)), true
	}

	return fmt.Sprintf("%s has stepped aside. The way %s is clear.", npcMob.Data.Name, DirectionToString(behavior.ExitToBlock)), true
}

func (behavior *BehaviorTroll) isBlockingDoor(npc *Npc, world *World) bool {
	npcMob := world.Mobs.Get(npc.mobHandle)
	npcRoom := &world.Rooms[npcMob.Data.Room]

	return npcRoom.ExitIsLocked[behavior.ExitToBlock]
}

func (behavior *BehaviorTroll) setIsBlockingDoor(npc *Npc, world *World, value bool) {
	npcMob := world.Mobs.Get(npc.mobHandle)
	npcRoom := &world.Rooms[npcMob.Data.Room]
	npcRoom.SetExitLocked(world, behavior.ExitToBlock, value)
}
