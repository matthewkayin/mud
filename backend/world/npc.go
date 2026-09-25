package world

import (
	"log"
	"fmt"
	"slices"
	"math/rand/v2"
	"mud/util"
)

// Rather than reset the NPC's sleepy timer after combat,
// we instead apply this adrenaline number to their sleepy timer,
// so sleepy time is delayed but not completely reset
const NPC_SLEEPY_ADRENALINE_DURATION int32 = (5 * 60) / WORLD_SECONDS_PER_UPDATE
const NPC_SURPRISE_DURATION int32 = 2

type NpcMode int
const (
	NPC_MODE_DEAD = iota
	NPC_MODE_IDLE
	NPC_MODE_SURPRISE
	NPC_MODE_AGGRO
	NPC_MODE_SLEEP
)

type NpcMovementType int
const (
	NPC_MOVEMENT_TYPE_SENTINEL = iota
	NPC_MOVEMENT_TYPE_WANDER
	// TODO? NPC_MOVEMENT_TYPE_HUNT and NPC_MOVEMENT_TYPE_PATROL
)

// This could be replaced with a fine-grained number later
type NpcDisposition int
const (
	NPC_DISPOSITION_NEUTRAL = iota
	NPC_DISPOSITION_HOSTILE
)

type Npc struct {
	// NPC "config" variables - These are public and saved to world JSON
	Type NpcType
	LevelRange util.Int32Range
	StartingDisposition NpcDisposition
	MovementType NpcMovementType
	SpawnRoom int
	RespawnDuration int32
	SleepDuration int32
	AwakeDuration int32
	MovementStepDuration int32
	DropCount int32
	Drops []NpcDrop

	// NPC "instance" variables - These are private and not saved to world JSON
	mobHandle MobHandle
	mode NpcMode
	disposition NpcDisposition
	timer int32
	sleepyTimer int32
}

//spawns the mob associated with the npc
func (npc *Npc) spawnMob(world *World) {
	npcData := NPC_DATA[npc.Type]

	// Create mob data
	level := npc.LevelRange.ChooseRandom()
	stats := calculateStatBlockAtLevel(&npcData.baseStats, &npcData.scaling, level)
	mobData := MobData {
		Name: npcData.name,
		Room: npc.SpawnRoom,

		Level: level,
		Experience: 0, // TODO

		Stats: stats,
		Spells: []Spell {},
		Inventory: npc.determineDrops(),
		Equipment: npcData.equipment,
	}

	mobData.Health = mobData.MaxHealth()
	mobData.Mana = mobData.MaxMana()

	// Init mob
	npcMob := MobInit(&mobData)
	npcMob.Npc = npc

	// Add mob to world
	npc.mobHandle = world.Mobs.Push(npcMob)
	npcRoom := &world.Rooms[npcMob.Data.Room]
	npcRoom.AddOccupant(world, npc.mobHandle)

	// Init behavior
	npc.setModeIdle()
	npc.disposition = npc.StartingDisposition
	npc.sleepyTimer = 1 + rand.Int32N(npc.AwakeDuration)

	world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s has spawned into this room.", npcMob.GetName()))
}

func (npc *Npc) determineDrops() Inventory {
	// Determine total drop chance among all NPC drops
	var dropChanceTotal int32 = 0
	for _, drop := range npc.Drops {
		dropChanceTotal += drop.dropChance
	}

	inventory := Inventory { Items: []Item {} }

	// Roll for an item and add it to the inventory
	for _ = range npc.DropCount {
		roll := rand.Int32N(dropChanceTotal)
		var dropChance int32 = 0
		var dropIndex int = 0

		for dropIndex < len(npc.Drops) {
			dropChance += npc.Drops[dropIndex].dropChance
			if roll <= dropChance {
				break
			}
			dropIndex++
		}

		if dropIndex == len(npc.Drops) {
			panic("Something bad happened during Npc drop determination")
		}

		drop := &npc.Drops[dropIndex]
		inventory.AddItem(Item {
			Id: drop.itemId,
			Amount: drop.amountRange.ChooseRandom(),
			Durability: drop.durabilityRange.ChooseRandom(),
		})
	}

	return inventory
}

func (npc *Npc) hasSleepCycle() bool {
	return npc.SleepDuration > 0 && npc.AwakeDuration > 0
}

func (npc *Npc) setModeIdle() {
	npc.mode = NPC_MODE_IDLE
	npc.timer = npc.MovementStepDuration
}

