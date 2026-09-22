package world

import (
	"fmt"
	"log"
	"slices"
	"math/rand/v2"
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

type MobMode int32
const (
	MOB_MODE_IDLE MobMode = iota
	MOB_MODE_ATTACK
	MOB_MODE_CAST
	MOB_MODE_USE_ITEM
)

type Mob struct {
	PlayerCharacter *Character
	Data MobData

	mode MobMode
	target MobHandle

	castSpell Spell
	castTimer int32
	useItemId ItemId
}

func MobInit(data *MobData) Mob {
	mob := Mob {
		PlayerCharacter: nil,
		Data: *data,
		mode: MOB_MODE_IDLE,
	}

	mob.Data.Equipment.CalculateStatBonuses()
	return mob
}

func MobInitFromCharacter(character *Character) Mob {
	mob := MobInit(&character.Data)
	mob.PlayerCharacter = character

	return mob
}

func (mob *Mob) IsDead() bool {
	return mob.Data.Health <= 0
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

			// Recalculate stats
			mob.PlayerCharacter.recalculateStats()
			mob.Data.Stats = mob.PlayerCharacter.Data.Stats

			// Announce level up message
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Level up! %s is now level %d.", mob.Data.Name, mob.Data.Level))
			continue
		}

		mob.Data.Experience += experience
		experience = 0
	}
}

func (mob *Mob) SetModeAttack(world *World, mobHandle MobHandle, targetHandle MobHandle) {
	mob.mode = MOB_MODE_ATTACK
	mob.target = targetHandle

	world.pushEvent(Event {
		EventType: EVENT_TYPE_MOB_SET_TARGET,
		Data: EventMobSetTarget {
			Attacker: mobHandle,
			Defender: targetHandle,
		},
	})
}

func (mob *Mob) SetModeCast(world *World, mobHandle MobHandle, spell Spell, targetHandle MobHandle) {
	mob.mode = MOB_MODE_CAST
	mob.target = targetHandle
	mob.castSpell = spell
	mob.castTimer = SPELL_DATA[spell].castTime

	world.pushEvent(Event {
		EventType: EVENT_TYPE_MOB_SET_TARGET,
		Data: EventMobSetTarget {
			Attacker: mobHandle,
			Defender: targetHandle,
		},
	})
}

func (mob *Mob) SetModeUseItem(world *World, mobHandle MobHandle, itemId ItemId, targetHandle MobHandle) {
	mob.mode = MOB_MODE_USE_ITEM
	mob.target = targetHandle
	mob.useItemId = itemId

	world.pushEvent(Event {
		EventType: EVENT_TYPE_MOB_SET_TARGET,
		Data: EventMobSetTarget {
			Attacker: mobHandle,
			Defender: targetHandle,
		},
	})
}

func (mob *Mob) Update(world *World) {
	switch mob.mode {
		case MOB_MODE_IDLE:
		case MOB_MODE_ATTACK:
			// Check if target exists
			targetMob, targetExists := mob.getTargetIfExists(world)
			if !targetExists {
				break
			}

			// Attack with weapon
			room := &world.Rooms[mob.Data.Room]
			mob.attackTargetWithWeapon(world, room, targetMob, EQUIPMENT_SLOT_MAIN_HAND)
			mob.attackTargetWithWeapon(world, room, targetMob, EQUIPMENT_SLOT_OFF_HAND)
		case MOB_MODE_CAST:
			// Check if target exists
			targetMob, targetExists := mob.getTargetIfExists(world)
			if !targetExists {
				break
			}

			// Cast spell
			mob.spellcast(world, targetMob)
			mob.mode = MOB_MODE_IDLE
		case MOB_MODE_USE_ITEM:
			// Check if target exists
			targetMob, targetExists := mob.getTargetIfExists(world)
			if !targetExists {
				break
			}

			// Use item
			mob.useItem(world, targetMob)
			mob.mode = MOB_MODE_IDLE
	}
}

func (mob *Mob) getTargetIfExists(world *World) (*Mob, bool) {
	targetMob, targetExists := world.Mobs.GetIfExists(mob.target)
	if !targetExists || targetMob.IsDead() || targetMob.Data.Room != mob.Data.Room {
		mob.mode = MOB_MODE_IDLE
		return nil, false
	}

	return targetMob, true
}

