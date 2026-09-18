package game

import (
	"log"
	"sort"
	"fmt"
	"errors"
	"slices"
	"math/rand"
)

const ROOM_NONE int = -1

const CHEST_DOES_NOT_DECAY = -1
const CHEST_CORPOSE_DECAY_DURATION = 30 / GAME_SECONDS_PER_UPDATE

type Direction int
const (
	DIRECTION_NORTH = iota
	DIRECTION_SOUTH
	DIRECTION_EAST
	DIRECTION_WEST
	DIRECTION_COUNT
)

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
	Inventory Inventory
	Chests []Chest

	occupants []MobHandle
}

func DirectionToString(direction Direction) string {
	switch direction {
		case DIRECTION_NORTH:
			return "north"
		case DIRECTION_SOUTH:
			return "south"
		case DIRECTION_EAST:
			return "east"
		case DIRECTION_WEST:
			return "west"
		default:
			panic(fmt.Sprintf("%d is not a valid direction.", direction))
	}
}

func DirectionFromString(directionStr string) (Direction, bool) {
	switch directionStr {
		case "north":
			return DIRECTION_NORTH, true
		case "south":
			return DIRECTION_SOUTH, true
		case "east":
			return DIRECTION_EAST, true
		case "west":
			return DIRECTION_WEST, true
		default:
			return 0, false
	}
}

func (room *Room) MoveOccupant(gameState *GameState, occupantHandle MobHandle, direction Direction) error {
	// Check if there is an exit in that direction
	newRoomIndex := room.Exits[direction]
	if newRoomIndex == ROOM_NONE {
		return errors.New("There is no exit in that direction.")
	}

	// Check if the exit is locked
	if room.ExitIsLocked[direction] {
		return fmt.Errorf("The %s exit is locked.", DirectionToString(direction))
	}

	// Get a pointer to the new room
	newRoom := &gameState.world.Rooms[newRoomIndex]
	occupantMob := gameState.world.Mobs.Get(occupantHandle)

	// Move the occupant

	room.RemoveOccupant(occupantHandle)

	// Note that the order matters here, we don't want to send these messages to the moving
	// player. Since the occupantHandle is in neither room at this point, the broadcast
	// function will not send the messages into the occupant's inbox
	room.broadcast(gameState, fmt.Sprintf("%s left the room.", occupantMob.data.Name))
	newRoom.broadcast(gameState, fmt.Sprintf("%s entered the room.", occupantMob.data.Name))

	newRoom.AddOccupant(occupantHandle)
	occupantMob.data.Room = newRoomIndex

	return nil
}

func (room *Room) AddOccupant(handle MobHandle) {
	room.occupants = append(room.occupants, handle)
}

func (room *Room) RemoveOccupant(handle MobHandle) {
	occupantIndex := -1
	for index, occupant := range room.occupants {
		if occupant.Equals(handle) {
			occupantIndex = index
			break
		}
	}
	if occupantIndex == -1 {
		log.Printf("Warning: Tried to remove occupant with handle %d:%d from room %s, but no such occupant was found.",
			handle.id, handle.generation, room.Name)
	}

	room.RemoveOccupantByIndex(occupantIndex)
}

func (room *Room) RemoveOccupantByIndex(index int) {
	lastIndex := len(room.occupants) - 1
	room.occupants[index] = room.occupants[lastIndex]
	room.occupants = room.occupants[:lastIndex]
}

func (room *Room) Update(gameState *GameState) {
	// Chest / Corpse decay
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

	// Sort combatants by initiative order
	// Combatants are a clone of the room occupant handles because otherwise
	// the sorting will mess up multi-target identifiers i.e. Goblin #2 might
	// not be Goblin #2 anymore after the initiative sort
	combatants := slices.Clone(room.occupants)
	sort.Slice(combatants, func(i int, j int) bool {
		mobI := gameState.world.Mobs.Get(room.occupants[i])
		mobJ := gameState.world.Mobs.Get(room.occupants[j])

		// If they have the same agility, choose a random one to go first
		if mobI.data.Agility() == mobJ.data.Agility() {
			return rand.Intn(2) == 0
		}

		return mobI.data.Agility() > mobJ.data.Agility()
	})

	// Occupant update and combat
	for _, occupantHandle := range combatants {
		// Get occupant mob
		occupantMob := gameState.world.Mobs.Get(occupantHandle)

		// Skip if dead
		if occupantMob.IsDead() {
			continue
		}

		// Update mob
		occupantMob.Update(gameState)
	}

	// Remove dead occupants
	occupantIndex := 0
	for occupantIndex < len(room.occupants) {
		// Get occupant mob
		occupantHandle := room.occupants[occupantIndex]
		occupantMob := gameState.world.Mobs.Get(occupantHandle)

		// Check for mob death
		if !occupantMob.IsDead() {
			occupantIndex++
			continue
		}

		// Remove occupant
		room.RemoveOccupantByIndex(occupantIndex)

		// Trigger player on death
		if occupantMob.player != nil {
			occupantMob.player.onDeath(gameState)
		}

		// If player mob, remove their equipment so that it goes into their corpse
		if occupantMob.player != nil {
			for slotIndex := range EQUIPMENT_SLOT_COUNT {
				slot := EquipmentSlot(slotIndex)
				item, success := occupantMob.data.EquippedItems.Unequip(slot)
				if !success {
					continue
				}

				occupantMob.data.Inventory.AddItem(item)
			}
		}

		// Create corpse in room
		room.Chests = append(room.Chests, Chest {
			Name: fmt.Sprintf("%s's Corpse", occupantMob.data.Name),
			DecayTimer: CHEST_CORPOSE_DECAY_DURATION,
			Inventory: occupantMob.data.Inventory,
		})

		// Remove from mob array
		gameState.world.Mobs.Remove(occupantHandle)
	}
}

func (room *Room) broadcast(gameState *GameState, message string) {
	for _, occupantHandle := range room.occupants {
		occupantMob := gameState.world.Mobs.Get(occupantHandle)
		if occupantMob.player == nil {
			continue
		}

		*occupantMob.player.inbox <- message
	}
}
