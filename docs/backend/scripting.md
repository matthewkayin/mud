# Backend Scripting

The backend loads Lua scripts and uses them both for definitions of game items and behaviors. Examples include item definitions, spell definitions, spell behavior, and NPC behavior. The goal of this design is to separate the engine from the data and to allow for flexible NPC behavior scripting in a way that is difficult using just Go.

## Requirements

1. Lua scripts as data - marshalling Lua tables into Go structs

Each script should represent a piece of data. Lua DoFile and DoString both return a result. All in game data scripts should return a table, and the backend should parse this table into the repsective data type.

Example using spells. Suppose the backend spell type is:

``` 
type SpellData struct {
  Name string
  Description string
  CastsToLearn int32
  
  ManaCost int32
  CastTime int32
  CanTargetPlayers bool

  OnHit func(world *World, caster *Mob, target *Mob)
}
```

It should be posible to write a Lua script to create an instance of SpellData:

```
-- firebolt.lua

local spell = {}

spell.Name = "Firebolt"
spell.Description = "Casts a bolt of fire toward the target"
spell.CastsToLearn = 50

spell.ManaCost = 5
spell.CastTime = 1
spell.CanTargetPlayers = false

spell.OnHit = function(caster, target)
  world.messageRoom(caster.room, "{target} took {damage} damage from the firebolt", {
    target = target.name,
    damage = 5
  })
end

-- Script returns the spell table to the host program
return spell
```

Note that in the above example, `world` refers to a global table exposed by the host program (the backend) which exposes an API that can be used by scripts to manipulate the world state.

To make this possible, there needs to be some way to marshal a Lua table into a Go type, even if this is done by manually assigning fields (which may be necessary especially in the case of ItemData which has an any-typed field whose type depends on which kind of item it is).

2. Validation - Runtime marshalling of all data

Once data has been written out into Lua scripts, the world will need to be changed so that, instead of defining data directly into an array in code, the data arrays will need to be populated by iterating through a directory of Lua scripts.

Taking spells as an example, there should be a directory called `backend/data/spells` where each item in that directory is expected to be a SpellData Lua file. Iterating through this directory and loading all the spells will then populate the SPELL_DATA array.

During backend startup, as the backend loads game data from Lua scripts, the backend should validate that each table has the fields necessary to populate the struct and should validate constraints. Examples of constraints include: ManaCost on a spell should not be less than 0, no two spells should have the same name, and a Spellbook item which casts a certain spell should refer to a valid spell.

This validation is necessary to prevent against the problems that may arise when moving data from Go (a type-aware language) to Lua (a type-unaware language, especially in the sense that each data script will not know about other data scripts)

3. Marshalling functions

Lua functions received from a file are not expected to be marshalled into a Go function type. Instead they should be stored in some form such that they can be called again by the Lua state during gameplay.

For example, SpellData.OnHit will no longer be a Go func but must instead be some either type. Either a raw string or, more likely, a lua.Value of kind function, if such a thing is possible. This way the Lua function can be called whenever the spell is cast.

This also means we need to make sure that the Lua function is not garbage collected as we keep it inside the SpellData table.

4. Identifiers

Currently, numeric identifiers are used to reference data objects, such as the `Spell` type or the `ItemId` type. These types should continue to exist, but they will not be able to be stored in permanent game data anymore. They will need to be generated automatically. The ID of the Spell firebolt will be whatever index it happens to be in the array once all the spell scripts have been loaded.

This poses a challenge for saved game data such as the world JSON and player save files. We will need to save string values instead of raw ID numbers when saving these files and then convert them back into numbers when loading the files.

Taking a player inventory for example. Currently each player inventory will have an Item such as Item { ItemId: 0, Amount: 25 }, and this represents that the player has 25 of ItemId 0. When saving the player's inventory, we will need to say that, if ItemId 0 is "Gold", that the saved JSON for this Item will be { ItemId: "Gold", Amount: 25 }.

Then when loading the player's inventory JSON, we will need to walk back using a string->int mapping which tells us that ItemId: "Gold" has the ID 0. This way if the item IDs get re-arranged during re-runs of the server, the player's data will stay consistent. This also ensures that comparisons between items remains fast during gameplay runtime, because ItemId, SpellId, etc. can all just be numbers.

