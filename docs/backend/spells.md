# Spells

Spells are Lua scripts in `backend/data/spells/`, loaded by `parseSpell` (`backend/world/spell.go`) into `SpellData`. See [scripting.md](scripting.md) for how data scripts are loaded and for the `world` script API.

## Spell scripts

| Field | Type | Notes |
| --- | --- | --- |
| `name` | string | Unique among spells |
| `description` | string | |
| `casts_to_learn` | integer | Casts needed to master the spell. Must not be negative |
| `mana_cost` | integer | Must be positive |
| `cast_time` | integer | Ticks spent charging before the spell is cast. Must not be negative. Use `world.SPELL_CAST_TIME_INSTANT` for no charge time |
| `target_type` | `world.TargetType` | What the spell is cast on. See [Spell targets](#spell-targets) |
| `on_cast` | function | Runs when the spell is cast. Its parameters depend on `target_type` |
| `validate_target` | function | Required when `target_type` is `ITEM` or `ARGUMENT`, and not read otherwise |

```lua
-- spells/firebolt.lua
local spell = {}

spell.name = "Firebolt"
spell.description = "Casts a bolt of fire toward the target"
spell.casts_to_learn = 50

spell.mana_cost = 5
spell.cast_time = 1
spell.target_type = world.TargetType.ENEMY

function spell.on_cast(caster, target)
    local damage = world.deal_magic_damage(caster, target, 10)
    local room = world.mob_get_room(caster)
    world.message_room(room, string.format("<y>%s</y> took <r>%d damage</r> from the firebolt.", world.mob_get_name(target), damage))
end

return spell
```

`on_cast` sends its own messages about what the spell did. The backend has already printed `<y>X</y> cast <g>Spell</g>!` before it runs.

## Spell targets

A spell's `target_type` field (`TargetType` in `target.go`) decides what a player must write after `on` in `cast <spell> [on <target>]` or `use <scroll> [on <target>]`, and what `on_cast` receives:

| `target_type` | Player writes | `on_cast` parameters |
| --- | --- | --- |
| `NONE` | nothing. Writing `on ...` is an error | `(caster)` |
| `ALLY` | a player in the room | `(caster, target)`, both `MobHandle`s |
| `ENEMY` | an NPC in the room | `(caster, target)`, both `MobHandle`s |
| `ITEM` | an item in their inventory or equipment. See [Item targets](#item-targets) | `(caster, item)`, where `item` is an [`ItemHandle`](scripting.md#item-handles) |
| `ARGUMENT` | any text | `(caster, argument)`, where `argument` is a string |

`findTarget` (`backend/game/menu_world.go`) checks what the player wrote and turns it into a [`MobTarget`](#mob-targets), which is queued with the cast or use action. The `cast` and `use` commands share it, so a spell scroll targets the same way as casting its spell. Consumable items have the same `target_type` and `validate_target` fields and target the same way.

### Argument targets

An `ARGUMENT` spell takes free text instead of a mob, such as the direction in `cast unlock on north`. The words after `on` are joined with single spaces and lowercased, so the script always receives lowercase text.

The script decides which arguments are valid with `validate_target(caster, argument)`:

- Return `nil` if the spell can be cast with that argument.
- Otherwise return a string. It's shown to the player as the error, so write it as a full sentence that tells them what they can write instead.

```lua
-- spells/unlock.lua
spell.target_type = world.TargetType.ARGUMENT

function spell.validate_target(caster, argument)
    local room = world.mob_get_room(caster)
    if argument ~= world.Direction.NORTH and argument ~= world.Direction.EAST and
        argument ~= world.Direction.SOUTH and argument ~= world.Direction.WEST then
        return string.format("'%s' is not a valid direction. The directions are 'north', 'east', 'south', and 'west'.", argument)
    end
    if not world.room_is_exit_locked(room, argument) then
        return string.format("The %s exit is not locked.", argument)
    end
    return nil
end

function spell.on_cast(caster, argument)
    world.room_set_exit_locked(world.mob_get_room(caster), argument, false)
end
```

Go calls it through `SpellData.ValidateTarget`. It runs twice:

1. When the player enters the command (`findTarget`). If it returns an error, the player sees it and nothing is queued.
2. Again in `Mob.spellcast` (`backend/world/mob.go`), right before `on_cast`, because the world can change while the spell charges or before the next tick. If it fails here, the room sees `But it failed!`.

`on_cast` can therefore assume its argument is valid. `validate_target` must not change the world or send messages, since it runs on every attempt. If it raises a Lua error, the error is logged and the player sees `The spell failed.`

### Item targets

An `ITEM` spell or consumable targets one of the player's items, such as the weapon in `use repair kit on sword`. `fuzzyFindMobItem` (`backend/game/fuzzy.go`) searches both the player's inventory and their equipment. If the words match an item in each, the target is ambiguous and the player has to add where to look: `use repair kit on sword in inventory`, or `in` followed by an equipment slot (`in main hand`). `inspect` reads the words after `in` the same way.

The hooks receive the item as an [`ItemHandle`](scripting.md#item-handles), and read and change it with the `world.item_*` functions. `validate_target(user, item)` runs at the same two points as for argument targets, and returns `nil` or an error message the same way:

```lua
-- items/consumables/repair_kit.lua
item.target_type = world.TargetType.ITEM

function item.validate_target(user, target)
    if not world.item_is_repairable(target) then
        return "That item cannot be repaired."
    end
    if world.item_get_durability_percent(target) >= 100 then
        return "That item is not damaged."
    end
    return nil
end

function item.on_use(user, target)
    world.item_repair(target)
end
```

The queued `MOB_TARGET_ITEM` holds the `ItemHandle`. The item can move in the inventory before the action runs, so `buildTargetHookArgs` (`backend/world/mob.go`) passes the hooks a handle relocated with `World.RelocateItemHandle`. If the item is gone, the action fails with `But it failed!`.

### Failed casts

A cast fails when the second `validate_target` check fails or when `on_cast` raises an error. A failed cast costs no mana and doesn't count toward mastering the spell, and a spell scroll that fails is put back in the caster's inventory.

## Mob targets

`Mob.Target` (`MobTarget` in `backend/world/mob.go`) is what a mob's current action is aimed at. `Target.Type` says what kind of value `Target.Value` holds:

| `Type` | `Value` | Used for |
| --- | --- | --- |
| `MOB_TARGET_NONE` | `nil` | Idle mobs, and spells that target `NONE` |
| `MOB_TARGET_MOB` | `MobHandle` | Attacking, taunting, `ALLY`/`ENEMY` spells and consumables |
| `MOB_TARGET_ITEM` | `ItemHandle` | `ITEM` spells and consumables |
| `MOB_TARGET_ARGUMENT` | `string` | `ARGUMENT` spells |

The mode setters keep the target in step with `Mob.Mode`. `SetModeAttack` and `SetModeTaunt` take a handle and build a `MOB_TARGET_MOB`, `SetModeCast` and `SetModeUseItem` take a `MobTarget`, and `SetModeIdle` resets it to `MOB_TARGET_NONE`. Return a mob to idle with `SetModeIdle` rather than assigning `Mode` directly, so the old target doesn't stay behind.

Each tick, `Mob.Update` calls `isTargetInvalid` before acting on a target. A `MOB_TARGET_MOB` target is invalid if the mob no longer exists, is dead, has left the room or is hidden. The mob then goes idle without acting. Other target types are always valid here. Item and argument targets are checked by `validate_target` instead, as described in [Item targets](#item-targets) and [Argument targets](#argument-targets).

Check `Target.Type` before type-asserting `Target.Value`. `getTargetMob` assumes a valid `MOB_TARGET_MOB` target, so call it only after `isTargetInvalid` returns `false` for one.

To add a target type, add a `MOB_TARGET_*` constant, handle it in `isTargetInvalid`, and handle it wherever `Target.Value` is read for the modes that can use it (`spellcast`, `useItem`, the `stop` command).
