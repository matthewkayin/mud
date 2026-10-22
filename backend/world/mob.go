package world

import (
	"fmt"
	"log"
	"math/rand/v2"
	"slices"
	"mud/util"
	"github.com/mmcdole/lunar"
)

const MOB_PLAYER_NONE = -1

// Making Evasion K higher makes evasion chance smaller
// Making this smaller makes evasion chance greater
// Evasion chance of target = T / (T + (A * K))
// Where T is target AGI and A is attacker AGI
const MOB_EVASION_K float32 = 6.0

// Making Crit K higher increases the likelihood of crits
// Crit chance =  (AGI / 50) * 0.33
//			   => AGI * (0.33 / 50)
// A mob with a base AGI of 12 and high crit scaling will have 50 AGI at level 20,
// so 50 is roughly the "max agility" a mob can have
const MOB_CRIT_K float32 = 0.33 / 50.0
const MOB_CRIT_DAMAGE_MULTIPLIER float32 = 1.5

// Making Stealth K smaller makes it harder for monsters to see players
// Stealth chance = (ThiefAGI / (ThiefAGI + K * EnemyINT))
//                    * EscapeChance * (1 - EnemyAwareness)
//                    + Armor Penalty + Room Brightness
const MOB_STEALTH_K float32 = 0.5

// Escape chance starts at 1.0, meaning you have full escape chance
// After an escape attempt fails, your escape chance gets dropped to 0.0
// It will recover a little each tick. At a rate of 0.1, your recovery chance recovers in 30s
const MOB_ESCAPE_CHANCE_MAX float32 = 1.0
const MOB_ESCAPE_CHANCE_RECOVERY_RATE float32 = 0.1

// Magic condition duration = Base * (1 + (FTH / 50.0))
//                         => Base * (1 + (FTH * K))
// Each spell condition lasts for at minimum, Base turns
// But FTH increases the duration by a percentage of "max faith"
// Once FTH is at 50.0, then the duration will be Base x 2
const MOB_MAGIC_CONDITION_DURATION_K float32 = 1.0 / 50.0

const MOB_ALERTNESS_MAX float32 = 1.0
const MOB_ALERTNESS_COOLDOWN_RATE float32 = 0.05

const MOB_TAUNT_COOLDOWN_MAX float32 = 1.0
const MOB_TAUNT_COOLDOWN_RATE float32 = 0.05

type MobTargetType int
const (
	MOB_TARGET_NONE MobTargetType = iota
	MOB_TARGET_MOB
	MOB_TARGET_INVENTORY_ITEM
	MOB_TARGET_ARGUMENT
)

type MobTarget struct {
	Type MobTargetType
	Value any
}

type MobItemTarget struct {
	Id ItemId
	Index int
}

type MobValidateArgumentParams struct {
	Hook *lua.Function
	Argument string
	Noun string
	Verb string
	Name string
}

type MobOnUseOrCastParams struct {
	targetType TargetType
	hook *lua.Function
	validateArgument *lua.Function
	noun string
	verb string
	name string
}

type MobFlag uint32
const (
	MOB_FLAG_HIDDEN MobFlag = 1 << iota
)

type MobMode int32
const (
	MOB_MODE_IDLE MobMode = iota
	MOB_MODE_ATTACK
	MOB_MODE_CAST
	MOB_MODE_USE_ITEM
	MOB_MODE_CRAFT_ITEM
	MOB_MODE_TAUNT
)

type Mob struct {
	PlayerCharacter *Character
	Npc *Npc
	Handle MobHandle
	Data MobData

	Mode MobMode
	Target MobTarget

	flags MobFlag

	castSpell SpellId
	castTimer int32
	itemToUse MobItemTarget

	craftItemRecipe RecipeId
	craftItemAmount int32

	escapeChance float32
	alertness float32
	tauntCooldown float32
	fuzzyNumber int

	// Created on first use by getLuaHandle so that each mob has one handle userdata
	luaHandle *lua.UserData
}

func MobInit(data *MobData) Mob {
	mob := Mob {
		PlayerCharacter: nil,
		Data: data.Clone(),
		Mode: MOB_MODE_IDLE,
		Target: MobTarget {
			Type: MOB_TARGET_NONE,
			Value: nil,
		},

		escapeChance: MOB_ESCAPE_CHANCE_MAX,
		fuzzyNumber: 1,
	}

	mob.Data.Equipment.CalculateStatBonuses()
	return mob
}

func MobInitFromCharacter(character *Character) Mob {
	mob := MobInit(&character.Data)
	mob.PlayerCharacter = character

	return mob
}

func (mob *Mob) GetPlayerId() int {
	if !mob.IsPlayer() {
		return MOB_PLAYER_NONE
	}

	return mob.PlayerCharacter.PlayerId
}