## Data scripts

Each folder under `backend/data/` holds one kind of data, and every `.lua` file in it returns a single table of that kind. `World.LoadData(dataFolder)` (`backend/world/world.go`) loads the folders in this order. `WorldInit` calls it with `./data`, and the world editor calls it with the backend's data folder so it can read world files. A script can only refer to data that was loaded before it, and those references are names that are checked during parsing:

| Order | Folder | Go type | Loader | References |
| --- | --- | --- | --- | --- |
| 1 | `races/` | `RaceData` | `loadRaceData` (`character_race.go`) | |
| 2 | `jobs/` | `JobData` | `loadJobData` (`character_job.go`) | |
| 3 | `spells/` | `SpellData` | `loadSpellData` (`spell.go`) | |
| 4 | `recipes/` | `RecipeData` | `loadRecipeTables` + `loadRecipeData` (`recipe.go`) | jobs, items |
| 5 | `items/` | `ItemData` | `loadItemData` (`item.go`) | spells, recipes |
| 6 | `classes/` | `ClassData` | `loadClassData` (`character_class.go`) | spells, items |
| 7 | `npcs/` | `NpcData` | `loadNpcData` (`npc_data.go`) | items |

Recipes are loaded in two passes. `loadRecipeTables` registers recipe names first so that items can refer to them. `loadRecipeData` parses the rest of each recipe once items exist. Classes and NPCs load after items so they can refer to items. NPC scripts, their behavior hooks and their behavior params are described in [npcs.md](npcs.md).

Each ID is the entry's position in load order, and files load in alphabetical order. Names must be unique within a kind, and the player-facing lookups (`RaceIdFromString` etc.) ignore case.

Stat blocks are tables keyed by stat abbreviation (`VIT`, `STR`, `AGI`, `INT`, `FTH`), and missing stats are 0. Race stats can be negative. All other stat blocks must not be.

Races, classes and jobs require a one-sentence `description`, which the character creator shows in a table when a player types `describe race`, `describe class` or `describe job`.

```lua
-- races/dwarf.lua
local race = {}
race.name = "Dwarf"
race.description = "Stout folk of the deep mountains, hardy and strong but slow on their feet."
race.stats = { VIT = 2, STR = 2, AGI = -1 }
return race

-- jobs/blacksmith.lua
local job = {}
job.name = "Blacksmith"
job.description = "A forger of steel who hammers ore into weapons and armor."
job.stats = {}
job.scaling = {}
return job

-- classes/warrior.lua
local class = {}
class.name = "Warrior"
class.description = "A frontline fighter who relies on strength and draws enemies' attention."
class.stats = { VIT = 8, STR = 10, AGI = 6, INT = 6, FTH = 8 }
class.scaling = { VIT = 8, STR = 10, AGI = 6, INT = 6, FTH = 8 }
-- Each unlock sets a level (1 to MOB_MAX_LEVEL - 1) and exactly one of `ability` (a MOB_ABILITY_DATA name) or `spell`
class.unlocks = {
	{ level = 2, ability = "Taunt" },
}
return class
```

Character saves (`saves/<Name>.json`, `CharacterJson` in `character.go`) store race, class, job, spell and recipe names instead of IDs. World files (`data/world.json`) store item names in drop tables (`DropTableEntry` in `drop_table.go`) and inventories (`Item` in `item.go`), and NPC keys in unique NPCs (`Npc` in `npc.go`). Both kinds of file store mob conditions by name (`Condition` in `mob_condition.go`). If a save names something that no longer exists, decoding returns an error and startup fails.

## Script API

Scripts reach the backend through the global `world` table, which `scriptInit` (`backend/world/script.go`) builds when the Lua state is created. It contains:

