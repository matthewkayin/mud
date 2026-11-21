package game

import (
	"log"
	"fmt"
	"slices"
	"strings"
	"strconv"
	"mud/world"
	"errors"
	"mud/util"
)

type FuzzyResult int
const (
	FUZZY_RESULT_OK FuzzyResult = iota
	FUZZY_RESULT_TARGET_NOT_SPECIFIED
	FUZZY_RESULT_NOT_FOUND
	FUZZY_RESULT_AMBIGUOUS
	FUZZY_RESULT_NUMBER_OUT_OF_RANGE
	FUZZY_RESULT_AMBIGUOUS_LOCATION
)

const FUZZY_NUMBER_NOT_SPECIFIED int = -1

// An error from a fuzzy search that keeps the FuzzyResult, so callers can tell a missing match from an ambiguous one
type FuzzyFindError struct {
	result FuzzyResult
	message string
}

// Where parseItemLocation says to look for an item. slot is set for ITEM_LOCATION_EQUIPMENT, chest for ITEM_LOCATION_CHEST
type ItemLocationResult struct {
	location world.ItemLocation
	slot world.EquipmentSlot
	chest *world.Chest
}

func (err *FuzzyFindError) Error() string {
	return err.message
}

// Splits args at the first occurrence of word, ignoring case
func splitArgsBy(args []string, word string) ([]string, []string, bool) {
	index := slices.IndexFunc(args, func(arg string) bool {
		return strings.EqualFold(arg, word)
	})
	if index == -1 {
		return args, []string{}, false
	}

	return args[:index], args[index + 1:], true
}

// Splits args into item words and location words at an 'in' whose following words name an item location
// (see parseItemLocation). The last such 'in' is used, so 'jack in the box in inventory' splits before 'inventory'.
// If no 'in' is followed by a location, all of args are item words, so names like 'light in window' still match.
func splitArgsByItemLocation(room *world.Room, args []string) ([]string, []string, bool) {
	for index := len(args) - 1; index >= 0; index-- {
		if !strings.EqualFold(args[index], "in") {
			continue
		}

		locationWords := args[index + 1:]
		if len(locationWords) == 0 {
			continue
		}

		_, ok := parseItemLocation(room, locationWords)
		if ok {
			return args[:index], locationWords, true
		}
	}

	return args, []string{}, false
}

