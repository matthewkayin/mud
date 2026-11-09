package world

import (
	"fmt"
	"github.com/mmcdole/lunar"
)

type Mob struct {
	PlayerCharacter *Character
	Handle MobHandle

	*MobData

	// Created on first use by getLuaHandle so that each mob has one handle userdata
	luaHandle *lua.UserData
}

func NewMob(data *MobData) Mob {
	return Mob {
		MobData: data,
	}
}

func NewMobFromCharacter(character *Character) Mob {
	mob := NewMob(&character.MobData)
	mob.PlayerCharacter = character

	return mob
}

func (mob *Mob) IsPlayer() bool {
	return mob.PlayerCharacter != nil
}

func (mob *Mob) InventoryCapacity() int32 {
	if mob.IsPlayer() {
		return INVENTORY_CAPACITY_PLAYER
	}

	return INVENTORY_CAPACITY_UNLIMITED
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
