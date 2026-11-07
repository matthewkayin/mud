package util

import (
	"strings"
)

func CombineNames(names []string) string {
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

// Returns the first sentence inside str which contains substr
// This function ignores case sensitivity, but it is not unicode safe
func FindSentenceContainingSubstr(str string, substr string) (string, bool) {
	substrLen := len(substr)
	if substrLen == 0 {
		return "", false
	}

	strLen := len(str)
	if strLen < substrLen {
		return "", false
	}

	sentenceStart := 0
	firstSubstrChar := charToLower(substr[0])
	for i := 0; i <= strLen - substrLen; i++ {
		// Check if we have reached a new sentence
		// A new sentence started after a punctuation mark and a space
		if i > 1 && str[i - 1] == ' ' && isTerminalPunctuation(str[i - 2]) {
			sentenceStart = i
		}

		// Check if we found the first char of the substr
		if charToLower(str[i]) != firstSubstrChar {
			continue
		}

		// Once we found the first char, check for the rest of the substr
		match := true
		for j := 1; j < substrLen; j++ {
			if charToLower(str[i + j]) != charToLower(substr[j]) {
				match = false
				break
			}
		}

		if match {
			// Find the end of the sentence (exclusive), which will be the end of the string by default
			sentenceEnd := strLen
			for j := i + substrLen + 1; j < strLen; j++ {
				if str[j] == ' ' && isTerminalPunctuation(str[j - 1]) {
					sentenceEnd = j
					break
				}
			}

			return str[sentenceStart:sentenceEnd], true
		}
	}

	return "", false
}

func charToLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}

func isTerminalPunctuation(b byte) bool {
	switch b {
		case '.', '!', '?':
			return true
		default:
			return false
	}
}
