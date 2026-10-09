package world

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestNpcJsonRoundTrip(t *testing.T) {
	primeTestData()

	npc := Npc {
		Id: TEST_NPC_TROLL,
		SpawnRoom: 1,
		LevelRange: Int32Range{ Min: 2, Max: 3 },
		MovementTypeOverride: NPC_MOVEMENT_TYPE_WANDER,
		DropTableOverride: DropTable {
			Entries: []DropTableEntry {
				{
					ItemId: TEST_ITEM_GOLD,
					AmountRange: Int32Range{ Min: 1, Max: 5 },
					DurabilityPercentRange: Int32Range{ Min: 100, Max: 100 },
					DropChancePercent: 50,
				},
			},
		},
		BehaviorParams: testTrollBehaviorParams(),
	}

	data, err := json.Marshal(&npc)
	if err != nil {
		t.Fatalf("Error marshalling NPC: %s", err.Error())
	}

	// Saved data refers to NPCs, items and directions by name
	for _, expected := range []string{ `"Id":"troll"`, `"MovementTypeOverride":"Wander"`, `"Id":"Gold"`, `"ExitToBlock":"east"` } {
		if !strings.Contains(string(data), expected) {
			t.Errorf("Expected NPC JSON to contain %s, got %s", expected, data)
		}
	}

	var loadedNpc Npc
	err = json.Unmarshal(data, &loadedNpc)
	if err != nil {
		t.Fatalf("Error unmarshalling NPC: %s", err.Error())
	}
	if !reflect.DeepEqual(npc, loadedNpc) {
		t.Errorf("Expected NPC %+v after round trip, got %+v", npc, loadedNpc)
	}
}

func TestNpcJsonNoMovementTypeOverride(t *testing.T) {
	primeTestData()

	var npc Npc
	err := json.Unmarshal([]byte(`{ "Id": "goblin", "MovementTypeOverride": "", "DropTableOverride": { "Entries": [] }, "BehaviorParams": {} }`), &npc)
	if err != nil {
		t.Fatalf("Error unmarshalling NPC: %s", err.Error())
	}
	if npc.MovementTypeOverride != NPC_MOVEMENT_TYPE_OVERRIDE_NONE {
		t.Errorf("Expected no movement type override, got %d", npc.MovementTypeOverride)
	}
	if npc.getMovementType() != NPC_MOVEMENT_TYPE_WANDER {
		t.Errorf("Expected the goblin's movement type to come from its NPC data, got %s", npc.getMovementType().String())
	}
}

func TestNpcJsonErrors(t *testing.T) {
	tests := []struct {
		name string
		json string
		expectedError string
	}{
		{ "unknown id", `{ "Id": "dragon" }`, "No NPC id matches 'dragon'" },
		{ "unknown movement type", `{ "Id": "goblin", "MovementTypeOverride": "Fly" }`, "No NPC movement type matches 'Fly'" },
		{ "undeclared param", `{ "Id": "goblin", "BehaviorParams": { "Toll": 1 } }`, "no behavior param named 'Toll'" },
		{ "wrong param type", `{ "Id": "troll", "BehaviorParams": { "Patience": "three" } }`, "behavior param 'Patience'" },
		{ "unknown item", `{ "Id": "troll", "BehaviorParams": { "Toll": { "Id": "Diamond", "Amount": 1 } } }`, "No item ID matches 'Diamond'" },
		{ "invalid direction", `{ "Id": "troll", "BehaviorParams": { "ExitToBlock": "up" } }`, "'up' is not a valid direction" },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			primeTestData()

			var npc Npc
			err := json.Unmarshal([]byte(test.json), &npc)
			if err == nil || !strings.Contains(err.Error(), test.expectedError) {
				t.Errorf("Expected an error containing %q, got %v", test.expectedError, err)
			}
		})
	}
}
