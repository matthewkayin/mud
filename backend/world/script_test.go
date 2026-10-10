package world

import (
	"strings"
	"testing"
)

func scriptTestWorld(t *testing.T) *World {
	world := &World{
		Mobs: MobArrayInit(),
	}
	world.scriptInit(t.TempDir())
	t.Cleanup(world.ScriptQuit)
	return world
}

func scriptTestMob(name string) Mob {
	return MobInit(&MobData{
		Name: name,
		Equipment: EquipmentInitEmpty(),
	})
}

func scriptTestSetHandle(t *testing.T, world *World, name string, handle MobHandle) {
	value, err := world.Mobs.Get(handle).getLuaHandle(world)
	if err != nil {
		t.Fatalf("getLuaHandle: %s", err.Error())
	}
	if err := world.luaState.SetGlobal(name, value); err != nil {
		t.Fatalf("SetGlobal %s: %s", name, err.Error())
	}
}

func scriptTestEval(t *testing.T, world *World, source string) string {
	results, err := world.luaState.DoString("@test.lua", source)
	if err != nil {
		t.Fatalf("%s: %s", source, err.Error())
	}
	if len(results) != 1 {
		t.Fatalf("%s: expected 1 result, got %d", source, len(results))
	}
	return results[0].String()
}

func TestScriptMobHandle(t *testing.T) {
	world := scriptTestWorld(t)
	bufoHandle := world.Mobs.Push(scriptTestMob("Bufo"))
	hodorHandle := world.Mobs.Push(scriptTestMob("Hodor"))

	scriptTestSetHandle(t, world, "a", bufoHandle)
	scriptTestSetHandle(t, world, "a2", bufoHandle)
	scriptTestSetHandle(t, world, "b", hodorHandle)

	cases := []struct {
		source string
		expected string
	}{
		{ "return world.mob_exists(a)", "true" },
		{ "return world.get_mob_data(a, { \"name\" }).name", "Bufo" },
		{ "return world.get_mob_data(b, { \"name\" }).name", "Hodor" },
		{ "return rawequal(a, a2)", "true" },
		{ "return a == a2", "true" },
		{ "return a == b", "false" },
		{ "return tostring(b)", "MobHandle(1:0)" },
	}
	for _, testCase := range cases {
		if result := scriptTestEval(t, world, testCase.source); result != testCase.expected {
			t.Errorf("%s: expected %s, got %s", testCase.source, testCase.expected, result)
		}
	}
}

func TestScriptMobHandleStale(t *testing.T) {
	world := scriptTestWorld(t)
	bufoHandle := world.Mobs.Push(scriptTestMob("Bufo"))
	scriptTestSetHandle(t, world, "a", bufoHandle)

	world.Mobs.Remove(bufoHandle)
	if result := scriptTestEval(t, world, "return world.mob_exists(a)"); result != "false" {
		t.Errorf("expected removed mob to not exist, got %s", result)
	}

	_, err := world.luaState.DoString("@test.lua", "return world.get_mob_data(a, { \"name\" })")
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected stale handle error, got %v", err)
	}

	// A new mob reusing the ID gets a different handle
	hodorHandle := world.Mobs.Push(scriptTestMob("Hodor"))
	scriptTestSetHandle(t, world, "b", hodorHandle)
	if result := scriptTestEval(t, world, "return a == b"); result != "false" {
		t.Errorf("expected stale handle to differ from reused ID, got %s", result)
	}
}

func TestScriptMobHandleRejectsTable(t *testing.T) {
	world := scriptTestWorld(t)
	world.Mobs.Push(scriptTestMob("Bufo"))

	_, err := world.luaState.DoString("@test.lua", "return world.mob_exists({ id = 0, generation = 0 })")
	if err == nil || !strings.Contains(err.Error(), "MobHandle expected") {
		t.Errorf("expected MobHandle argument error, got %v", err)
	}
}
