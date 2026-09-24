package game

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"strconv"
	"mud/world"
)

const FUZZY_FIND_NUMBER_NONE = -1

const FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED = -1
const FUZZY_FIND_RESULT_NOT_FOUND = -2
const FUZZY_FIND_RESULT_AMBIGUOUS = -3
const FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE = -4

func splitArgsBy(args []string, word string) ([]string, []string, bool) {
	index := slices.Index(args, word)
	if index == -1 {
		return args, []string{}, false
	}

	return args[:index], args[index + 1:], true
}

func getFuzzyNumberFromArgs(args []string) (int, []string) {
	if len(args) == 0 {
		return 1, args
	}

	lastIndex := len(args) - 1
	number, err := strconv.Atoi(args[lastIndex])
	if err != nil {
		return 1, args
	}

	return number, args[:lastIndex]
}

func combineNames(names []string) string {
	switch len(names) {
		case 0:
			return ""
		case 1:
			return names[0]
		case 2:
			return names[0] + " and " + names[1]
		default:
			return strings.Join(names[:len(names) - 1], ", ") + ", and " + names[len(names) - 1]
	}
}

func scoreFuzzyMatch(name string, searchWords []string) int {
	nameLower := strings.ToLower(name)
	searchLower := strings.ToLower(strings.Join(searchWords, " "))

	// Check for exact match
	if nameLower == searchLower {
		return 100
	}

	// Check for continuous substring match
	substrIndex := strings.Index(nameLower, searchLower)
	if substrIndex != -1 {
		searchAlignsWithWordBoundary := substrIndex == 0 || nameLower[substrIndex - 1] == ' '
		if searchAlignsWithWordBoundary {
			return 99
		}

		return 98
	}

	// Check for order-independent word match
	score := 0
	nameWords := strings.Fields(nameLower)

	for _, searchWord := range searchWords {
		searchWordHasMatch := false

		// Check the search word against each name word
		for _, nameWord := range nameWords {
			// Exact matches count for 2
			if searchWord == nameWord {
				score += 2
				searchWordHasMatch = true
				break
			// Substr matches count for 1
			} else if strings.Contains(nameWord, searchWord) {
				score++
				searchWordHasMatch = true
				break
			}
		}

		// Don't match searches with irrelevant words
		if !searchWordHasMatch {
			return 0
		}
	}

	return score
}

func fuzzyFind(names []string, searchWords []string, fuzzyNumber int) int {
	bestScore := 0
	bestIndices := make([]int, 0, 1)

	// Score all names and keep a list of all indices who are high-scorers
	for index := range len(names) {
		score := scoreFuzzyMatch(names[index], searchWords)

		if score > bestScore {
			bestScore = score
			bestIndices = make([]int, 0, 1)
		}
		if score == bestScore {
			bestIndices = append(bestIndices, index)
		}
	}

	// If there's no winners, then no result was found
	if bestScore == 0 {
		return FUZZY_FIND_RESULT_NOT_FOUND
	}

	// If there's only 1 winner, then return that
	if len(bestIndices) == 1 && (fuzzyNumber == 1 || fuzzyNumber == FUZZY_FIND_NUMBER_NONE) {
		return bestIndices[0]
	}

	// If no number provided, don't try to disambiguate
	if fuzzyNumber == FUZZY_FIND_NUMBER_NONE {
		return FUZZY_FIND_RESULT_AMBIGUOUS
	}

	// Check if the best names are all the same
	bestNamesAreSame := true
	for index := 1; index < len(bestIndices); index++ {
		if !strings.EqualFold(names[bestIndices[0]], names[bestIndices[index]]) {
			bestNamesAreSame = false
			break
		}
	}

	// If they don't have the same name, then the result is ambiguous
	if !bestNamesAreSame {
		return FUZZY_FIND_RESULT_AMBIGUOUS
	}

	// Ensure the fuzzy number is in range
	if fuzzyNumber < 1 || fuzzyNumber > len(bestIndices) {
		return FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE
	}

	// If they do have the same name and a number was provided, use that number to disambiguate
	return bestIndices[fuzzyNumber - 1]
}