func getFuzzyNumberFromArgs(args []string) (int, []string) {
	if len(args) == 0 {
		return FUZZY_NUMBER_NOT_SPECIFIED, args
	}

	lastIndex := len(args) - 1
	number, err := strconv.Atoi(args[lastIndex])
	if err != nil {
		return FUZZY_NUMBER_NOT_SPECIFIED, args
	}

	// Numbers below 1 are out of range. Use 0 for all of them so that a typed number
	// can't collide with FUZZY_NUMBER_NOT_SPECIFIED
	if number < 1 {
		number = 0
	}

	return number, args[:lastIndex]
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

// Scores all names and returns the best score and the indices of every name with that score
func fuzzyFindBestIndices(names []string, searchWords []string) (int, []int) {
	bestScore := 0
	bestIndices := make([]int, 0, 1)

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

	return bestScore, bestIndices
}

func fuzzyFind(names []string, searchWords []string, fuzzyNumber int) (int, FuzzyResult) {
	bestScore, bestIndices := fuzzyFindBestIndices(names, searchWords)

	// If there's no winners, then no result was found
	if bestScore == 0 {
		return 0, FUZZY_RESULT_NOT_FOUND
	}

	// If there's only 1 winner, then return that
	if len(bestIndices) == 1 && (fuzzyNumber == 1 || fuzzyNumber == FUZZY_NUMBER_NOT_SPECIFIED) {
		return bestIndices[0], FUZZY_RESULT_OK
	}

	// If no number provided, don't try to disambiguate
	if fuzzyNumber == FUZZY_NUMBER_NOT_SPECIFIED {
		return 0, FUZZY_RESULT_AMBIGUOUS
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
		return 0, FUZZY_RESULT_AMBIGUOUS
	}

	// Ensure the fuzzy number is in range
	if fuzzyNumber < 1 || fuzzyNumber > len(bestIndices) {
		return 0, FUZZY_RESULT_NUMBER_OUT_OF_RANGE
	}

	// If they do have the same name and a number was provided, use that number to disambiguate
	return bestIndices[fuzzyNumber - 1], FUZZY_RESULT_OK
}

func fuzzyFindInventoryItem(inventory *world.Inventory, searchWords []string) (int, FuzzyResult) {
	if len(searchWords) == 0 {
		return 0, FUZZY_RESULT_TARGET_NOT_SPECIFIED
	}
	if len(inventory.Items) == 0 {
		return 0, FUZZY_RESULT_NOT_FOUND
	}

	// Get fuzzy number
	var fuzzyNumber int
	fuzzyNumber, searchWords = getFuzzyNumberFromArgs(searchWords)

	// Re-check search words length after getting fuzzy number
	if len(searchWords) == 0 {
		return 0, FUZZY_RESULT_TARGET_NOT_SPECIFIED
	}

	// Put all item names into an array
	itemNames := make([]string, len(inventory.Items))
	for index := range len(inventory.Items) {
		itemNames[index] = inventory.Items[index].NameWithCondition()
	}

	// Fuzzy find the item
	return fuzzyFind(itemNames, searchWords, fuzzyNumber)
}

func fuzzyFindInventoryItemResultString(result FuzzyResult, inventoryName string, searchWords []string) string {
	switch result {
		case FUZZY_RESULT_TARGET_NOT_SPECIFIED:
			return "You must specify an item"
		case FUZZY_RESULT_NOT_FOUND:
			return fmt.Sprintf("There is no item called '%s' in %s.", strings.Join(searchWords, " "), inventoryName)
		case FUZZY_RESULT_AMBIGUOUS:
			return fmt.Sprintf("There are multiple items matching '%s' in %s.", strings.Join(searchWords, " "), inventoryName)
		case FUZZY_RESULT_NUMBER_OUT_OF_RANGE:
			return fmt.Sprintf("There is no '%s' matching that number in %s.", strings.Join(searchWords, " "), inventoryName)
		case FUZZY_RESULT_OK:
			log.Printf("warn: you should not call fuzzyFindInventoryItemResultString() on FUZZY_RESULT_OK.")
			return "Something strange happened."
		default:
			log.Printf("warn: fuzzy result type %d not handled in fuzzyFindInventoryItemResultString().", result)
			return "Something strange happened."
	}
}

func fuzzyFindEquipmentSlot(searchWords []string) (world.EquipmentSlot, error) {
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify an equipment slot.")
	}

	// Equipment slot names
	slotNames := make([]string, world.EQUIPMENT_SLOT_COUNT)
	for index := range world.EQUIPMENT_SLOT_COUNT {
		slot := world.EquipmentSlot(index)
		slotNames[index] = slot.String()
	}

	slotIndex, fuzzyResult := fuzzyFind(slotNames, searchWords, FUZZY_NUMBER_NOT_SPECIFIED)

	// Handle edge cases
	if fuzzyResult == FUZZY_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("'%s' is not an equipment slot. Valid slot names are: %s",
			strings.Join(searchWords, " "), util.CombineNames(slotNames))
	}
	if fuzzyResult == FUZZY_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The equipment slot string '%s' is ambiguous.",
			strings.Join(searchWords, " "))
	}
	if fuzzyResult == FUZZY_RESULT_NUMBER_OUT_OF_RANGE {
		panic("Received fuzzy number out of range when no fuzzy number was provided.")
	}

	return world.EquipmentSlot(slotIndex), nil
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

		itemNames = append(itemNames, item.Name)
		equipmentSlots = append(equipmentSlots, slot)
	}

	// Check if they even have any items
	if len(itemNames) == 0 {
		return 0, errors.New("You have no items equipped.")
	}

	index, result := fuzzyFind(itemNames, searchWords, FUZZY_NUMBER_NOT_SPECIFIED)

	// Handle edge cases
	if result == FUZZY_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("'%s' is not an item you have equipped.",
			strings.Join(searchWords, " "))
	}
	if result == FUZZY_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The item string '%s' is ambiguous.",
			strings.Join(searchWords, " "))
	}
	if result == FUZZY_RESULT_NUMBER_OUT_OF_RANGE {
		panic("Received fuzzy number out of range when no fuzzy number was provided.")
	}

	return equipmentSlots[index], nil
}

