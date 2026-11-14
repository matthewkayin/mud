# World Data

The game's content is loaded from the data folder, `backend/data/`, when the server starts. Most of it is Lua scripts, each returning one table. Room positions and exits are stored in `layout.json` instead.

If any file fails to load, the server refuses to start and logs the error. Each error names the file it came from.

`world.NewWorld` reads the data folder through the `fs.FS` it is given, so the same loader can run over files that aren't on disk. The server passes `os.DirFS("data")`. `NewWorld` loads only the script data; the server then loads saved characters from `saves/` with `World.LoadCharacters`.

## Data folder

```
data/
    items/          item types
    races/          playable races
    backgrounds/    playable backgrounds
    rooms/          room contents
    layout.json     room positions, exits and the start room
    world.d.lua     generated LuaLS definitions for the `world` script API
```

Files load in this order: items, races, backgrounds, rooms (including the layout). A loader can refer to data loaded before it. For example, room loot tables name items. Every `.lua` file under a core folder is loaded, including files in subfolders. Other files are skipped.

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

The same room file can be used by more than one room. Each copy gets its own inventory and state. Because this is usually a mistake, the server logs a warning when it happens.

Rooms are indexed in the order they appear in `rooms`.

The loader is in `backend/world/layout.go` (parsing and validation) and `backend/world/room_parse.go` (building the rooms).
