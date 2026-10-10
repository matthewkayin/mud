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
	value, err := world.getMobLuaHandle(handle)
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
		{ "return world.mob_get_name(a)", "Bufo" },
		{ "return world.mob_get_name(b)", "Hodor" },
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

	_, err := world.luaState.DoString("@test.lua", "return world.mob_get_name(a)")
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

func TestScriptMobGetters(t *testing.T) {
	world := scriptTestWorld(t)
	mob := scriptTestMob("Bufo")
	mob.Data.Room = 3
	mob.Data.Level = 4
	mob.Data.Health = 36
	mob.Data.Mana = 12
	mob.Data.Stats.Values[STAT_VIT] = 10
	mob.Data.Stats.Values[STAT_INT] = 8
	mob.Data.Stats.Values[STAT_STR] = 7
	scriptTestSetHandle(t, world, "a", world.Mobs.Push(mob))

	cases := []struct {
		source string
		expected string
	}{
		{ "return world.mob_get_name(a)", "Bufo" },
		{ "return world.mob_get_room(a)", "3" },
		{ "return world.mob_get_level(a)", "4" },
		{ "return world.mob_get_health(a)", "36" },
		{ "return world.mob_get_max_health(a)", "50" },
		{ "return world.mob_get_mana(a)", "12" },
		{ "return world.mob_get_max_mana(a)", "40" },
		{ "return world.mob_get_stat(a, world.Stat.STR)", "7" },
		{ "return world.mob_get_equipment(a, world.EquipmentSlot.MAIN_HAND)", "nil" },
	}
	for _, testCase := range cases {
		if result := scriptTestEval(t, world, testCase.source); result != testCase.expected {
			t.Errorf("%s: expected %s, got %s", testCase.source, testCase.expected, result)
		}
	}

	invalidCases := []struct {
		source string
		expectedError string
	}{
		{ "return world.mob_get_stat(a, \"XYZ\")", "not a valid stat" },
		{ "return world.mob_get_equipment(a, \"Nowhere\")", "not a valid equipment slot" },
	}
	for _, testCase := range invalidCases {
		_, err := world.luaState.DoString("@test.lua", testCase.source)
		if err == nil || !strings.Contains(err.Error(), testCase.expectedError) {
			t.Errorf("%s: expected error containing %q, got %v", testCase.source, testCase.expectedError, err)
		}
	}
}
