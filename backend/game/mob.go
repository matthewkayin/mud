package game

import (
	"fmt"
	"log"
	"math/rand/v2"
)

const MOB_MAX_LEVEL int32 = 20
const MOB_EXP_PER_LEVEL int32 = 300

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

// Making Cast K higher increases how effective intelligence is at reducing spell learn-time
// Spell Casts to learn = Base + (1 - (INT / 50.0) * 0.75)
// 					    => Base * (1 - INT * (0.75 / 50.0))
// Where Base is the spell's base casts to learn
// Higher intelligence therefore reduces the number of casts it takes to learn the spell
const MOB_CASTS_TO_LEARN_K float32 = 0.75 / 50.0

type MobBaseStats struct {
	Vitality int32
	Strength int32
	Agility int32
	Intelligence int32
	Faith int32
}

type MobData struct {
	Name string
	Room uint

	Level int32
	Experience int32
	ExperienceToNextLevel int32

	Stats MobBaseStats

	Health int32
	Mana int32

	Spells []Spell

	Inventory ItemList
	EquippedItems Equipment
}

type MobMode int32
const (
	MOB_MODE_IDLE MobMode = iota
	MOB_MODE_ATTACK
	MOB_MODE_CAST
)

type Mob struct {
	player *Player
	Data MobData

	Mode MobMode
	Target MobHandle
	CastSpell Spell
	CastTimer int32
}

func (stats *MobBaseStats) Add(other *MobBaseStats) MobBaseStats {
	return MobBaseStats {
		Vitality: stats.Vitality + other.Vitality,
		Strength: stats.Strength + other.Strength,
		Agility: stats.Agility + other.Agility,
		Intelligence: stats.Intelligence + other.Intelligence,
		Faith: stats.Faith + other.Faith,
	}
}

// Returns true if stats >= other
func (stats *MobBaseStats) Meets(other *MobBaseStats) bool {
	return !(stats.Vitality < other.Vitality ||
			stats.Strength < other.Strength ||
			stats.Agility < other.Agility ||
			stats.Intelligence < other.Intelligence ||
			stats.Faith < other.Faith)
}

func MobInitFromCharacter(player *Player, character *Character) Mob {
	playerMob := Mob {
		player: player,
		Data: character.Data,

		Mode: MOB_MODE_IDLE,
	}

	// Calculate equipment stat bonuses
	playerMob.Data.EquippedItems.CalculateStatBonuses()

	return playerMob
}

func (mob *Mob) IsDead() bool {
	return mob.Data.Health <= 0
}

func (mobData *MobData) GetExpToNextLevel() int32 {
	if mobData.Level == MOB_MAX_LEVEL {
		return 0
	}

	return mobData.Level * MOB_EXP_PER_LEVEL
}

func (mobData *MobData) MaxHealth() int32 {
	return mobData.Vitality() * 5
}

func (mobData *MobData) MaxMana() int32 {
	return mobData.Intelligence() * 5
}

func (mobData *MobData) Armor() int32 {
	outfit := mobData.EquippedItems.Get(EQUIPMENT_SLOT_OUTFIT)
	if outfit == nil {
		return 0
	}

	itemData := ITEM_DATA[outfit.Id]
	outfitData := itemData.data.(*ItemDataOutfit)
	return outfitData.armor
}

func (mobData *MobData) Vitality() int32 {
	return mobData.Stats.Vitality + mobData.EquippedItems.statBonuses.Vitality
}

func (mobData *MobData) Strength() int32 {
	return mobData.Stats.Strength + mobData.EquippedItems.statBonuses.Strength
}

func (mobData *MobData) Agility() int32 {
	return mobData.Stats.Agility + mobData.EquippedItems.statBonuses.Agility
}

func (mobData *MobData) Intelligence() int32 {
	return mobData.Stats.Intelligence + mobData.EquippedItems.statBonuses.Intelligence
}

func (mobData *MobData) Faith() int32 {
	return mobData.Stats.Faith + mobData.EquippedItems.statBonuses.Faith
}

