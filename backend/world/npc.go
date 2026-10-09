package world

import (
	"fmt"
	"log"
	"slices"
	"math/rand/v2"
	"encoding/json"
)

// Rather than reset the NPC's sleepy timer after combat,
// we instead apply this adrenaline number to their sleepy timer,
// so sleepy time is delayed but not completely reset
const NPC_SLEEPY_ADRENALINE_DURATION int32 = (5 * 60) / WORLD_SECONDS_PER_UPDATE
const NPC_SURPRISE_DURATION int32 = 1

const NPC_RESPAWN_DURATION = 60 / WORLD_SECONDS_PER_UPDATE
const NPC_MOVEMENT_STEP_DURATION = 60 / WORLD_SECONDS_PER_UPDATE

const NPC_MOVEMENT_TYPE_OVERRIDE_NONE = NPC_MOVEMENT_TYPE_COUNT

type NpcMode int
const (
	NPC_MODE_DEAD = iota
	NPC_MODE_IDLE
	NPC_MODE_SURPRISE
	NPC_MODE_AGGRO
)

type NpcMovementType int
const (
	NPC_MOVEMENT_TYPE_SENTINEL = iota
	NPC_MOVEMENT_TYPE_WANDER
	NPC_MOVEMENT_TYPE_COUNT
)

func (movementType NpcMovementType) String() string {
	switch movementType {
		case NPC_MOVEMENT_TYPE_SENTINEL:
			return "Sentinel"
		case NPC_MOVEMENT_TYPE_WANDER:
			return "Wander"
		default:
			return ""
	}
}

// This could be replaced with a fine-grained number later
type NpcDisposition int
const (
	NPC_DISPOSITION_NEUTRAL = iota
	NPC_DISPOSITION_HOSTILE
	NPC_DISPOSITION_FRIENDLY
	NPC_DISPOSITION_COUNT
)

func (disposition NpcDisposition) String() string {
	switch disposition {
		case NPC_DISPOSITION_NEUTRAL:
			return "Neutral"
		case NPC_DISPOSITION_HOSTILE:
			return "Hostile"
		case NPC_DISPOSITION_FRIENDLY:
			return "Friendly"
		default:
			return ""
	}
}

type Npc struct {
	// NPC "config" variables - tells us how to make a mob based on this NPC
	Id NpcId
	SpawnRoom int
	LevelRange Int32Range

	// Overrides - for things that would otherwise be specified by NPC data
	MovementTypeOverride NpcMovementType
	DropTableOverride DropTable

	// NPC "instance" variables - keeps track of the NPC's current state
	mobHandle MobHandle
	mode NpcMode
	disposition NpcDisposition
	timer int32
	shouldReset bool
}

type NpcJson struct {
	Key string
	SpawnRoom int
	LevelRange Int32Range

	MovementTypeOverride string
	DropTableOverride DropTable
}

func (npc *Npc) MarshalJSON() ([]byte, error) {
	npcJson := NpcJson {
		Key: NPC_DATA[npc.Id].Key,
		SpawnRoom: npc.SpawnRoom,
		LevelRange: npc.LevelRange,

		MovementTypeOverride: npc.MovementTypeOverride.String(),
		DropTableOverride: npc.DropTableOverride,
	}
	return json.Marshal(&npcJson)
}

func (npc *Npc) UnmarshalJSON(data []byte) error {
	var npcJson NpcJson
	err := json.Unmarshal(data, &npcJson)
	if err != nil {
		return err
	}

	var exists bool
	npc.Id, exists = NPC_KEY_TO_ID[npcJson.Key]
	if !exists {
		return fmt.Errorf("No NPC key matches '%s'.", npcJson.Key)
	}

	npc.SpawnRoom = npcJson.SpawnRoom
	npc.LevelRange = npcJson.LevelRange

	if npcJson.MovementTypeOverride == "" {
		npc.MovementTypeOverride = NPC_MOVEMENT_TYPE_OVERRIDE_NONE
	} else {
		npc.MovementTypeOverride, exists = EnumFromString(npcJson.MovementTypeOverride, NpcMovementType(NPC_MOVEMENT_TYPE_COUNT))
		if !exists {
			return fmt.Errorf("No NPC movement type matches '%s'.", npcJson.MovementTypeOverride)
		}
	}

	npc.DropTableOverride = npcJson.DropTableOverride

	return nil
}

func (npc *Npc) tryReset(world *World) {
	// No need to reset if the NPC is already dead,
	// it will just respawn after respawn timer is up
	if npc.mode == NPC_MODE_DEAD {
		npc.shouldReset = false
	}

	// If NPC is not dead, try to despawn mob
	npcMob, npcMobExists := world.Mobs.GetIfExists(npc.mobHandle)
	if !npcMobExists {
		npc.timer = 0
		npc.shouldReset = false
		return
	}

	// If players are still in the room, then don't despawn
	npcRoom := &world.Rooms[npcMob.Data.Room]
	if npcRoom.hasPlayerOccupants(world) {
		return
	}

	// Otherwise, despawn
	npcRoom.RemoveOccupant(npc.mobHandle)
	npc.mode = NPC_MODE_DEAD
	npc.timer = 0 // Trigger a respawn
	npc.shouldReset = false
}