// Finds an item that a player can target: in their inventory or equipment, on the room floor, or in a chest in the room.
// itemWords and locationWords are the words before and after 'in' (see splitArgsByItemLocation). Without a location, the
// player's inventory, equipment and the room floor are searched together. If the best matches are in more than one of
// those locations, the search fails with FUZZY_RESULT_AMBIGUOUS_LOCATION rather than guessing which one the player means.
// Chests are only searched when named as the location, since opening a chest rolls its loot.
// Fuzzy search failures are returned as *FuzzyFindError.
func fuzzyFindTargetItem(gamestate *GameState, player *Player, itemWords []string, locationWords []string, userSpecifiedLocation bool) (*world.Item, world.ItemTarget, error) {
	if len(itemWords) == 0 {
		return nil, world.ItemTarget{}, fmt.Errorf("You must specify an item.")
	}

	room := &gamestate.world.Rooms[player.character.Room]
	playerTarget := world.ItemTarget {
		Holder: world.ITEM_TARGET_HOLDER_MOB,
		Mob: player.mobHandle,
	}
	roomTarget := world.ItemTarget {
		Holder: world.ITEM_TARGET_HOLDER_ROOM,
		Room: player.character.Room,
	}

	// If user specified location, search only that location
	if userSpecifiedLocation {
		location, ok := parseItemLocation(room, locationWords)
		if !ok {
			return nil, world.ItemTarget{}, fmt.Errorf("'%s' is not 'room', 'inventory', an equipment slot, or a container in the room.", strings.Join(locationWords, " "))
		}

		switch location.location {
			case world.ITEM_LOCATION_INVENTORY: {
				return fuzzyFindTargetItemInInventory(&player.character.Inventory, "your inventory", playerTarget, itemWords)
			}

			case world.ITEM_LOCATION_ROOM: {
				return fuzzyFindTargetItemInInventory(&room.Inventory, "the room", roomTarget, itemWords)
			}

			case world.ITEM_LOCATION_CHEST: {
				chestInventory := location.chest.GetInventory(gamestate.world, player.character)
				chestTarget := world.ItemTarget {
					Holder: world.ITEM_TARGET_HOLDER_CHEST,
					Room: player.character.Room,
				}
				return fuzzyFindTargetItemInInventory(chestInventory, location.chest.Name, chestTarget, itemWords)
			}

			case world.ITEM_LOCATION_EQUIPMENT: {
				item := player.character.Equipment.Get(location.slot)
				// Known issue, left as-is on purpose: itemWords may end in a fuzzy number (as in 'ring 1 in accessory'),
				// which matches no word in the item's name, so the score is 0 and the item isn't found. A slot holds
				// at most one item, so players have no reason to add a number here.
				if item == nil || scoreFuzzyMatch(item.NameWithCondition(), itemWords) == 0 {
					return nil, world.ItemTarget{}, fmt.Errorf("There is no item called '%s' in your %s.", strings.Join(itemWords, " "), strings.ToLower(location.slot.String()))
				}

				playerTarget.Id = item.Id
				return item, playerTarget, nil
			}

			default: {
				log.Printf("warn: item location %d not handled in fuzzyFindTargetItem()", location.location)
				return nil, world.ItemTarget{}, fmt.Errorf("Something strange happened.")
			}
		}
	}

	// If user did not specify location, search inventory, equipment and the room floor
	fuzzyNumber, itemWords := getFuzzyNumberFromArgs(itemWords)
	if len(itemWords) == 0 {
		return nil, world.ItemTarget{}, fmt.Errorf("You must specify an item.")
	}

	// Create parallel arrays of item name, item, target and location, in the order inventory, equipment, room floor.
	// The location is the word the player would type after 'in' to search only there
	capacity := len(player.character.Inventory.Items) + world.EQUIPMENT_SLOT_COUNT + len(room.Inventory.Items)
	itemNames := make([]string, 0, capacity)
	items := make([]*world.Item, 0, capacity)
	targets := make([]world.ItemTarget, 0, capacity)
	locationWordsByItem := make([]string, 0, capacity)
	for index := range len(player.character.Inventory.Items) {
		item := &player.character.Inventory.Items[index]
		itemNames = append(itemNames, item.NameWithCondition())
		items = append(items, item)
		targets = append(targets, playerTarget)
		locationWordsByItem = append(locationWordsByItem, "inventory")
	}

	for index := range world.EQUIPMENT_SLOT_COUNT {
		// Get returns nil for empty slots, and for the off hand while a two-handed weapon is equipped
		slot := world.EquipmentSlot(index)
		item := player.character.Equipment.Get(slot)
		if item == nil {
			continue
		}
		itemNames = append(itemNames, item.NameWithCondition())
		items = append(items, item)
		targets = append(targets, playerTarget)
		locationWordsByItem = append(locationWordsByItem, strings.ToLower(slot.String()))
	}

	for index := range len(room.Inventory.Items) {
		item := &room.Inventory.Items[index]
		itemNames = append(itemNames, item.NameWithCondition())
		items = append(items, item)
		targets = append(targets, roomTarget)
		locationWordsByItem = append(locationWordsByItem, "room")
	}

	// If the best matches are in more than one location, make the player name the location.
	// This is checked even when a number was given, so that a number never picks between locations
	// searchedLocationsDescription names where the player is told the search looked. It's narrowed to a single
	// location when all the best matches are there, so that "multiple items matching" names the right place
	searchedLocationsDescription := "your inventory, your equipment, or the room"
	bestScore, bestIndices := fuzzyFindBestIndices(itemNames, itemWords)
	if bestScore > 0 {
		matchingLocations := make([]string, 0, 1)
		for _, bestIndex := range bestIndices {
			if !slices.Contains(matchingLocations, locationWordsByItem[bestIndex]) {
				matchingLocations = append(matchingLocations, locationWordsByItem[bestIndex])
			}
		}

		if len(matchingLocations) > 1 {
			locationDescriptions := make([]string, 0, len(matchingLocations))
			locationSuggestions := make([]string, 0, len(matchingLocations))
			for _, locationWord := range matchingLocations {
				locationDescriptions = append(locationDescriptions, describeItemLocation(locationWord))
				locationSuggestions = append(locationSuggestions, fmt.Sprintf("'in %s'", locationWord))
			}

			return nil, world.ItemTarget{}, &FuzzyFindError {
				result: FUZZY_RESULT_AMBIGUOUS_LOCATION,
				message: fmt.Sprintf("'%s' matches items in %s. Add %s to choose one.",
					strings.Join(itemWords, " "), util.CombineNames(locationDescriptions), util.CombineNamesWithConjunction(locationSuggestions, "or")),
			}
		}

		searchedLocationsDescription = describeItemLocation(matchingLocations[0])
	}

	index, result := fuzzyFind(itemNames, itemWords, fuzzyNumber)
	if result != FUZZY_RESULT_OK {
		return nil, world.ItemTarget{}, &FuzzyFindError {
			result: result,
			message: fuzzyFindInventoryItemResultString(result, searchedLocationsDescription, itemWords),
		}
	}

	target := targets[index]
	target.Id = items[index].Id
	return items[index], target, nil
}

