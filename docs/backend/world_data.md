# World Data

The game's content is loaded from the data folder, `backend/data/`, when the server starts. Most of it is Lua scripts, each returning one table. Room positions and exits are stored in `layout.json` instead.

`world.NewWorld` reads the data folder through the `fs.FS` it is given, so the same loader can run over files that aren't on disk. The server passes `os.DirFS("data")`. The validation language server (`lsp.md`) passes a filesystem that overlays unsaved editor buffers on the folder. `NewWorld` also takes the writer that scripts calling `print` write to: stdout for the server, stderr for the language server. `NewWorld` loads only the script data; the server then loads saved characters from `saves/` with `World.LoadCharacters`.

`World.RoomFileNames` maps the path of every room file that loaded, relative to `rooms/`, to its `name`. It is filled even when the layout has errors.

## Problems

Loading doesn't stop at the first bad file. `NewWorld` returns every problem it finds as a list of `ScriptProblem`s (`backend/world/script_problem.go`), along with the world. Each problem has:

- `Path`: the file it is about, relative to the data folder (or the core folder, if the folder is missing).
- `Key`: the field it is about, if any (for example `item` in a loot entry). Problems inside `attributes`, `skills`, `details`, `interactions` and `loot` use that table's key.
- `Line`: the line Lua reported, for syntax errors and errors raised while the script runs. 0 otherwise.
- `RoomIndex`: for `layout.json` problems, the index of the room entry. -1 otherwise.
- `Severity`: error or warning.

A file with an error is left out of the world, so later loaders see it as missing. For example, a room whose loot names an item from a broken item file gets an "unknown item" error too. A file whose name duplicates an earlier file's is also left out. If the layout has any errors, no rooms are built.

The server logs every problem. It refuses to start if any of them is an error. Warnings don't stop it.

## Data folder

```
data/
    items/          item types
    npcs/           NPC types
    races/          playable races
    backgrounds/    playable backgrounds
    rooms/          room contents
    layout.json     room positions, exits and the start room
    world.d.lua     generated LuaLS definitions for the `world` script API
```

Files load in this order: items, npcs, races, backgrounds, rooms (including the layout). Each of these core folders must exist, or it's an error. A loader can refer to data loaded before it. For example, room loot tables name items. Every `.lua` file under a core folder is loaded, including files in subfolders. Other files are skipped.

Scripts run in a restricted Lua state. The base, math, string and table libraries are available, along with the `world` table (see `backend/world/script_library.go` and the generated `world.d.lua`). Reading or creating an undefined global is an error, so declare variables `local`.

Required fields must be present. There are no defaults for them.

## Core folders

### items

| Field | Type | Notes |
|---|---|---|
| `name` | string | Other files refer to the item by this name. |
| `description` | string | |
| `kind` | `world.ItemKind.*` | See the kinds below. |
| `price` | integer | Not negative. |
| `size` | integer | Not negative. |

Item kinds:

| Kind | Equips to | Required skill | Extra fields |
|---|---|---|---|
| `CONSUMABLE`, `MISC` | Can't be equipped | None | None |
| `ONE_HANDED_SWORD`, `ONE_HANDED_AXE`, `ONE_HANDED_BLUNT`, `DAGGER` | Main hand or off hand | Swords, Axes, Bludgeoning, Daggers | Weapon |
| `TWO_HANDED_SWORD`, `TWO_HANDED_AXE`, `TWO_HANDED_BLUNT`, `BOW` | Both hands | Swords, Axes, Bludgeoning, Marksmanship | Weapon |
| `SHIELD` | Off hand | Shields | Armor |
| `CLOTHING` | Outfit | None | Armor |
| `LIGHT_ARMOR`, `HEAVY_ARMOR` | Outfit | Light Armor, Heavy Armor | Armor |
| `ACCESSORY` | Accessory | None | Armor |

Weapon and armor kinds have durability and need these fields:

| Field | Type | Notes |
|---|---|---|
| `max_durability` | integer | Not negative. An item below half its max durability is shown as "Damaged". |
| `reinforced_prefix` | string | Shown before the item's name while its durability is above its max durability, for example "Sharpened". |