func fuzzyFindTarget(gamestate *GameState, player *Player, searchWords []string) (world.MobHandle, error) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		return world.MobHandle{}, errors.New("You must specify a target.")
	}

	// Handle when user targets "self"
	if len(searchWords) == 1 && strings.EqualFold(searchWords[0], "self") {
		return player.mobHandle, nil
	}

	// Get handle to player room
	playerMob := gamestate.world.Mobs.Get(player.mobHandle)
	room := gamestate.world.Rooms[playerMob.Data.Room]

	// Put all room occupant names into an array
	mobNames := make([]string, len(room.Occupants))
	for index, occupantHandle := range room.Occupants {
		occupant := gamestate.world.Mobs.Get(occupantHandle)
		mobNames[index] = occupant.GetName()
	}

	// Fuzzy find the target mob
	targetIndex := fuzzyFind(mobNames, searchWords, FUZZY_FIND_NUMBER_NONE)

	// Handle edge cases
	if targetIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return world.MobHandle{}, fmt.Errorf("No target in the room matches the name '%s'.", strings.Join(searchWords, " "))
	}
	if targetIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return world.MobHandle{}, fmt.Errorf("The target string '%s' is ambiguous.", strings.Join(searchWords, " "))
	}
	if targetIndex == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
		panic("Received fuzzy result number out of range but no fuzzy number was specified.")
	}

	return room.Occupants[targetIndex], nil
}

func fuzzyFindPreparedSpell(gamestate *GameState, player *Player, searchWords []string) (world.Spell, error) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify a spell.")
	}

	// Get handle to player mob
	playerMob := gamestate.world.Mobs.Get(player.mobHandle)

	// Put all spell names into an array
	spellNames := make([]string, len(playerMob.Data.Spells))
	for index, spell := range playerMob.Data.Spells {
		spellData := world.SPELL_DATA[spell]
		spellNames[index] = spellData.Name
	}

	// Fuzzy find the target spell
	spellIndex := fuzzyFind(spellNames, searchWords, FUZZY_FIND_NUMBER_NONE)

	// Handle edge cases
	if spellIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("You have no prepared spell called '%s'.", strings.Join(searchWords, " "))
	}
	if spellIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The spell string '%s' is ambiguous.", strings.Join(searchWords, " "))
	}
	if spellIndex == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
		panic("Received fuzzy number out of range when no fuzzy number was provided.")
	}

	return playerMob.Data.Spells[spellIndex], nil
}

func fuzzyFindKnownOrEquippedSpell(player *Player, searchWords []string) (world.Spell, error) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify a spell.")
	}

	// Get equipped spells into a flat array
	spellsEquipped := make([]world.Spell, 0, 2)
	for spell, _ := range player.character.SpellsEquipped {
		spellsEquipped = append(spellsEquipped, spell)
	}

	// Put all spell names into an array
	spellNames := make([]string, 0, len(player.character.SpellsKnown) + len(player.character.SpellsEquipped))
	for _, spell := range player.character.SpellsKnown {
		spellData := world.SPELL_DATA[spell]
		spellNames = append(spellNames, spellData.Name)
	}
	for _, spell := range spellsEquipped {
		spellData := world.SPELL_DATA[spell]
		spellNames = append(spellNames, spellData.Name)
	}

	// Fuzzy find the target spell
	spellIndex := fuzzyFind(spellNames, searchWords, FUZZY_FIND_NUMBER_NONE)

	// Handle edge cases
	if spellIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("You have no known or equipped spell called '%s'", strings.Join(searchWords, " "))
	}
	if spellIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The spell string '%s' is ambiguous.", strings.Join(searchWords, " "))
	}
	if spellIndex == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
		panic("Received fuzzy number out of range when no fuzzy number was provided.")
	}

	// If spell is a known spell
	if spellIndex < len(player.character.SpellsKnown) {
		return player.character.SpellsKnown[spellIndex], nil
	}

	// If the spell is an equipped spell
	return spellsEquipped[spellIndex - len(player.character.SpellsKnown)], nil
}

func fuzzyFindInventoryItemIndex(inventory *world.Inventory, searchWords []string) int {
	if len(searchWords) == 0 {
		return FUZZY_FIND_RESULT_NOT_FOUND
	}

	// Get fuzzy number
	var fuzzyNumber int
	fuzzyNumber, searchWords = getFuzzyNumberFromArgs(searchWords)

	// Put all item names into an array
	itemNames := make([]string, len(inventory.Items))
	for index := range len(inventory.Items) {
		itemNames[index] = inventory.Items[index].GetNameWithCondition()
	}

	// Fuzzy find the item
	return fuzzyFind(itemNames, searchWords, fuzzyNumber)
}