// Fuzzy finds an item in a single inventory. target is the holder of the inventory, and gets the found item's ID
func fuzzyFindTargetItemInInventory(inventory *world.Inventory, inventoryName string, target world.ItemTarget, itemWords []string) (*world.Item, world.ItemTarget, error) {
	index, result := fuzzyFindInventoryItem(inventory, itemWords)
	if result != FUZZY_RESULT_OK {
		return nil, world.ItemTarget{}, &FuzzyFindError {
			result: result,
			message: fuzzyFindInventoryItemResultString(result, inventoryName, itemWords),
		}
	}

	item := &inventory.Items[index]
	target.Id = item.Id
	return item, target, nil
}

// Describes a location word from fuzzyFindTargetItem ('inventory', 'room' or a lowercase equipment slot name) for a message
func describeItemLocation(locationWord string) string {
	switch locationWord {
		case "inventory":
			return "your inventory"
		case "room":
			return "the room"
		default:
			return "your " + locationWord
	}
}

func parseItemLocation(room *world.Room, locationWords []string) (ItemLocationResult, bool) {
	if len(locationWords) == 1 && strings.EqualFold(locationWords[0], "inventory") {
		return ItemLocationResult { location: world.ITEM_LOCATION_INVENTORY }, true
	}

	if len(locationWords) == 1 && strings.EqualFold(locationWords[0], "room") {
		return ItemLocationResult { location: world.ITEM_LOCATION_ROOM }, true
	}

	slot, err := fuzzyFindEquipmentSlot(locationWords)
	if err == nil {
		return ItemLocationResult { location: world.ITEM_LOCATION_EQUIPMENT, slot: slot }, true
	}

	chest, err := fuzzyFindChest(room, locationWords)
	if err == nil {
		return ItemLocationResult { location: world.ITEM_LOCATION_CHEST, chest: chest }, true
	}

	return ItemLocationResult{}, false
}