//spawns the mob associated with the npc
func (npc *Npc) spawnMob(world *World) {
	// Check if the room is empty of players before spawning
	npcRoom := &world.Rooms[npc.SpawnRoom]
	if npcRoom.hasPlayerOccupants(world) {
		return
	}

	npcData := NPC_DATA[npc.Id]

	// Create mob data
	level := npc.LevelRange.ChooseRandom()
	stats := calculateStatBlockAtLevel(&npcData.stats, &npcData.scaling, level)
	mobData := MobData {
		Name: npcData.name,
		Room: npc.SpawnRoom,

		Level: level,
		Experience: npcData.experienceWorth + (npcData.experienceWorthScaling * (level - 1)),

		Stats: stats,
		Spells: []SpellId {},
		Inventory: npcData.dropTable.getLoot(),
		Equipment: npcData.equipment,
	}

	mobData.Health = mobData.MaxHealth()
	mobData.Mana = mobData.MaxMana()

	// Init mob
	npcMob := MobInit(&mobData)
	npcMob.Npc = npc

	// Add mob to world
	npc.mobHandle = world.Mobs.Push(npcMob)
	npcRoom.AddOccupant(world, npc.mobHandle)

	// Init behavior
	npc.setModeIdle()
	npc.disposition = npcData.startingDisposition

	if npcData.init != nil {
		world.luaState.Call(npcData.init.Value())
	}
}

func (npc *Npc) setModeIdle() {
	npc.mode = NPC_MODE_IDLE
	npc.timer = NPC_MOVEMENT_STEP_DURATION
}

func (npc *Npc) setModeSurprise(world *World) {
	npcMob := world.Mobs.Get(npc.mobHandle)

	npc.mode = NPC_MODE_SURPRISE
	npc.timer = NPC_SURPRISE_DURATION
	npcMob.alertness = MOB_ALERTNESS_MAX
}

func (npc *Npc) update(world *World) {
	// Try reset
	if npc.shouldReset {
		npc.tryReset(world)
	}

	// Respawn
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
		npc.timer = NPC_RESPAWN_DURATION
		return
	}

	npcData := NPC_DATA[npc.Id]
	if npcData.update != nil {
		world.luaState.Call(npcData.update.Value())
	}

	switch npc.mode {
		case NPC_MODE_IDLE: {
			// Check if player in room
			if npc.disposition == NPC_DISPOSITION_HOSTILE {
				// Check if there is a player in the room
				roomHasPlayer := slices.ContainsFunc(world.Rooms[npcMob.Data.Room].Occupants, func (handle MobHandle) bool {
					mob := world.Mobs.Get(handle)
					if mob.CheckFlag(MOB_FLAG_HIDDEN) {
						return false
					}
					if mob.PlayerCharacter == nil {
						return false
					}
					return true
				})

				// If room has player, get ready to fight
				if roomHasPlayer {
					npc.setModeSurprise(world)
					world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s is getting ready to fight!", npcMob.Data.Name))
					break
				}
			}

			// Update movement
			if npc.getMovementType() == NPC_MOVEMENT_TYPE_WANDER {
				npc.timer--

				if npc.timer <= 0 {
					npc.movementStep(world)
					npc.timer = NPC_MOVEMENT_STEP_DURATION
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

				// Don't attack hidden players
				if targetMob.CheckFlag(MOB_FLAG_HIDDEN) {
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
			}
		}
	}
}

func (npc *Npc) getMovementType() NpcMovementType {
	if npc.MovementTypeOverride != NPC_MOVEMENT_TYPE_OVERRIDE_NONE {
		return npc.MovementTypeOverride
	}

	return NPC_DATA[npc.Id].movementType
}

func (npc *Npc) movementStep(world *World) {
	npcMob := world.Mobs.Get(npc.mobHandle)
	npcRoom := &world.Rooms[npcMob.Data.Room]

	switch npc.getMovementType() {
		case NPC_MOVEMENT_TYPE_SENTINEL: {
			log.Printf("Warn - movementStep() called on a sentinel NPC with key %s, mob name %s, and mob handle %d:%d.",
				NPC_DATA[npc.Id].Key, npcMob.Data.Name, npc.mobHandle.Id, npc.mobHandle.Generation)
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

				// Don't walk into safe rooms
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

func (npc *Npc) OnEvent(world *World, event BehaviorEvent) {
	// First, try event through behavior
	// TODO: fire the appropriate NPC on event hook

	// If behavior did not handle event, fallback to default
	npcMob := world.Mobs.Get(npc.mobHandle)
	switch event.Type {
		case BEHAVIOR_EVENT_TYPE_ATTACKED: {
			// TODO: for friendly NPCs, like town guards,
			// make the NPC hostile only to the attacker, not
			// all players?
			if npc.disposition == NPC_DISPOSITION_NEUTRAL {
				npc.disposition = NPC_DISPOSITION_HOSTILE
			}
		}

		case BEHAVIOR_EVENT_TYPE_ITEM_GIVEN: {
			eventData := event.Data.(BehaviorEventItemGiven)

			playerMob := world.Mobs.Get(eventData.PlayerHandle)
			item := npcMob.Data.Inventory.RemoveItems(eventData.AddedToIndex, eventData.Amount)
			playerMob.Data.Inventory.AddItem(item)

			world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s is uninterested in this item. They returned it to %s.", npcMob.Data.Name, playerMob.Data.Name))
		}
	}
}

func (npc *Npc) GetDescription(world *World) string {
	return NPC_DATA[npc.Id].description
}

func (npc *Npc) GetStatusDescription(world *World) (string, bool) {
	npcData := NPC_DATA[npc.Id]
	if npcData.getStatusDescription == nil {
		return "", false
	}

	results, err := world.luaState.Call(npcData.getStatusDescription.Value())
	if err != nil {
		log.Printf("Warn - Error calling NPC get_description(): %s", err.Error())
		return "", false
	}

	if len(results) != 0 {
		log.Printf("Warn - NPC get_description() returned no result.")
		return "", false
	}

	result, ok := results[0].AsString()
	if !ok {
		log.Printf("Warn - NPC get_description() returned a non-string value.")
		return "", false
	}

	return result, true
}
