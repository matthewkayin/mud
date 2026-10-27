package game

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"strconv"
	"mud/world"
	"mud/util"
)

const FUZZY_FIND_NUMBER_NONE = -1

const FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED = -1
const FUZZY_FIND_RESULT_NOT_FOUND = -2
const FUZZY_FIND_RESULT_AMBIGUOUS = -3
const FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE = -4
// The item matched in both the mob's inventory and its equipment
const FUZZY_FIND_RESULT_AMBIGUOUS_LOCATION = -5

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
	mobNames := make([]string, 0, len(room.Occupants))
	for _, occupantHandle := range room.Occupants {
		occupant := gamestate.world.Mobs.Get(occupantHandle)

		// Skip hidden occupants, except you can still target yourself of course
		if occupant.CheckFlag(world.MOB_FLAG_HIDDEN) && occupantHandle != player.mobHandle {
			continue
		}

		mobNames = append(mobNames, occupant.GetName())
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

func fuzzyFindFixture(gamestate *GameState, player *Player, searchWords []string) (*world.Fixture, error) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		return nil, errors.New("You must specify a target.")
	}

	// Get handle to player room
	playerMob := gamestate.world.Mobs.Get(player.mobHandle)
	room := gamestate.world.Rooms[playerMob.Data.Room]

	// Check that there are any fixtures
	if len(room.Fixtures) == 0 {
		return nil, fmt.Errorf("No target in the room matches the name '%s'.", strings.Join(searchWords, " "))
	}

	// Put all fixture names into an array
	fixtureNames := make([]string, 0, len(room.Fixtures))
	for index := range room.Fixtures {
		fixtureNames = append(fixtureNames, room.Fixtures[index].GetName())
	}

	// Fuzzy find the fixture
	// TODO: fuzzy numbers on the fixtures, but just not gonna worry about it for tuesday
	index := fuzzyFind(fixtureNames, searchWords, FUZZY_FIND_NUMBER_NONE)

	if index == FUZZY_FIND_RESULT_NOT_FOUND {
		return nil, fmt.Errorf("")
	}
	// Handle edge cases
	if index == FUZZY_FIND_RESULT_NOT_FOUND {
		return nil, fmt.Errorf("No target in the room matches the name '%s'.", strings.Join(searchWords, " "))
	}
	if index == FUZZY_FIND_RESULT_AMBIGUOUS {
		return nil, fmt.Errorf("The target string '%s' is ambiguous.", strings.Join(searchWords, " "))
	}
	if index == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
		panic("Received fuzzy result number out of range but no fuzzy number was specified.")
	}

	return &room.Fixtures[index], nil
}

// Chooses either a prepared spell or a class spell
func fuzzyFindCastableSpell(gamestate *GameState, player *Player, searchWords []string) (world.SpellId, error) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify a spell.")
	}

	// Get handle to player mob
	playerMob := gamestate.world.Mobs.Get(player.mobHandle)

	// Put all spell names into an array
	spellNames := make([]string, 0, len(playerMob.Data.Spells) + len(player.character.ClassSpells))
	for _, spell := range playerMob.Data.Spells {
		spellData := world.SPELL_DATA[spell]
		spellNames = append(spellNames, spellData.Name)
	}
	for _, spell := range player.character.ClassSpells {
		spellData := world.SPELL_DATA[spell]
		spellNames = append(spellNames, spellData.Name)
	}

	// Fuzzy find the spell to cast
	spellIndex := fuzzyFind(spellNames, searchWords, FUZZY_FIND_NUMBER_NONE)

	// Handle edge cases
	if spellIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("You have no castable spell called '%s'.", strings.Join(searchWords, " "))
	}
	if spellIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return 0, fmt.Errorf("The spell name '%s' is ambiguous.", strings.Join(searchWords, " "))
	}
	if spellIndex == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
		panic("Received fuzzy number out of range when no fuzzy number was provided.")
	}

	if spellIndex < len(playerMob.Data.Spells) {
		return playerMob.Data.Spells[spellIndex], nil
	}

	return player.character.ClassSpells[spellIndex - len(playerMob.Data.Spells)], nil
}

