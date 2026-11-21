package world

import (
	"fmt"
	"log"
	"strings"
	"math/rand/v2"
	"github.com/mmcdole/lunar"
)

// Escape chance starts at 1.0, meaning you have full escape chance
// After an escape attempt fails, your escape chance gets dropped to 0.0
// It will recover a little each tick. At a rate of 0.1, your recovery chance recovers in 30s
const MOB_ESCAPE_CHANCE_MAX float32 = 1.0
const MOB_ESCAPE_CHANCE_RECOVERY_RATE float32 = 0.1

const MOB_ALERTNESS_MAX float32 = 1.0
const MOB_ALERTNESS_COOLDOWN_RATE float32 = 0.05

const MOB_TAUNT_COOLDOWN_MAX float32 = 1.0
const MOB_TAUNT_COOLDOWN_RATE float32 = 0.1

// Each point of deflection above the attack's penetration adds this much percent chance to deflect
const MOB_DEFLECTION_PER_POINT = 5

type MobMode int
const (
	MOB_MODE_IDLE MobMode = iota
	MOB_MODE_ATTACK
	MOB_MODE_CAST
	MOB_MODE_USE_ITEM
	MOB_MODE_TAUNT
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

type MobValidateTargetParams struct {
	Hook *lua.Function
	Target MobTarget
	Noun string
	Verb string
	Name string
}

type MobAttack struct {
	damage int32
	penetration int32

	// Set for attacks made with an equipped weapon, so that the weapon can lose durability
	hasWeaponSlot bool
	weaponSlot EquipmentSlot
}

type Mob struct {
	// Exactly one of these is set
	Character *Character
	Npc *Npc

	Handle MobHandle

	*MobData

	Mode MobMode
	Target MobTarget

	// Created on first use by getLuaHandle so that each mob has one handle userdata
	luaHandle *lua.UserData

	// The item that MOB_MODE_USE_ITEM will use
	itemToUse ItemId

	// MOB_MODE_CAST state
	spellTimer int

	// Set when an action is queued and cleared on the mob's turn (Mob.update)
	HasPendingAction bool

	escapeChance float32
	alertness float32
	tauntCooldown float32
}

func NewCharacterMob(character *Character) Mob {
	mob := newMob(&character.MobData)
	mob.Character = character
	return mob
}

func NewNpcMob(npc *Npc) Mob {
	mob := newMob(&npc.MobData)
	mob.Npc = npc
	return mob
}

func newMob(mobData *MobData) Mob {
	return Mob {
		MobData: mobData,

		Mode: MOB_MODE_IDLE,
		Target: MobTarget {
			Type: MOB_TARGET_NONE,
			Value: nil,
		},
		HasPendingAction: false,
	}
}

// Owner
// These functions hide whether the mob is a character or an NPC

func (mob *Mob) IsPlayer() bool {
	return mob.Character != nil
}

func (mob *Mob) IsNpc() bool {
	return mob.Npc != nil
}

func (mob *Mob) IsEvil() bool {
	if mob.IsPlayer() {
		return false
	}
	return mob.Npc.disposition != NPC_DISPOSITION_FRIENDLY
}

func (mob *Mob) IsHostile() bool {
	if mob.IsPlayer() {
		return false
	}
	return mob.Npc.disposition == NPC_DISPOSITION_HOSTILE
}

// Sends a message to the mob's player. Does nothing for NPCs
func (mob *Mob) Message(world *World, message string) {
	if mob.Character == nil {
		return
	}
	world.MessagePlayer(mob.Character.PlayerKey, message)
}

func (mob *Mob) DescribeToPlayer(world *World, playerKey string) {
	if mob.IsPlayer() {
		mob.Character.DescribeToPlayer(world, playerKey)
	} else {
		mob.Npc.DescribeToPlayer(world, playerKey)
	}

	world.MessagePlayer(playerKey, fmt.Sprintf("%s is %s.", mob.Name, strings.ToLower(mob.Condition().String())))
}

// SECONDARY STATS

// Stats live on MobData. Characters calculate them from their attributes and gear, NPCs copy them from their data on spawn
func (mob *Mob) Stat(stat Stat) int32 {
	return mob.Stats.Values[stat]
}

func (mob *Mob) InventoryCapacity() int32 {
	if mob.IsPlayer() {
		return mob.Character.InventoryCapacity()
	}
	return mob.Npc.InventoryCapacity()
}

func (mob *Mob) GetAttacks() []MobAttack {
	if mob.IsPlayer() {
		return mob.Character.GetAttacks()
	}
	return mob.Npc.GetAttacks()
}

func (mob *Mob) isDead() bool {
	return mob.Health <= 0
}

func (mob *Mob) IsInCombat(world *World) bool {
	room := &world.Rooms[mob.Room]
	for _, occupantHandle := range room.Occupants {
		occupant := world.Mobs.Get(occupantHandle)
		if mob.IsPlayer() && occupant.IsHostile() {
			return true
		}
		if mob.IsHostile() && occupant.IsPlayer() {
			return true
		}
	}

	return false
}

// Set Mode

// Silently resets the mob to idle. Use Stop instead when the mob chose to cancel its action
func (mob *Mob) SetModeIdle() {
	mob.Mode = MOB_MODE_IDLE
	mob.Target = MobTarget {
		Type: MOB_TARGET_NONE,
		Value: nil,
	}
	mob.HasPendingAction = false
}

// Cancels the mob's current action, announcing the cancellation if the action was in progress
func (mob *Mob) Stop(world *World) {
	mob.announceActionCancelled(world)
	mob.SetModeIdle()
}

func (mob *Mob) SetModeAttack(world *World, targetHandle MobHandle) {
	mob.beginAction(world, MOB_MODE_ATTACK, MobTarget {
		Type: MOB_TARGET_MOB,
		Value: targetHandle,
	})
}

func (mob *Mob) SetModeUseItem(world *World, itemToUse ItemId, target MobTarget) {
	mob.beginAction(world, MOB_MODE_USE_ITEM, target)
	mob.itemToUse = itemToUse
}

func (mob *Mob) SetModeTaunt(world *World, targetHandle MobHandle) {
	mob.beginAction(world, MOB_MODE_TAUNT, MobTarget {
		Type: MOB_TARGET_MOB,
		Value: targetHandle,
	})
}

// Every mode the mob chooses to start (attack, cast, use item) must be set through here,
// so that replacing an in-progress action always announces its cancellation
func (mob *Mob) beginAction(world *World, mode MobMode, target MobTarget) {
	mob.announceActionCancelled(world)
	mob.Mode = mode
	mob.Target = target
	mob.HasPendingAction = true
}

// Announces that the mob cancelled its current action. Says nothing if the action hasn't started yet,
// if its target is gone (dead or left the room), or if the action is an item use
func (mob *Mob) announceActionCancelled(world *World) {
	if mob.HasPendingAction {
		return
	}
	if !mob.isTargetValid(world) {
		return
	}

	switch mob.Mode {
		case MOB_MODE_ATTACK: {
			targetMob := world.Mobs.Get(mob.Target.Value.(MobHandle))
			world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> stopped attacking <y>%s</y>.", mob.Name, targetMob.Name))
		}

		case MOB_MODE_CAST: {
			world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> cancelled their spell.", mob.Name))
		}
	}
}

func (mob *Mob) IsAttacking(targetHandle MobHandle) bool {
	if mob.Mode != MOB_MODE_ATTACK || mob.Target.Type != MOB_TARGET_MOB {
		return false
	}
	return mob.Target.Value.(MobHandle) == targetHandle
}

// Behavior

func (mob *Mob) Move(world *World, newRoomIndex int) {
	// Clear the mob's action
	if mob.Mode == MOB_MODE_ATTACK || mob.Mode == MOB_MODE_CAST {
		mob.SetModeIdle()
	}

	oldRoom := &world.Rooms[mob.Room]
	oldRoom.RemoveOccupant(mob.Handle)

	world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> left the room.", mob.Name))
	mob.Room = newRoomIndex

	newRoom := &world.Rooms[newRoomIndex]
	newRoom.AddOccupant(mob.Handle)

	world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> entered the room.", mob.Name))
}

// UPDATE

func (mob *Mob) update(world *World) {
	if mob.IsNpc() {
		mob.Npc.update(world)
	}

	// This is the mob's turn, so any queued action resolves now
	mob.HasPendingAction = false

	// Update cooldowns
	mob.escapeChance = min(mob.escapeChance + MOB_ESCAPE_CHANCE_RECOVERY_RATE, MOB_ESCAPE_CHANCE_MAX)
	mob.alertness = max(mob.alertness - MOB_ALERTNESS_COOLDOWN_RATE, 0.0)
	mob.tauntCooldown = max(mob.tauntCooldown - MOB_TAUNT_COOLDOWN_RATE, 0.0)

	switch mob.Mode {
		// ATTACK

		case MOB_MODE_ATTACK: {
			// Check if target exists
			if !mob.isTargetValid(world) {
				mob.SetModeIdle()
				break
			}

			// Attack
			targetMob := mob.getTargetMobIfExists(world)
			attacks := mob.GetAttacks()
			for _, attack := range attacks {
				mob.attackTarget(world, targetMob, attack)
			}
		}

		// USE ITEM

		case MOB_MODE_USE_ITEM: {
			mob.useItem(world)
			mob.SetModeIdle()
		}

		// TAUNT

		case MOB_MODE_TAUNT: {
			// Check if target exists
			if !mob.isTargetValid(world) {
				mob.SetModeIdle()
				break
			}

			// Check to make sure taunter is a player
			if !mob.IsPlayer() {
				log.Printf("warn: mob %s tried to taunt, but they are not a player.", mob.Name)
				mob.SetModeIdle()
				break
			}

			// Check to make sure target is evil
			targetMob := mob.getTargetMobIfExists(world)
			if !targetMob.IsEvil() {
				log.Printf("warn: tried to taunt against %s, but they are not an evil NPC.", targetMob.Name)
				mob.SetModeIdle()
				break
			}

			// Narrate taunt
			world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> taunted <y>%s</y>!", mob.Name, targetMob.Name))

			// Roll taunt
			success := mob.rollTaunt(targetMob)
			if success {
				targetMob.Npc.setHighestThreat(world, mob.Handle, mob.Character.Attribute(ATTR_STR))
				targetMob.SetModeAttack(world, mob.Handle)
				world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> is now angry at <y>%s</y>!", targetMob.Name, mob.Name))
			} else {
				world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> ignored the taunt.", targetMob.Name))
			}

			// The taunt resolves once. Taunting again right away has no chance until the cooldown recovers
			mob.tauntCooldown = MOB_TAUNT_COOLDOWN_MAX
			mob.SetModeIdle()
		}
	}
}

func (mob *Mob) damage(attackerHandle MobHandle, damage int32) {
	mob.Health -= damage
	mob.alertness = MOB_ALERTNESS_MAX
	// targetMob.rollForConcentration(world, damage)

	if mob.IsNpc() {
		mob.Npc.onAttacked(attackerHandle, damage)
	}
}

func (mob *Mob) attackTarget(world *World, targetMob *Mob, attack MobAttack) {
	// Roll to hit
	hit := mob.rollToHit(targetMob)
	if !hit {
		world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> dodged <y>%s's</y> attack!", targetMob.Name, mob.Name))
		return
	}

	// Roll crit
	crit := mob.rollCrit()

	// Check for deflection
	deflect := false
	if !crit {
		deflectChance := (targetMob.Stat(STAT_DEFLECTION) - attack.penetration) * MOB_DEFLECTION_PER_POINT
		deflectChance = max(min(deflectChance, 100), 0)
		deflect = rand.Float32() < (float32(deflectChance) / 100.0)
	}

	// On deflect, subtract armor
	if deflect {
		attack.damage -= (attack.damage * targetMob.Stat(STAT_DEFLECTION_AMOUNT)) / 100
	}

	// Subtract flat-damage resistance
	attack.damage -= targetMob.Stat(STAT_RESISTANCE)

	// Calculate final damage
	attack.damage = max(attack.damage, 1)

	// Deal damage
	targetMob.damage(mob.Handle, attack.damage)

	// Broadcast result to room
	var message string
	if crit {
		message = fmt.Sprintf("<r>Critical hit!</r> <y>%s</y> struck <y>%s</y> for <r>%d damage</r>.", mob.Name, targetMob.Name, attack.damage)
	} else if deflect {
		message = fmt.Sprintf("<y>%s's</y> attack glanced off <y>%s</y> for <r>%d damage</r>.", mob.Name, targetMob.Name, attack.damage)
	} else {
		message = fmt.Sprintf("<y>%s</y> struck <y>%s</y> for <r>%d damage</r>.", mob.Name, targetMob.Name, attack.damage)
	}
	world.MessageRoom(mob.Room, message)
	if targetMob.isDead() {
		world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> has <r>slain</r> <y>%s</y>.", mob.Name, targetMob.Name))
	}

	// Reduce weapon durability, unequipping the weapon if it breaks
	if attack.hasWeaponSlot && mob.Character != nil {
		mob.Character.subtractDurabilityFromEquipment(world, attack.weaponSlot)
	}

	if targetMob.IsPlayer() {
		// Subtract durability from Armor and Shield

		// We don't have to check if outfit is non-nil because
		// subtractDurabilityFromEquipment will do this for us
		targetMob.Character.subtractDurabilityFromEquipment(world, EQUIPMENT_SLOT_OUTFIT)

		// But we do have to check for shield, because the offhand
		// item might not be a shield
		shield := targetMob.Character.Equipment.GetShield()
		if shield != nil {
			targetMob.Character.subtractDurabilityFromEquipment(world, EQUIPMENT_SLOT_OFF_HAND)
		}
	}
}

func (mob *Mob) useItem(world *World) {
	if !mob.IsPlayer() {
		log.Printf("warn: NPC %s tried to call useItem()", mob.Name)
		return
	}

	// Find the item
	item, found := mob.Character.FindItem(mob.itemToUse)
	if !found {
		mob.Message(world, "You no longer have the item you were going to use.")
		return
	}
	itemType := item.ItemType

	if !mob.isTargetValid(world) {
		mob.Message(world, fmt.Sprintf("You couldn't use <g>%s</g> because your target is gone.", itemType.Name))
		return
	}

	world.MessageRoom(mob.Room, fmt.Sprintf("<y>%s</y> used <g>%s</g>!", mob.Name, itemType.Name))

	success := false
	switch itemType.Kind {
		case ITEM_KIND_CONSUMABLE: {
			consumableData := itemType.Data.(*ItemDataConsumable)
			success = mob.callOnUse(world, consumableData, itemType.Name)
		}
		default: {
			log.Printf("warn: %s tried to use item %s of kind %s, which can't be used.", mob.Name, itemType.Name, itemType.Kind.String())
		}
	}

	if !success {
		world.MessageRoom(mob.Room, "But it failed!")
		return
	}

	// Consume the item. Find it again in case the hook changed the mob's inventory
	index, found := mob.Character.Inventory.FindItemById(mob.itemToUse)
	if found {
		mob.Character.Inventory.RemoveItem(world, index)
		if mob.Character != nil {
			world.PlayerInventoryChanged(mob.Character.PlayerKey)
		}
	}
}

// Calls a consumable's on_use() on the mob's target. Returns false if the target was refused or the hook failed
func (mob *Mob) callOnUse(world *World, consumableData *ItemDataConsumable, itemName string) bool {
	// Validate the target again, since the world may have changed since the item was queued
	if consumableData.ValidateTarget != nil {
		err := mob.ValidateTarget(world, MobValidateTargetParams {
			Hook: consumableData.ValidateTarget,
			Target: mob.Target,
			Verb: "use",
			Name: itemName,
			Noun: "item",
		})
		if err != nil {
			return false
		}
	}

	args, ok := mob.buildTargetHookArgs(world, mob.Target)
	if !ok {
		return false
	}

	functionDesc := fmt.Sprintf("item.%s.on_use()", itemName)
	_, ok = world.ScriptCall(consumableData.onUse.Value(), functionDesc, args...)
	return ok
}

// Target

// Returns false if the mob's target no longer exists, such as a target mob that died or left the room
func (mob *Mob) isTargetValid(world *World) bool {
	switch mob.Target.Type {
		case MOB_TARGET_MOB: {
			targetMob, exists := world.Mobs.GetIfExists(mob.Target.Value.(MobHandle))
			return exists && !targetMob.isDead() && targetMob.Room == mob.Room
		}

		case MOB_TARGET_ITEM: {
			_, found := mob.findTargetItem(world, mob.Target.Value.(ItemTarget))
			return found
		}

		default:
			return true
	}
}

func (mob *Mob) getTargetMobIfExists(world *World) *Mob {
	if mob.Target.Type == MOB_TARGET_MOB {
		targetMob, exists := world.Mobs.GetIfExists(mob.Target.Value.(MobHandle))
		if exists {
			return targetMob
		}
	}

	return nil
}

func (mob *Mob) isTargeting(mobHandle MobHandle) bool {
	return mob.Target.Type == MOB_TARGET_MOB && mob.Target.Value.(MobHandle) == mobHandle
}

func (mob *Mob) ValidateTarget(world *World, params MobValidateTargetParams) error {
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
			itemTarget := target.Value.(ItemTarget)
			_, found := mob.findTargetItem(world, itemTarget)
			if !found {
				return args, false
			}

			args = append(args, lua.Number(float64(itemTarget.Id)))
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
