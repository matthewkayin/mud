package world

import (
	"fmt"
	"github.com/mmcdole/lunar"
)

type MobOwner interface {
	IsPlayer() bool
	IsNpc() bool
	GetPlayerKey() string
	getMobData() *MobData

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

	// Created on first use by getLuaHandle so that each mob has one handle userdata
	luaHandle *lua.UserData
}

func NewMob(owner MobOwner) Mob {
	return Mob {
		Owner: owner,
		MobData: owner.getMobData(),
	}
}

func (mob *Mob) MaxHealth() int32 {
	return mob.Owner.MaxHealth()
}

func (mob *Mob) MaxMana() int32 {
	return mob.Owner.MaxMana()
}

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
