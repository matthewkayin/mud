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

Each folder under `backend/data/` holds one kind of data, and every `.lua` file in it returns a single table of that kind. `WorldInit` (`backend/world/world.go`) loads the folders in this order. A script can only refer to data that was loaded before it, and those references are names that are checked during parsing:

| Order | Folder | Go type | Loader | References |
| --- | --- | --- | --- | --- |
| 1 | `races/` | `RaceData` | `loadRaceData` (`character_race.go`) | |
| 2 | `jobs/` | `JobData` | `loadJobData` (`character_job.go`) | |
| 3 | `spells/` | `SpellData` | `loadSpellData` (`spell.go`) | |
| 4 | `recipes/` | `RecipeData` | `loadRecipeTables` + `loadRecipeData` (`recipe.go`) | jobs, items |
| 5 | `items/` | `ItemData` | `loadItemData` (`item.go`) | spells, recipes |
| 6 | `classes/` | `ClassData` | `loadClassData` (`character_class.go`) | spells, items |

Recipes are loaded in two passes. `loadRecipeTables` registers recipe names first so that items can refer to them. `loadRecipeData` parses the rest of each recipe once items exist. Classes load last so they can refer to items.

Each ID is the entry's position in load order, and files load in alphabetical order. Names must be unique within a kind, and the player-facing lookups (`RaceIdFromString` etc.) ignore case.

Stat blocks are tables keyed by stat abbreviation (`VIT`, `STR`, `AGI`, `INT`, `FTH`), and missing stats are 0. Race stats can be negative. All other stat blocks must not be.

```lua
-- races/dwarf.lua
local race = {}
race.name = "Dwarf"
race.stats = { VIT = 2, STR = 2, AGI = -1 }
return race

-- jobs/blacksmith.lua
local job = {}
job.name = "Blacksmith"
job.stats = {}
job.scaling = {}
return job

-- classes/warrior.lua
local class = {}
class.name = "Warrior"
class.stats = { VIT = 8, STR = 10, AGI = 6, INT = 6, FTH = 8 }
class.scaling = { VIT = 8, STR = 10, AGI = 6, INT = 6, FTH = 8 }
-- Each unlock sets a level (1 to MOB_MAX_LEVEL - 1) and exactly one of `ability` (a MOB_ABILITY_DATA name) or `spell`
class.unlocks = {
	{ level = 2, ability = "Taunt" },
}
return class
```

Character saves (`saves/<Name>.json`, `CharacterJson` in `character.go`) store race, class, job, spell and recipe names instead of IDs. If a save names something that no longer exists, startup fails.

## Script API

Scripts reach the backend through the global `world` table, which `scriptInit` (`backend/world/script.go`) builds when the Lua state is created. It contains:

- **Functions** from `SCRIPT_LIBRARY` in `backend/world/script_library.go`. Each entry maps a Lua name to a `lua.NativeFunc`, so `"log"` becomes `world.log`.
- **Constant tables** from `ScriptConstantTables()` in `backend/world/script.go`. Each `ScriptConstantTable` becomes `world.<Name>`, with string values. For example, an item script sets `item.kind = world.ItemKind.CONSUMABLE`, and the parser maps that string back to the Go enum. To add a table, append it to the slice that `ScriptConstantTables()` returns.

### Language server definitions

The `world` table only exists at runtime, so `backend/data/world.d.lua` describes it to the Lua language server (LuaLS). It is generated, so don't edit it by hand. After changing `SCRIPT_LIBRARY` or `ScriptConstantTables()`, regenerate it from `backend/`:

```
go generate
```

The generator (`backend/luadefs_gen.go`, run through the `-generate-lua-defs` flag in `backend/main.go`) writes:
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