func fuzzyFindMob(gamestate *GameState, player *Player, searchWords []string) (world.MobHandle, error) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		return world.MobHandle{}, fmt.Errorf("You must specify a target.")
	}

	// Handle when user targets "self"
	if len(searchWords) == 1 && strings.EqualFold(searchWords[0], "self") {
		return player.mobHandle, nil
	}

	// Get fuzzy number
	var fuzzyNumber int
	fuzzyNumber, searchWords = getFuzzyNumberFromArgs(searchWords)

	// Re-check search words length after getting fuzzy number
	if len(searchWords) == 0 {
		return world.MobHandle{}, fmt.Errorf("You must specify a target.")
	}

	// Put all room occupants in an array
	room := &gamestate.world.Rooms[player.character.Room]
	mobNames := make([]string, 0, len(room.Occupants))
	for _, handle := range room.Occupants {
		occupant := gamestate.world.Mobs.Get(handle)

		// TODO: check hidden occupants
		mobNames = append(mobNames, occupant.Name)
	}

	// Fuzzy find the mob
	index, result := fuzzyFind(mobNames, searchWords, fuzzyNumber)

	// Handle result
	switch result {
		case FUZZY_RESULT_OK:
			return room.Occupants[index], nil
		case FUZZY_RESULT_NOT_FOUND:
			return world.MobHandle{}, fmt.Errorf("No target in the room matches the name '%s'.", strings.Join(searchWords, " "))
		case FUZZY_RESULT_AMBIGUOUS:
			return world.MobHandle{}, fmt.Errorf("There are multiple targets in the room matching the name '%s'.", strings.Join(searchWords, " "))
		case FUZZY_RESULT_NUMBER_OUT_OF_RANGE:
			return world.MobHandle{}, fmt.Errorf("There is no '%s' matching that number in the room.", strings.Join(searchWords, " "))
		default:
			log.Printf("warn: fuzzy result %d not handled in fuzzyFindMob().", result)
			return world.MobHandle{}, fmt.Errorf("Something strange happened.")
	}
}

func fuzzyFindChest(room *world.Room, searchWords []string) (*world.Chest, error) {
	// Get fuzzy number
	var fuzzyNumber int
	fuzzyNumber, searchWords = getFuzzyNumberFromArgs(searchWords)

	// Check for empty search words
	if len(searchWords) == 0 {
		return nil, fmt.Errorf("You must specify a container.")
	}

	// Build chest names
	chestNames := make([]string, 0, len(room.Chests))
	for index := range len(room.Chests) {
		chest := &room.Chests[index]
		chestNames = append(chestNames, chest.Name)
	}

	// Fuzzy find
	index, result := fuzzyFind(chestNames, searchWords, fuzzyNumber)

	// Handle result
	switch result {
		case FUZZY_RESULT_OK:
			return &room.Chests[index], nil
		case FUZZY_RESULT_NOT_FOUND:
			return nil, fmt.Errorf("No container in the room matches the name '%s'.", strings.Join(searchWords, " "))
		case FUZZY_RESULT_AMBIGUOUS:
			return nil, fmt.Errorf("There are multiple containers in the room matching the name '%s'.", strings.Join(searchWords, " "))
		case FUZZY_RESULT_NUMBER_OUT_OF_RANGE:
			return nil, fmt.Errorf("There is no '%s' matching that number in the room.", strings.Join(searchWords, " "))
		default:
			log.Printf("warn: fuzzy result %d not handled in fuzzyFindChest().", result)
			return nil, fmt.Errorf("Something stranged happened.")
	}
}
