package world

import (
	"fmt"
	"log"
	"math/rand"
	"slices"
	"sort"
	"strings"
	"mud/bitset"
)

const ROOM_NONE int = -1

const CHEST_DOES_NOT_DECAY = -1
const CHEST_CORPOSE_DECAY_DURATION = 30 / WORLD_SECONDS_PER_UPDATE

type Chest struct {
	Name string
	DecayTimer int
	Inventory Inventory
}

type Room struct {
	Name string
	Description string
	Exits [DIRECTION_COUNT]int
	ExitIsLocked [DIRECTION_COUNT]bool
	IsSafeZone bool
	Inventory Inventory
	Chests []Chest

	Occupants []MobHandle `json:"-"`
}

func (room *Room) MoveOccupant(world *World, occupantHandle MobHandle, newRoomIndex int) {
	// Get a pointer to the new room
	newRoom := &world.Rooms[newRoomIndex]
	occupantMob := world.Mobs.Get(occupantHandle)
	oldRoomIndex := occupantMob.Data.Room

	// Move the occupant
	room.RemoveOccupant(occupantHandle)
	world.messageRoom(oldRoomIndex, fmt.Sprintf("%s left the room.", occupantMob.GetName()))

	newRoom.AddOccupant(world, occupantHandle)
	world.messageRoomWithOptions(MessageRoomOptions {
		roomIndex: newRoomIndex,
		ignore: []MobHandle { occupantHandle },
		message: fmt.Sprintf("%s entered the room.", occupantMob.GetName()),
	})

	occupantMob.Data.Room = newRoomIndex

	// Player room discovery
	if occupantMob.PlayerCharacter != nil {
		bitset.Set(occupantMob.PlayerCharacter.RoomsDiscovered, newRoomIndex, true)
	}

	// Fire event
	world.pushEvent(Event {
		EventType: EVENT_TYPE_MOB_MOVE,
		Data: EventMobMove {
			MobHandle: occupantHandle,
			FromRoom: oldRoomIndex,
			ToRoom: newRoomIndex,
		},
	})
}

func (room *Room) AddOccupant(world *World, handle MobHandle) {
	// Determine the fuzzy numbers in use by other mobs of the same name
	mob := world.Mobs.Get(handle)
	fuzzyNumbersInUse := []int{}
	for _, occupantHandle := range room.Occupants {
		occupantMob := world.Mobs.Get(occupantHandle)
		if strings.EqualFold(mob.Data.Name, occupantMob.Data.Name) {
			fuzzyNumbersInUse = append(fuzzyNumbersInUse, occupantMob.fuzzyNumber)
		}
	}

	// For the joining mob, choose the first fuzzy number not in use
	mob.fuzzyNumber = 1
	for slices.Contains(fuzzyNumbersInUse, mob.fuzzyNumber) {
		mob.fuzzyNumber++
	}

	room.Occupants = append(room.Occupants, handle)
}

func (room *Room) RemoveOccupant(handle MobHandle) {
	occupantIndex := -1
	for index, occupant := range room.Occupants {
		if occupant == handle {
			occupantIndex = index
			break
		}
	}
	if occupantIndex == -1 {
		log.Printf("Warning: Tried to remove occupant with handle %d:%d from room %s, but no such occupant was found.",
			handle.Id, handle.Generation, room.Name)
	}

	room.RemoveOccupantByIndex(occupantIndex)
}

func (room *Room) RemoveOccupantByIndex(index int) {
	lastIndex := len(room.Occupants) - 1
	room.Occupants[index] = room.Occupants[lastIndex]
	room.Occupants = room.Occupants[:lastIndex]
}

func (room *Room) updateChestDecay() {
	chestIndex := 0
	for chestIndex < len(room.Chests) {
		chest := &room.Chests[chestIndex]
		if chest.DecayTimer == CHEST_DOES_NOT_DECAY {
			chestIndex++
			continue
		}

		chest.DecayTimer--
		if chest.DecayTimer == 0 {
			room.Chests[chestIndex] = room.Chests[len(room.Chests) - 1]
			room.Chests = room.Chests[:len(room.Chests) - 1]
			continue
		}

		chestIndex++
	}
}

// Sorts combatants by initiative order
// Combatants are a clone of the room occupant handles because otherwise
// the sorting will mess up multi-target identifiers i.e. Goblin #2 might
// not be Goblin #2 anymore after the initiative sort
func (room *Room) sortOccupantsByInitiativeOrder(world *World) []MobHandle {
	combatants := slices.Clone(room.Occupants)
	sort.Slice(combatants, func(i int, j int) bool {
		mobI := world.Mobs.Get(room.Occupants[i])
		mobJ := world.Mobs.Get(room.Occupants[j])

		// If they have the same agility, choose a random one to go first
		if mobI.Data.Agility() == mobJ.Data.Agility() {
			return rand.Intn(2) == 0
		}

		return mobI.Data.Agility() > mobJ.Data.Agility()
	})

	return combatants
}

func (room *Room) removeDeadOccupants(world *World) {
	// Remove dead occupants
	occupantIndex := 0
	for occupantIndex < len(room.Occupants) {
		// Get occupant mob
		occupantHandle := room.Occupants[occupantIndex]
		occupantMob := world.Mobs.Get(occupantHandle)

		// Check for mob death
		if !occupantMob.IsDead() {
			occupantIndex++
			continue
		}

		// Remove occupant
		room.RemoveOccupantByIndex(occupantIndex)

		// Fire event
		world.pushEvent(Event {
			EventType: EVENT_TYPE_MOB_DEATH,
			Data: EventMobDeath {
				PlayerId: occupantMob.GetPlayerId(),
				MobHandle: occupantHandle,
			},
		})

		// If NPC mob, distribute experience to players in the room
		if occupantMob.PlayerCharacter == nil {
			// Get a list of all player mobs
			playersInRoom := []*Mob{}
			for _, handle := range room.Occupants {
				mob := world.Mobs.Get(handle)
				if mob.PlayerCharacter != nil {
					playersInRoom = append(playersInRoom, mob)
				}
			}

			// Dole out EXP to each of them
			for _, player := range playersInRoom {
				dispursedExp := occupantMob.Data.Experience / int32(len(playersInRoom))
				player.GrantExperience(world, dispursedExp)
			}
		}

		// If player mob, remove their equipment so that it goes into their corpse
		if occupantMob.PlayerCharacter != nil {
			for slotIndex := range EQUIPMENT_SLOT_COUNT {
				slot := EquipmentSlot(slotIndex)
				item, success := occupantMob.Data.Equipment.Unequip(slot)
				if !success {
					continue
				}

				// Perform random durability damage to the player's equipped items on death
				halfMaxDurability := item.GetMaxDurability() / 2
				durabilityDamage := halfMaxDurability + int32(rand.Intn(int(halfMaxDurability)))
				item.Durability -= durabilityDamage
				if item.Durability <= 0 {
					continue
				}

				occupantMob.Data.Inventory.AddItem(item)
			}
		}

		// Create corpse in room
		room.Chests = append(room.Chests, Chest {
			Name: fmt.Sprintf("%s's Corpse", occupantMob.GetName()),
			DecayTimer: CHEST_CORPOSE_DECAY_DURATION,
			Inventory: occupantMob.Data.Inventory,
		})

		// Remove from mob array
		world.Mobs.Remove(occupantHandle)
	}
}
