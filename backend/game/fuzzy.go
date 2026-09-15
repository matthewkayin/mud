package game

import (
	"fmt"
	"slices"
	"strings"
)

const FUZZY_FIND_RESULT_NOT_FOUND = -1
const FUZZY_FIND_RESULT_AMBIGUOUS = -2

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

func fuzzyFindTarget(gameState *GameState, player *Player, searchWords []string) (MobHandle, bool) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		*player.inbox <- "You must specify a target."
		return MobHandle{}, false
	}

	// Handle when user targets "self"
	if len(searchWords) == 1 && strings.EqualFold(searchWords[0], "self") {
		return player.mobHandle, true
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
		*player.inbox <- fmt.Sprintf("No target in the room matches the name '%s'.", strings.Join(searchWords, " "))
		return MobHandle{}, false
	}
	if targetIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		*player.inbox <- fmt.Sprintf("The target string '%s' is ambiguous.", strings.Join(searchWords, " "))
		return MobHandle{}, false
	}

	return room.occupants[targetIndex], true
}

func fuzzyFindPreparedSpell(gameState *GameState, player *Player, searchWords []string) (Spell, bool) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		*player.inbox <- "You must specify a spell."
		return 0, false
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
		*player.inbox <- fmt.Sprintf("You have no prepared spell called '%s'.", strings.Join(searchWords, " "))
		return 0, false
	}
	if spellIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		*player.inbox <- fmt.Sprintf("The spell string '%s' is ambiguous.", strings.Join(searchWords, " "))
		return 0, false
	}

	return playerMob.data.Spells[spellIndex], true
}

func fuzzyFindKnownOrEquippedSpell(player *Player, searchWords []string) (Spell, bool) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		*player.inbox <- "You must specify a spell."
		return 0, false
	}

	// Get equipped spells into a flat array
	spellsEquipped := make([]Spell, 0, 2)
	for spell, _ := range player.character.SpellsEquipped {
		spellsEquipped = append(spellsEquipped, spell)
	}

	// Put all spell names into an array
	spellNames := make([]string, len(player.character.SpellsKnown) + len(player.character.SpellsEquipped))
	for _, spell := range player.character.SpellsKnown{
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
		*player.inbox <- fmt.Sprintf("You have no known or equipped spell called '%s'", strings.Join(searchWords, " "))
		return 0, false
	}
	if spellIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		*player.inbox <- fmt.Sprintf("The spell string '%s' is ambiguous.", strings.Join(searchWords, " "))
		return 0, false
	}

	// If spell is a known spell
	if spellIndex < len(player.character.SpellsKnown) {
		return player.character.SpellsKnown[spellIndex], true
	}

	// if the spell is an equipped spell
	return spellsEquipped[spellIndex - len(player.character.SpellsKnown)], true
}

func fuzzyFindInventoryItemIndex(inventory *ItemList, searchWords []string) int {
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

func fuzzyFindEquipmentSlotByItem(player *Player, equipment *Equipment, searchWords []string) (EquipmentSlot, bool) {
	if len(searchWords) == 0 {
		*player.inbox <- "You must specify an item."
		return 0, false
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
		*player.inbox <- "You have no items equipped."
		return 0, false
	}

	index := fuzzyFind(itemNames, searchWords)

	// Handle edge cases
	if index == FUZZY_FIND_RESULT_NOT_FOUND {
		*player.inbox <- fmt.Sprintf("'%s' is not an item you have equipped.",
			strings.Join(searchWords, " "))
		return 0, false
	}
	if index == FUZZY_FIND_RESULT_AMBIGUOUS {
		*player.inbox <- fmt.Sprintf("The item string '%s' is ambiguous.",
			strings.Join(searchWords, " "))
		return 0, false
	}

	return equipmentSlots[index], true
}

func fuzzyFindEquipmentSlot(player *Player, searchWords []string) (EquipmentSlot, bool) {
	if len(searchWords) == 0 {
		*player.inbox <- "You must specify an equipment slot."
		return 0, false
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
		*player.inbox <- fmt.Sprintf("'%s' is not an equipment slot. Valid slot names are: %s",
			strings.Join(searchWords, " "), combineNames(slotNames))
		return 0, false
	}
	if slotIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		*player.inbox <- fmt.Sprintf("The equipment slot string '%s' is ambiguous.",
			strings.Join(searchWords, " "))
		return 0, false
	}

	return EquipmentSlot(slotIndex), true
}