Kinds with a required skill (all weapons, `SHIELD`, `LIGHT_ARMOR` and `HEAVY_ARMOR`) also need `skill_requirement`: an integer, not negative, giving the rank a character needs in the required skill to equip the item. `CLOTHING` and `ACCESSORY` don't have it.

Weapon kinds also need `damage` and `penetration`, both integers and not negative.

Armor kinds also need `deflection`, `deflection_bonus` and `resistance` (integers, not negative), plus `evade_modifier`, `sneak_modifier` and `initiative_modifier` (integers). A character's stats add these up over every equipped item with armor data (see [stats.md](stats.md)).
Each item instance in the game has an `ItemId` that is unique while the server runs. IDs aren't saved: loaded items get new ones. Consumables stack.

#### Consumables

Items with `kind = world.ItemKind.CONSUMABLE` also have:

| Field | Type | Notes |
|---|---|---|
| `target_type` | `world.TargetType.*` | What the player names after `on` in `use <item> [on <target>]`. |
| `validate_target` | function | Required when `target_type` is `ITEM` or `ARGUMENT`. `validate_target(self, target)` returns an error message string to refuse the target, or `nil` to accept it. |
| `on_use` | function | `on_use(self, target)`. Called when the item is used. |

`self` is the user's `MobHandle`. What `target` holds depends on `target_type`:

| `target_type` | `target` |
|---|---|
| `NONE` | Not passed. The player can't name a target. |
| `ALLY`, `ENEMY` | The target's `MobHandle`. Must be in the user's faction (`ALLY`) or not in it (`ENEMY`). |
| `ITEM` | The `ItemId` (an integer) of an item in the user's inventory or equipment, on the room floor, or in a container in the room. Without `in <location>`, the player's inventory, equipment and the room floor are searched. Containers are only searched when named, as in `use kit on sword in chest`. If the best matches are in more than one location (say, a sword in the inventory and one on the floor), the command fails and tells the player to add `in <location>`. A number such as `sword 2` only picks between matches in the same location. |
| `ARGUMENT` | The words after `on`, lowercased, as a string. |

Using an item is a combat action. The `use` command checks the target (calling `validate_target` if the item has it) and sets the mob to `MOB_MODE_USE_ITEM`. On the next world tick, `Mob.useItem` checks that the item and target still exist and calls `validate_target` again. Then it calls `on_use`, and if that doesn't error, removes one of the item from the user's inventory.

An item target is a `world.ItemTarget`: the item's `ItemId` and its holder, which is a mob (inventory or equipment), a room's floor, or the chests in a room. Chests are matched by searching every chest in the room for the ID, because chest indices change when corpses rot. `Mob.findTargetItem` resolves a target and also checks reach: a floor or chest item must be in the user's current room, and a mob holder other than the user must be in that room. So an item that is picked up, moved or left behind makes the queued use fail.

Players can also `look at <item> [in <location>]` to see an item's name, kind, description and weapon or armor numbers (not its durability). It finds items the same way, with `fuzzyFindTargetItem` in `backend/game/fuzzy.go`. The location is `inventory`, `room`, an equipment slot, or a container in the room. `in` (in any case) only starts a location when the words after it name one (checked by `splitArgsByItemLocation`), so names containing "in", such as a room detail "light in window", are searched whole. Without a location, `look at` tries, in order: a mob, a container, an exact room detail, an item, and then a sentence of the room description. If items match but are ambiguous, the player is told so instead of the search moving on.

Queuing an action (`SetModeUseItem`, `SetModeAttack`) sets `Mob.HasPendingAction`, and the mob's turn (`Mob.update`, once per tick) clears it before resolving the mode. The player's status heartbeat reports it as `action`, which the client shows as an animated prompt.

### npcs

