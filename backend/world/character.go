package world

import (
	"fmt"
	"log"
	"strings"
	"slices"
	"mud/bitset"
	"encoding/json"
)

const CHARACTER_ROOMS_DISCOVERED_BYTE_SIZE int = WORLD_MAX_ROOMS / 8

type CharacterSheet struct {
	Name string
	Race RaceId
	Class ClassId
	Job JobId
}

type CharacterEquippedSpell struct {
	EquipCount int32
	Casts int32
	IsKnown bool
}

type Character struct {
	PlayerId int
	Race RaceId
	Class ClassId
	Job JobId

	SpellsEquipped map[SpellId]*CharacterEquippedSpell
	SpellsKnown []SpellId
	ClassSpells []SpellId
	RecipesKnown []RecipeId
	RoomsDiscovered []byte
	Data MobData
}

// CharacterJson represents how Character is stored when saved to a file
type CharacterJson struct {
	PlayerId int
	Race string
	Class string
	Job string

	SpellsEquipped map[string]*CharacterEquippedSpell
	SpellsKnown []string
	ClassSpells []string
	RecipesKnown []string
	RoomsDiscovered []byte
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
	character.Data.ExperienceToNextLevel = character.Data.GetExpToNextLevel()

	// Get handles to race/class/job data
	raceData := RACE_DATA[character.Race]
	classData := CLASS_DATA[character.Class]
	jobData := JOB_DATA[character.Job]

	// Determine base stats
	character.Data.Stats = raceData.Stats
	character.Data.Stats = character.Data.Stats.Add(&classData.Stats)
	character.Data.Stats = character.Data.Stats.Add(&jobData.Stats)

	// Determine derived stats
	character.Data.Health = character.Data.MaxHealth()
	character.Data.Mana = character.Data.MaxMana()

	// Init spell list
	character.Data.Spells = make([]SpellId, 0, 1)
	character.SpellsEquipped = make(map[SpellId]*CharacterEquippedSpell)
	character.SpellsKnown = make([]SpellId, 0, 1)
	character.ClassSpells = make([]SpellId, 0, 1)
	character.RecipesKnown = make([]RecipeId, 0, 1)

	// Init inventory
	character.Data.Inventory = Inventory {
		Items: make([]Item, 0, 1),
	}

	// Init equipment
	character.Data.Equipment = EquipmentInitEmpty()

	// Init rooms discovered
	character.RoomsDiscovered = bitset.New(CHARACTER_ROOMS_DISCOVERED_BYTE_SIZE)
	bitset.Set(character.RoomsDiscovered, character.Data.Room, true)

	// Grant first level unlocks
	for _, unlock := range classData.UnlocksAtLevel[character.Data.Level] {
		character.grantClassUnlock(unlock)
	}

	return character
}

func (character *Character) MarshalJSON() ([]byte, error) {
	characterJson := CharacterJson {
		PlayerId: character.PlayerId,
		Race: RACE_DATA[character.Race].Name,
		Class: CLASS_DATA[character.Class].Name,
		Job: JOB_DATA[character.Job].Name,

		SpellsEquipped: make(map[string]*CharacterEquippedSpell),
		SpellsKnown: spellsToStringArray(character.SpellsKnown),
		ClassSpells: spellsToStringArray(character.ClassSpells),
		RecipesKnown: recipesToStringArray(character.RecipesKnown),
		RoomsDiscovered: character.RoomsDiscovered,
		Data: character.Data,
	}
	return json.Marshal(&characterJson)
}

