package world

import (
	"fmt"
	"log"
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
	MOB_MODE_USE_ITEM
)

// Target

type MobTargetType int
const (
	MOB_TARGET_NONE MobTargetType = iota
	MOB_TARGET_MOB
	MOB_TARGET_ITEM
	MOB_TARGET_ARGUMENT
)

type MobTarget struct {
	Type MobTargetType
	Value any
}

type MobItemTarget struct {
	Mob MobHandle
	Location ItemLocation
	Id ItemId
}

// Owner

type MobOwner interface {
	IsPlayer() bool
	IsNpc() bool
	GetPlayerKey() string
	getMobData() *MobData
	GetFaction() MobFaction
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
	itemToUse ItemId
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

// Target

type MobValidateTargetParams struct {
	Hook *lua.Function
	Target MobTarget
	Argument string
	Noun string
	Verb string
	Name string
}

func (mob *Mob) ValidateTarget(world*World, params MobValidateTargetParams) error {
	args, ok := mob.buildTargetHookArgs(world, params.Target)
	if !ok {
		return fmt.Errorf("Something went wrong.")
	}

	functionDesc := fmt.Sprintf("%s.%s.validate_target()", params.Noun, params.Name)
	result, ok := world.ScriptCall(params.Hook.Value(), functionDesc, args...)
	if !ok {
		return fmt.Errorf("Something went wrong.")
	}

	errorString, hasError := result.AsString()
	if hasError {
		return fmt.Errorf("%s", errorString)
	}

	return nil
}

// Returns a list of lua values representing the mob's target, so that the target can be passed in to a lua hook
func (mob *Mob) buildTargetHookArgs(world *World, target MobTarget) ([]lua.Value, bool) {
	args := make([]lua.Value, 0, 1)

	// Mob lua handle
	mobLuaHandle, ok := world.GetMobLuaHandle(mob.Handle)
	if !ok {
		return args, false
	}
	args = append(args, mobLuaHandle)

	// Target
	switch target.Type {
		case MOB_TARGET_NONE: {}

		case MOB_TARGET_MOB: {
			targetHandle := target.Value.(MobHandle)
			targetLuaHandle, ok := world.GetMobLuaHandle(targetHandle)
			if !ok {
				return args, false
			}

			args = append(args, targetLuaHandle)
		}

		case MOB_TARGET_ITEM: {
			itemHandle := target.Value.(ItemHandle)
			itemHandle, found := world.relocateItemHandle(itemHandle)
			if !found {
				return args, false
			}

			itemLuaHandle, err := world.itemHandleType.New(itemHandle)
			if err != nil {
				return args, false
			}

			args = append(args, itemLuaHandle.Value())
		}

		case MOB_TARGET_ARGUMENT: {
			argument := target.Value.(string)
			args = append(args, lua.String(argument))
		}

		default: {
			log.Printf("warn: target type %d not handled in buildTargetHookArgs", target.Type)
			return args, false
		}
	}

	return args, true
}