func (mob *Mob) GetName() string {
	if mob.fuzzyNumber == 1 {
		return mob.Data.Name
	}
	return fmt.Sprintf("%s %d", mob.Data.Name, mob.fuzzyNumber)
}

func (mob *Mob) GetNameWithInjury() string {
	name := mob.GetName()

	if mob.Data.Health <= 0 {
		return name
	}

	if mob.Data.Health < mob.Data.MaxHealth() / 5 {
		return fmt.Sprintf("%s (Near Death)", name)
	}

	if mob.Data.Health < mob.Data.MaxHealth() / 2 {
		return fmt.Sprintf("%s (Injured)", name)
	}

	return name
}

func (mob *Mob) IsDead() bool {
	return mob.Data.Health <= 0
}

func (mob *Mob) IsInCombat(world *World) bool {
	// If player is hidden, then they are not in combat
	if mob.CheckFlag(MOB_FLAG_HIDDEN) {
		return false
	}

	// A mob is in combat if at least one occupant in the room is a hostile, non-sleeping NPC
	return slices.ContainsFunc(world.Rooms[mob.Data.Room].Occupants, func(handle MobHandle) bool {
		occupant := world.Mobs.Get(handle)
		return occupant.IsNpc() &&
			occupant.Npc.disposition == NPC_DISPOSITION_HOSTILE
	})
}

func (mob *Mob) CheckFlag(flag MobFlag) bool {
	return (mob.flags & flag) == flag
}

func (mob *Mob) SetFlag(flag MobFlag, value bool) {
	if value {
		mob.flags |= flag
	} else {
		mob.flags &= ^flag
	}
}

func (mob *Mob) GrantExperience(world *World, experience int32) {
	// This function is only meant for player mobs at this time
	if !mob.IsPlayer() {
		return
	}

	for experience > 0 && mob.Data.Level < MOB_MAX_LEVEL {
		if mob.Data.Experience + experience >= mob.Data.ExperienceToNextLevel {
			experience -= mob.Data.ExperienceToNextLevel

			// Increase level
			mob.Data.Experience = 0
			mob.Data.ExperienceToNextLevel = mob.Data.GetExpToNextLevel()
			mob.Data.Level++
			mob.PlayerCharacter.Data.Level = mob.Data.Level

			// Announce level up message
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Level up! <y>%s</y> is now level %d.", mob.Data.Name, mob.Data.Level))

			// Grant class unlocks
			classData := CLASS_DATA[mob.PlayerCharacter.Class]
			for _, unlock := range classData.UnlocksAtLevel[mob.Data.Level] {
				message := mob.PlayerCharacter.grantClassUnlock(unlock)
				world.messagePlayer(mob.PlayerCharacter.PlayerId, message)
			}

			// Recalculate stats
			mob.PlayerCharacter.recalculateStats()
			mob.Data.Stats = mob.PlayerCharacter.Data.Stats
			mob.Data.Abilities = mob.PlayerCharacter.Data.Abilities

			// Save the character to disk
			mob.SavePlayerCharacter()

			continue
		}

		mob.Data.Experience += experience
		experience = 0
	}
}

func (mob *Mob) SetModeIdle() {
	mob.Mode = MOB_MODE_IDLE
	mob.Target = MobTarget {
		Type: MOB_TARGET_NONE,
		Value: nil,
	}
}

func (mob *Mob) SetModeAttack(world *World, targetHandle MobHandle) {
	mob.Mode = MOB_MODE_ATTACK
	mob.Target = MobTarget {
		Type: MOB_TARGET_MOB,
		Value: targetHandle,
	}
}

func (mob *Mob) SetModeCast(world *World, mobHandle MobHandle, spell SpellId, target MobTarget) {
	mob.Mode = MOB_MODE_CAST
	mob.Target = target
	mob.castSpell = spell
	mob.castTimer = SPELL_DATA[spell].CastTime
}

func (mob *Mob) SetModeUseItem(world *World, mobHandle MobHandle, itemToUse MobItemTarget, target MobTarget) {
	mob.Mode = MOB_MODE_USE_ITEM
	mob.Target = target
	mob.itemToUse = itemToUse
}

func (mob *Mob) SetModeCraftItem(world *World, mobHandle MobHandle, recipe RecipeId, amount int32) {
	mob.Mode = MOB_MODE_CRAFT_ITEM
	mob.craftItemRecipe = recipe
	mob.craftItemAmount = amount
}

func (mob *Mob) SetModeTaunt(world *World, mobHandle MobHandle, targetHandle MobHandle) {
	mob.Mode = MOB_MODE_TAUNT
	mob.Target = MobTarget {
		Type: MOB_TARGET_MOB,
		Value: targetHandle,
	}
}