| Field | Type | Notes |
|---|---|---|
| `id` | string | Required. Rooms refer to the NPC by this id. Must be unique. |
| `name` | string | Required. |
| `description` | string | Required. |
| `starting_disposition` | `world.NpcDisposition.*` | Required. |
| `stats` | table | Optional. Keys are `world.Stat.*`, values are integers (see [stats.md](stats.md)). Stats not given, or given as 0, are 0, except `MAX_HEALTH` (defaults to 1) and `DEFLECTION_AMOUNT` (defaults to 40, the percent of damage deflected on a successful deflect, from 1 to 100). No stat may be negative. |
| `damage` | integer | Optional, defaults to 1. Not negative. |
| `penetration` | integer | Optional, defaults to 0. Armor penetration score. Not negative. |
| `attacks_per_round` | integer | Optional, defaults to 1. Greater than 0. Each attack deals `damage` with `penetration`. |
| `experience_worth` | integer | Optional, defaults to 0. Not negative. |
| `loot` | array | Optional. Items the NPC drops, in the same format as a room's `loot`. |
| `wares` | array | Optional. Item names the NPC sells. |
| `init` | function | Optional. `(self)`, called when the NPC spawns. Returns the NPC's instance state table. |
| `update` | function | Optional. `(self, instance)`, called each world update with the table `init` returned. |

### races and backgrounds

Races and backgrounds have the same fields. A new character's attributes and skills are the sum of its race's and its background's.

| Field | Type | Notes |
|---|---|---|
| `name` | string | Save files refer to the race or background by this name. |
| `description` | string | |
| `attributes` | table | Keys are `world.Attribute.*`, values are integers. |
| `skills` | table | Keys are `world.Skill.*`, values are integers. |

### rooms

A room file describes what is in a room, not where the room is. `layout.json` places room files on the map.

| Field | Type | Notes |
|---|---|---|
| `name` | string | Required. |
| `description` | string | Required. |
| `details` | table | Optional. Maps a detail name to the text shown when a player looks at it. |
| `interactions` | table | Optional. Maps a lowercase verb to a function `(state, player, arg)` that runs when a player in the room types that verb. |
| `reset` | function | Optional. Called each time the room resets, every 5 minutes. |
| `loot` | array | Optional. Items the room restocks on reset. Each entry has `item` (an item name), `amount` (an integer or `{ min, max }`), `drop_chance_percent`, and an optional `durability_percent` (defaults to 100). |
| `npcs` | array | Optional. NPCs that spawn in the room. Each entry has `id` (an NPC id) and an optional `respawn_duration` in world updates (defaults to 5 minutes). Each NPC spawns when the server starts and respawns after it is gone. |

Rooms with a `reset` function get a `state` table, which is passed to their interactions.

## layout.json

`layout.json` holds the map: where each room is, which rooms connect, and where new characters start. Editors write it, so it is plain JSON rather than Lua.

```json
{
    "start": { "x": 0, "y": 0, "z": 0 },
    "rooms": [
        { "position": { "x": 0, "y": 0, "z": 0 }, "path": "presentation_space.lua", "exits": ["west", "up"] },
        { "position": { "x": -1, "y": 0, "z": 0 }, "path": "kitchen.lua", "exits": ["east"] }
    ]
}
```

- `start`: the position of the room new characters start in.
- `rooms`: one entry per room on the map.
  - `position`: integer `x`, `y`, `z`. North is `-y`, east is `+x`, up is `+z`. A room's position is its identity. Save files refer to rooms by position, so moving a room moves any character standing in it.
  - `path`: the room's Lua file, relative to `rooms/` (for example `kitchen.lua`).
  - `exits`: directions from `north`, `east`, `south`, `west`, `up`, `down`. Each one leads to the room next to this one in that direction.

All fields are required, and unknown fields are an error. The layout is also invalid if:

- two rooms share a position,
- `start` isn't the position of a room,
- an exit isn't a valid direction, or is listed twice on the same room,
- an exit leads to a position with no room,
- an exit has no matching exit back. Exits are two-way, so a room's `east` exit needs a `west` exit on the room to its east,
- a room's `path` is empty, or isn't a room file that loaded.

The same room file can be used by more than one room. Each copy gets its own inventory and state. Because this is usually a mistake, it is reported as a warning.

Rooms are indexed in the order they appear in `rooms`.

The loader is in `backend/world/layout.go` (parsing and validation) and `backend/world/room_parse.go` (building the rooms).
