package world

import (
	"log"
	"fmt"
	"slices"
	"github.com/mmcdole/lunar"
)

type NpcMovementType int
const (
	NPC_MOVEMENT_TYPE_STATIONARY NpcMovementType = iota
	NPC_MOVEMENT_TYPE_WANDER
	NPC_MOVEMENT_TYPE_COUNT
)

type Npc struct {
	data *NpcData
	MobData

	disposition NpcDisposition
	movementType NpcMovementType
	mobHandle MobHandle
	instance lua.Value
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

			Health: params.data.maxHealth,
			Mana: params.data.maxMana,

			Inventory: params.data.loot.getLoot(world),
		},

		disposition: params.data.startingDisposition,
		movementType: params.movementType,
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

func (npc *Npc) MaxHealth() int32 {
	return npc.data.maxHealth
}

func (npc *Npc) MaxMana() int32 {
	return npc.data.maxMana
}

func (npc *Npc) InventoryCapacity() int32 {
	return INVENTORY_CAPACITY_UNLIMITED
}

// UPDATE

func (npc *Npc) update(world *World) {
	npc.callUpdate(world)

	npcMob := world.Mobs.Get(npc.mobHandle)

	// If hostile, find a target to attack
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

func (npc *Npc) getInitiative() int32 {
	return npc.data.initiative
}

func (npc *Npc) getDeflection() int32 {
	return npc.data.deflection
}

func (npc *Npc) getDeflectionAmount() int32 {
	return npc.data.deflectionAmount
}

func (npc *Npc) getResistance() int32 {
	return npc.data.resistance
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
		log.Printf("Warn: %s returned a %s instead of a table.", functionDesc, result.Kind().String())
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