func (mobData *MobData) SpellSlots() int32 {
	// Spell slots is based on base INT, not INT bonus,
	// otherwise they could equip INT attributes to increase
	// their spell slots, prepare the spells, and then unequip the items
	//
	// There is a case to be made that spell slots should be class determined
	// and separate from the int stat entirely

	return int32(float32(mobData.Stats.Intelligence) / 3.0)
}

func (mobData *MobData) CastsToLearn(spell Spell) int32 {
	spellData := SPELL_DATA[spell]
	spellCastsToLearn := float32(spellData.castsToLearn)
	mobInt := float32(mobData.Intelligence())

	return int32(spellCastsToLearn * (1.0 - (mobInt * MOB_CASTS_TO_LEARN_K)))
}

func (mobData *MobData) RemoveSpell(toRemove Spell) {
	for index, spell := range mobData.Spells {
		if spell == toRemove {
			lastIndex := len(mobData.Spells) - 1
			mobData.Spells[index] = mobData.Spells[lastIndex]
			mobData.Spells = mobData.Spells[:lastIndex]
			return
		}
	}
}

func (mob *Mob) GrantExperience(experience int32) {
	// This function is only meant for player mobs at this time
	if mob.player == nil {
		return
	}

	for experience > 0 && mob.Data.Level < MOB_MAX_LEVEL {
		if mob.Data.Experience + experience >= mob.Data.ExperienceToNextLevel {
			experience -= mob.Data.ExperienceToNextLevel

			// Increase level
			mob.Data.Experience = 0
			mob.Data.ExperienceToNextLevel = mob.Data.GetExpToNextLevel()
			mob.Data.Level++

			// Increase stats
			classData := CLASS_DATA[mob.player.character.Class]
			raceData := RACE_DATA[mob.player.character.Race]
			baseStats := classData.Stats.Add(&raceData.Stats)

			mob.Data.Stats.Vitality = CharacterStatAtLevel(baseStats.Vitality, classData.Stats.Vitality, mob.Data.Level)
			mob.Data.Stats.Strength = CharacterStatAtLevel(baseStats.Strength, classData.Stats.Strength, mob.Data.Level)
			mob.Data.Stats.Agility = CharacterStatAtLevel(baseStats.Agility, classData.Stats.Agility, mob.Data.Level)
			mob.Data.Stats.Intelligence = CharacterStatAtLevel(baseStats.Intelligence, classData.Stats.Intelligence, mob.Data.Level)
			mob.Data.Stats.Faith = CharacterStatAtLevel(baseStats.Faith, classData.Stats.Faith, mob.Data.Level)

			// Announce level up message
			*mob.player.inbox <- fmt.Sprintf("Level up! %s is now level %d", mob.Data.Name, mob.Data.Level)

			continue
		}

		mob.Data.Experience += experience
		experience = 0
	}
}

func (mob *Mob) SetModeAttack(targetHandle MobHandle) {
	mob.Mode = MOB_MODE_ATTACK
	mob.Target = targetHandle
}

func (mob *Mob) SetModeCast(spell Spell, targetHandle MobHandle) {
	mob.Mode = MOB_MODE_CAST
	mob.Target = targetHandle
	mob.CastSpell = spell
	mob.CastTimer = SPELL_DATA[spell].castTime
}

