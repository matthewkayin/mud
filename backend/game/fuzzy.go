package game

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

const FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED = -1
const FUZZY_FIND_RESULT_NOT_FOUND = -2
const FUZZY_FIND_RESULT_AMBIGUOUS = -3

func splitArgsBy(args []string, word string) ([]string, []string, bool) {
	index := slices.Index(args, word)
	if index == -1 {
		return args, []string{}, false
	}

	return args[:index], args[index + 1:], true
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

func fuzzyFind(names []string, searchWords []string) int {
	bestScore := 0
	bestIndex := 0
	itemsWithBestScore := 0

	for index := range len(names) {
		score := scoreFuzzyMatch(names[index], searchWords)

		if score > bestScore {
			bestScore = score
			bestIndex = index
			itemsWithBestScore = 0
		}
		if score == bestScore {
			itemsWithBestScore++
		}
	}

	if bestScore == 0 {
		return FUZZY_FIND_RESULT_NOT_FOUND
	}
	if itemsWithBestScore > 1 {
		return FUZZY_FIND_RESULT_AMBIGUOUS
	}

	return bestIndex
}

func fuzzyFindTarget(gameState *GameState, player *Player, searchWords []string) (MobHandle, error) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		*player.inbox <- "You must specify a target."
		return MobHandle{}, errors.New("You must specify a target")
	}

	// Handle when user targets "self"
	if len(searchWords) == 1 && strings.EqualFold(searchWords[0], "self") {
		return player.mobHandle, nil
	}

	// Get handle to player room
	playerMob := gameState.world.Mobs.Get(player.mobHandle)
	room := gameState.world.Rooms[playerMob.data.Room]

	// Put all room occupant names into an array
	mobNames := make([]string, len(room.occupants))
	for index, occupantHandle := range room.occupants {
		occupant := gameState.world.Mobs.Get(occupantHandle)
		mobNames[index] = occupant.data.Name
	}

	// Fuzzy find the target mob
	targetIndex := fuzzyFind(mobNames, searchWords)

	// Handle edge cases
	if targetIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return MobHandle{}, fmt.Errorf("No target in the room matches the name '%s'.", strings.Join(searchWords, " "))
	}
	if targetIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return MobHandle{}, fmt.Errorf("The target string '%s' is ambiguous.", strings.Join(searchWords, " "))
	}

	return room.occupants[targetIndex], nil
}

func fuzzyFindPreparedSpell(gameState *GameState, player *Player, searchWords []string) (Spell, error) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify a spell.")
	}

	// Get handle to player mob
	playerMob := gameState.world.Mobs.Get(player.mobHandle)

	// Put all spell names into an array
	spellNames := make([]string, len(playerMob.data.Spells))
	for index, spell := range playerMob.data.Spells {
		spellData := SPELL_DATA[spell]
		spellNames[index] = spellData.name
	}

	// Fuzzy find the target spell
	spellIndex := fuzzyFind(spellNames, searchWords)

	// Handle edge cases
	if spellIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("You have no prepared spell called '%s'.", strings.Join(searchWords, " "))
	}
	if spellIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The spell string '%s' is ambiguous.", strings.Join(searchWords, " "))
	}

	return playerMob.data.Spells[spellIndex], nil
}

func fuzzyFindKnownOrEquippedSpell(player *Player, searchWords []string) (Spell, error) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify a spell.")
	}

	// Get equipped spells into a flat array
	spellsEquipped := make([]Spell, 0, 2)
	for spell, _ := range player.character.SpellsEquipped {
		spellsEquipped = append(spellsEquipped, spell)
	}

	// Put all spell names into an array
	spellNames := make([]string, 0, len(player.character.SpellsKnown) + len(player.character.SpellsEquipped))
	for _, spell := range player.character.SpellsKnown {
		spellData := SPELL_DATA[spell]
		spellNames = append(spellNames, spellData.name)
	}
	for _, spell := range spellsEquipped {
		spellData := SPELL_DATA[spell]
		spellNames = append(spellNames, spellData.name)
	}

	// Fuzzy find the target spell
	spellIndex := fuzzyFind(spellNames, searchWords)

	// Handle edge cases
	if spellIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("You have no known or equipped spell called '%s'", strings.Join(searchWords, " "))
	}
	if spellIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The spell string '%s' is ambiguous.", strings.Join(searchWords, " "))
	}

	// If spell is a known spell
	if spellIndex < len(player.character.SpellsKnown) {
		return player.character.SpellsKnown[spellIndex], nil
	}

	// If the spell is an equipped spell
	return spellsEquipped[spellIndex - len(player.character.SpellsKnown)], nil
}

