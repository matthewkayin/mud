package game

import (
	"testing"
)

func TestRemoval(t *testing.T) {
	mobArray := MobArrayInit()
	a := mobArray.Push(Mob {
		Data: MobData {
			Name: "A",
		},
	})
	b := mobArray.Push(Mob {
		Data: MobData {
			Name: "B",
		},
	})

	if len(mobArray.data) != 2 {
		t.Errorf("Mob array data length not 2")
	}
	if len(mobArray.idToIndex) != 2 {
		t.Errorf("Mob array ID to index length not 2")
	}

	mobArray.Remove(a)

	_, aExists := mobArray.GetIfExists(a)
	if aExists {
		t.Errorf("Mob A exists, but it was removed!")
	}

	_, bExists := mobArray.GetIfExists(b)
	if !bExists {
		t.Errorf("Mob B does not exist, but it should!")
	}

	c := mobArray.Push(Mob {
		Data: MobData {
			Name: "C",
		},
	})
	_, cExists := mobArray.GetIfExists(c)
	if !cExists {
		t.Errorf("Mob C does not exist, but it should!")
	}

	cIndex := mobArray.idToIndex[c.id].index
	if c.id != 0 || c.generation != 1 || cIndex != 1 {
		t.Errorf("Mob C exists, but isn't stored in the array correctly!")
	}
}
