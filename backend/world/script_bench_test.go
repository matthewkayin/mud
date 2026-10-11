package world

import (
	"fmt"
	"testing"

	"github.com/mmcdole/lunar"
)

// Compares two designs for reading mob state from scripts. The table design
// (bench_get_mob_table) takes a table of field names and returns a table of
// values, which is how the script API used to work. The getter design is the
// world.mob_get_* functions in SCRIPT_LIBRARY. See docs/backend/scripting.md.

// Test-only copy of the table-returning design, limited to scalar fields
func benchGetMobTable(frame lua.Frame) lua.Outcome {
	mob := frameGetMobArg(&frame, 0)

	// Get requested fields from args
	fieldsTable, ok := frame.Table(1)
	if !ok {
		frame.ThrowArgTypeError(1, lua.TableKind)
	}

	// Fill out requested data
	fieldCount := fieldsTable.RawLen()
	returnTable, err := frame.State().NewTableWithCapacity(0, fieldCount)
	if err != nil {
		frame.ThrowError(err)
	}
	for index := 1; index <= fieldCount; index++ {
		field, ok := fieldsTable.RawGetInt(index).AsString()
		if !ok {
			frame.ThrowArgError(1, "Provided data fields must be strings.")
		}

		switch field {
			case "name":
				returnTable.RawSetString(field, lua.String(mob.Data.Name))
			case "room":
				returnTable.RawSetString(field, lua.Number(float64(mob.Data.Room)))
			case "health":
				returnTable.RawSetString(field, lua.Number(float64(mob.Data.Health)))
			case "mana":
				returnTable.RawSetString(field, lua.Number(float64(mob.Data.Mana)))
			case "max_health":
				returnTable.RawSetString(field, lua.Number(float64(mob.Data.MaxHealth())))
			default:
				frame.ThrowArgError(1, fmt.Sprintf("Field '%s' is not a valid mob field.", field))
		}
	}

	return frame.ReturnValue(returnTable.Value())
}

// Runs body b.N times inside a single Lua call, with a handle to a mob bound to `a`
// and the total of the values read accumulated into `acc`
func benchScriptMobData(b *testing.B, body string) {
	world := &World {
		Mobs: MobArrayInit(),
	}
	world.scriptInit(b.TempDir())
	b.Cleanup(world.ScriptQuit)

	mob := MobInit(&MobData {
		Name: "Bufo",
		Equipment: EquipmentInitEmpty(),
		Room: 3,
		Health: 36,
		Mana: 12,
	})
	handle := world.Mobs.Push(mob)
	handleValue, err := world.getMobLuaHandle(handle)
	if err != nil {
		b.Fatal(err)
	}
	world.luaState.SetGlobal("a", handleValue)

	tableFunc, err := world.luaState.NewNativeFunction(benchGetMobTable)
	if err != nil {
		b.Fatal(err)
	}
	world.luaState.SetGlobal("bench_get_mob_table", tableFunc.Value())

	source := fmt.Sprintf(`
		local fields_2 = { "name", "room" }
		return function(n)
			local acc = 0
			for i = 1, n do
				%s
			end
			return acc
		end`, body)
	results, err := world.luaState.DoString("@bench.lua", source)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	if _, err := world.luaState.Call(results[0], lua.Number(float64(b.N))); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkScriptMobDataTable1(b *testing.B) {
	benchScriptMobData(b, `
		local mob = bench_get_mob_table(a, { "room" })
		acc = acc + mob.room`)
}

func BenchmarkScriptMobDataGetters1(b *testing.B) {
	benchScriptMobData(b, `
		acc = acc + world.mob_get_room(a)`)
}

func BenchmarkScriptMobDataTable2(b *testing.B) {
	benchScriptMobData(b, `
		local mob = bench_get_mob_table(a, { "name", "room" })
		acc = acc + mob.room + #mob.name`)
}

// The fields table is built once outside the loop, so only the result table is allocated per call
func BenchmarkScriptMobDataTable2ReusedFields(b *testing.B) {
	benchScriptMobData(b, `
		local mob = bench_get_mob_table(a, fields_2)
		acc = acc + mob.room + #mob.name`)
}

func BenchmarkScriptMobDataGetters2(b *testing.B) {
	benchScriptMobData(b, `
		acc = acc + world.mob_get_room(a) + #world.mob_get_name(a)`)
}

func BenchmarkScriptMobDataTable5(b *testing.B) {
	benchScriptMobData(b, `
		local mob = bench_get_mob_table(a, { "name", "room", "health", "mana", "max_health" })
		acc = acc + mob.room + mob.health + mob.mana + mob.max_health + #mob.name`)
}

func BenchmarkScriptMobDataGetters5(b *testing.B) {
	benchScriptMobData(b, `
		acc = acc + world.mob_get_room(a) + world.mob_get_health(a) + world.mob_get_mana(a)
			+ world.mob_get_max_health(a) + #world.mob_get_name(a)`)
}
