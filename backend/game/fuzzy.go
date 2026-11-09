package game

import (
	"fmt"
	"slices"
	"strings"
	"strconv"
	"mud/world"
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

	// Put all item names into an array
	itemNames := make([]string, len(inventory.Items))
	for index := range len(inventory.Items) {
		itemNames[index] = inventory.Items[index].Name
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
			panic("You shouldn't call this function on an OK result.")
		default:
			panic(fmt.Sprintf("Fuzzy result type %d not handled", result))
	}
}
