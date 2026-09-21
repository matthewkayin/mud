package world

import (
	"log"
	"strings"
)

type CharacterSheet struct {
	Name string
	Race RaceId
	Class ClassId
	Job JobId
}

type Character struct {
	PlayerId int
	Race RaceId
	Class ClassId
	Job JobId

	// SpellsEquipped map[Spell]*CharacterEquippedSpell
	// SpellsKnown []Spell
	// RecipesKnown []Recipe
	// RoomsDiscovered []byte
	Data MobData
}

func CharacterInitEmpty(playerId int, characterSheet *CharacterSheet) *Character {
	var character *Character = &Character{}

	character.PlayerId = playerId
	character.Race = characterSheet.Race
	character.Class = characterSheet.Class
	character.Job = characterSheet.Job

	character.Data.Name = characterSheet.Name
	character.Data.Room = 0

	character.Data.Level = 1
	character.Data.Experience = 0
	// character.Data.ExperienceToNextLevel = character.Data.GetExpToNextLevel()

	// Get handles to race/class/job data
	raceData := RACE_DATA[character.Race]
	classData := CLASS_DATA[character.Class]
	jobData := JOB_DATA[character.Job]

	// Determine base stats
	character.Data.Stats = raceData.Stats
	character.Data.Stats = character.Data.Stats.Add(&classData.Stats)
	character.Data.Stats = character.Data.Stats.Add(&jobData.Stats)

	// Determine derived stats
	// character.Data.Health = character.Data.MaxHealth()
	// character.Data.Mana = character.Data.MaxMana()

	// Init spell list

	// Init inventory

	// Init equipment

	// Init rooms discovered

	return character
}

func (world *World) GetCharacterIfExists(name string) (*Character, bool) {
	character, exists := world.Characters[strings.ToLower(name)]
	return character, exists
}

func (world *World) AddCharacter(playerId int, character *Character) {
	world.Characters[strings.ToLower(character.Data.Name)] = character

	_, playerCharactersListExists := world.PlayerCharacters[playerId]
	if !playerCharactersListExists {
		world.PlayerCharacters[playerId] = make([]string, 0, 1)
	}

	oldCharacterList := world.PlayerCharacters[playerId]
	world.PlayerCharacters[playerId] = append(oldCharacterList, character.Data.Name)
}

func (world *World) RemoveCharacter(character *Character) {
	// Find the index of the character's name in the PlayerCharacters[playerId] array
	var index int
	for index = 0; index < len(world.PlayerCharacters[character.PlayerId]) - 1; index++ {
		if world.PlayerCharacters[character.PlayerId][index] == character.Data.Name {
			break
		}
	}

	// Remove the character's name at the index we just found
	if index < len(world.PlayerCharacters[character.PlayerId]) {
		// I'm choosing to do an ordered removal here because
		// 1. Player death is not a per-turn action, so we can afford the cost
		// 2. I think it'd be nice to preserve the order of the player character login list
		world.PlayerCharacters[character.PlayerId] = append(
			world.PlayerCharacters[character.PlayerId][:index],
			world.PlayerCharacters[character.PlayerId][index + 1:]...)
	} else {
		log.Printf("Warning - Character %s does not exist in the PlayerCharacters list for player %d", character.Data.Name, character.PlayerId)
	}

	delete(world.Characters, character.Data.Name)
}