- **Functions** from `SCRIPT_LIBRARY` in `backend/world/script_library.go`. Each entry maps a Lua name to a `lua.NativeFunc`, so `"log"` becomes `world.log`.
- **Constant tables** from `ScriptConstantTables()` in `backend/world/script_constants.go`. Each `ScriptConstantTable` becomes `world.<Name>`, with string values. For example, an item script sets `item.kind = world.ItemKind.CONSUMABLE`, and the parser maps that string back to the Go enum. To add a table, append it to the slice that `ScriptConstantTables()` returns.
- **Free constants** from `ScriptFreeConstants()` in the same file. Each `ScriptFreeConstant` is set directly on the table as `world.<Name>`, and its `Value` is any scalar `lua.Value` (`lua.Number`, `lua.String`, `lua.Bool`). For example, a spell script can set `spell.cast_time = world.SPELL_CAST_TIME_INSTANT`. To add one, append it to the slice that `ScriptFreeConstants()` returns.

Messages that scripts send to players (`message_room`, `message_player`, `say_to_room`, `say_to_player`) use color tags. Follow the rules in [text_colors.md](text_colors.md).

### Globals are locked

At the end of `scriptInit`, the global table gets a metatable whose `__newindex` and `__index` raise errors. Scripts therefore cannot create new globals (`x = 5` without `local` fails with `Cannot create global variable 'x'`) or read undefined ones (a typo like `wrold.log` fails with `Cannot read undefined global variable 'wrold'`). Declare everything `local`. Existing globals such as `world` and the standard library still work. Go code that needs to add a global after `scriptInit`, such as test setup, must use `RawSetGlobal`, since `SetGlobal` goes through `__newindex`.

### Script errors

When Go calls a script function (a spell's `on_hit`, an item's `on_use`, an NPC hook) and the script raises an error, the error is logged as a `Warn` along with Lua's stack traceback, e.g.

```
Warn - NPC 'troll' update() failed: bad argument #1 (MobHandle expected)
stack traceback:
	[Go]: in function 'native function'
	npcs/troll.lua:68: in function <anonymous>
```

A `[Go]` frame on top means a `world` library function rejected the call; the frame below it is the script line that called it. Call script functions from Go through `World.scriptCall` (`backend/world/script.go`) so that their errors are logged the same way.

Lua can't be re-entered while a `world` library function is running: lunar's `State.Call*` fails with `ErrRunning`. `scriptCall` marks a script as running for the duration of the call, so NPC events fired from inside it (see `docs/backend/npcs.md`) are deferred and handled once it returns. Calling `luaState` directly at runtime skips both the error logging and the deferral.

### Mob handles