func (mob *Mob) Update(world *World) {
	if mob.IsDead() {
		return
	}

	mob.escapeChance = min(mob.escapeChance + MOB_ESCAPE_CHANCE_RECOVERY_RATE, MOB_ESCAPE_CHANCE_MAX)
	mob.alertness = max(mob.alertness - MOB_ALERTNESS_COOLDOWN_RATE, 0.0)
	mob.tauntCooldown = max(mob.tauntCooldown - MOB_TAUNT_COOLDOWN_RATE, 0.0)

	// Update conditions
	conditionIndex := 0
	for conditionIndex < len(mob.Data.Conditions) {
		condition := &mob.Data.Conditions[conditionIndex]

		// Update condition TTL
		if condition.Duration != CONDITION_DURATION_INFINITE {
			condition.Timer++

			// Remove condition
			if condition.Timer >= condition.Duration {
				lastIndex := len(mob.Data.Conditions) - 1
				mob.Data.Conditions[conditionIndex] = mob.Data.Conditions[lastIndex]
				mob.Data.Conditions = mob.Data.Conditions[:lastIndex]
				continue
			}

			// Update condition damage over time effects
			conditionData := CONDITION_DATA[condition.Id]
			for effectIndex := range len(conditionData.Effects) {
				effect := &conditionData.Effects[effectIndex]

				if effect.Type == CONDITION_EFFECT_TYPE_DAMAGE_OVER_TIME {
					effectData := effect.Data.(ConditionDamageOverTime)
					if condition.Timer % effectData.Frequency == 0 {
						// This intetionally does not use the damage() function
						// it does not trigger NPC onAttacked (we don't know the attacker handle at this point)
						// and it does not raise mob alertness
						mob.Data.Health -= effectData.Amount
						world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> took <r>%d damage</r> from <r>%s</r>.", mob.GetName(), effectData.Amount, conditionData.Name))

						if mob.IsDead() {
							world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> has died from <r>%s</r>.", mob.GetName(), conditionData.Name))
							return
						}
					}
				}
			}
		}

		conditionIndex++
	}

	switch mob.Mode {
		case MOB_MODE_IDLE:

		case MOB_MODE_ATTACK: {
			// Check if target exists
			if mob.isTargetInvalid(world) {
				mob.SetModeIdle()
				break
			}

			// Attack with weapon
			targetMob := mob.getTargetMob(world)
			mob.attackTargetWithWeapon(world, targetMob, EQUIPMENT_SLOT_MAIN_HAND)
			mob.attackTargetWithWeapon(world, targetMob, EQUIPMENT_SLOT_OFF_HAND)

			// Update status
			mob.alertness = MOB_ALERTNESS_MAX
			mob.SetFlag(MOB_FLAG_HIDDEN, false)
		}

		case MOB_MODE_CAST: {
			// Check if target exists
			if mob.isTargetInvalid(world) {
				mob.SetModeIdle()
				break
			}

			// Check if caster has enough mana
			spellData := SPELL_DATA[mob.castSpell]
			if mob.Data.Mana < spellData.ManaCost {
				mob.SetModeIdle()
				world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> tried to cast <g>%s</g>, but they don't have enough mana.", mob.GetName(), spellData.Name))
				break
			}

			// Check spell timer
			if mob.castTimer > 0 {
				mob.castTimer--
				world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> is charging a spell...", mob.GetName()))
				break
			}

			// Cast spell
			world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> cast <g>%s</g>!", mob.GetName(), spellData.Name))
			success := mob.spellcast(world)
			if success {
				mob.Data.Mana -= spellData.ManaCost

				// Progress spell mastery
				if mob.IsPlayer() {
					mob.progressPlayerSpellMastery(world, mob.castSpell)
				}
			}

			mob.SetModeIdle()
		}

		case MOB_MODE_USE_ITEM: {
			// Check if target exists
			if mob.isTargetInvalid(world) {
				mob.SetModeIdle()
				break
			}

			// Use item
			mob.useItem(world)
			mob.Mode = MOB_MODE_IDLE
		}

		case MOB_MODE_CRAFT_ITEM: {
			//craft the item
			hadIngredients := mob.CraftItem(world, mob.craftItemRecipe)

			//halt if something was amiss
			if !hadIngredients {
				mob.SetModeIdle()
				break
			}

			//increment amount left to craftItemAmount
			mob.craftItemAmount--

			//halt if no more are required
			if mob.craftItemAmount <= 0 {
				mob.SetModeIdle()
				break
			}
		}

		case MOB_MODE_TAUNT: {
			// Check if target exists
			if mob.isTargetInvalid(world) {
				mob.SetModeIdle()
				break
			}

			targetMob := mob.getTargetMob(world)
			world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> taunted <y>%s</y>!", mob.Data.Name, targetMob.Data.Name))
			success := mob.rollForTaunt(targetMob)
			if success {
				targetMob.SetModeAttack(world, mob.Handle)
				if targetMob.IsNpc() {
					targetMob.Npc.disposition = NPC_DISPOSITION_HOSTILE
				}

				world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> grew angry and is now attacking <y>%s</y>!", targetMob.Data.Name, mob.Data.Name))
			} else {
				world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> ignored the taunt.", targetMob.Data.Name))
			}

			mob.Mode = MOB_MODE_IDLE
			mob.tauntCooldown = MOB_TAUNT_COOLDOWN_MAX
		}
	}
}

