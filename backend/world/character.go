package world

import (
	"fmt"
	"math"
	"mud/bitset"
)

type Character struct {
	PlayerKey string

	Race RaceId
	Background BackgroundId

	Level int32
	Experience int32
	ExpTotal int32

	Attributes AttributeBlock
	Skills SkillBlock

	RoomsDiscovered []byte

	MobData
}

type CharacterSheet struct {
	Name string

	Race RaceId
	Background BackgroundId
}

func (world *World) NewCharacter(playerKey string, characterSheet *CharacterSheet) *Character {
	character := &Character{}

	character.PlayerKey = playerKey
	character.Name = characterSheet.Name
	character.Race = characterSheet.Race
	character.Background = characterSheet.Background

	// Set attributes and skills
	raceData := world.RaceData[character.Race]
	backgroundData := world.BackgroundData[character.Background]
	character.Attributes = raceData.Attributes.Add(&backgroundData.Attributes)
	character.Skills = raceData.Skills.Add(&backgroundData.Skills)

	character.RoomsDiscovered = bitset.New(len(world.Rooms))

	// Set the character's room to the start room
	origin := RoomPosition { X: 0, Y: 0, Z: 0 }
	character.Room = world.RoomPositionToIndex[origin]
	bitset.Set(character.RoomsDiscovered, character.Room, true)

	// TODO set based on stats and such
	character.Health = character.MaxHealth()
	character.Mana = character.MaxMana()

	return character
}

// MOB OWNER FUNCTIONS

func (character *Character) IsPlayer() bool {
	return true
}

func (character *Character) IsNpc() bool {
	return false
}

func (character *Character) GetPlayerKey() string {
	return character.PlayerKey
}

func (character *Character) getMobData() *MobData {
	return &character.MobData
}

func (character *Character) MaxHealth() int32 {
	return 40 + (character.Level * 10) + (character.Skills.Values[SKILL_FORTITUDE] * 25)
}

func (character *Character) MaxMana() int32 {
	return 40 + (character.Skills.Values[SKILL_SPELLCRAFT] * 10)
}

func (character *Character) InventoryCapacity() int32 {
	return INVENTORY_CAPACITY_PLAYER
}

func (character *Character) Message(world *World, message string) {
	world.MessagePlayer(character.PlayerKey, message)
}

func (character *Character) update(world *World) {
	// npc-only function, does nothing
}

// CHARACTER FUNCTIONS

func (character *Character) MaxSpellSlots() int32 {
	return 2 + character.Skills.Values[SKILL_SPELLCRAFT]
}

func ExpToNextLevel(level int32) int32 {
	return int32(756.0 * math.Pow(1.26, float64(level - 1)))
}

func LevelFromExpTotal(expTotal int32) int32 {
	level := int32(1)
	for expTotal > ExpToNextLevel(level) {
		level++
	}
	return level
}

func MaxAttributeAtLevel(level int32) int32 {
	return level + 4
}

func (character *Character) grantExperience(world *World, amount int32) {
	character.Experience += amount
	character.ExpTotal += amount
	for character.ExpTotal > ExpToNextLevel(character.Level) {
		character.Level++
		world.MessagePlayer(character.PlayerKey, fmt.Sprintf("Level up! You are now level %d.", character.Level))
	}
}
