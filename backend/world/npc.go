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
	mob := NewMob(npc)
	world.Mobs.Push(mob)
	npc.mobHandle = mob.Handle

	// Add to room
	world.Rooms[npc.Room].AddOccupant(npc.mobHandle)

	// Call init
	npc.instance = npc.callInit(world)

	log.Printf("Spawned NPC %s in room %s.", npc.data.id, world.Rooms[npc.Room].Name)
	return npc.mobHandle
}

// MOB OWNER FUNCTIONS

func (npc *Npc) IsPlayer() bool {
	return false
}

func (npc *Npc) IsNpc() bool {
	return true
}

func (npc *Npc) GetPlayerKey() string {
	// player-only function, returns nothing
	return ""
}

func (npc *Npc) getMobData() *MobData {
	return &npc.MobData
}

func (npc *Npc) getFaction() MobFaction {
	if npc.disposition == NPC_DISPOSITION_FRIENDLY {
		return MOB_FACTION_ALLIANCE
	}
	return MOB_FACTION_HORDE
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

func (npc *Npc) Message(world *World, message string) {
	// player-only function, does nothing for NPCs
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
			return !occupant.isDead() && occupant.Owner.getFaction() != npc.getFaction()
		})

		// If target enemy found, set mode to attack
		if targetIndex != -1 {
			npcMob.SetModeAttack(room.Occupants[targetIndex])
		}
	}
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
	world.ScriptCall(npc.data.init.Value(), functionDesc, selfLuaHandle, npc.instance)
}
