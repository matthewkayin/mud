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
)

const FUZZY_NUMBER_NOT_SPECIFIED int = 1

func splitArgsBy(args []string, word string) ([]string, []string, bool) {
	index := slices.Index(args, word)
	if index == -1 {
		return args, []string{}, false
	}

	return args[:index], args[index + 1:], true
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

func fuzzyFind(names []string, searchWords []string, fuzzyNumber int) (int, FuzzyResult) {
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
		return 0, FUZZY_RESULT_NOT_FOUND
	}

	// If there's only 1 winner, then return that
	if len(bestIndices) == 1 && fuzzyNumber == 1 {
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
			log.Printf("Warn: You should not call fuzzyFindInventoryItemResultString() on FUZZY_RESULT_OK.")
			return "Something strange happened."
		default:
			log.Printf("Warn: Fuzzy result type %d not handled in fuzzyFindInventoryItemResultString().", result)
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

func fuzzyFindPlayerItem(player *Player, searchWords []string) (world.MobItemTarget, error) {
	// Split args
	itemWords, locationWords, userSpecifiedLocation := splitArgsBy(searchWords, "in")
	if len(itemWords) == 0 {
		return world.MobItemTarget{}, fmt.Errorf("You must specify an item.")
	}

	// If user specified location, search just inventory or just equipment
	if userSpecifiedLocation {
		location, slot, isLocation := parseItemLocation(locationWords)
		if !isLocation {
			return world.MobItemTarget{}, fmt.Errorf("'%s' is not your 'inventory' or an equipment slot.", strings.Join(locationWords, " "))
		}

		switch location {
			case world.ITEM_LOCATION_INVENTORY: {
				index, result := fuzzyFindInventoryItem(&player.character.Inventory, itemWords)
				if result != FUZZY_RESULT_OK {
					return world.MobItemTarget{}, fmt.Errorf("%s", fuzzyFindInventoryItemResultString(result, "your inventory", itemWords))
				}

				return world.MobItemTarget {
					Mob: player.mobHandle,
					Id: player.character.Inventory.Items[index].Id,
				}, nil
			}

			case world.ITEM_LOCATION_EQUIPMENT: {
				item := player.character.Equipment.Get(slot)
				if item == nil {
					return world.MobItemTarget{}, fmt.Errorf("There is no item called '%s' in your %s.", strings.Join(itemWords, " "), strings.ToLower(slot.String()))
				}

				return world.MobItemTarget {
					Mob: player.mobHandle,
					Id: item.Id,
				}, nil
			}
		}
	}

	// If user did not specify location, search both inventory and equipment
	fuzzyNumber, itemWords := getFuzzyNumberFromArgs(itemWords)
	if len(itemWords) == 0 {
		return world.MobItemTarget{}, fmt.Errorf("You must specify an item.")
	}

	itemNames := make([]string, 0, len(player.character.Inventory.Items) + world.EQUIPMENT_SLOT_COUNT)
	for index := range len(player.character.Inventory.Items) {
		itemNames = append(itemNames, player.character.Inventory.Items[index].NameWithCondition())
	}

	for index := range world.EQUIPMENT_SLOT_COUNT {
		slot := world.EquipmentSlot(index)
		item := player.character.Equipment.Get(slot)
		if slot == world.EQUIPMENT_SLOT_OFF_HAND && item.IsTwoHandedWeapon() {
			continue
		}
		itemNames = append(itemNames, item.NameWithCondition())
	}

	index, result := fuzzyFind(itemNames, searchWords, fuzzyNumber)
	if result == FUZZY_RESULT_OK {
		var item *world.Item = nil
		if index < len(player.character.Inventory.Items) {
			item = &player.character.Inventory.Items[index]
		} else {
			slotIndex := index - len(player.character.Inventory.Items)
			if player.character.Equipment.IsTwoHandedWeaponEquipped() && slotIndex >= world.EQUIPMENT_SLOT_OFF_HAND {
				slotIndex++
			}
			slot := world.EquipmentSlot(slotIndex)
			item = player.character.Equipment.Get(slot)
		}

		if item == nil {
			log.Printf("warn: fuzzyFindPlayerItem failed, item is nil")
			return world.MobItemTarget{}, fmt.Errorf("Something went wrong.")
		}

		return world.MobItemTarget {
			Mob: player.mobHandle,
			Id: item.Id,
		}, nil
	}

	return world.MobItemTarget{}, fmt.Errorf("%s", fuzzyFindInventoryItemResultString(result, "your inventory", itemWords))
}

func parseItemLocation(locationWords []string) (world.ItemLocation, world.EquipmentSlot, bool) {
	if len(locationWords) == 1 && strings.EqualFold(locationWords[0], "inventory") {
		return world.ITEM_LOCATION_INVENTORY, 0, true
	}

	slot, err := fuzzyFindEquipmentSlot(locationWords)
	if err == nil {
	    return world.ITEM_LOCATION_EQUIPMENT, slot, true
	}

	return 0, 0, false
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
			log.Printf("Warn: Fuzzy result %d not handled in fuzzyFindMob().", result)
			return world.MobHandle{}, fmt.Errorf("Something strange happened.")
	}
}
