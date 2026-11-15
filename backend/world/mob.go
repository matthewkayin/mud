package world

import (
	"fmt"
	"github.com/mmcdole/lunar"
)

type MobFaction int
const (
	// Consists of players and friendly NPCs, apologies in advance to all orc players
	MOB_FACTION_ALLIANCE MobFaction = iota
	// Consists of non-friendly NPCs
	MOB_FACTION_HORDE
)

type MobMode int
const (
	MOB_MODE_IDLE MobMode = iota
	MOB_MODE_ATTACK
)

type MobTargetType int
const (
	MOB_TARGET_NONE MobTargetType = iota
	MOB_TARGET_MOB
)

type MobTarget struct {
	Type MobTargetType
	Value any
}

type MobOwner interface {
	IsPlayer() bool
	IsNpc() bool
	GetPlayerKey() string
	getMobData() *MobData
	getFaction() MobFaction
	DescribeToPlayer(world *World, playerKey string)

	MaxHealth() int32
	MaxMana() int32
	InventoryCapacity() int32

	Message(world *World, message string)
	update(world *World)
}

type Mob struct {
	Owner MobOwner
	Handle MobHandle

	*MobData

	Mode MobMode
	Target MobTarget

	// Created on first use by getLuaHandle so that each mob has one handle userdata
	luaHandle *lua.UserData
}

func NewMob(owner MobOwner) Mob {
	return Mob {
		Owner: owner,
		MobData: owner.getMobData(),

		Mode: MOB_MODE_IDLE,
		Target: MobTarget {
			Type: MOB_TARGET_NONE,
			Value: nil,
		},
	}
}

func (mob *Mob) MaxHealth() int32 {
	return mob.Owner.MaxHealth()
}

func (mob *Mob) MaxMana() int32 {
	return mob.Owner.MaxMana()
}

func (mob *Mob) isDead() bool {
	return mob.Health <= 0
}

// Set Mode

func (mob *Mob) SetModeIdle() {
	mob.Mode = MOB_MODE_IDLE
	mob.Target = MobTarget {
		Type: MOB_TARGET_NONE,
		Value: nil,
	}
}

func (mob *Mob) SetModeAttack(targetHandle MobHandle) {
	mob.Mode = MOB_MODE_ATTACK
	mob.Target = MobTarget {
		Type: MOB_TARGET_MOB,
		Value: targetHandle,
	}
}

// Behavior

func (mob *Mob) Move(world *World, newRoomIndex int) {
	oldRoom := &world.Rooms[mob.Room]
	oldRoom.RemoveOccupant(mob.Handle)

	world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> left the room.", mob.Name))
	mob.Room = newRoomIndex

	newRoom := &world.Rooms[newRoomIndex]
	newRoom.AddOccupant(mob.Handle)

	world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> entered the room.", mob.Name))
}

func (mob *Mob) update(world *World) {
	mob.Owner.update(world)
}