func (npc *Npc) update(world *World) {
	if npc.mode == NPC_MODE_DEAD {
		npc.timer--
		if npc.timer <= 0 {
			npc.spawnMob(world)
		}

		return
	}

	// Check for mob death
	npcMob, npcMobExists := world.Mobs.GetIfExists(npc.mobHandle)
	if !npcMobExists {
		npc.mode = NPC_MODE_DEAD
		npc.timer = npc.RespawnDuration
		return
	}

	switch npc.mode {
		case NPC_MODE_IDLE: {
			// Check if player in room
			if npc.disposition == NPC_DISPOSITION_HOSTILE {
				// Check if there is a player in the room
				roomHasPlayer := slices.ContainsFunc(world.Rooms[npcMob.Data.Room].Occupants, func (handle MobHandle) bool {
					mob := world.Mobs.Get(handle)
					return mob.PlayerCharacter != nil
				})

				// If room has player, get ready to fight
				if roomHasPlayer {
					npc.mode = NPC_MODE_SURPRISE
					npc.timer = NPC_SURPRISE_DURATION
					world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s is getting ready to fight!", npcMob.Data.Name))
					break
				}
			}

			// Decrement sleep timer
			if npc.hasSleepCycle() {
				npc.sleepyTimer--

				// If sleepy, sleep
				if npc.sleepyTimer <= 0 {
					npc.mode = NPC_MODE_SLEEP
					npc.sleepyTimer = max(npc.SleepDuration, npcMob.Data.MaxHealth() - npcMob.Data.Health)
					world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s lied down and went to sleep.", npcMob.Data.Name))
				}
			}

			// Update movement
			if npc.MovementType == NPC_MOVEMENT_TYPE_WANDER {
				npc.timer--

				if npc.timer <= 0 {
					npc.movementStep(world)
					npc.timer = npc.MovementStepDuration
				}
			}
		}

		case NPC_MODE_SURPRISE: {
			// Countdown surpise timer, then switch to aggro
			npc.timer--
			if npc.timer <= 0 {
				npc.mode = NPC_MODE_AGGRO
			}
		}

		case NPC_MODE_AGGRO: {
			// If mob is doing something, then don't interrupt it
			if npcMob.Mode != MOB_MODE_IDLE {
				break
			}

			// Find a target to attack
			npcRoom := &world.Rooms[npcMob.Data.Room]
			for _, targetHandle := range npcRoom.Occupants {
				// Don't attack yourself
				if targetHandle == npc.mobHandle {
					continue
				}

				// For now, only attack players
				targetMob := world.Mobs.Get(targetHandle)
				if targetMob.PlayerCharacter == nil {
					continue
				}

				// Found target, set to attack
				npcMob.SetModeAttack(world, npc.mobHandle, targetHandle)
				break
			}

			// If mob is still idle at this point, it means no
			// target was found, so go back to idle
			if npcMob.Mode == MOB_MODE_IDLE {
				npc.mode = NPC_MODE_IDLE
				npc.sleepyTimer += NPC_SLEEPY_ADRENALINE_DURATION
			}
		}

		case NPC_MODE_SLEEP: {
			// Decrement sleepy timer
			npc.sleepyTimer--

			// Heal mob
			if npcMob.Data.Health < npcMob.Data.MaxHealth() {
				npcMob.Data.Health++
			}

			// If sleepy timer is over, wake up!
			if npc.sleepyTimer < 0 {
				npc.setModeIdle()
				npc.sleepyTimer = npc.AwakeDuration
				world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s has woken up!", npcMob.Data.Name))
			}
		}
	}
}

func (npc *Npc) movementStep(world *World) {
	npcMob := world.Mobs.Get(npc.mobHandle)
	npcRoom := &world.Rooms[npcMob.Data.Room]

	switch npc.MovementType {
		case NPC_MOVEMENT_TYPE_SENTINEL: {
			log.Printf("Warn - movementStep() called on a sentinel NPC with type %d, mob name %s, and mob handle %d:%d.",
				npc.Type, npcMob.Data.Name, npc.mobHandle.Id, npc.mobHandle.Generation)
		}

		case NPC_MOVEMENT_TYPE_WANDER: {
			// Collect an array of all possible exits
			exitRoomIndices := make([]int, 0, 4)
			for directionIndex := range DIRECTION_COUNT {
				// Don't walk into non-existing or locked rooms
				adjacentRoomIndex := npcRoom.Exits[directionIndex]
				if adjacentRoomIndex == ROOM_NONE || npcRoom.ExitIsLocked[directionIndex] {
					continue
				}

				// Don't walk into adjacent rooms
				adjacentRoom := &world.Rooms[adjacentRoomIndex]
				if adjacentRoom.IsSafeZone {
					continue
				}

				exitRoomIndices = append(exitRoomIndices, adjacentRoomIndex)
			}

			// If there are no exits, don't move
			if len(exitRoomIndices) == 0 {
				return
			}

			// Otherwise, choose a random exit
			index := rand.IntN(len(exitRoomIndices))
			newRoomIndex := exitRoomIndices[index]

			npcRoom.MoveOccupant(world, npc.mobHandle, newRoomIndex)
		}
	}
}

func (npc *Npc) onAttacked(world *World) {
	npcMob := world.Mobs.Get(npc.mobHandle)
	if npc.mode == NPC_MODE_SLEEP {
		npc.mode = NPC_MODE_SURPRISE
		npc.timer = NPC_SURPRISE_DURATION * 2
		world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s was violently awoken from their nap! They seem disgruntled.", npcMob.Data.Name))
	}
}

func (npc *Npc) GetDescription() string {
	return NPC_DATA[npc.Type].description
}

func (npc *Npc) GetStatusDescription() (string, bool) {
	if npc.mode == NPC_MODE_SLEEP {
		return "is taking a nap.", true
	}
	return "", false
}