Scripts refer to mobs through `MobHandle` values, such as the `caster` and `target` arguments of a spell's `on_hit`. A handle is opaque userdata with no readable fields, so scripts read mob state through the `world.mob_get_*` functions, one per value (`mob_get_name`, `mob_get_room`, `mob_get_health`, `mob_get_max_health` etc.). `mob_get_stat(handle, stat)` takes a `world.Stat` value (`world.Stat.STR`), and `mob_get_equipment(handle, slot)` takes a `world.EquipmentSlot` value and returns the equipped item's name, or `nil` if the slot is empty. Handles compare with `==`, and `tostring` gives `MobHandle(<id>:<generation>)` for logging. NPC state is read and set through handles too, with the `npc_*` functions described in [NPC state](npcs.md#npc-state).

A handle can outlive its mob. Library functions that take a mob throw an error when given a handle to a mob that no longer exists, so a script holding onto a handle (for example in an NPC instance table) should check `world.mob_exists` before using it.

On the Go side, `scriptInit` registers the `MobHandle` userdata type as `World.mobHandleType`. Pass a mob to Lua with `Mob.getLuaHandle`, which creates the userdata on first use and caches it on the mob. Passing a mob to a script therefore doesn't allocate, and every handle for a mob is the same Lua value. Read a handle argument in a `SCRIPT_LIBRARY` function with `frameGetMobHandleArg`, or with `frameGetMobArg` to also look up the mob. Document handle parameters as `@param name MobHandle`.

### Inventories

Scripts read and change a mob's inventory through the `world.mob_*inventory*` functions. Items are addressed by their **0-based** index in the inventory, the same index that an NPC's `on_item_given` hook receives as `added_to_index`. An index that isn't an integer in `[0, length)` throws an error.

| Function | Returns | Notes |
| --- | --- | --- |
| `mob_get_inventory_length(handle)` | integer | Number of item stacks, for looping from `0` to `length - 1` |
| `mob_get_inventory_item_name(handle, index)` | string | |
| `mob_get_inventory_item_amount(handle, index)` | integer | |
| `mob_get_inventory_item_durability(handle, index)` | integer? | `nil` for items that have no durability, so `0` always means broken |
| `mob_get_inventory_amount_of(handle, item_name)` | integer | Total across all stacks. Throws an error if no item has that name |
| `mob_drop_inventory_item(handle, index, amount)` | integer | Moves the item into the mob's room. Returns the amount dropped |
| `mob_give_inventory_item(handle, index, target_handle, amount)` | integer | Moves the item to another mob. Returns the amount given, which is `0` if none of it fits in the target's inventory |

The read functions follow the getter guideline in [Performance](#performance). There's one function per item property, not a function that returns an item table.

`amount` must be a positive integer. It's limited to the size of the stack, and for giving, also to the space left in the target's inventory (players have a capacity, NPCs don't). An item that doesn't stack can only be dropped or given with an amount of `1`.

Removing a whole stack moves the last stack in the inventory into its index (`Inventory.RemoveItems`), so after a drop or give, read the indices again before using them. These functions don't send any messages, so use `message_room` to tell players what happened. Giving an item to an NPC doesn't trigger its `on_item_given` hook, because that hook is only for items that players give.

### Conditions

A condition is a timed effect on a mob, such as the buff from the Blessing of Might spell. Conditions aren't scripted data. They're defined in Go in `backend/world/mob_condition.go`: each `ConditionType` constant has an entry in `CONDITION_DATA` with a `Name` and a `Description`, and the condition's effect is checked in Go wherever it applies (for example, `CONDITION_TYPE_BLESS_MIGHT` is checked in the weapon damage calculation in `mob.go`). Scripts only add, check and remove conditions.

Scripts name conditions with the `world.ConditionType` table. Its keys are the condition names in upper case with spaces replaced by underscores, and its values are the names, so `world.ConditionType.BLESSING_OF_MIGHT` is `"Blessing of Might"`.

| Function | Returns | Notes |
| --- | --- | --- |
| `mob_has_condition(handle, condition)` | boolean | |
| `mob_add_condition(handle, condition, duration)` | | `duration` is in ticks and must be positive or `world.CONDITION_DURATION_INFINITE`. If the mob already has the condition, it keeps whichever duration is longer, with infinite counting as the longest |
| `mob_remove_condition(handle, condition)` | | Does nothing if the mob doesn't have the condition |
| `mob_calc_magic_condition_duration(caster, base_duration)` | integer | Scales a positive base duration by the caster's FTH: `base * (1 + FTH / 50)` |

Durations are counted in world updates (ticks, `WORLD_SECONDS_PER_UPDATE` seconds each), so use `world.time_minutes` to convert from real time. Each `Mob.Update` lowers every finite duration by one and removes a condition once it reaches 0. Conditions with `CONDITION_DURATION_INFINITE` (-1) last until they are removed. These functions don't send any messages, so use `message_room` to tell players what happened. A spell that applies a condition typically looks like:

```lua
function spell.on_hit(caster, target)
    local duration = world.mob_calc_magic_condition_duration(caster, world.time_minutes(1))
    world.mob_add_condition(target, world.ConditionType.BLESSING_OF_MIGHT, duration)
end
```

Conditions are stored on `MobData.Conditions`, so they are saved with characters and NPCs. `Condition`'s custom JSON methods save the condition by name along with its remaining duration.

To add a condition, add a `CONDITION_TYPE_*` constant before `CONDITION_TYPE_COUNT`, give it a `CONDITION_DATA` entry, implement its effect in Go, and run `go generate` so that `world.ConditionType` in `world.d.lua` picks it up.

### Language server definitions

The `world` table only exists at runtime, so `backend/data/world.d.lua` describes it to the Lua language server (LuaLS). It is generated, so don't edit it by hand. After changing `SCRIPT_LIBRARY`, `ScriptConstantTables()` or `ScriptFreeConstants()`, regenerate it from `backend/`:

```
go generate
```

The generator (`backend/luadefs_gen.go`, run through the `-generate-lua-defs` flag in `backend/main.go`) writes:
- the opaque `MobHandle` class
- the `Item` class, matching the `{ name, amount, durability }` table that `npcBehaviorParamToLua` passes for Item behavior params. Update both together.
- each free constant as `world.<Name> = <value>` with a `---@type` annotation for its Lua type
- each constant table as a LuaLS `---@enum`, in the order it's declared
- a stub `function world.<name>(...) end` for each `SCRIPT_LIBRARY` entry, sorted by name

Each stub is documented with the `//` comment directly above its `SCRIPT_LIBRARY` entry, and generation fails if an entry has no comment. Plain comment lines become `---` description lines. Lines that start with `@` are copied through as LuaLS annotations, and the stub's parameter list comes from the `@param` lines in order. Mark optional parameters with `?`:

```go
// Sends a message to the specified room
//
// @param room integer
// @param message string
// @param args? table
"messageRoom": func(frame lua.Frame) lua.Outcome {
```

### Performance

The backend runs Lua with [lunar](https://github.com/mmcdole/lunar) (`github.com/mmcdole/lunar`), a Lua 5.1 interpreter written in pure Go. Calling a `SCRIPT_LIBRARY` function from Lua is an ordinary native call inside the interpreter, with no cgo or FFI cost. With the C Lua implementation, the usual advice is to batch work into fewer Lua/C calls, but that doesn't apply here: one call to a simple getter costs around 90 ns. Allocating Lua tables is what costs time, whether the script allocates them or Go does. Each table is a Go heap allocation that has to be garbage collected later, and that dominates the cost of a call that returns one.

Library functions that read state should therefore:
- Return scalars, one value per function (`mob_get_health`), rather than returning a table that holds a few fields.
- Take an enum parameter for a family of similar values (`mob_get_stat(handle, world.Stat.STR)`), rather than taking a table of field names.
- Return a table only when the data really is a collection, such as a list of mobs in a room.

Separate functions are also better for scripting. The generated LuaLS definitions give each getter its own return type, and a misspelled function name gets flagged by the language server, whereas a misspelled field name inside a table of field names only fails at runtime.

#### Measurements

`backend/world/script_bench_test.go` compares the two designs. `bench_get_mob_table(handle, fields)` is a test-only native that takes a table of field names and returns a table of values. This is how a function returning several mob fields at once would work. The other side uses the real `world.mob_get_*` getters. Each benchmark runs a Lua `for` loop inside a single Lua call, so the Go→Lua call that starts the loop isn't counted. Each iteration reads the given number of fields from a real mob and adds them to an accumulator.

Results (median of 3 runs at 2,000,000 iterations each; lunar v0.1.3, Go 1.26.7, Apple Silicon aarch64 on Asahi Linux):

| Fields read | Table design | Getter design | Getters faster by |
| --- | ---: | ---: | ---: |
| 1 | 4,296 ns/op, 10 allocs | 93 ns/op, 0 allocs | ~46× |
| 2 | 4,775 ns/op, 10 allocs | 186 ns/op, 0 allocs | ~26× |
| 2, fields table built once outside the loop | 2,746 ns/op, 5 allocs | 186 ns/op, 0 allocs | ~15× |
| 5 | 5,158 ns/op, 10 allocs | 480 ns/op, 0 allocs | ~11× |

What this shows:
- A getter call costs about 90 ns, and getters never allocate, so their cost grows linearly with the number of fields read.
- A table-design call costs roughly 4–5 µs, almost all of it in allocation and collection. Each call creates two tables, the script's `{ "name", "room" }` literal and the result table that Go builds, at about 5 Go allocations each. Building the field-name table once and reusing it removes half of that, but the result table still costs more than a dozen getter calls.
- Reading 5 fields with getters is still about 11× faster than one table call. The table design would only win if a script read dozens of fields at once.

Scripts currently run on events (spell hits, item use, NPC hooks) rather than on every tick, so neither design would dominate the 3-second update today. This guideline keeps the API cheap and typed as more scripts get added.

To reproduce, run from `backend/` (takes about 2 minutes):

```
go test ./world -run '^$' -bench ScriptMobData -benchmem -benchtime 2000000x -count 3
```

Absolute numbers depend on the machine. The ratios between the two designs are the part to compare.
