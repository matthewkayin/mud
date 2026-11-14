package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mud/world"
	"regexp"
	"unicode/utf16"
)

// Returns the range a problem's diagnostic covers in the source text of its file
func problemRange(problem world.ScriptProblem, source []byte) Range {
	// Lua errors know their line
	if problem.Line > 0 {
		return lineRange(source, problem.Line - 1)
	}

	// Field problems go on the key
	if problem.Key != "" {
		keyRange, found := keyRange(source, problem.Key)
		if found {
			return keyRange
		}
	}

	// Layout room problems go on the room's entry
	if problem.RoomIndex >= 0 {
		line, found := layoutRoomLine(source, problem.RoomIndex)
		if found {
			return lineRange(source, line)
		}
	}

	return lineRange(source, 0)
}

// Returns the range of the first place the key is written as a table key: key =, ["key"] = or ['key'] =
func keyRange(source []byte, key string) (Range, bool) {
	quotedKey := regexp.QuoteMeta(key)
	pattern := fmt.Sprintf(`(?:^|[^\w])(%s|\["%s"\]|\['%s'\])\s*=(?:[^=]|$)`, quotedKey, quotedKey, quotedKey)
	keyRegexp := regexp.MustCompile(pattern)

	match := keyRegexp.FindSubmatchIndex(source)
	if match == nil {
		return Range{}, false
	}

	return Range {
		Start: offsetToPosition(source, match[2]),
		End: offsetToPosition(source, match[3]),
	}, true
}

// Returns the 0-based line of the room entry at roomIndex in layout.json's rooms array
func layoutRoomLine(source []byte, roomIndex int) (int, bool) {
	decoder := json.NewDecoder(bytes.NewReader(source))

	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return 0, false
	}

	// Find the rooms key, skipping the values of other keys
	for {
		if !decoder.More() {
			return 0, false
		}
		token, err = decoder.Token()
		if err != nil {
			return 0, false
		}
		if token == "rooms" {
			break
		}
		var skippedValue json.RawMessage
		err = decoder.Decode(&skippedValue)
		if err != nil {
			return 0, false
		}
	}

	token, err = decoder.Token()
	if err != nil || token != json.Delim('[') {
		return 0, false
	}

	for index := 0; decoder.More(); index++ {
		// The offset is just past the previous token, so skip to the start of this entry
		offset := int(decoder.InputOffset())
		for offset < len(source) && isJsonSpaceOrComma(source[offset]) {
			offset++
		}
		if index == roomIndex {
			return offsetToPosition(source, offset).Line, true
		}

		var skippedRoom json.RawMessage
		err = decoder.Decode(&skippedRoom)
		if err != nil {
			return 0, false
		}
	}

	return 0, false
}

func isJsonSpaceOrComma(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == ','
}

// Returns the range covering the whole of a 0-based line. A line past the end of the source covers nothing
func lineRange(source []byte, line int) Range {
	lines := bytes.Split(source, []byte("\n"))
	if line >= len(lines) {
		return Range {
			Start: Position { Line: line, Character: 0 },
			End: Position { Line: line, Character: 0 },
		}
	}

	lineText := bytes.TrimRight(lines[line], "\r")
	return Range {
		Start: Position { Line: line, Character: 0 },
		End: Position { Line: line, Character: utf16Length(lineText) },
	}
}

// Converts a byte offset into the source into a line and UTF-16 column
func offsetToPosition(source []byte, offset int) Position {
	before := source[:offset]
	line := bytes.Count(before, []byte("\n"))
	lineStart := bytes.LastIndexByte(before, '\n') + 1
	return Position {
		Line: line,
		Character: utf16Length(before[lineStart:]),
	}
}

func utf16Length(text []byte) int {
	return len(utf16.Encode([]rune(string(text))))
}
