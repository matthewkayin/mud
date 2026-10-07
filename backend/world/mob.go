package world

import (
	"fmt"
	"log"
	"math/rand/v2"
	"slices"
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

const MOB_ALERTNESS_MAX float32 = 1.0
const MOB_ALERTNESS_COOLDOWN_RATE float32 = 0.05

const MOB_TAUNT_COOLDOWN_MAX float32 = 1.0
const MOB_TAUNT_COOLDOWN_RATE float32 = 0.05

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
	Target MobHandle

	flags MobFlag

	castSpell Spell
	castTimer int32
	useItemId ItemId

	craftItemRecipe Recipe
	craftItemAmount int32

	escapeChance float32
	alertness float32
	tauntCooldown float32
	fuzzyNumber int
}

func MobInit(data *MobData) Mob {
	mob := Mob {
		PlayerCharacter: nil,
		Data: *data,
		Mode: MOB_MODE_IDLE,

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
	if mob.PlayerCharacter == nil {
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
		return occupant.Npc != nil &&
			occupant.Npc.mode != NPC_MODE_SLEEP &&
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
	if mob.PlayerCharacter == nil {
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
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Level up! %s is now level %d.", mob.Data.Name, mob.Data.Level))

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
			SaveCharacter(mob.PlayerCharacter)

			continue
		}

		mob.Data.Experience += experience
		experience = 0
	}
}

func (mob *Mob) SetModeIdle() {
	mob.Mode = MOB_MODE_IDLE
}

func (mob *Mob) SetModeAttack(world *World, mobHandle MobHandle, targetHandle MobHandle) {
	mob.Mode = MOB_MODE_ATTACK
	mob.Target = targetHandle
}

func (mob *Mob) SetModeCast(world *World, mobHandle MobHandle, spell Spell, targetHandle MobHandle) {
	mob.Mode = MOB_MODE_CAST
	mob.Target = targetHandle
	mob.castSpell = spell
	mob.castTimer = SPELL_DATA[spell].CastTime
}

func (mob *Mob) SetModeUseItem(world *World, mobHandle MobHandle, itemId ItemId, targetHandle MobHandle) {
	mob.Mode = MOB_MODE_USE_ITEM
	mob.Target = targetHandle
	mob.useItemId = itemId
}

func (mob *Mob) SetModeCraftItem(world *World, mobHandle MobHandle, recipe Recipe, amount int32) {
	mob.Mode = MOB_MODE_CRAFT_ITEM
	mob.craftItemRecipe = recipe
	mob.craftItemAmount = amount
}

func (mob *Mob) SetModeTaunt(world *World, mobHandle MobHandle, targetHandle MobHandle) {
	mob.Mode = MOB_MODE_TAUNT
	mob.Target = targetHandle
}

func (mob *Mob) Update(world *World) {
	if mob.IsDead() {
		return
	}

	mob.escapeChance = min(mob.escapeChance + MOB_ESCAPE_CHANCE_RECOVERY_RATE, MOB_ESCAPE_CHANCE_MAX)
	mob.alertness = max(mob.alertness - MOB_ALERTNESS_COOLDOWN_RATE, 0.0)
	mob.tauntCooldown = max(mob.tauntCooldown - MOB_TAUNT_COOLDOWN_RATE, 0.0)

	switch mob.Mode {
		case MOB_MODE_IDLE:

		case MOB_MODE_ATTACK: {
			// Check if target exists
			targetMob, targetExists := mob.getTargetIfExists(world)
			if !targetExists {
				break
			}

			// Attack with weapon
			mob.attackTargetWithWeapon(world, targetMob, EQUIPMENT_SLOT_MAIN_HAND)
			mob.attackTargetWithWeapon(world, targetMob, EQUIPMENT_SLOT_OFF_HAND)

			// Update status
			mob.alertness = MOB_ALERTNESS_MAX
			mob.SetFlag(MOB_FLAG_HIDDEN, false)
		}

		case MOB_MODE_CAST: {
			// Check if target exists
			targetMob, targetExists := mob.getTargetIfExists(world)
			if !targetExists {
				break
			}

			// Check if caster has enough mana
			spellData := SPELL_DATA[mob.castSpell]
			if mob.Data.Mana < spellData.ManaCost {
				mob.Mode = MOB_MODE_IDLE
				world.messageRoom(mob.Data.Room, fmt.Sprintf("%s tried to cast %s, but they don't have enough mana.", mob.GetName(), spellData.Name))
				break
			}

			// Check spell timer
			if mob.castTimer > 0 {
				mob.castTimer--
				world.messageRoom(mob.Data.Room, fmt.Sprintf("%s is charging a spell...", mob.GetName()))
				break
			}

			// Cast spell
			mob.spellcast(world, targetMob)
			mob.Mode = MOB_MODE_IDLE

			// Update status
			mob.alertness = MOB_ALERTNESS_MAX
			mob.SetFlag(MOB_FLAG_HIDDEN, false)
		}

		case MOB_MODE_USE_ITEM: {
			// Check if target exists
			targetMob, targetExists := mob.getTargetIfExists(world)
			if !targetExists {
				break
			}

			// Use item
			mob.useItem(world, targetMob)
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
			targetMob, targetExists := mob.getTargetIfExists(world)
			if !targetExists {
				break
			}

			world.messageRoom(mob.Data.Room, fmt.Sprintf("%s taunted %s!", mob.Data.Name, targetMob.Data.Name))
			success := mob.rollForTaunt(targetMob)
			if success {
				targetMob.SetModeAttack(world, mob.Target, mob.Handle)
				if targetMob.Npc != nil {
					targetMob.Npc.disposition = NPC_DISPOSITION_HOSTILE
				}

				world.messageRoom(mob.Data.Room, fmt.Sprintf("%s grew angry and is now attacking %s!", targetMob.Data.Name, mob.Data.Name))
			} else {
				world.messageRoom(mob.Data.Room, fmt.Sprintf("%s ignored the taunt.", targetMob.Data.Name))
			}

			mob.Mode = MOB_MODE_IDLE
			mob.tauntCooldown = MOB_TAUNT_COOLDOWN_MAX
		}
	}
}

func (mob *Mob) getTargetIfExists(world *World) (*Mob, bool) {
	targetMob, targetExists := world.Mobs.GetIfExists(mob.Target)
	targetIsInvalid :=
		!targetExists ||
		targetMob.IsDead() ||
		targetMob.Data.Room != mob.Data.Room ||
		targetMob.CheckFlag(MOB_FLAG_HIDDEN)

	if targetIsInvalid {
		mob.Mode = MOB_MODE_IDLE
		return nil, false
	}

	return targetMob, true
}

func (mob *Mob) damage(world *World, attackerHandle MobHandle, damage int32) {
	mob.Data.Health -= damage
	mob.alertness = MOB_ALERTNESS_MAX

	if mob.Npc != nil {
		mob.Npc.OnEvent(world, BehaviorEvent {
			Type: BEHAVIOR_EVENT_TYPE_ATTACKED,
			Data: BehaviorEventAttacked {
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
		heldItemIsWeapon =
			itemData.ItemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED ||
			itemData.ItemType == ITEM_TYPE_EQUIPMENT_TWO_HANDED
	}

	// Don't attack with off-hand unless there is a weapon in off-hand
	if slot == EQUIPMENT_SLOT_OFF_HAND && !heldItemIsWeapon {
		return
	}

	// Check for evasion
	targetAgility := float32(targetMob.Data.Agility())
	mobAgility := float32(mob.Data.Agility())
	evasionChance := targetAgility / (targetAgility + (mobAgility * MOB_EVASION_K))
	evasionRoll := rand.Float32()
	if evasionRoll < evasionChance {
		world.messageRoom(mob.Data.Room, fmt.Sprintf("%s dodged %s's attack!", targetMob.GetName(), mob.GetName()))
		return
	}

	// Get weapon damage from the item
	var damage int32 = 0
	if heldItemIsWeapon {
		weaponData := itemData.Data.(*ItemDataWeapon)
		damage = weaponData.Damage
	}

	// Add strength to the damage
	if (slot == EQUIPMENT_SLOT_MAIN_HAND) {
		damage += mob.Data.Strength() / 2
	} else {
		damage += mob.Data.Strength() / 4
	}

	// Check for critical hit
	critChance := mobAgility * MOB_CRIT_K
	critRoll := rand.Float32()
	crit := critRoll < critChance
	if crit {
		damage = int32(float32(damage) * MOB_CRIT_DAMAGE_MULTIPLIER)
	}

	// Subtract armor
	damage -= mob.Data.Armor() / 2.0

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
	world.messageRoom(mob.Data.Room, fmt.Sprintf("%s%s struck %s for %d damage.", critStr, mob.GetName(), targetMob.GetName(), damage))
	if targetMob.IsDead() {
		world.messageRoom(mob.Data.Room, fmt.Sprintf("%s has slain %s.", mob.GetName(), targetMob.GetName()))
	} else {
		targetMob.rollForConcentration(world, damage)
	}

	// Reduce weapon durability
	mob.subtractDurabilityFromEquipment(world, slot)
	if !targetMob.IsDead() {
		targetMob.subtractDurabilityFromEquipment(world, EQUIPMENT_SLOT_OUTFIT)
		// TODO: shield durability gets subtracted here as well
	}
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
	world.messageRoom(mob.Data.Room, fmt.Sprintf("%s lost concentration on their spell!", mob.GetName()))
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
		// TODO: change to !occupant.isNpc()
		if occupant.Npc == nil || occupant.Npc.disposition != NPC_DISPOSITION_HOSTILE {
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
		if occupant.Npc == nil || occupant.Npc.disposition != NPC_DISPOSITION_HOSTILE {
			continue
		}

		enemyIntelligence = max(enemyIntelligence, float32(occupant.Data.Intelligence()))
		enemyAlertness = max(enemyAlertness, occupant.alertness)
	}

	// TODO: Room Brightness modifier

	// Determine armor stealth penalty
	var armorStealthPenalty float32 = 0.0
	outfit := mob.Data.Equipment.Get(EQUIPMENT_SLOT_OUTFIT)
	if outfit != nil {
		outfitData := ITEM_DATA[outfit.Id].Data.(*ItemDataOutfit)
		armorStealthPenalty = outfitData.StealthPenality
	}

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

	if targetMob.Npc != nil && targetMob.Npc.disposition == NPC_DISPOSITION_FRIENDLY {
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
	if item.Durability == 0 {
		world.messageRoom(mob.Data.Room, fmt.Sprintf("%s's %s broke!", mob.GetName(), itemData.Name))
		_, messages, _ := mob.unequip(slot)
		for _, message := range messages {
			world.messagePlayer(mob.PlayerCharacter.PlayerId, message)
		}

		return
	}

	// The rest of these messages are only sent to players holding the item
	if mob.PlayerCharacter == nil {
		return
	}

	// If item has become damaged, tell the user
	maxDurability := itemData.GetMaxDurability()
	itemWasDamaged := (item.Durability + 1) < maxDurability / 2
	itemIsDamaged := item.Durability < maxDurability / 2
	if itemIsDamaged && !itemWasDamaged {
		world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Your %s is now damaged.", itemData.Name))
		return
	}

	// If an item has lost its sharpness, tell the user
	itemWasSharp := (item.Durability + 1) > maxDurability
	itemIsSharp := item.Durability > maxDurability
	if itemWasSharp && !itemIsSharp {
		itemIsWeapon :=
			itemData.ItemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED ||
			itemData.ItemType == ITEM_TYPE_EQUIPMENT_TWO_HANDED
		if itemIsWeapon {
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Your %s has lost its sharpness.", itemData.Name))
		} else {
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Your %s has lost its fortification.", itemData.Name)) }
	}
}

func (mob *Mob) spellcast(world *World, targetMob *Mob) {
	spellData := SPELL_DATA[mob.castSpell]

	// Cast spell
	world.messageRoom(mob.Data.Room, fmt.Sprintf("%s cast %s!", mob.GetName(), spellData.Name))
	mob.Data.Mana -= spellData.ManaCost

	// TODO: wrap a context around this to timeout calls?
	// spellData.onHit(world, mob, targetMob)
	_, err := world.luaState.Call(spellData.OnHit.Value(), lua.Nil(), lua.Nil())
	if err != nil {
		log.Printf("Warn - Error during spell OnHit: %s", err.Error())
	}

	if mob.PlayerCharacter != nil {
		equippedSpell, spellIsEquipped := mob.PlayerCharacter.SpellsEquipped[mob.castSpell]
		if spellIsEquipped && !equippedSpell.IsKnown {

			// Spell mastery progress
			equippedSpell.Casts++
			if equippedSpell.Casts >= mob.Data.CastsToLearn(mob.castSpell) {
				equippedSpell.IsKnown = true
				mob.PlayerCharacter.SpellsKnown = append(mob.PlayerCharacter.SpellsKnown, mob.castSpell)
				world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("You have mastered %s!", spellData.Name))
			}

			// Spellbook durability
			slot, _ := mob.getEquipmentWhichProvidesSpell(mob.castSpell)
			mob.subtractDurabilityFromEquipment(world, slot)
		}
	}
}

func (mob *Mob) calculateMagicDamage(baseDamage int32, target *Mob) int32 {
	return baseDamage + (mob.Data.Faith() / 2) + (target.Data.Faith() / 4)
}

func (mob *Mob) getEquipmentWhichProvidesSpell(spell Spell) (EquipmentSlot, bool) {
	slots := []EquipmentSlot { EQUIPMENT_SLOT_MAIN_HAND, EQUIPMENT_SLOT_OFF_HAND }
	for _, slot := range slots {
		slotSpell, slotProvidesSpell := mob.getSpellProvidedBySlot(slot)
		if slotProvidesSpell && slotSpell == spell {
			return slot, true
		}
	}

	return 0, false
}

func (mob *Mob) getSpellProvidedBySlot(slot EquipmentSlot) (Spell, bool) {
	item := mob.Data.Equipment.Get(slot)
	if item == nil {
		return 0, false
	}

	itemData := ITEM_DATA[item.Id]
	if itemData.ItemType != ITEM_TYPE_EQUIPMENT_SPELLBOOK {
		return 0, false
	}

	spellbookData := itemData.Data.(*ItemDataSpellbook)
	return spellbookData.Spell, true
}

func (mob *Mob) useItem(world *World, targetMob *Mob) {
	// Find item in mob inventory
	var itemIndex int = -1
	for index := range len(mob.Data.Inventory.Items) {
		if mob.Data.Inventory.Items[index].Id == mob.useItemId {
			itemIndex = index
			break
		}
	}

	// Check if the item still exists
	// This is a legit edge case - player might drop the item from their inventory before their turn happens
	if itemIndex == -1 {
		mob.Mode = MOB_MODE_IDLE
		return
	}

	// Remove item from their inventory
	item := mob.Data.Inventory.RemoveItem(itemIndex)

	// Use item
	itemData := ITEM_DATA[item.Id]
	world.messageRoom(mob.Data.Room, fmt.Sprintf("%s used %s!", mob.GetName(), itemData.Name))

	switch itemData.ItemType {
		case ITEM_TYPE_CONSUMABLE: {
			consumableData := itemData.Data.(*ItemDataConsumable)
			consumableData.onUse(world, targetMob)
		}
		case ITEM_TYPE_SPELL_SCROLL: {
			scrollData := itemData.Data.(*ItemDataSpellScroll)
			spellData := SPELL_DATA[scrollData.Spell]
			spellData.onHit(world, mob, targetMob)
		}
		default:
			panic(fmt.Sprintf("Unhandled item type %s. This item type should never have been allowed to be used here.", ItemTypeToString(itemData.ItemType)))
	}
}

// Returns an error if the mob is unable to craft batchAmount of the recipe
func (mob *Mob) CanCraft(recipe Recipe, batchAmount int32) error {
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
	if mob.PlayerCharacter == nil {
		return INVENTORY_CAPACITY_UNLIMITED
	}

	return INVENTORY_CAPACITY_PLAYER
}

func (mob *Mob) CraftItem(world *World, recipe Recipe) bool {
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
	if mob.PlayerCharacter != nil {
		world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("You crafted %s.", recipeOutput.GetNameWithAmount()))
	}
	return true
}
