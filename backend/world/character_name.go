package world

import (
	"fmt"
	"strings"
	"errors"
	"unicode"
)

const CHARACTER_NAME_MAX int = 32
const CHARACTER_NAME_MIN_LETTERS int = 2

var CHARACTER_NAME_BANNED_KEYWORDS = []string {
	"self",
	"in",
	"at",
	"on",
	"from",
	"all",
	"friends",
}

func CharacterNameValidate(world *World, name string) (string, error) {
	if len(name) > CHARACTER_NAME_MAX {
		return "", fmt.Errorf("Character names must be no more than %d characters.", CHARACTER_NAME_MAX)
	}

	nameTrimmed := strings.TrimSpace(name)
	nameTrimmedParts := strings.Fields(name)
	nameTrimmed = strings.Join(nameTrimmedParts, " ")

	// Check keywords
	for _, keyword := range CHARACTER_NAME_BANNED_KEYWORDS {
		for _, part := range nameTrimmedParts {
			if strings.EqualFold(keyword, part) {
				return "", fmt.Errorf("Your name cannot contain the word '%s'. It is a reserved keyword.", keyword)
			}
		}
	}

	// Count letters
	letterCount := 0
	for _, c := range nameTrimmed {
		if c == ' ' {
			continue
		}

		if !unicode.IsLetter(c) {
			return "", errors.New("Your name can only contain letters and spaces.")
		}

		letterCount++
	}

	if letterCount < CHARACTER_NAME_MIN_LETTERS {
		return "", fmt.Errorf("You name must contain at least %d letters.", CHARACTER_NAME_MIN_LETTERS)
	}

	_, nameIsTaken := world.GetCharacterIfExists(nameTrimmed)
	if nameIsTaken {
		return "", fmt.Errorf("A character named '%s' already exists.", nameTrimmed)
	}

	return nameTrimmed, nil
}
