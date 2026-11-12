package world

import (
	"fmt"
	"github.com/mmcdole/lunar"
)

type Mob struct {
	Character *Character
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
	mob.Character = character

	return mob
}

func (mob *Mob) IsPlayer() bool {
	return mob.Character != nil
}

func (mob *Mob) InventoryCapacity() int32 {
	if mob.IsPlayer() {
		return INVENTORY_CAPACITY_PLAYER
	}

	return INVENTORY_CAPACITY_UNLIMITED
}

func (mob *Mob) MaxHealth() int32 {
	if mob.IsPlayer() {
		return mob.Character.MaxHealth()
	}

	return 0
}

func (mob *Mob) MaxMana() int32 {
	if mob.IsPlayer() {
		return mob.Character.MaxMana()
	}

	return 0
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