func (mob *Mob) IsNpc() bool {
	return mob.Npc != nil
}

func (mob *Mob) IsPlayer() bool {
	return mob.PlayerCharacter != nil
}

func (mob *Mob) isTargetInvalid(world *World) bool {
	if mob.Target.Type == MOB_TARGET_MOB {
		targetHandle := mob.Target.Value.(MobHandle)
		targetMob, targetExists := world.Mobs.GetIfExists(targetHandle)

		return !targetExists ||
			targetMob.IsDead() ||
			targetMob.Data.Room != mob.Data.Room ||
			targetMob.CheckFlag(MOB_FLAG_HIDDEN)
	}

	return false
}

// Returns target mob, assumes that the mob exists
func (mob *Mob) getTargetMob(world *World) *Mob {
	handle := mob.Target.Value.(MobHandle)
	return world.Mobs.Get(handle)
}

func (mob *Mob) damage(world *World, attackerHandle MobHandle, damage int32) {
	mob.Data.Health -= damage
	mob.alertness = MOB_ALERTNESS_MAX

	if mob.IsNpc() {
		mob.Npc.FireEvent(world, NpcEvent {
			Type: NPC_EVENT_TYPE_ATTACKED,
			Data: NpcEventAttacked {
				AttackerHandle: attackerHandle,
			},
		})
	}
}

func (mob *Mob) attackTargetWithWeapon(world *World, targetMob *Mob, slot EquipmentSlot) {
	// Check for weapon
	weapon := mob.Data.Equipment.Get(slot)
	var itemData *ItemData = nil
	heldItemIsWeapon := false
	if weapon != nil {
		itemData = ITEM_DATA[weapon.Id]
		heldItemIsWeapon = itemData.ItemIsWeapon()
	}

	// Don't attack with off-hand unless there is a weapon in off-hand
	if slot == EQUIPMENT_SLOT_OFF_HAND && !heldItemIsWeapon {
		return
	}

	// Check for evasion
	targetAgility := float32(targetMob.Data.Agility())
	mobAgility := float32(mob.Data.Agility())
	evasionChance := targetAgility / (targetAgility + (mobAgility * MOB_EVASION_K))
	evasionChance += targetMob.Data.GetConditionChanceModifier(evasionChance, CONDITION_EFFECT_TYPE_STAT_CHANGE_DODGE_CHANCE)
	evasionChance -= mob.Data.GetConditionChanceModifier(evasionChance, CONDITION_EFFECT_TYPE_STAT_CHANGE_HIT_CHANCE)
	evasionChance = max(min(evasionChance, float32(100.0)), float32(0.0))

	evasionRoll := rand.Float32()
	if evasionRoll < evasionChance {
		world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> dodged <y>%s's</y> attack!", targetMob.GetName(), mob.GetName()))
		return
	}

	// Get weapon damage from the item
	var damage int32 = 0
	if heldItemIsWeapon {
		damage = weapon.GetDamage()
	}

	// Add strength to the damage
	if (slot == EQUIPMENT_SLOT_MAIN_HAND) {
		damage += mob.Data.Strength() / 2
	} else {
		damage += mob.Data.Strength() / 4
	}

	// Apply condition damage changes
	damage = mob.Data.GetConditionStatChange(damage, CONDITION_EFFECT_TYPE_STAT_CHANGE_DAMAGE)

	// Check for critical hit
	critChance := mobAgility * MOB_CRIT_K
	critRoll := rand.Float32()
	crit := critRoll < critChance
	if crit {
		damage = int32(float32(damage) * MOB_CRIT_DAMAGE_MULTIPLIER)
	}

	// Subtract armor
	damage -= targetMob.Data.Armor() / 2.0

	// Calculate final damage
	attackerMinDamage := max(1, mob.Data.Level / 2)
	damage = max(damage, attackerMinDamage)

	// Deal damage
	targetMob.damage(world, mob.Handle, damage)
	// Broadcast result to room
	critStr := ""
	if crit {
		critStr = "Critical hit! "
	}
	world.messageRoom(mob.Data.Room, fmt.Sprintf("<r>%s</r><y>%s</y> struck <y>%s</y> for <r>%d damage</r>.", critStr, mob.GetName(), targetMob.GetName(), damage))
	if targetMob.IsDead() {
		world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> has <r>slain</r> <y>%s</y>.", mob.GetName(), targetMob.GetName()))
	} else {
		targetMob.rollForConcentration(world, damage)
	}

	// Handle attack infusion
	infusion := mob.getInfusionInSlot(slot)
	if infusion != nil {
		// Infusion uses should always decrement because the attacker used the weapon
		// But we shouldn't roll to apply condition if the target is already dead

		// Roll to apply condition
		if !targetMob.IsDead() {
			shouldApplyCondition := rand.Float32() < infusion.ApplyChance
			if shouldApplyCondition {
				if !targetMob.Data.HasCondition(infusion.ConditionId) {
					world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> afflicted <y>%s</y> with <r>%s</r>!",
						mob.GetName(), targetMob.GetName(), CONDITION_DATA[infusion.ConditionId].Name))
				}

				targetMob.Data.AddCondition(infusion.ConditionId, infusion.ApplyDuration)
			}
		}

		// Decrement uses remaining
		if infusion.UsesRemaining != ITEM_INFUSION_USES_UNLIMITED {
			infusion.UsesRemaining--
			if mob.IsPlayer() && infusion.UsesRemaining == 0 {
				world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Your <g>%s's</g> infusion wore off.", itemData.Name))
			}
		}
	}

	// Reduce weapon durability
	mob.subtractDurabilityFromEquipment(world, slot)
	// and from defender's weapon or armor
	targetMob.subtractDurabilityFromEquipment(world, EQUIPMENT_SLOT_OFF_HAND)
	targetMob.subtractDurabilityFromEquipment(world, EQUIPMENT_SLOT_OUTFIT)
}

