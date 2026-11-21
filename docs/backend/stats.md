# Attributes and stats

Mobs have two kinds of numbers:

- **Attributes** (`Attribute`, `AttributeBlock`, `backend/world/character_attributes.go`) are a character's primary values: `VIT`, `STR`, `AGI`, `INT`, `FTH`, `LCK`. Only characters have them. They don't affect game rules directly except where noted below (attack damage uses STR, spell slots use INT, loot rolls use LCK).
- **Stats** (`Stat`, `StatBlock`, `backend/world/stats.go`) are the derived values that game rules read. Both characters and NPCs have them.

| Stat | Lua key | Meaning |
|---|---|---|
| `STAT_MAX_HEALTH` | `world.Stat.MAX_HEALTH` | Maximum health. |
| `STAT_MAX_MANA` | `world.Stat.MAX_MANA` | Maximum mana. |
| `STAT_INITIATIVE` | `world.Stat.INITIATIVE` | Who goes first in combat. Room occupants update in descending initiative. |
| `STAT_ACCURACY` | `world.Stat.ACCURACY` | How likely an attack is to hit. |
| `STAT_EVASION` | `world.Stat.EVASION` | How likely an attack is to be dodged. |
| `STAT_DEFLECTION` | `world.Stat.DEFLECTION` | Chance to deflect an attack. Each point above the attack's penetration adds `MOB_DEFLECTION_PER_POINT` percent. |
| `STAT_DEFLECTION_AMOUNT` | `world.Stat.DEFLECTION_AMOUNT` | Percent of damage removed on a deflect. |
| `STAT_RESISTANCE` | `world.Stat.RESISTANCE` | Flat damage reduction on each attack. |
| `STAT_MAGIC_RESISTANCE` | `world.Stat.MAGIC_RESISTANCE` | Flat damage reduction on magic damage. |
| `STAT_FOCUS` | `world.Stat.FOCUS` | Keeping concentration and resisting mental effects such as taunt or charm. |

`Stat.String()` gives the display name ("Max Health"), which is also the value of the Lua constant. Lua constant names are the upper-case display name with spaces replaced by underscores. Attributes use their abbreviation for both.

`StatBlock.ToStringArray(isBonus)` formats a block for players as lines of at most `STATS_PER_LINE` (6) stats joined by ` | `, so the caller can indent each line. Each column is padded to its widest entry so the ` | ` separators line up from line to line. With `isBonus` true (equipment), zero stats are left out and the rest read `+X Stat` or `-X Stat`. With it false (the character sheet), every stat is shown as `Stat X`.

## Reading stats

Every mob's current stats are in `MobData.Stats`, which `Character`, `Npc` and `Mob` all reach through `MobData`. Read one stat with `mob.Stat(STAT_X)`, or the whole block with `mob.Stats`. Lua reads a stat with `world.mob_get_stat(handle, world.Stat.X)`.

`MobData.Stats` is never saved. Characters recalculate it, and NPCs copy it from their data when they spawn.

## Characters

A character's attributes (`Character.Attributes`) and stats (`MobData.Stats`) are cached and are not saved in the character's JSON. Both are recalculated on load.

- `calculateAttributes()`: the race's and class's attributes added together, then scaled to the character's level (`AttributeBlock.ScaledToLevel`). Called on creation, load and level up.
- `calculateStats()`: built from attributes and equipment. Must be called whenever attributes or equipment change. It is called on creation, load (after equipment is restored), level up, `EquipFromInventory`, `UnequipToInventory`, when an equipped item breaks, and when equipment is dumped on death.

| Stat | Formula |
|---|---|
| Max health | 10 × VIT |
| Max mana | 10 × INT |
| Initiative | AGI + Σ armor `initiative_modifier` |
| Accuracy | AGI |
| Evasion | AGI + Σ armor `evade_modifier` |
| Deflection | Σ armor `deflection` |
| Deflection amount | 40 + Σ armor `deflection_bonus` |
| Resistance | Σ armor `resistance` |
| Magic resistance | FTH / 2 |
| Focus | INT / 2 |

Σ armor sums over every equipped item with armor data (outfits, shields, accessories).

## NPCs

An NPC's base stats come from the `stats` table in its data file and are stored on `NpcData.stats`, shared by every NPC of that type. `spawnNpc` copies them into the new NPC's `MobData.Stats` (and sets its health and mana to the maximums), so each NPC instance can have its stats changed on its own. `ScriptParser.getOptionalStatBlock` parses a stat table, and every stat it isn't given is 0. The NPC parser then treats 0 as "not given" for the two stats with other defaults: max health becomes 1 and deflection amount becomes 40. See [world_data.md](world_data.md#npcs).

```lua
npc.stats = {
    [world.Stat.MAX_HEALTH] = 10,
    [world.Stat.DEFLECTION] = 2,
}
```
