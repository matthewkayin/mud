package world

import (
	"testing"
	"mud/bitset"
)

// Three rooms in a row along X, with one race, class and job for characters to reference
func characterTestWorld() *World {
	primeTestData()
	RACE_DATA = []*RaceData { { Name: "Human" } }
	RACE_NAME_TO_ID = map[string]RaceId { "Human": 0 }
	CLASS_DATA = []*ClassData { { Name: "Wizard" } }
	CLASS_NAME_TO_ID = map[string]ClassId { "Wizard": 0 }
	JOB_DATA = []*JobData { { Name: "Farmer" } }
	JOB_NAME_TO_ID = map[string]JobId { "Farmer": 0 }

	world := &World {
		Rooms: []Room {
			validateTestRoom("Start", 0, 0),
			validateTestRoom("Middle", 1, 0),
			validateTestRoom("End", 2, 0),
		},
	}
	world.initRoomPositionLookup()

	return world
}

func characterTestCharacter() *Character {
	character := &Character {
		SpellsEquipped: map[SpellId]*CharacterEquippedSpell{},
		RoomsDiscovered: bitset.New(WORLD_MAX_ROOMS),
		Data: MobData {
			Name: "Tester",
			Room: 2,
			Conditions: []Condition{},
			Spells: []SpellId{},
		},
	}
	bitset.Set(character.RoomsDiscovered, 0, true)
	bitset.Set(character.RoomsDiscovered, 2, true)

	return character
}

func TestCharacterJsonRoundTrip(t *testing.T) {
	world := characterTestWorld()

	characterJson := world.characterToJson(characterTestCharacter())
	if characterJson.Data.Room != (RoomPosition{ X: 2, Y: 0 }) {
		t.Errorf("Expected the room to be saved as (2, 0, 0), got %v", characterJson.Data.Room)
	}
	if len(characterJson.RoomsDiscovered) != 2 {
		t.Errorf("Expected 2 discovered rooms to be saved, got %v", characterJson.RoomsDiscovered)
	}

	loadedCharacter, err := world.characterFromJson(&characterJson)
	if err != nil {
		t.Fatalf("Expected the character to load, got %s", err.Error())
	}
	if loadedCharacter.Data.Room != 2 {
		t.Errorf("Expected the character to load in room 2, got %d", loadedCharacter.Data.Room)
	}
	for roomIndex, isExpectedDiscovered := range []bool { true, false, true } {
		if bitset.Check(loadedCharacter.RoomsDiscovered, roomIndex) != isExpectedDiscovered {
			t.Errorf("Expected room %d discovered to be %t", roomIndex, isExpectedDiscovered)
		}
	}
}

// Saves reference rooms by position, so a character comes back in the same room even if room indices change
func TestCharacterJsonSurvivesRoomReorder(t *testing.T) {
	world := characterTestWorld()
	characterJson := world.characterToJson(characterTestCharacter())

	world.Rooms[0], world.Rooms[2] = world.Rooms[2], world.Rooms[0]
	world.initRoomPositionLookup()

	loadedCharacter, err := world.characterFromJson(&characterJson)
	if err != nil {
		t.Fatalf("Expected the character to load, got %s", err.Error())
	}
	if world.Rooms[loadedCharacter.Data.Room].Name != "End" {
		t.Errorf("Expected the character to load in the End room, got %s", world.Rooms[loadedCharacter.Data.Room].Name)
	}
}

func TestCharacterJsonMissingRoom(t *testing.T) {
	world := characterTestWorld()
	characterJson := world.characterToJson(characterTestCharacter())

	// Remove the End room, which the character is in and has discovered
	world.Rooms = world.Rooms[:2]
	world.initRoomPositionLookup()

	loadedCharacter, err := world.characterFromJson(&characterJson)
	if err != nil {
		t.Fatalf("Expected the character to load, got %s", err.Error())
	}
	if loadedCharacter.Data.Room != WORLD_START_ROOM {
		t.Errorf("Expected the character to fall back to the start room, got %d", loadedCharacter.Data.Room)
	}
	if !bitset.Check(loadedCharacter.RoomsDiscovered, 0) {
		t.Errorf("Expected the start room to still be discovered")
	}
	if bitset.Check(loadedCharacter.RoomsDiscovered, 2) {
		t.Errorf("Expected the removed room not to be discovered")
	}
}