func (mob *Mob) getInfusionInSlot(slot EquipmentSlot) *ItemInfusion {
	weapon := mob.Data.Equipment.Get(slot)
	if weapon != nil && weapon.HasInfusion() {
		return &weapon.Infusion
	}

	if mob.IsNpc() {
		npcData := NPC_DATA[mob.Npc.Id]
		if npcData.infusion.UsesRemaining != 0 {
			return &npcData.infusion
		}
	}

	return nil
}

func (mob *Mob) rollForConcentration(world *World, damage int32) {
	// Not concentrating
	if mob.Mode != MOB_MODE_CAST {
		return
	}

	// Don't break concentration for instants or spells that have already been charged
	if mob.castTimer == 0 {
		return
	}

	mobFaith := float32(mob.Data.Faith())
	attackDamage := float32(damage)

	concentrationChance := mobFaith / (mobFaith + (attackDamage / 2))
	concentrationRoll := rand.Float32()
	if concentrationRoll < concentrationChance {
		// Concentration maintained!
		return
	}

	// Concentration broken!
	mob.Mode = MOB_MODE_IDLE
	world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> lost concentration on their spell!", mob.GetName()))
}

func (mob *Mob) RollForEscape(world *World) bool {
	// All the escape chances are multiplied together
	// This is mathematically the same as doing individual escape rolls
	// for each monster
	var escapeChance float32 = mob.escapeChance

	mobAgility := float32(mob.Data.Agility())

	room := &world.Rooms[mob.Data.Room]
	for _, handle := range room.Occupants {
		occupant := world.Mobs.Get(handle)

		if !occupant.IsNpc() || occupant.Npc.disposition != NPC_DISPOSITION_HOSTILE {
			continue
		}

		occupantAgility := float32(occupant.Data.Agility())
		escapeChance *= mobAgility / (mobAgility + occupantAgility)
	}

	// Roll to escape
	escaped := rand.Float32() < escapeChance

	// On failed roll, drop escape chance to 0.0
	if !escaped {
		mob.escapeChance = 0.0
	}

	return escaped
}

func (mob *Mob) RollForStealth(world *World) bool {
	var enemyIntelligence float32 = 0.0
	var enemyAlertness float32 = 0.0

	// Determine the highest intelligence of all hostile mobs in the room
	room := &world.Rooms[mob.Data.Room]
	for _, handle := range room.Occupants {
		occupant := world.Mobs.Get(handle)
		if !occupant.IsNpc() || occupant.Npc.disposition != NPC_DISPOSITION_HOSTILE {
			continue
		}

		enemyIntelligence = max(enemyIntelligence, float32(occupant.Data.Intelligence()))
		enemyAlertness = max(enemyAlertness, occupant.alertness)
	}

	// TODO: Room Brightness modifier

	armorStealthPenalty := mob.GetStealthPenalty()

	// Roll to hide
	mobAgility := float32(mob.Data.Agility())
	stealthChance := mobAgility / (mobAgility + (MOB_STEALTH_K * enemyIntelligence))
	stealthChance *= mob.escapeChance
	stealthChance *= (1.0 - enemyAlertness)
	stealthChance *= (1.0 - armorStealthPenalty)
	hidden := rand.Float32() < stealthChance

	// On failed roll, drop escape chance to 0.0
	if !hidden {
		mob.escapeChance = 0.0
	}

	return hidden
}

