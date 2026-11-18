package world

import (
	"fmt"
	"log"
	"math/rand/v2"
	"github.com/mmcdole/lunar"
)

// Each point of deflection above the attack's penetration adds this much percent chance to deflect
const MOB_DEFLECTION_PER_POINT = 5

type MobMode int
const (
	MOB_MODE_IDLE MobMode = iota
	MOB_MODE_ATTACK
	MOB_MODE_CAST
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

// An item in a mob's inventory or equipment, found by its instance ID
type MobItemTarget struct {
	Mob MobHandle
	Id ItemId
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
		return
	}
	mob.Npc.DescribeToPlayer(world, playerKey)
}

func (mob *Mob) MaxHealth() int32 {
	if mob.IsPlayer() {
		return mob.Character.MaxHealth()
	}
	return mob.Npc.MaxHealth()
}

func (mob *Mob) MaxMana() int32 {
	if mob.IsPlayer() {
		return mob.Character.MaxMana()
	}
	return mob.Npc.MaxMana()
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

func (mob *Mob) getInitiative() int32 {
	if mob.IsPlayer() {
		return mob.Character.getInitiative()
	}
	return mob.Npc.getInitiative()
}

func (mob *Mob) getDeflection() int32 {
	if mob.IsPlayer() {
		return mob.Character.getDeflection()
	}
	return mob.Npc.getDeflection()
}

func (mob *Mob) getDeflectionAmount() int32 {
	if mob.IsPlayer() {
		return mob.Character.getDeflectionAmount()
	}
	return mob.Npc.getDeflectionAmount()
}

func (mob *Mob) getResistance() int32 {
	if mob.Character != nil {
		return mob.Character.getResistance()
	}
	return mob.Npc.getResistance()
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

// Items

// Returns the item with the given ID from the mob's inventory or equipment
func (mob *Mob) FindItem(id ItemId) (*Item, bool) {
	// Check inventory
	index, found := mob.Inventory.FindItemById(id)
	if found {
		return &mob.Inventory.Items[index], true
	}

	// Check equipment
	if mob.IsPlayer() {
		for index := range EQUIPMENT_SLOT_COUNT {
			slot := EquipmentSlot(index)
			item := mob.Character.Equipment.Get(slot)
			if item != nil {
				return item, true
			}
		}
	}

	return nil, false
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

func (mob *Mob) update(world *World) {
	if mob.IsNpc() {
		mob.Npc.update(world)
	}

	// This is the mob's turn, so any queued action resolves now
	mob.HasPendingAction = false

	switch mob.Mode {
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

		case MOB_MODE_USE_ITEM: {
			mob.useItem(world)
			mob.SetModeIdle()
		}
	}
}

func (mob *Mob) attackTarget(world *World, targetMob *Mob, attack MobAttack) {
	// TODO:
	// - check for evasion
	// - check for crit
	crit := false

	// Check for deflection
	deflect := false
	if !crit {
		deflectChance := (targetMob.getDeflection() - attack.penetration) * MOB_DEFLECTION_PER_POINT
		deflectChance = max(min(deflectChance, 100), 0)
		deflect = rand.Float32() < (float32(deflectChance) / 100.0)
	}

	// On deflect, subtract armor
	if deflect {
		attack.damage = (attack.damage * targetMob.getDeflectionAmount()) / 100
	}

	// Subtract flat-damage resistance
	attack.damage -= targetMob.getResistance()

	// Calculate final damage
	attack.damage = max(attack.damage, 1)

	// Deal damage
	targetMob.Health -= attack.damage

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
	// else {
		// should this be moved into mob.damage()?
		// targetMob.rollForConcentration(world, damage)
	// }

	// Reduce weapon durability, unequipping the weapon if it breaks
	if attack.hasWeaponSlot && mob.Character != nil {
		mob.Character.subtractDurabilityFromEquipment(world, attack.weaponSlot)
	}

	if targetMob.IsPlayer() {
		// Subtract durability from Armor and Shield

		// We don't have to check if outfit is non-nil because
		// subtractDurabilityFromEquipment will do this for us
		mob.Character.subtractDurabilityFromEquipment(world, EQUIPMENT_SLOT_OUTFIT)

		// But we do have to check for shield, because the offhand
		// item might not be a shield
		shield := targetMob.Character.Equipment.GetShield()
		if shield != nil {
			mob.Character.subtractDurabilityFromEquipment(world, EQUIPMENT_SLOT_OFF_HAND)
		}
	}
}

func (mob *Mob) useItem(world *World) {
	// Find the item
	item, found := mob.FindItem(mob.itemToUse)
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
			log.Printf("Warn - %s tried to use item %s of kind %s, which can't be used.", mob.Name, itemType.Name, itemType.Kind.String())
		}
	}

	if !success {
		world.MessageRoom(mob.Room, "But it failed!")
		return
	}

	// Consume the item. Find it again in case the hook changed the mob's inventory
	index, found := mob.Inventory.FindItemById(mob.itemToUse)
	if found {
		mob.Inventory.RemoveItem(world, index)
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
			itemTarget := mob.Target.Value.(MobItemTarget)
			targetMob, exists := world.Mobs.GetIfExists(itemTarget.Mob)
			if !exists {
				return false
			}

			_, found := targetMob.FindItem(itemTarget.Id)
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
			itemTarget := target.Value.(MobItemTarget)
			targetMob, exists := world.Mobs.GetIfExists(itemTarget.Mob)
			if !exists {
				return args, false
			}

			_, found := targetMob.FindItem(itemTarget.Id)
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
			log.Printf("Warn - target type %d not handled in buildTargetHookArgs", target.Type)
			return args, false
		}
	}

	return args, true
}
