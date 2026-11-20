package world

import (
	"fmt"
	"log"
	"mud/util"
	"slices"

	"github.com/mmcdole/lunar"
)

// This offset is added to threat levels so that NPCs do not forget attacks as easily
// This offset also helps to ensure NPCs do not ping-pong between targets too easily
const NPC_THREAT_OFFSET int32 = (3 * 60) / WORLD_SECONDS_PER_UPDATE

// This percent is used to decide whether to switch threats. In order for an NPC to switch to a higher
// threat target, the target must have 10% more threat than the NPC's current target
const NPC_THREAT_SWITCH_PERCENT float32 = 0.10

type NpcMovementType int
const (
	NPC_MOVEMENT_TYPE_STATIONARY NpcMovementType = iota
	NPC_MOVEMENT_TYPE_WANDER
	NPC_MOVEMENT_TYPE_COUNT
)

type NpcThreat struct {
	handle MobHandle
	level int32
}

type Npc struct {
	data *NpcData
	MobData

	disposition NpcDisposition
	movementType NpcMovementType
	mobHandle MobHandle
	instance lua.Value

	threats []NpcThreat
}

type NpcSpawnParams struct {
	data *NpcData
	room int
	movementType NpcMovementType
}

func (world *World) spawnNpc(params NpcSpawnParams) MobHandle {
	// Init NPC
	npc := &Npc {
		data: params.data,
		MobData: MobData {
			Name: params.data.name,
			Room: params.room,

			Health: params.data.stats.Values[STAT_MAX_HEALTH],
			Mana: params.data.stats.Values[STAT_MAX_MANA],
			Stats: params.data.stats,
		},

		disposition: params.data.startingDisposition,
		movementType: params.movementType,

		threats: make([]NpcThreat, 0, 1),
	}

	// Spawn mob
	mob := NewNpcMob(npc)
	npc.mobHandle = world.Mobs.Push(mob)

	// Add to room
	world.Rooms[npc.Room].AddOccupant(npc.mobHandle)

	// Call init
	npc.instance = npc.callInit(world)

	log.Printf("Spawned NPC %s in room %s.", npc.data.id, world.Rooms[npc.Room].Name)
	return npc.mobHandle
}

// MOB FUNCTIONS
// Called through the Mob functions of the same name

func (npc *Npc) DescribeToPlayer(world *World, playerKey string) {
	world.MessagePlayer(playerKey, npc.data.description)
	switch npc.disposition {
		case NPC_DISPOSITION_FRIENDLY:
			world.MessagePlayer(playerKey, "They appear <g>friendly</g>.")
		case NPC_DISPOSITION_NEUTRAL:
			world.MessagePlayer(playerKey, "They appear <y>neutral</y>.")
		case NPC_DISPOSITION_HOSTILE:
			world.MessagePlayer(playerKey, "They appear <r>hostile</r>.")
	}
}

func (npc *Npc) InventoryCapacity() int32 {
	return INVENTORY_CAPACITY_UNLIMITED
}

// UPDATE

func (npc *Npc) update(world *World) {
	// Update threats
	index := 0
	for index < len(npc.threats) {
		// Decrement threat level
		npc.threats[index].level--

		// Check if mob is still alive
		threatMob, threatExists := world.Mobs.GetIfExists(npc.threats[index].handle)
		threatIsAlive := threatExists && !threatMob.isDead()

		if npc.threats[index].level <= 0 || !threatIsAlive {
			npc.threats = util.RemoveAtUnordered(npc.threats, index)
		} else {
			index++
		}
	}

	npc.callUpdate(world)
	npcMob := world.Mobs.Get(npc.mobHandle)

	// Acquire a target from the threat table
	target, shouldSwitchTargets := npc.shouldSwitchTargets(world)
	if shouldSwitchTargets {
		npcMob.SetModeAttack(world, target)
	}

	// If hostile and idle, find a target to attack
	if npc.disposition == NPC_DISPOSITION_HOSTILE && npcMob.Target.Type == MOB_TARGET_NONE {
		room := &world.Rooms[npc.Room]

		// Find a target enemy
		targetIndex := slices.IndexFunc(room.Occupants, func (handle MobHandle) bool {
			occupant := world.Mobs.Get(handle)
			return !occupant.isDead() && occupant.IsPlayer()
		})

		// If target enemy found, set mode to attack
		if targetIndex != -1 {
			npcMob.SetModeAttack(world, room.Occupants[targetIndex])
		}
	}
}

// COMBAT

func (npc *Npc) GetAttacks() []MobAttack {
	attacks := make([]MobAttack, 0, npc.data.attacksPerRound)
	for _ = range npc.data.attacksPerRound {
		attacks = append(attacks, MobAttack {
			damage: npc.data.damage,
			penetration: npc.data.penetration,
		})
	}

	return attacks
}

