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
| `targets` | `world.SpellTarget` | What the spell is cast on. See [Spell targets](#spell-targets) |
| `on_cast` | function | Runs when the spell is cast. Its parameters depend on `targets` |
| `validate_argument` | function | Required when `targets` is `ARGUMENT`, and not read otherwise |

```lua
-- spells/firebolt.lua
local spell = {}

spell.name = "Firebolt"
spell.description = "Casts a bolt of fire toward the target"
spell.casts_to_learn = 50

spell.mana_cost = 5
spell.cast_time = 1
spell.targets = world.SpellTarget.ENEMY

function spell.on_cast(caster, target)
    local damage = world.deal_magic_damage(caster, target, 10)
    local room = world.mob_get_room(caster)
    world.message_room(room, string.format("<y>%s</y> took <r>%d damage</r> from the firebolt.", world.mob_get_name(target), damage))
end

return spell
```

`on_cast` sends its own messages about what the spell did. The backend has already printed `<y>X</y> cast <g>Spell</g>!` before it runs.

## Spell targets

A spell's `targets` field (`SpellTarget` in `spell.go`) decides what a player must write after `on` in `cast <spell> [on <target>]` or `use <scroll> [on <target>]`, and what `on_cast` receives:

| `targets` | Player writes | `on_cast` parameters |
| --- | --- | --- |
| `NONE` | nothing. Writing `on ...` is an error | `(caster)` |
| `ALLY` | a player in the room | `(caster, target)`, both `MobHandle`s |
| `ENEMY` | an NPC in the room | `(caster, target)`, both `MobHandle`s |
| `ARGUMENT` | any text | `(caster, argument)`, where `argument` is a string |

`findSpellTarget` (`backend/game/menu_world.go`) checks what the player wrote and turns it into a [`MobTarget`](#mob-targets), which is queued with the cast or use action. The `cast` and `use` commands share it, so a spell scroll targets the same way as casting its spell.

### Argument targets

An `ARGUMENT` spell takes free text instead of a mob, such as the direction in `cast unlock on north`. The words after `on` are joined with single spaces and lowercased, so the script always receives lowercase text.

The script decides which arguments are valid with `validate_argument(caster, argument)`:

- Return `nil` if the spell can be cast with that argument.
- Otherwise return a string. It's shown to the player as the error, so write it as a full sentence that tells them what they can write instead.

```lua
-- spells/unlock.lua
spell.targets = world.SpellTarget.ARGUMENT

function spell.validate_argument(caster, argument)
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

Go calls it through `SpellData.ValidateArgument`. It runs twice:

1. When the player enters the command (`findSpellTarget`). If it returns an error, the player sees it and nothing is queued.
2. Again in `Mob.spellcast` (`backend/world/mob.go`), right before `on_cast`, because the world can change while the spell charges or before the next tick. If it fails here, the room sees `But it failed!`.

`on_cast` can therefore assume its argument is valid. `validate_argument` must not change the world or send messages, since it runs on every attempt. If it raises a Lua error, the error is logged and the player sees `The spell failed.`

### Failed casts

A cast fails when the second `validate_argument` check fails or when `on_cast` raises an error. A failed cast costs no mana and doesn't count toward mastering the spell, and a spell scroll that fails is put back in the caster's inventory.

## Mob targets

`Mob.Target` (`MobTarget` in `backend/world/mob.go`) is what a mob's current action is aimed at. `Target.Type` says what kind of value `Target.Value` holds:

| `Type` | `Value` | Used for |
| --- | --- | --- |
| `MOB_TARGET_NONE` | `nil` | Idle mobs, and spells that target `NONE` |
| `MOB_TARGET_MOB` | `MobHandle` | Attacking, taunting, `ALLY`/`ENEMY` spells and consumables |
| `MOB_TARGET_ARGUMENT` | `string` | `ARGUMENT` spells |

The mode setters keep the target in step with `Mob.Mode`. `SetModeAttack` and `SetModeTaunt` take a handle and build a `MOB_TARGET_MOB`, `SetModeCast` and `SetModeUseItem` take a `MobTarget`, and `SetModeIdle` resets it to `MOB_TARGET_NONE`. Return a mob to idle with `SetModeIdle` rather than assigning `Mode` directly, so the old target doesn't stay behind.

Each tick, `Mob.Update` calls `isTargetInvalid` before acting on a target. A `MOB_TARGET_MOB` target is invalid if the mob no longer exists, is dead, has left the room or is hidden. The mob then goes idle without acting. Other target types are always valid here. Argument targets are checked by the spell itself, as described in [Argument targets](#argument-targets).

Check `Target.Type` before type-asserting `Target.Value`. `getTargetMob` assumes a valid `MOB_TARGET_MOB` target, so call it only after `isTargetInvalid` returns `false` for one.

To add a target type, add a `MOB_TARGET_*` constant, handle it in `isTargetInvalid`, and handle it wherever `Target.Value` is read for the modes that can use it (`spellcast`, `useItem`, the `stop` command).

## Spell List - List and requirements for each spell

Each entry in this list contains the spell's Type, Description, and Behavior. Below are some details about eaech property.

*Type* - Indicates if a spell is a Class spells or Learned spell. If a spell is a Learned spell, you should create a Spellbook and Spell Scroll item for the spell. If it is a Class spell, you should not create a Spellbook or Spell Scroll item for the spell. Spellbooks should always be called "Spellbook of <Spell Name>". Scrolls should always be called "Scroll of <Spell Name>".

*Targets* - Describes the SpellTarget

*Description* - Provides brief flavor description of the spell. Use this as the Spell's description field.

*Behavior* - Describes the spell's behavior including messages and edge cases.

*Notes* - Any additional implementation notes which you may want to keep in mind when implementing the spell.

Entries in this list do not describe the spell's casts_to_learn, mana_cost, cast_time, or damage value. For these fields, choose reasonable values and don't think about it too much. These values can be changed later. You main focus should be on implementing the spell and its behavior.

All spell damages and healing amounts should scale with the caster's FTH using provided Lua API functions (script_library.go) unless stated otherwise in the spell description.

### Firebolt

Type: Learned Spell
Targets: Enemy 
Description: Casts a bolt of fire toward the target.
Behavior: Deals damage to the target and messages the room: "<target> took X damage from the firebolt." on hit. If the target died, also messages: "<target> as burnt to a crisp."

### Cure

Type: Learned Spell
Targets: Ally
Description: Heal the target with holy magic
Behavior: Heals the target and messages the room: "<target> regained X HP." The spell does not heal beyond max health. If the amount healed is 0, the message is still sent.

### Unlock

Type: Learned Spell
Targets: Argument
Description: Unlocks a physical or magical lock
Behavior: The argument is a direction. Validates that the direction is valid and that the exit in the specified direction is locked. Unlocks the exit and says "<caster> unlocked the %s exit!"

### Blessing of Might

Type: Class Spell
Targets: None
Description: Grants increased power to you and your allies
Behavior: Provides an increased damage buff to all allies in the room (including the caster) for 1 minute (scaling).

### Purify

Type: Class Spell
Targets: Ally
Description: Cures a target of their ailments
Behavior: Removes malicious conditions from the target. For each condition removed, messages the room "<target> was cured of <condition>"! If the target has no conditions to cure, it should simply message "It had no effect"
Notes: To implement this, you will need to make engine changes. Add a bool property on ConditionData that marks each condition as isAilment true/false. Then add scripts to the API to query mob conditions. These functions should be `mob_get_condition_count(mob)`, `mob_get_condition_name(mob, index)`, `mob_condition_is_ailment(mob, index)`.

### Mage Weapon

Type: Learned Spell
Targets: None
Description: Imbunes your attacks with arcane strength
Behavior: Adds a buff to the caster which increases their attack damage. Can only be cast on self. Damage increase should be a percentage similar to blessing of might's percentage, but it should be a different buff from blessing of might to allow stacking of these two buffs side-by-side. Sends no room message.

### Mage Armor

Type: Learned Spell
Targets: None
Description: Shield yourself with arcane protection
Behavior: Adds a buff to the caster which increases their defense damage. Can only be cast on self. Defense increase should be a percentage similar to blessing of might's percentage. Sends no room message.

### Invisibility

Type: Learned Spell
Targets: Ally
Description: Turn yourself invisible
Behavior: Sets the target invisible by adding an INVISIBLE condition which lasts 1 minute (scaled). Invisible targets remain hidden without needing a stealth check. Performing actions while invisible does not remove the invisibility. Messages the room: "<target> disappeared from sight!"
Notes: To implement, you will need to add a new getting to check if the Mob isHidden() which checks both for the invisibility condition and the hidden flag.

### Chain Lightning

Type: Learned Spell
Targets: None
Description: Casts a chain of lightning against all targets in the room
Behavior: Does damage to each enemy in the room. Messages the room: "Lightning sparks from <caster>'s fingertips!". Messages the room "<target> took X damage from the lightning!". If a target dies, messages: "<target> has been reduced to a pile of ash!". This spell can still be cast when there are no enemies in the room, it just won't do anything.
Notes: It can be assumed for now that NPCs are enemies, as friendly NPCs only exist in safe zones.