func fuzzyFindEquipmentSlotByItem(equipment *world.Equipment, searchWords []string) (world.EquipmentSlot, error) {
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify an item.")
	}

	// Create parallel arrays of equipped item name and equipment slot
	itemNames := make([]string, 0, world.EQUIPMENT_SLOT_COUNT)
	equipmentSlots := make([]world.EquipmentSlot, 0, world.EQUIPMENT_SLOT_COUNT)
	for slotIndex := range world.EQUIPMENT_SLOT_COUNT {
		slot := world.EquipmentSlot(slotIndex)
		item := equipment.Get(slot)
		if item == nil {
			continue
		}

		itemData := world.ITEM_DATA[item.Id]
		itemNames = append(itemNames, itemData.Name)
		equipmentSlots = append(equipmentSlots, slot)
	}

	// Check if they even have any items
	if len(itemNames) == 0 {
		return 0, errors.New("You have no items equipped.")
	}

	index := fuzzyFind(itemNames, searchWords, FUZZY_FIND_NUMBER_NONE)

	// Handle edge cases
	if index == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("'%s' is not an item you have equipped.",
			strings.Join(searchWords, " "))
	}
	if index == FUZZY_FIND_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The item string '%s' is ambiguous.",
			strings.Join(searchWords, " "))
	}
	if index == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
		panic("Received fuzzy number out of range when no fuzzy number was provided.")
	}

	return equipmentSlots[index], nil
}

func fuzzyFindEquipmentSlot(searchWords []string) (world.EquipmentSlot, error) {
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify an equipment slot.")
	}

	// Equipment slot names
	slotNames := make([]string, world.EQUIPMENT_SLOT_COUNT)
	for index := range world.EQUIPMENT_SLOT_COUNT {
		slot := world.EquipmentSlot(index)
		slotNames[index] = world.EquipmentSlotToString(slot)
	}

	slotIndex := fuzzyFind(slotNames, searchWords, FUZZY_FIND_NUMBER_NONE)

	// Handle edge cases
	if slotIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("'%s' is not an equipment slot. Valid slot names are: %s",
			strings.Join(searchWords, " "), combineNames(slotNames))
	}
	if slotIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The equipment slot string '%s' is ambiguous.",
			strings.Join(searchWords, " "))
	}
	if slotIndex == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
		panic("Received fuzzy number out of range when no fuzzy number was provided.")
	}

	return world.EquipmentSlot(slotIndex), nil
}

func fuzzyFindChestInventory(room *world.Room, searchWords []string) (*world.Inventory, string, error) {
	if len(searchWords) == 0 {
		return nil, "", errors.New("You must specify a container.")
	}

	// Check for "room"
	searchString := strings.Join(searchWords, " ")
	if strings.EqualFold(searchString, "room") {
		return &room.Inventory, "room", nil
	}

	// Get fuzzy number
	var fuzzyNumber int
	fuzzyNumber, searchWords = getFuzzyNumberFromArgs(searchWords)

	// Chest names
	chestNames := make([]string, 0, len(room.Chests))
	for index := range len(room.Chests) {
		chest := &room.Chests[index]
		chestNames = append(chestNames, chest.Name)
	}

	chestIndex := fuzzyFind(chestNames, searchWords, fuzzyNumber)

	// Handle edge cases
	if chestIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return nil, "", fmt.Errorf("There are no chests called '%s' in the room.",
			strings.Join(searchWords, " "))
	}
	if chestIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return nil, "", fmt.Errorf("The chest name '%s' is ambiguous.",
			strings.Join(searchWords, " "))
	}

	return &room.Chests[chestIndex].Inventory, room.Chests[chestIndex].Name, nil
}

func fuzzyFindKnownRecipe(character *world.Character, searchWords []string) (world.Recipe, error) {
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify a recipe.")
	}

	recipeNames := make([]string, len(character.RecipesKnown))
	for index, recipe := range character.RecipesKnown {
		recipeNames[index] = world.RECIPE_DATA[recipe].Name
	}

	index := fuzzyFind(recipeNames, searchWords, FUZZY_FIND_NUMBER_NONE)

	if index == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("You do not know a recipe named '%s'.",
			strings.Join(searchWords, " "))
	}
	if index == FUZZY_FIND_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The recipe string '%s' is ambiguous.",
			strings.Join(searchWords, " "))
	}

	return character.RecipesKnown[index], nil
}
