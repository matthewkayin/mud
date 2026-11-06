package world

import (
	"log"
	"slices"
	"strings"
)

type Character struct {
	PlayerKey string

	Race RaceId
	Class ClassId

	MobData
}

type CharacterSheet struct {
	Name string

	Race RaceId
	Class ClassId
}

func NewCharacter(playerKey string, characterSheet *CharacterSheet) *Character {
	character := &Character{}

	character.PlayerKey = playerKey
	character.Name = characterSheet.Name
	character.Race = characterSheet.Race
	character.Class = characterSheet.Class

	return character
}

func (world *World) GetCharacterIfExists(name string) (*Character, bool) {
	character, exists := world.Characters[strings.ToLower(name)]
	return character, exists
}

func (world *World) AddCharacter(playerKey string, character *Character) {
	// Set the character's room to the start room
	origin := RoomPosition { X: 0, Y: 0, Z: 0 }
	character.Room = world.RoomPositionToIndex[origin]

	world.Characters[strings.ToLower(character.Name)] = character

	// Alloc a characters list for this player if it does not exist
	_, exists := world.PlayerCharacters[playerKey]
	if !exists {
		world.PlayerCharacters[playerKey] = make([]string, 0, 1)
	}

	// Add the character to the player's characters list
	oldCharacterList := world.PlayerCharacters[playerKey]
	world.PlayerCharacters[playerKey] = append(oldCharacterList, character.Name)
}

func (world *World) RemoveCharacter(character *Character) {
	// Find the index of the character's name in the PlayerCharacters[playerKey] array
	playerCharacters := world.PlayerCharacters[character.PlayerKey]
	index := slices.IndexFunc(playerCharacters, func (characterName string) bool {
		return characterName == character.Name
	})

	// Remove the character's name at the index we just found, using ordered removal
	if index < len(world.PlayerCharacters[character.PlayerKey]) {
		world.PlayerCharacters[character.PlayerKey] = append(playerCharacters[:index], playerCharacters[index + 1:]...)
	} else {
		log.Printf("Warn: Character %s is being removed but does not exist in playerCharacters list for player %s.", character.Name, character.PlayerKey)
	}

	delete(world.Characters, character.Name)
	deleteCharacter(character)
}
