package world

import (
	"fmt"
)

type Mob struct {
	PlayerCharacter *Character
	Handle MobHandle

	*MobData
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

func (mob *Mob) Move(world *World, newRoomIndex int) {
	oldRoom := &world.Rooms[mob.Room]
	oldRoom.RemoveOccupant(mob.Handle)

	world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> left the room.", mob.Name))
	mob.Room = newRoomIndex

	newRoom := &world.Rooms[newRoomIndex]
	newRoom.AddOccupant(mob.Handle)

	world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> entered the room.", mob.Name))
}