func fuzzyFindInventoryItemIndex(inventory *Inventory, searchWords []string) int {
	if len(searchWords) == 0 {
		return FUZZY_FIND_RESULT_NOT_FOUND
	}

	// Put all item names into an array
	itemNames := make([]string, len(inventory.Items))
	for index := range len(inventory.Items) {
		itemData := ITEM_DATA[inventory.Items[index].Id]
		itemNames[index] = itemData.name
	}

	// Fuzzy find the item
	return fuzzyFind(itemNames, searchWords)
}

func fuzzyFindEquipmentSlotByItem(equipment *Equipment, searchWords []string) (EquipmentSlot, error) {
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify an item.")
	}

	// Create parallel arrays of equipped item name and equipment slot
	itemNames := make([]string, 0, EQUIPMENT_SLOT_COUNT)
	equipmentSlots := make([]EquipmentSlot, 0, EQUIPMENT_SLOT_COUNT)
	for slotIndex := range EQUIPMENT_SLOT_COUNT {
		slot := EquipmentSlot(slotIndex)
		item := equipment.Get(slot)
		if item == nil {
			continue
		}

		itemData := ITEM_DATA[item.Id]
		itemNames = append(itemNames, itemData.name)
		equipmentSlots = append(equipmentSlots, slot)
	}

	// Check if they even have any items
	if len(itemNames) == 0 {
		return 0, errors.New("You have no items equipped.")
	}

	index := fuzzyFind(itemNames, searchWords)

	// Handle edge cases
	if index == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("'%s' is not an item you have equipped.",
			strings.Join(searchWords, " "))
	}
	if index == FUZZY_FIND_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The item string '%s' is ambiguous.",
			strings.Join(searchWords, " "))
	}

	return equipmentSlots[index], nil
}

func fuzzyFindEquipmentSlot(searchWords []string) (EquipmentSlot, error) {
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify an equipment slot.")
	}

	// Equipment slot names
	slotNames := make([]string, EQUIPMENT_SLOT_COUNT)
	for index := range EQUIPMENT_SLOT_COUNT {
		slot := EquipmentSlot(index)
		slotNames[index] = EquipmentSlotToString(slot)
	}

	slotIndex := fuzzyFind(slotNames, searchWords)

	// Handle edge cases
	if slotIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("'%s' is not an equipment slot. Valid slot names are: %s",
			strings.Join(searchWords, " "), combineNames(slotNames))
	}
	if slotIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The equipment slot string '%s' is ambiguous.",
			strings.Join(searchWords, " "))
	}

	return EquipmentSlot(slotIndex), nil
}

func fuzzyFindChestInventory(room *Room, searchWords []string) (*Inventory, string, error) {
	if len(searchWords) == 0 {
		return nil, "", errors.New("You must specify a container.")
	}

	// Check for "room"
	searchString := strings.Join(searchWords, " ")
	if strings.EqualFold(searchString, "room") {
		return &room.Inventory, "room", nil
	}

	// Chest names
	chestNames := make([]string, 0, len(room.Chests))
	for index := range len(room.Chests) {
		chest := &room.Chests[index]
		chestNames = append(chestNames, chest.Name)
	}

	chestIndex := fuzzyFind(chestNames, searchWords)

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
