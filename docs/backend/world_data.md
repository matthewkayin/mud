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
| `kind` | `world.ItemKind.*` | |
| `price` | integer | Not negative. |
| `size` | integer | Not negative. |

### npcs

| Field | Type | Notes |
|---|---|---|
| `id` | string | Required. Rooms refer to the NPC by this id. Must be unique. |
| `name` | string | Required. |
| `description` | string | Required. |
| `starting_disposition` | `world.NpcDisposition.*` | Required. |
| `max_health` | integer | Optional, defaults to 1. Greater than 0. |
| `max_mana` | integer | Optional, defaults to 0. Not negative. |
| `damage` | integer | Optional, defaults to 1. Not negative. |
| `penetration` | integer | Optional, defaults to 0. Armor penetration score. Not negative. |
| `resistance` | integer | Optional, defaults to 0. Flat armor resistance. Not negative. |
| `deflection` | integer | Optional, defaults to 0. Deflection score. Not negative. |
| `deflection_amount` | integer | Optional, defaults to 60. Percent of damage deflected on a successful deflect, from 1 to 100. |
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
