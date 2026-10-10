# NPC

NPCs are non-player characters, the computer-controlled "players" of the game. An NPC enters the world by creating an instance of itself as a mob, so an NPC is both a mob spawner and a mob controller.

There are two halves to an NPC:

- **NpcData** (`backend/world/npc_data.go`) is the canonical definition, such as "Goblin". It is loaded from a Lua script in `backend/data/npcs/`, and NPC_DATA is indexed by `NpcId`.
- **Npc** (`backend/world/npc.go`) is a unique NPC placed in the world. Unique NPCs are saved in the world JSON (`World.Npcs`), so the world decides which NPCs spawn where. Each one points back to its NpcData through `Id` and uses it for starting stats and behavior.

The NPC state machine in `Npc.update()` handles core behavior, including respawning, resets, wandering, and getting ready to fight and then attacking players. Scripts can add to it by implementing behavior hooks.

## NPC scripts

Each `.lua` file in `backend/data/npcs/` returns one NpcData table. `parseNpc` reads it. NPCs load after items, so equipment, drop tables and Item behavior params can refer to items by name.

| Field | Type | Notes |
| --- | --- | --- |
| `id` | string | Unique ID that the world JSON uses to refer to this NPC |
| `name` | string | Display name of the NPC's mobs |
| `description` | string | Shown when a player looks at the NPC |
| `experience_worth`, `experience_worth_scaling` | integer | Experience at level 1, plus the amount added per level after that |
| `stats`, `scaling` | stat block | Stats at level 1 and stats gained per level |
| `equipment` | table | Slot name (e.g. `main_hand`) → item name |
| `drop_table` | list | The same entries as other drop tables: `{ item, amount, durability_percent, drop_chance_percent }` |
| `starting_disposition` | `world.NpcDisposition` | Hostile NPCs attack players who enter their room |
| `movement_type` | `world.NpcMovementType` | Sentinel NPCs stay put and wanderers move between rooms |
| `behavior_params` | list (optional) | See [Behavior params](#behavior-params) |
| hooks | function (optional) | See [Behavior hooks](#behavior-hooks) |

```lua
local npc = {}

npc.id = "troll"
npc.name = "Troll"
npc.description = "A hairy beast towers over you."
npc.experience_worth = 100
npc.experience_worth_scaling = 25
npc.stats = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.scaling = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.equipment = {}
npc.drop_table = {}
npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.SENTINEL

npc.behavior_params = {
	{ name = "Toll", type = world.NpcBehaviorParamType.ITEM },
	{ name = "ExitToBlock", type = world.NpcBehaviorParamType.DIRECTION },
}

npc.init = function(self, params)
	return { toll = params.Toll, exit_to_block = params.ExitToBlock }
end

return npc
```

## Behavior hooks

The script table is loaded only once and shared by every NPC of that type, so it can't hold per-NPC state. Instead, `init` returns an **instance table**, and the NPC passes that table back to each of its other hooks. When the mob dies or is despawned on a world reset, the instance table is dropped, and the next spawn calls `init` again.

Every hook also receives `self`, the `MobHandle` of the NPC's own mob, so a script can pass it to the `world` library (`world.mob_get_room(self)`, `world.npc_set_mode(self, world.NpcMode.AGGRO)`). `self` and `instance` live exactly as long as the mob, so a hook never sees a stale `self`.

| Hook | Called | Arguments | Return |
| --- | --- | --- | --- |
| `init` | When the NPC spawns its mob | `self, params` | The instance table, or nil |
| `update` | Every world update while the mob is alive, before the state machine runs | `instance, self` | |
| `on_attacked` | When the mob takes damage | `instance, self, attacker` | |
| `on_player_entered` | When a player moves into the mob's room | `instance, self, player` | `true` to prevent the default |
| `on_item_given` | When a player gives the mob an item | `instance, self, player, added_to_index, amount` | `true` to prevent the default |
| `get_status_description` | When a player looks at the mob, after its description | `instance, self` | A string, or nil for no status line |

`attacker` and `player` are `MobHandle`s. `added_to_index` is the index in the NPC's inventory that the given item was added to.

Every hook is optional. Each event has its own hook, so an NPC only crosses the Go/Lua boundary for the events it handles. If an event hook returns `true` (`world.NPC_EVENT_PREVENT_DEFAULT`), the Go default is skipped:
- `on_player_entered`: by default, a neutral NPC becomes hostile.
- `on_item_given`: by default, the NPC gives the item back to the player.
- `on_attacked` has no Go default, so its return value is ignored.

Hook errors are logged as warnings and treated as unhandled.

On the Go side, events are queued on the NPC with `Npc.PushEvent` and handled at the start of the NPC's next `update` by `Npc.onEvent`, so a hook never runs in the middle of another hook. `Npc.callHook` passes `instance` and `self` before the hook's own arguments, and `Npc.callEventHook` also reads the prevent-default result. To add an event:
1. Add a `*lua.Function` field to `NpcData` and read it in `parseNpc` with `getOptionalFunction`.
2. Add an `NPC_EVENT_TYPE_*` constant and an event data struct, and push the event from wherever it happens.
3. Handle the event in `onEvent`: call the hook with `callHook` or `callEventHook`, then run any Go default.

## NPC state

Scripts read and change NPC state through the `world` library, passing the NPC's mob handle. These functions throw an error if the mob doesn't exist or isn't an NPC, so check `world.mob_is_npc(handle)` first when the handle might be a player.

| Function | Notes |
| --- | --- |
| `mob_is_npc(handle)` | |
| `npc_get_mode(handle)`, `npc_set_mode(handle, mode)` | `world.NpcMode`: `IDLE`, `SURPRISE` or `AGGRO`. Setting a mode goes through the same Go setters as the state machine, so `SURPRISE` alerts the mob and `IDLE` restarts the wander timer. Dead isn't a mode that scripts can see or set, because an NPC is dead exactly when its mob doesn't exist. |
| `npc_get_disposition(handle)`, `npc_set_disposition(handle, disposition)` | `world.NpcDisposition`: `NEUTRAL`, `HOSTILE` or `FRIENDLY` |

A mode set in `update` takes effect on the same tick, because `update` runs before the state machine.

## Behavior params

Behavior params let the world configure each unique NPC. For example, the troll's script declares that each troll has a `Toll` and an `ExitToBlock`, and the world editor sets those values on each troll it places. The script defines the schema in `behavior_params`, which `NpcData.BehaviorParams` stores as a map of name → `NpcBehaviorParamType`. Each `Npc` stores its values in `BehaviorParams map[string]any`.

| Type | Go value | World JSON | Lua value in `params` |
| --- | --- | --- | --- |
| `String` | `string` | string | string |
| `Number` | `float64` | number | number |
| `Boolean` | `bool` | boolean | boolean |
| `Item` | `Item` | `{ "Id": "<item name>", "Amount": n, "Durability": n }` | `{ name, amount, durability }` |
| `Direction` | `Direction` | `"north"`, `"east"`, `"south"` or `"west"` | the same string |

Params are passed only to `init`. If an NPC needs them later, `init` should copy them into its instance table.

A unique NPC must have a value for every declared param, and it can't have params that aren't declared. Decoding rejects undeclared params and wrong types, and `World.Validate()` reports missing params and invalid values.

## World JSON

```json
{
  "Id": "troll",
  "SpawnRoom": 2,
  "LevelRange": { "Min": 2, "Max": 3 },
  "MovementTypeOverride": "",
  "DropTableOverride": { "Entries": [] },
  "BehaviorParams": {
    "Toll": { "Id": "Gold", "Amount": 10, "Durability": 0 },
    "ExitToBlock": "north"
  }
}
```

- `Id` is the NPC script's `id`.
- `MovementTypeOverride` is `""` to use the script's `movement_type`, or a movement type name such as `"Wander"`.
- `DropTableOverride` replaces the script's `drop_table` if it has any entries.

In the world editor, the NPCs section of a room's sidebar edits all of these fields. The editor builds the behavior param inputs from the schema that `GET /api/npcs` returns.
