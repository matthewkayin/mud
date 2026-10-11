package world

import (
	"strings"
	"testing"

	"github.com/mmcdole/lunar"
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

// Pushes a mob with an attached NPC and returns the NPC
func scriptTestPushNpc(world *World, name string) *Npc {
	npc := &Npc{
		Id: TEST_NPC_GOBLIN,
		mode: NPC_MODE_IDLE,
		disposition: NPC_DISPOSITION_NEUTRAL,
	}
	mob := scriptTestMob(name)
	mob.Npc = npc
	npc.mobHandle = world.Mobs.Push(mob)
	return npc
}

func TestScriptNpcState(t *testing.T) {
	primeTestData()
	world := scriptTestWorld(t)
	npc := scriptTestPushNpc(world, "Goblin")
	scriptTestSetHandle(t, world, "goblin", npc.mobHandle)
	scriptTestSetHandle(t, world, "player", world.Mobs.Push(scriptTestMob("Bufo")))

	cases := []struct {
		source string
		expected string
	}{
		{ "return world.mob_is_npc(goblin)", "true" },
		{ "return world.mob_is_npc(player)", "false" },
		{ "return world.npc_get_mode(goblin)", "Idle" },
		{ "world.npc_set_mode(goblin, world.NpcMode.AGGRO) return world.npc_get_mode(goblin)", "Aggro" },
		{ "world.npc_set_mode(goblin, world.NpcMode.IDLE) return world.npc_get_mode(goblin)", "Idle" },
		{ "world.npc_set_mode(goblin, world.NpcMode.SURPRISE) return world.npc_get_mode(goblin)", "Surprise" },
		{ "return world.npc_get_disposition(goblin)", "Neutral" },
		{ "world.npc_set_disposition(goblin, world.NpcDisposition.FRIENDLY) return world.npc_get_disposition(goblin)", "Friendly" },
	}
	for _, testCase := range cases {
		if result := scriptTestEval(t, world, testCase.source); result != testCase.expected {
			t.Errorf("%s: expected %s, got %s", testCase.source, testCase.expected, result)
		}
	}

	// Setting surprise goes through setModeSurprise, which alerts the mob
	if alertness := world.Mobs.Get(npc.mobHandle).alertness; alertness != MOB_ALERTNESS_MAX {
		t.Errorf("Expected surprised NPC to have max alertness, got %f", alertness)
	}
}

func TestScriptNpcErrors(t *testing.T) {
	primeTestData()
	world := scriptTestWorld(t)
	npc := scriptTestPushNpc(world, "Goblin")
	scriptTestSetHandle(t, world, "goblin", npc.mobHandle)
	scriptTestSetHandle(t, world, "player", world.Mobs.Push(scriptTestMob("Bufo")))

	cases := []struct {
		source string
		expectedError string
	}{
		{ "return world.npc_get_mode(player)", "is not an NPC" },
		{ "world.npc_set_disposition(player, world.NpcDisposition.HOSTILE)", "is not an NPC" },
		{ "world.npc_set_mode(goblin, \"Dead\")", "not a valid NPC mode" },
		{ "world.npc_set_mode(goblin, \"Asleep\")", "not a valid NPC mode" },
		{ "world.npc_set_disposition(goblin, \"Angry\")", "not a valid NPC disposition" },
	}
	for _, testCase := range cases {
		_, err := world.luaState.DoString("@test.lua", testCase.source)
		if err == nil || !strings.Contains(err.Error(), testCase.expectedError) {
			t.Errorf("%s: expected error containing %q, got %v", testCase.source, testCase.expectedError, err)
		}
	}
}

func TestNpcHookArguments(t *testing.T) {
	primeTestData()
	world := scriptTestWorld(t)

	hooks, err := world.luaState.DoString("@test.lua", `
		return {
			init = function(self, params)
				return { self = self, greeting = params.Greeting }
			end,
			get_status_description = function(self, instance)
				return instance.greeting .. " " .. world.mob_get_name(self) .. " " .. tostring(instance.self == self)
			end,
		}
	`)
	if err != nil {
		t.Fatalf("Error loading hooks: %s", err.Error())
	}
	hooksTable, _ := hooks[0].AsTable()
	parser := ScriptParser{}
	npcData := NPC_DATA[TEST_NPC_GOBLIN]
	npcData.BehaviorParams = map[string]NpcBehaviorParamType{ "Greeting": NPC_BEHAVIOR_PARAM_TYPE_STRING }
	npcData.init = parser.getOptionalFunction(hooksTable, "init")
	npcData.getStatusDescription = parser.getOptionalFunction(hooksTable, "get_status_description")
	if parser.getError() != nil {
		t.Fatalf("Error parsing hooks: %s", parser.getError().Error())
	}

	npc := scriptTestPushNpc(world, "Goblin")
	npc.BehaviorParams = map[string]any{ "Greeting": "Hello" }
	npc.callInit(world)
	if npc.instance.Kind() != lua.TableKind {
		t.Fatalf("Expected init() to set the instance table, got %s", npc.instance.Kind().String())
	}

	statusDescription, ok := npc.GetStatusDescription(world)
	if !ok || statusDescription != "Hello Goblin true" {
		t.Errorf("Expected status description %q, got %q (ok = %t)", "Hello Goblin true", statusDescription, ok)
	}
}

func TestScriptInventory(t *testing.T) {
	primeTestData()
	ITEM_DATA[TEST_ITEM_SWORD].Data.(*ItemDataWeapon).MaxDurability = 20

	world := scriptTestWorld(t)
	world.Rooms = []Room{ {}, {} }

	npcMob := scriptTestMob("Bufo")
	npcMob.Data.Room = 1
	npcMob.Data.Inventory.Items = []Item {
		{ Id: TEST_ITEM_GOLD, Amount: 30 },
		{ Id: TEST_ITEM_SWORD, Amount: 1, Durability: 12 },
		{ Id: TEST_ITEM_POTION_HEALTH, Amount: 50 },
		{ Id: TEST_ITEM_AXE, Amount: 1 },
	}
	npcHandle := world.Mobs.Push(npcMob)
	scriptTestSetHandle(t, world, "a", npcHandle)

	playerMob := scriptTestMob("Player")
	playerMob.PlayerCharacter = &Character{}
	playerMob.Data.Inventory.Items = []Item {
		{ Id: TEST_ITEM_POTION_HEALTH, Amount: 5 },
	}
	playerHandle := world.Mobs.Push(playerMob)
	scriptTestSetHandle(t, world, "p", playerHandle)

	cases := []struct {
		source string
		expected string
	}{
		{ "return world.mob_get_inventory_length(a)", "4" },
		{ "return world.mob_get_inventory_item_name(a, 1)", "Sword" },
		{ "return world.mob_get_inventory_item_amount(a, 0)", "30" },
		{ "return world.mob_get_inventory_item_durability(a, 1)", "12" },
		{ "return world.mob_get_inventory_item_durability(a, 0)", "nil" },
		{ "return world.mob_get_inventory_amount_of(a, \"Potion of Health\")", "50" },
		{ "return world.mob_get_inventory_amount_of(a, \"Spellbook of Firebolt\")", "0" },
		// Drop part of a stack, then more than the rest of it
		{ "return world.mob_drop_inventory_item(a, 0, 10)", "10" },
		{ "return world.mob_get_inventory_item_amount(a, 0)", "20" },
		{ "return world.mob_drop_inventory_item(a, 0, 100)", "20" },
		// The axe moved from the last index into the removed gold stack's index
		{ "return world.mob_get_inventory_item_name(a, 0)", "Axe" },
		// The player has 100 - 5 * 5 = 75 space left, which fits 15 potions
		{ "return world.mob_give_inventory_item(a, 2, p, 40)", "15" },
		{ "return world.mob_get_inventory_amount_of(a, \"Potion of Health\")", "35" },
		{ "return world.mob_get_inventory_amount_of(p, \"Potion of Health\")", "20" },
		{ "return world.mob_give_inventory_item(a, 2, p, 1)", "0" },
		{ "return world.mob_give_inventory_item(p, 0, a, 20)", "20" },
		{ "return world.mob_get_inventory_length(p)", "0" },
	}
	for _, testCase := range cases {
		if result := scriptTestEval(t, world, testCase.source); result != testCase.expected {
			t.Errorf("%s: expected %s, got %s", testCase.source, testCase.expected, result)
		}
	}

	roomInventory := &world.Rooms[1].Inventory
	if roomInventory.Length() != 1 || roomInventory.Items[0].Id != TEST_ITEM_GOLD || roomInventory.Items[0].Amount != 30 {
		t.Errorf("expected 30 gold in room 1, got %+v", roomInventory.Items)
	}

	invalidCases := []struct {
		source string
		expectedError string
	}{
		{ "return world.mob_get_inventory_item_name(a, 3)", "out of range" },
		{ "return world.mob_get_inventory_item_name(a, -1)", "out of range" },
		{ "return world.mob_get_inventory_item_name(a, 0.5)", "must be an integer" },
		{ "return world.mob_get_inventory_amount_of(a, \"Nothing\")", "not a valid item" },
		{ "return world.mob_drop_inventory_item(a, 0, 0)", "must be positive" },
		{ "return world.mob_give_inventory_item(a, 0, p, 2)", "does not stack" },
		{ "return world.mob_give_inventory_item(a, 0, a, 1)", "different mob" },
	}
	for _, testCase := range invalidCases {
		_, err := world.luaState.DoString("@test.lua", testCase.source)
		if err == nil || !strings.Contains(err.Error(), testCase.expectedError) {
			t.Errorf("%s: expected error containing %q, got %v", testCase.source, testCase.expectedError, err)
		}
	}
}