func (mob *Mob) rollForTaunt(targetMob *Mob) bool {
	mobStrength := float32(mob.Data.Strength())
	targetIntelligence := float32(targetMob.Data.Intelligence())
	tauntChance := mobStrength / targetIntelligence
	tauntChance *= (1.0 - mob.tauntCooldown)

	if targetMob.IsNpc() && targetMob.Npc.disposition == NPC_DISPOSITION_FRIENDLY {
		tauntChance = 0.0
	}

	return rand.Float32() < tauntChance
}

func (mob *Mob) subtractDurabilityFromEquipment(world *World, slot EquipmentSlot) {
	item := mob.Data.Equipment.Get(slot)
	if item == nil {
		return
	}

	itemData := ITEM_DATA[item.Id]

	item.Durability--

	if item.Durability <= 0 {
		world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s's</y> <g>%s</g> broke!", mob.GetName(), itemData.Name))
		_, messages, _ := mob.unequip(slot)
		for _, message := range messages {
			world.messagePlayer(mob.PlayerCharacter.PlayerId, message)
		}

		return
	}

	// The rest of these messages are only sent to players holding the item
	if !mob.IsPlayer() {
		return
	}

	// If item has become damaged, tell the user
	maxDurability := itemData.GetMaxDurability()
	itemWasDamaged := (item.Durability + 1) < maxDurability / 2
	itemIsDamaged := item.Durability < maxDurability / 2
	if itemIsDamaged && !itemWasDamaged {
		world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Your <g>%s</g> is now damaged.", itemData.Name))
		return
	}

	// If an item has lost its sharpness, tell the user
	itemWasSharp := (item.Durability + 1) > maxDurability
	itemIsSharp := item.Durability > maxDurability
	if itemWasSharp && !itemIsSharp {
		itemIsWeapon :=
			itemData.Kind == ITEM_KIND_EQUIPMENT_ONE_HANDED ||
			itemData.Kind == ITEM_KIND_EQUIPMENT_TWO_HANDED
		if itemIsWeapon {
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Your <g>%s</g> has lost its sharpness.", itemData.Name))
		} else {
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Your <g>%s</g> has lost its fortification.", itemData.Name)) }
	}
}

func (mob *Mob) ValidateArgument(world *World, params MobValidateArgumentParams) error {
	userLuaHandle, err := world.getMobLuaHandle(mob.Handle)
	if err != nil {
		log.Printf("Warn - Failed to get lua handle during ValidateArgument: %s", err.Error())
		return fmt.Errorf("Failed to %s %s because of a technical issue.", params.Verb, params.Noun)
	}

	result, ok := world.scriptCall(params.Hook.Value(), fmt.Sprintf("%s %s validate_argument()", params.Noun, params.Name), userLuaHandle, lua.String(params.Argument))
	if !ok {
		return fmt.Errorf("Failed to %s %s because of a technical issue.", params.Verb, params.Noun)
	}

	errorString, hasError := result.AsString()
	if hasError {
		return fmt.Errorf("%s", errorString)
	}

	return nil
}

// Calls the on_use() or on_cast() of an item or spell
func (mob *Mob) onUseOrCast(world *World, params MobOnUseOrCastParams) bool {
	// Double check argument validation before use/cast
	if params.targetType == TARGET_TYPE_ARGUMENT {
		argument := mob.Target.Value.(string)
		err := mob.ValidateArgument(world, MobValidateArgumentParams {
			Hook: params.validateArgument,
			Argument: argument,
			Verb: params.verb,
			Name: params.name,
			Noun: params.noun,
		})
		if err != nil {
			world.messageRoom(mob.Data.Room, "But it failed!")
			return false
		}
	}

	// Build hook arguments array
	args := make([]lua.Value, 0, 1)

	// User handle
	mobLuaHandle, err := world.getMobLuaHandle(mob.Handle)
	if err != nil {
		log.Printf("Failed to get mob lua handle during onUseOrCast: %s", err.Error())
		world.messageRoom(mob.Data.Room, "But it failed!")
		return false
	}
	args = append(args, mobLuaHandle)

	// Target args
	switch params.targetType {
		case TARGET_TYPE_NONE: {}

		case TARGET_TYPE_ALLY, TARGET_TYPE_ENEMY: {
			targetMob := mob.getTargetMob(world)
			targetLuaHandle, err := world.getMobLuaHandle(targetMob.Handle)
			if err != nil {
				log.Printf("Failed to get target lua handle during onUseOrCast: %s", err.Error())
				world.messageRoom(mob.Data.Room, "But it failed!")
				return false
			}

			args = append(args, targetLuaHandle)
		}

		case TARGET_TYPE_INVENTORY_ITEM: {
			itemTarget := mob.Target.Value.(MobItemTarget)
			index, found := mob.Data.Inventory.FindExpectedItem(itemTarget.Id, itemTarget.Index)
			if !found {
				world.messageRoom(mob.Data.Room, "But it failed!")
				return false
			}

			args = append(args, lua.Number(float64(index)))
		}

		case TARGET_TYPE_ARGUMENT: {
			argument := mob.Target.Value.(string)
			args = append(args, lua.String(argument))
		}

		default: {
			log.Printf("Target type %s not handled on onUseOrCast.", params.targetType.String())
			world.messageRoom(mob.Data.Room, "But it failed!")
			return false
		}
	}

	// Invoke hook
	_, ok := world.scriptCall(params.hook.Value(), fmt.Sprintf("%s '%s' on_%s()", params.noun, params.name, params.verb), args...)
	if !ok {
		world.messageRoom(mob.Data.Room, "But it failed!")
		return false
	}

	return true
}

func (mob *Mob) spellcast(world *World) bool {
	spellData := SPELL_DATA[mob.castSpell]

	// Update mob status
	mob.alertness = MOB_ALERTNESS_MAX
	mob.SetFlag(MOB_FLAG_HIDDEN, false)

	return mob.onUseOrCast(world, MobOnUseOrCastParams {
		targetType: spellData.TargetType,
		hook: spellData.OnCast,
		validateArgument: spellData.ValidateArgument,
		noun: "spell",
		verb: "cast",
		name: spellData.Name,
	})
}

func (mob *Mob) calculateMagicDamage(baseDamage int32, target *Mob) int32 {
	return baseDamage + (mob.Data.Faith() / 2) + (target.Data.Faith() / 4)
}

func (mob *Mob) calculateMagicConditionDuration(baseDuration int) int {
	percent := 1.0 + (float32(mob.Data.Faith()) * MOB_MAGIC_CONDITION_DURATION_K)
	return int(float32(baseDuration) * percent)
}

func (mob *Mob) progressPlayerSpellMastery(world *World, spell SpellId) {
	spellData := SPELL_DATA[spell]
	equippedSpell, spellIsEquipped := mob.PlayerCharacter.SpellsEquipped[mob.castSpell]
	if spellIsEquipped && !equippedSpell.IsKnown {

		// Spell mastery progress
		equippedSpell.Casts++
		if equippedSpell.Casts >= mob.Data.CastsToLearn(mob.castSpell) {
			equippedSpell.IsKnown = true
			mob.PlayerCharacter.SpellsKnown = append(mob.PlayerCharacter.SpellsKnown, mob.castSpell)
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("You have mastered <g>%s</g>!", spellData.Name))

			// If player is an enchanter, grant recipe to their known recipes
			if mob.PlayerCharacter.Job == JOB_ENCHANTER {
				spellName := spellData.Name
				spellbookName := fmt.Sprintf("Spellbook of %s", spellName)
				spellScrollName := fmt.Sprintf("Scroll of %s", spellName)

				newRecipeNames := make([]string, 0, 2)

				spellbookRecipe, exists := RECIPE_NAME_TO_ID[spellbookName]
				if exists {
					mob.PlayerCharacter.RecipesKnown = append(mob.PlayerCharacter.RecipesKnown, spellbookRecipe)
					newRecipeNames = append(newRecipeNames, fmt.Sprintf("<g>%s</g>", spellbookName))
				}

				spellScrollRecipe, exists := RECIPE_NAME_TO_ID[spellScrollName]
				if exists {
					mob.PlayerCharacter.RecipesKnown = append(mob.PlayerCharacter.RecipesKnown, spellScrollRecipe)
					newRecipeNames = append(newRecipeNames, fmt.Sprintf("<g>%s</g>", spellScrollName))
				}

				if len(newRecipeNames) != 0 {
					world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("You can now craft %s.", util.CombineNames(newRecipeNames)))
				}
			}
		}

		// Spellbook durability
		slot, _ := mob.getEquipmentWhichProvidesSpell(mob.castSpell)
		mob.subtractDurabilityFromEquipment(world, slot)
	}
}

