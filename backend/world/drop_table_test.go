package world

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDropTableEntryJsonUsesItemName(t *testing.T) {
	primeTestData()

	dropTable := DropTable {
		Entries: []DropTableEntry {
			{
				ItemId: TEST_ITEM_GOLD,
				AmountRange: Int32Range{ Min: 1, Max: 5 },
				DurabilityPercentRange: Int32Range{ Min: 100, Max: 100 },
				DropChancePercent: 50,
			},
		},
	}

	data, err := json.Marshal(&dropTable)
	if err != nil {
		t.Fatalf("Error encoding drop table: %s", err.Error())
	}
	if !strings.Contains(string(data), `"ItemId":"Gold"`) {
		t.Errorf("Expected the entry to be saved by item name, got %s", data)
	}

	loadedDropTable := DropTable{}
	err = json.Unmarshal(data, &loadedDropTable)
	if err != nil {
		t.Fatalf("Error decoding drop table: %s", err.Error())
	}
	if len(loadedDropTable.Entries) != 1 || loadedDropTable.Entries[0] != dropTable.Entries[0] {
		t.Errorf("Drop table changed after a round trip.\nExpected: %+v\nActual: %+v", dropTable, loadedDropTable)
	}
}

func TestDropTableEntryJsonWithUnknownItem(t *testing.T) {
	primeTestData()

	entry := DropTableEntry{}
	err := json.Unmarshal([]byte(`{ "ItemId": "Not An Item", "DropChancePercent": 50 }`), &entry)
	if err == nil {
		t.Errorf("Expected an unknown item name to be an error")
	}
}

func TestItemJsonWithUnknownItem(t *testing.T) {
	primeTestData()

	item := Item{}
	err := json.Unmarshal([]byte(`{ "Id": "Not An Item", "Amount": 1 }`), &item)
	if err == nil {
		t.Errorf("Expected an unknown item name to be an error")
	}
}