func (npc *Npc) onAttacked(attackerHandle MobHandle, damage int32) {
	npc.addThreat(attackerHandle, damage)
}

func (npc *Npc) getThreat(mobHandle MobHandle) int32 {
	for index := range len(npc.threats) {
		if npc.threats[index].handle == mobHandle {
			return npc.threats[index].level
		}
	}

	return 0
}

func (npc *Npc) addThreat(mobHandle MobHandle, amount int32) {
	amount += NPC_THREAT_OFFSET

	for index := range len(npc.threats) {
		if npc.threats[index].handle == mobHandle {
			npc.threats[index].level += amount
			return
		}
	}

	npc.threats = append(npc.threats, NpcThreat {
		handle: mobHandle,
		level: amount,
	})
}

func (npc *Npc) setThreat(mobHandle MobHandle, amount int32) {
	for index := range len(npc.threats) {
		if npc.threats[index].handle == mobHandle {
			npc.threats[index].level = amount
			return
		}
	}

	npc.threats = append(npc.threats, NpcThreat {
		handle: mobHandle,
		level: amount,
	})
}

// Returns the highest threat that is alive and in the NPC's room, or a level of 0 if there is none.
// Threats outside the room stay in the table so that the NPC remembers them if they come back
func (npc *Npc) getHighestThreat(world *World) (MobHandle, int32) {
	highestIndex := -1
	for index := range len(npc.threats) {
		threatMob, threatExists := world.Mobs.GetIfExists(npc.threats[index].handle)
		if !threatExists || threatMob.isDead() || threatMob.Room != npc.Room {
			continue
		}

		if highestIndex == -1 || npc.threats[index].level > npc.threats[highestIndex].level {
			highestIndex = index
		}
	}

	if highestIndex == -1 {
		return MobHandle{}, 0
	}

	return npc.threats[highestIndex].handle, npc.threats[highestIndex].level
}

func (npc *Npc) setHighestThreat(world *World, mobHandle MobHandle, bonus int32) {
	_, highestThreatLevel := npc.getHighestThreat(world)
	threatLevel := int32(float32(highestThreatLevel) * (1.0 + NPC_THREAT_SWITCH_PERCENT)) + bonus
	npc.setThreat(mobHandle, threatLevel)
}

func (npc *Npc) shouldSwitchTargets(world *World) (MobHandle, bool) {
	highestThreat, highestThreatLevel := npc.getHighestThreat(world)

	// If we have no highest threat, then don't switch targets
	if highestThreatLevel == 0 {
		return MobHandle{}, false
	}

	// If we are already targeting the highest threat, then don't switch targets
	npcMob := world.Mobs.Get(npc.mobHandle)
	if npcMob.isTargeting(highestThreat) {
		return MobHandle{}, false
	}

	// If the NPC has a current target whose threat is close to the highest threat, don't switch targets
	if npcMob.Target.Type == MOB_TARGET_MOB && npcMob.isTargetValid(world) {
		targetThreatLevel := npc.getThreat(npcMob.Target.Value.(MobHandle))
		if targetThreatLevel != 0 {
			difference := highestThreatLevel - targetThreatLevel
			percentDifference := float32(difference) / float32(targetThreatLevel)
			if percentDifference < NPC_THREAT_SWITCH_PERCENT {
				return MobHandle{}, false
			}
		}
	}

	return highestThreat, true
}

// SCRIPT

func (npc *Npc) getFunctionDesc(name string) string {
	return fmt.Sprintf("npc.%s.%s()", npc.data.id, name)
}

func (npc *Npc) callInit(world *World) lua.Value {
	// Skip if init() was not defined
	if npc.data.init == nil {
		return lua.Nil()
	}

	// Get lua handle
	selfLuaHandle, ok := world.GetMobLuaHandle(npc.mobHandle)
	if !ok {
		return lua.Nil()
	}

	// Call hook
	functionDesc := npc.getFunctionDesc("init")
	result, ok := world.ScriptCall(npc.data.init.Value(), functionDesc, selfLuaHandle)
	if !ok || result.IsNil() {
		return lua.Nil()
	}
	if result.Kind() != lua.TableKind {
		log.Printf("warn: %s returned a %s instead of a table.", functionDesc, result.Kind().String())
		return lua.Nil()
	}

	return result
}

func (npc *Npc) callUpdate(world *World) {
	if npc.data.update == nil {
		return
	}

	// Get lua handle
	selfLuaHandle, ok := world.GetMobLuaHandle(npc.mobHandle)
	if !ok {
		return
	}

	// Call hook
	functionDesc := npc.getFunctionDesc("update")
	world.ScriptCall(npc.data.update.Value(), functionDesc, selfLuaHandle, npc.instance)
}