func (mob *Mob) getEquipmentWhichProvidesSpell(spell SpellId) (EquipmentSlot, bool) {
	slots := []EquipmentSlot { EQUIPMENT_SLOT_MAIN_HAND, EQUIPMENT_SLOT_OFF_HAND }
	for _, slot := range slots {
		slotSpell, slotProvidesSpell := mob.getSpellProvidedBySlot(slot)
		if slotProvidesSpell && slotSpell == spell {
			return slot, true
		}
	}

	return 0, false
}

func (mob *Mob) getSpellProvidedBySlot(slot EquipmentSlot) (SpellId, bool) {
	item := mob.Data.Equipment.Get(slot)
	if item == nil {
		return 0, false
	}

	itemData := ITEM_DATA[item.Id]
	if itemData.Kind != ITEM_KIND_EQUIPMENT_SPELLBOOK {
		return 0, false
	}

	spellbookData := itemData.Data.(*ItemDataSpellbook)
	return spellbookData.Spell, true
}

func (mob *Mob) useItem(world *World) {
	// Find item in mob inventory
	itemIndex, itemFound := mob.Data.Inventory.FindExpectedItem(mob.itemToUse.Id, mob.itemToUse.Index)
	if !itemFound {
		mob.Mode = MOB_MODE_IDLE
		return
	}

	// Use item
	item := &mob.Data.Inventory.Items[itemIndex]
	itemData := ITEM_DATA[item.Id]
	world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> used <g>%s</g>!", mob.GetName(), itemData.Name))

	success := false
	switch itemData.Kind {
		case ITEM_KIND_CONSUMABLE: {
			consumableData := itemData.Data.(*ItemDataConsumable)

			success = mob.onUseOrCast(world, MobOnUseOrCastParams {
				targetType: consumableData.TargetType,
				hook: consumableData.onUse,
				validateArgument: consumableData.ValidateArgument,
				noun: "item",
				verb: "use",
				name: itemData.Name,
			})
		}
		case ITEM_KIND_SPELL_SCROLL: {
			scrollData := itemData.Data.(*ItemDataSpellScroll)

			mob.castSpell = scrollData.Spell
			success = mob.spellcast(world)
		}
		default:
			panic(fmt.Sprintf("Unhandled item type %s. This item type should never have been allowed to be used here.", itemData.Kind.String()))
	}

	// If used successfully, remove the item from their inventory
	if success {
		// Re-check the itemIndex in case the hook changed their inventory in some way
		itemIndex, itemFound = mob.Data.Inventory.FindExpectedItem(mob.itemToUse.Id, mob.itemToUse.Index)
		if itemFound {
			mob.Data.Inventory.RemoveItem(itemIndex)
		}
	}
}