func (character *Character) UnmarshalJSON(data []byte) error {
	var characterJson CharacterJson
	err := json.Unmarshal(data, &characterJson)
	if err != nil {
		return err
	}

	character.PlayerId = characterJson.PlayerId
	var exists bool
	character.Race, exists = RACE_NAME_TO_ID[characterJson.Race]
	if !exists {
		log.Fatalf("No race ID matches race '%s'.", characterJson.Race)
	}
	character.Class, exists = CLASS_NAME_TO_ID[characterJson.Class]
	if !exists {
		log.Fatalf("No class ID matches class '%s'.", characterJson.Class)
	}
	character.Job, exists = JOB_NAME_TO_ID[characterJson.Job]
	if !exists {
		log.Fatalf("No job ID matches job '%s'.", characterJson.Job)
	}

	character.SpellsEquipped = make(map[SpellId]*CharacterEquippedSpell)
	for spellName, equippedSpell := range characterJson.SpellsEquipped {
		spellId, exists := SPELL_NAME_TO_ID[spellName]
		if !exists {
			log.Fatalf("No spell ID matches the spell '%s'.", spellName)
		}

		character.SpellsEquipped[spellId] = equippedSpell
	}

	character.SpellsKnown = spellsFromStringArray(characterJson.SpellsKnown)
	character.ClassSpells = spellsFromStringArray(characterJson.ClassSpells)
	character.RecipesKnown = recipesFromStringArray(characterJson.RecipesKnown)
	character.RoomsDiscovered = characterJson.RoomsDiscovered
	character.Data = characterJson.Data

	return nil
}

func spellsToStringArray(spellIds []SpellId) []string {
	result := make([]string, len(spellIds))
	for index := range len(spellIds) {
		result[index] = SPELL_DATA[spellIds[index]].Name
	}

	return result
}

func spellsFromStringArray(spellNames []string) []SpellId {
	result := make([]SpellId, len(spellNames))
	for index := range len(spellNames) {
		spellId, exists := SPELL_NAME_TO_ID[spellNames[index]]
		if !exists {
			log.Fatalf("No spell ID matches spell '%s'.", spellNames[index])
		}
		result[index] = spellId
	}

	return result
}

func recipesToStringArray(recipeIds []RecipeId) []string {
	result := make([]string, len(recipeIds))
	for index := range len(recipeIds) {
		result[index] = RECIPE_DATA[recipeIds[index]].Name
	}

	return result
}

func recipesFromStringArray(recipeNames []string) []RecipeId {
	result := make([]RecipeId, len(recipeNames))
	for index := range len(recipeNames) {
		recipeId, exists := RECIPE_NAME_TO_ID[recipeNames[index]]
		if !exists {
			log.Fatalf("No recipe ID matches recipe '%s'.", recipeNames[index])
		}
		result[index] = recipeId
	}

	return result
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

	// Delete the character from disk
	deleteCharacter(character)
}

func (character *Character) HasSpell(spell SpellId) bool {
	return slices.Contains(character.SpellsKnown, spell) ||
		slices.Contains(character.ClassSpells, spell)
}

func (character *Character) recalculateStats() {
	raceData := RACE_DATA[character.Race]
	classData := CLASS_DATA[character.Class]
	jobData := JOB_DATA[character.Job]

	baseStats := raceData.Stats
	baseStats = baseStats.Add(&classData.Stats)
	baseStats = baseStats.Add(&jobData.Stats)

	scaling := classData.Scaling
	scaling = scaling.Add(&jobData.Scaling)

	character.Data.Stats = calculateStatBlockAtLevel(&baseStats, &scaling, character.Data.Level)
}

func calculateStatBlockAtLevel(base *StatBlock, scaling *StatBlock, level int32) StatBlock {
	stats := StatBlock { Values: [STAT_COUNT]int32{} }
	for index := range STAT_COUNT {
		stats.Values[index] = calculateStatAtLevel(base.Values[index], scaling.Values[index], level)
	}

	return stats
}

func calculateStatAtLevel(base int32, scaling int32, level int32) int32 {
	return base + int32(2.0 * float32(level - 1) * (float32(scaling) / 10.0))
}

func (character *Character) grantClassUnlock(unlock ClassUnlock) string {
	switch unlock.Type {
		case CLASS_UNLOCK_TYPE_ABILITY: {
			ability := unlock.Data.(MobAbility)
			abilityData := MOB_ABILITY_DATA[ability]

			character.Data.SetHasAbility(ability, true)
			return fmt.Sprintf("You got the ability %s!", abilityData.Name)
		}

		case CLASS_UNLOCK_TYPE_SPELL: {
			spell := unlock.Data.(SpellId)
			spellData := SPELL_DATA[spell]

			character.ClassSpells = append(character.ClassSpells, spell)
			return fmt.Sprintf("You learned the spell %s!", spellData.Name)
		}

		default:
			panic(fmt.Sprintf("Unhandled class unlock type %d.", unlock.Type))
	}
}