func (mob *Mob) Update(gameState *GameState) {
	switch mob.Mode {
		case MOB_MODE_IDLE:
		case MOB_MODE_ATTACK:
			// Check if target exists
			targetMob, targetExists := gameState.world.Mobs.GetIfExists(mob.Target)
			if !targetExists || targetMob.IsDead() || targetMob.Data.Room != mob.Data.Room {
				mob.Mode = MOB_MODE_IDLE
				break
			}

			room := &gameState.world.Rooms[mob.Data.Room]
			mob.AttackTargetWithWeapon(gameState, room, targetMob, EQUIPMENT_SLOT_MAIN_HAND)
			mob.AttackTargetWithWeapon(gameState, room, targetMob, EQUIPMENT_SLOT_OFF_HAND)
		case MOB_MODE_CAST:
			// Check if target exists
			targetMob, targetExists := gameState.world.Mobs.GetIfExists(mob.Target)
			if !targetExists || targetMob.Data.Health == 0 || targetMob.Data.Room != mob.Data.Room {
				mob.Mode = MOB_MODE_IDLE
				break
			}

			// Check if caster has enough mana
			spellData := SPELL_DATA[mob.CastSpell]
			room := gameState.world.Rooms[mob.Data.Room]
			if mob.Data.Mana < spellData.manaCost {
				mob.Mode = MOB_MODE_IDLE
				room.broadcast(gameState, fmt.Sprintf("%s tried to cast %s, but they don't have enough mana.", mob.Data.Name, spellData.name))
				break
			}

			// Check spell timer
			if mob.CastTimer > 0 {
				mob.CastTimer--
				room.broadcast(gameState, fmt.Sprintf("%s is charging a spell...", mob.Data.Name))
				break
			}

			// Cast spell
			room.broadcast(gameState, fmt.Sprintf("%s cast %s!", mob.Data.Name, spellData.name))
			mob.Data.Mana -= spellData.manaCost
			spellData.onHit(gameState, mob, targetMob)

			// Spell mastery progress
			if mob.player != nil {
				equippedSpell, spellIsEquipped := mob.player.character.SpellsEquipped[mob.CastSpell]
				if spellIsEquipped && !equippedSpell.IsKnown {
					equippedSpell.Casts++
					if equippedSpell.Casts >= mob.Data.CastsToLearn(mob.CastSpell) {
						equippedSpell.IsKnown = true
						mob.player.character.SpellsKnown = append(mob.player.character.SpellsKnown, mob.CastSpell)
						*mob.player.inbox <- fmt.Sprintf("You have mastered %s!", spellData.name)
					}
				}
			}

			mob.Mode = MOB_MODE_IDLE
		default:
			log.Printf("Mob mode %d not handled.", mob.Mode)
	}
}

func (mob *Mob) AttackTargetWithWeapon(gameState *GameState, room *Room, targetMob *Mob, slot EquipmentSlot) {
	// Check for weapon
	weapon := mob.Data.EquippedItems.Get(slot)
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
		room.broadcast(gameState, fmt.Sprintf("%s dodged %s's attack!", targetMob.Data.Name, mob.Data.Name))
		return
	}

	// Check for critical hit
	critChance := mobAgility * MOB_CRIT_K
	critRoll := rand.Float32()
	crit := critRoll < critChance

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

	// Subtract target armor from damage
	// Crits ignore half armor
	if crit {
		damage -= mob.Data.Armor() / 2.0
	} else {
		damage -= mob.Data.Armor()
	}

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
	room.broadcast(gameState, fmt.Sprintf("%s%s struck %s for %d damage.", critStr, mob.Data.Name, targetMob.Data.Name, damage))
	if targetMob.IsDead() {
		room.broadcast(gameState, fmt.Sprintf("%s has slain %s.", mob.Data.Name, targetMob.Data.Name))
	} else {
		targetMob.RollForConcentration(gameState, damage)
	}
}

func (mob *Mob) CalculateMagicDamage(baseDamage int32, target *Mob) int32 {
	return baseDamage + (mob.Data.Faith() / 2) + (target.Data.Faith() / 4)
}

func (mob *Mob) RollForConcentration(gameState *GameState, damage int32) {
	// Not concentrating
	if mob.Mode != MOB_MODE_CAST {
		return
	}

	// Don't break concentration for instants or spells that have already been charged
	if mob.CastTimer == 0 {
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
	mobRoom := &gameState.world.Rooms[mob.Data.Room]
	mobRoom.broadcast(gameState, fmt.Sprintf("%s lost concentration on their spell!", mob.Data.Name))
}