func (mob *Mob) attackTargetWithWeapon(world *World, room *Room, targetMob *Mob, slot EquipmentSlot) {
	// Check for weapon
	weapon := mob.Data.Equipment.Get(slot)
	var itemData *ItemData = nil
	heldItemIsWeapon := false
	if weapon != nil {
		itemData = ITEM_DATA[weapon.Id]
		heldItemIsWeapon =
			itemData.itemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED ||
			itemData.itemType == ITEM_TYPE_EQUIPMENT_TWO_HANDED
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
		world.messageRoom(mob.Data.Room, fmt.Sprintf("%s dodged %s's attack!", targetMob.Data.Name, mob.Data.Name))
		return
	}

	// Get weapon damage from the item
	var damage int32 = 0
	if heldItemIsWeapon {
		weaponData := itemData.data.(*ItemDataWeapon)
		damage = weaponData.damage
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
	targetMob.Data.Health -= damage

	// Broadcast result to room
	critStr := ""
	if crit {
		critStr = "Critical hit! "
	}
	world.messageRoom(mob.Data.Room, fmt.Sprintf("%s%s struck %s for %d damage.", critStr, mob.Data.Name, targetMob.Data.Name, damage))
	if targetMob.IsDead() {
		world.messageRoom(mob.Data.Room, fmt.Sprintf("%s has slain %s.", mob.Data.Name, targetMob.Data.Name))
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
	if mob.mode != MOB_MODE_CAST {
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
	mob.mode = MOB_MODE_IDLE
	world.messageRoom(mob.Data.Room, fmt.Sprintf("%s lost concentration on their spell!", mob.Data.Name))
}

func (mob *Mob) subtractDurabilityFromEquipment(world *World, slot EquipmentSlot) {
	item := mob.Data.Equipment.Get(slot)
	if item == nil {
		return
	}

	itemData := ITEM_DATA[item.Id]

	item.Durability--
	if item.Durability == 0 {
		world.messageRoom(mob.Data.Room, fmt.Sprintf("%s's %s broke!", mob.Data.Name, itemData.name))
		item, _ := mob.Data.Equipment.Unequip(slot)
		mob.onPlayerItemUnequipped(world, item)
		return
	}

	// The rest of these messages are only sent to players holding the item
	if mob.PlayerCharacter == nil {
		return
	}

	// If item has become damaged, tell the user
	maxDurability := item.getMaxDurability()
	itemWasDamaged := (item.Durability + 1) < maxDurability / 2
	itemIsDamaged := item.Durability < maxDurability / 2
	if itemIsDamaged && !itemWasDamaged {
		world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Your %s is now damaged.", itemData.name))
		return
	}

	// If an item has lost its sharpness, tell the user
	itemWasSharp := (item.Durability + 1) > maxDurability
	itemIsSharp := item.Durability > maxDurability
	if itemWasSharp && !itemIsSharp {
		itemIsWeapon :=
			itemData.itemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED ||
			itemData.itemType == ITEM_TYPE_EQUIPMENT_TWO_HANDED
		if itemIsWeapon {
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Your %s has lost its sharpness.", itemData.name))
		} else {
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Your %s has lost its fortification.", itemData.name)) }
	}
}

func (mob *Mob) onPlayerItemUnequipped(world *World, item Item) {
	if mob.PlayerCharacter == nil {
		log.Printf("Warn - onPlayerItemUnequipped was called on a non-player mob.")
		return
	}

	itemData := ITEM_DATA[item.Id]
	if itemData.itemType == ITEM_TYPE_EQUIPMENT_SPELLBOOK {
		spellbookData := itemData.data.(*ItemDataSpellbook)

		// Decrement the equip count for this spell
		mob.PlayerCharacter.SpellsEquipped[spellbookData.spell].EquipCount--

		// If the equip count is now 0, delete the entry and remove the spell
		if mob.PlayerCharacter.SpellsEquipped[spellbookData.spell].EquipCount == 0 {
			delete(mob.PlayerCharacter.SpellsEquipped, spellbookData.spell)

			isSpellPrepared := slices.Contains(mob.Data.Spells, spellbookData.spell)
			isSpellKnown := slices.Contains(mob.PlayerCharacter.SpellsKnown, spellbookData.spell)
			if isSpellPrepared && !isSpellKnown {
				mob.Data.RemoveSpell(spellbookData.spell)
				spellData := SPELL_DATA[spellbookData.spell]
				world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("You lost the spell %s.", spellData.name))
			}
		}
	}
}

func (mob *Mob) spellcast(world *World, targetMob *Mob) {
	// Check if caster has enough mana
	spellData := SPELL_DATA[mob.castSpell]
	if mob.Data.Mana < spellData.manaCost {
		mob.mode = MOB_MODE_IDLE
		world.messageRoom(mob.Data.Room, fmt.Sprintf("%s tried to cast %s, but they don't have enough mana.", mob.Data.Name, spellData.name))
		return
	}

	// Check spell timer
	if mob.castTimer > 0 {
		mob.castTimer--
		world.messageRoom(mob.Data.Room, fmt.Sprintf("%s is charging a spell...", mob.Data.Name))
		return
	}

	// Cast spell
	world.messageRoom(mob.Data.Room, fmt.Sprintf("%s cast %s!", mob.Data.Name, spellData.name))
	mob.Data.Mana -= spellData.manaCost
	spellData.onHit(world, mob, targetMob)

	if mob.PlayerCharacter != nil {
		equippedSpell, spellIsEquipped := mob.PlayerCharacter.SpellsEquipped[mob.castSpell]
		if spellIsEquipped && !equippedSpell.IsKnown {

			// Spell mastery progress
			equippedSpell.Casts++
			if equippedSpell.Casts >= mob.Data.CastsToLearn(mob.castSpell) {
				equippedSpell.IsKnown = true
				mob.PlayerCharacter.SpellsKnown = append(mob.PlayerCharacter.SpellsKnown, mob.castSpell)
				world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("You have mastered %s!", spellData.name))
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
	if itemData.itemType != ITEM_TYPE_EQUIPMENT_SPELLBOOK {
		return 0, false
	}

	spellbookData := itemData.data.(*ItemDataSpellbook)
	return spellbookData.spell, true
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
		mob.mode = MOB_MODE_IDLE
		return
	}

	// Remove item from their inventory
	item := mob.Data.Inventory.RemoveItem(itemIndex)

	// Use item
	itemData := ITEM_DATA[item.Id]
	world.messageRoom(mob.Data.Room, fmt.Sprintf("%s used %s!", mob.Data.Name, itemData.name))

	switch itemData.itemType {
		case ITEM_TYPE_CONSUMABLE:
			consumableData := itemData.data.(*ItemDataConsumable)
			consumableData.onUse(world, targetMob)
		case ITEM_TYPE_SPELL_SCROLL:
			scrollData := itemData.data.(*ItemDataSpellScroll)
			spellData := SPELL_DATA[scrollData.spell]
			spellData.onHit(world, mob, targetMob)
		default:
			panic(fmt.Sprintf("Unhandled item type %s. This item type should never have been allowed to be used here.", ItemTypeToString(itemData.itemType)))
	}
}