func fuzzyFindPreparedSpell(gamestate *GameState, player *Player, searchWords []string) (world.SpellId, error) {
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

func fuzzyFindKnownOrEquippedSpell(player *Player, searchWords []string) (world.SpellId, error) {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify a spell.")
	}

	// Get equipped spells into a flat array
	spellsEquipped := make([]world.SpellId, 0, 2)
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

func fuzzyFindClassSpellIndex(player *Player, searchWords []string) int {
	// Check that there are any arguments
	if len(searchWords) == 0 {
		return FUZZY_FIND_RESULT_NOT_FOUND
	}

	// Put all spell names into an array
	spellNames := make([]string, 0, len(player.character.ClassSpells))
	for _, spell := range player.character.ClassSpells {
		spellData := world.SPELL_DATA[spell]
		spellNames = append(spellNames, spellData.Name)
	}

	// Fuzzy find spell
	return fuzzyFind(spellNames, searchWords, FUZZY_FIND_NUMBER_NONE)
}

func fuzzyFindInventoryItemIndex(inventory *world.Inventory, searchWords []string) int {
	if len(searchWords) == 0 || inventory.Length() == 0 {
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

// inventoryName is used in error messages, e.g. "your inventory" or "the room"
func fuzzyFindInventoryItem(inventory *world.Inventory, inventoryName string, searchWords []string) (int, error) {
	itemIndex := fuzzyFindInventoryItemIndex(inventory, searchWords)
	switch itemIndex {
		case FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED:
			return 0, errors.New("You must specify an item.")
		case FUZZY_FIND_RESULT_NOT_FOUND:
			return 0, fmt.Errorf("There is no item called '%s' in %s.", strings.Join(searchWords, " "), inventoryName)
		case FUZZY_FIND_RESULT_AMBIGUOUS:
			return 0, fmt.Errorf("There are multiple items matching '%s' in %s.", strings.Join(searchWords, " "), inventoryName)
		case FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE:
			return 0, fmt.Errorf("There is no item matching that number in %s.", inventoryName)
	}

	return itemIndex, nil
}

// Searches both the mob's inventory and equipment for the item.
// The result is non-negative when the item was found, otherwise it is a FUZZY_FIND_RESULT.
// This function does not check for the "in" keyword to disambiguate betweeen item "in inventory" or "in <slot>", for that, use fuzzyFindMobItem
func fuzzyFindMobInventoryOrEquippedItem(mob *world.Mob, searchWords []string) (world.ItemHandle, int) {
	inventoryIndex := fuzzyFindInventoryItemIndex(&mob.Data.Inventory, searchWords)
	slot, equipmentResult := fuzzyFindEquippedItemSlot(&mob.Data.Equipment, searchWords)

	if inventoryIndex >= 0 && equipmentResult >= 0 {
		return world.ItemHandle{}, FUZZY_FIND_RESULT_AMBIGUOUS_LOCATION
	}
	if inventoryIndex >= 0 {
		return world.ItemHandle {
			Mob: mob.Handle,
			Location: world.ITEM_LOCATION_INVENTORY,
			Index: inventoryIndex,
			Id: mob.Data.Inventory.Items[inventoryIndex].Id,
		}, inventoryIndex
	}
	if equipmentResult >= 0 {
		return world.ItemHandle {
			Mob: mob.Handle,
			Location: world.ITEM_LOCATION_EQUIPMENT,
			Index: int(slot),
			Id: mob.Data.Equipment.Get(slot).Id,
		}, equipmentResult
	}

	// Report the inventory's result unless only the equipment search has something to say
	if inventoryIndex == FUZZY_FIND_RESULT_NOT_FOUND || inventoryIndex == FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED {
		return world.ItemHandle{}, equipmentResult
	}
	return world.ItemHandle{}, inventoryIndex
}

// Returns the slot of the equipped item matching searchWords, along with the fuzzyFind result
func fuzzyFindEquippedItemSlot(equipment *world.Equipment, searchWords []string) (world.EquipmentSlot, int) {
	if len(searchWords) == 0 {
		return 0, FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED
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

		itemNames = append(itemNames, item.GetNameWithCondition())
		equipmentSlots = append(equipmentSlots, slot)
	}

	result := fuzzyFind(itemNames, searchWords, FUZZY_FIND_NUMBER_NONE)
	if result < 0 {
		return 0, result
	}

	return equipmentSlots[result], result
}

// Reads the words after "in" as the player's inventory or one of their equipment slots
func parseItemLocation(locationWords []string) (world.ItemLocation, world.EquipmentSlot, bool) {
	if len(locationWords) == 1 && strings.EqualFold(locationWords[0], "inventory") {
		return world.ITEM_LOCATION_INVENTORY, 0, true
	}

	slot, err := fuzzyFindEquipmentSlot(locationWords)
	if err != nil {
		return 0, 0, false
	}

	return world.ITEM_LOCATION_EQUIPMENT, slot, true
}

// Finds an item in the mob's inventory or equipment.
// searchWords may end with "in inventory" or "in <equipment slot>" to search only there.
func fuzzyFindMobItem(mob *world.Mob, searchWords []string) (world.ItemHandle, error) {
	itemWords, locationWords, userSpecifiedLocation := splitArgsBy(searchWords, "in")
	if len(itemWords) == 0 {
		return world.ItemHandle{}, errors.New("You must specify an item.")
	}

	if userSpecifiedLocation {
		location, slot, isLocation := parseItemLocation(locationWords)
		if !isLocation {
			return world.ItemHandle{}, fmt.Errorf("'%s' is not your inventory or an equipment slot.", strings.Join(locationWords, " "))
		}

		switch location {
			// If they specified an item "in inventory", then fuzzy find in their inventory and return the result
			case world.ITEM_LOCATION_INVENTORY: {
				index, err := fuzzyFindInventoryItem(&mob.Data.Inventory, "your inventory", itemWords)
				if err != nil {
					return world.ItemHandle{}, err
				}

				return world.ItemHandle {
					Mob: mob.Handle,
					Location: world.ITEM_LOCATION_INVENTORY,
					Index: index,
					Id: mob.Data.Inventory.Items[index].Id,
				}, nil
			}

			// If they specified "in <slot>", then fuzzy find against the equipped item in that slot and return the result
			case world.ITEM_LOCATION_EQUIPMENT: {
				item := mob.Data.Equipment.Get(slot)
				if item == nil || fuzzyFind([]string { item.GetNameWithCondition() }, itemWords, FUZZY_FIND_NUMBER_NONE) < 0 {
					return world.ItemHandle{}, fmt.Errorf("There is no item called '%s' equipped in your %s.",
						strings.Join(itemWords, " "), strings.ToLower(slot.String()))
				}

				return world.ItemHandle {
					Mob: mob.Handle,
					Location: world.ITEM_LOCATION_EQUIPMENT,
					Index: int(slot),
					Id: item.Id,
				}, nil
			}
		}
	}

	// Fuzzy find in both equipment and inventory and return the result
	handle, result := fuzzyFindMobInventoryOrEquippedItem(mob, itemWords)
	switch result {
		case FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED:
			return handle, errors.New("You must specify an item.")
		case FUZZY_FIND_RESULT_NOT_FOUND:
			return handle, fmt.Errorf("You have no item called '%s'.", strings.Join(itemWords, " "))
		case FUZZY_FIND_RESULT_AMBIGUOUS:
			return handle, fmt.Errorf("There are multiple items matching '%s'.", strings.Join(itemWords, " "))
		case FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE:
			return handle, errors.New("There is no item matching that number in your inventory.")
		case FUZZY_FIND_RESULT_AMBIGUOUS_LOCATION:
			return handle, ambiguousItemLocationError(itemWords)
	}

	return handle, nil
}

func ambiguousItemLocationError(itemWords []string) error {
	return fmt.Errorf("There are items matching '%s' in both your inventory and your equipment. Add 'in inventory' or 'in <slot>' to choose one.",
		strings.Join(itemWords, " "))
}

func fuzzyFindWare(wares []world.ItemId, merchantName string, searchWords []string) (world.ItemId, error) {
	if len(wares) == 0 {
		panic("This function should not be called with an empty wares array.")
	}
	if len(searchWords) == 0 {
		return 0, errors.New("You must specify an item.")
	}

	// Put all ware names into an array
	itemNames := make([]string, len(wares))
	for index := range len(wares) {
		itemNames[index] = world.ITEM_DATA[wares[index]].Name
	}

	index := fuzzyFind(itemNames, searchWords, FUZZY_FIND_NUMBER_NONE)
	switch index {
		case FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED:
			return 0, errors.New("You must specify an item.")
		case FUZZY_FIND_RESULT_NOT_FOUND:
			return 0, fmt.Errorf("%s is not selling an item called '%s'.", merchantName, strings.Join(searchWords, " "))
		case FUZZY_FIND_RESULT_AMBIGUOUS:
			return 0, fmt.Errorf("%s has multiple items for sale matching '%s'.", merchantName, strings.Join(searchWords, " "))
		case FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE:
			panic("fuzzyFindWare hit number out of range even though no fuzzy number was specified")
	}

	return wares[index], nil
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
		slotNames[index] = slot.String()
	}

	slotIndex := fuzzyFind(slotNames, searchWords, FUZZY_FIND_NUMBER_NONE)

	// Handle edge cases
	if slotIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return 0, fmt.Errorf("'%s' is not an equipment slot. Valid slot names are: %s",
			strings.Join(searchWords, " "), util.CombineNames(slotNames))
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

func fuzzyFindKnownRecipe(character *world.Character, searchWords []string) (world.RecipeId, error) {
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