// Returns an error if the mob is unable to craft batchAmount of the recipe
func (mob *Mob) CanCraft(recipe RecipeId, batchAmount int32) error {
	recipeData := RECIPE_DATA[recipe]

	// Check for the materials
	for _, ingredient := range recipeData.Materials {
		amountOfIngredient := mob.Data.Inventory.AmountOf(ingredient.Id)
		if amountOfIngredient < batchAmount * ingredient.Amount {
			return fmt.Errorf("You lack the ingredients to craft %s.", recipeData.Name)
		}
	}

	// Check for inventory space
	if !mob.Data.Inventory.HasSpaceFor(recipeData.NetItemSize(batchAmount), mob.InventoryCapacity()) {
		return fmt.Errorf("You don't have enough space in your inventory to craft %s.", recipeData.Name)
	}

	return nil
}

func (mob *Mob) InventoryCapacity() int32 {
	if !mob.IsPlayer() {
		return INVENTORY_CAPACITY_UNLIMITED
	}

	return INVENTORY_CAPACITY_PLAYER
}

func (mob *Mob) CraftItem(world *World, recipe RecipeId) bool {
	recipeData := RECIPE_DATA[recipe]

	err := mob.CanCraft(recipe, 1)
	if err != nil {
		if mob.PlayerCharacter != nil {
			world.messagePlayer(mob.PlayerCharacter.PlayerId, err.Error())
		}
		return false
	}

	// Remove the materials from the player's inventory
	for _, ingredient := range recipeData.Materials {
		amountToRemove := ingredient.Amount
		for amountToRemove > 0 {
			index, _ := mob.Data.Inventory.FindItem(ingredient.Id)
			item := mob.Data.Inventory.RemoveItems(index, amountToRemove)
			amountToRemove -= item.Amount
		}
	}

	// Add the crafted item to the player's inventory
	recipeOutput := recipeData.CreateOutput()
	mob.Data.Inventory.AddItem(recipeOutput)
	if mob.IsPlayer() {
		world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("You crafted <g>%s</g>.", recipeOutput.GetNameWithAmount()))
	}
	return true
}

func (mob *Mob) GetStealthPenalty() float32 {
		var armorStealthPenalty float32 = 0.0
		outfit := mob.Data.Equipment.Get(EQUIPMENT_SLOT_OUTFIT)
		if outfit != nil {
			outfitData := ITEM_DATA[outfit.Id].Data.(*ItemDataArmor)
			armorStealthPenalty += outfitData.StealthPenalty
		}

		shield := mob.Data.Equipment.GetShield()
		if shield != nil {
			shieldData := ITEM_DATA[shield.Id].Data.(*ItemDataArmor)
			armorStealthPenalty += shieldData.StealthPenalty
		}

		return armorStealthPenalty
}

func (mob *Mob) Reveal(world *World) {
	mob.SetFlag(MOB_FLAG_HIDDEN, false)
	world.messageRoom(mob.Data.Room, fmt.Sprintf("<y>%s</y> was hiding in the room and has revealed themselves!", mob.Data.Name))
}
