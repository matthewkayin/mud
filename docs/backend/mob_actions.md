# Mob actions

## Mob ownership

A `Mob` (`backend/world/mob.go`) belongs to exactly one `Character` or `Npc`, held in `Mob.Character` or `Mob.Npc`; the other is nil. `Mob.MobData` points to the owner's `MobData` (name, room, health, mana, inventory).

`Mob` methods hide the difference where both kinds of mob have an answer: `IsPlayer`, `IsEvil`, `IsHostile`, `Message` (does nothing for NPCs), `DescribeToPlayer`, `MaxHealth`, `MaxMana`, `InventoryCapacity`, `GetAttacks` and the combat stat getters. The stat getters read `Mob.GetStats()`: a character's cached `Stats`, calculated from its attributes and equipment, or the NPC's `NpcData.stats` (see [stats.md](stats.md)). Logic that applies to only one kind of mob, such as equipment durability or the player key, checks `mob.Character` or `mob.Npc` directly.

## Attacks

`Mob.attackTarget` resolves one `MobAttack` (damage, penetration, and the equipment slot of the weapon, if any):

1. The deflect chance is `(target deflection - penetration) * MOB_DEFLECTION_PER_POINT` percent, clamped to 0–100.
2. A deflected attack deals `deflectionAmount` percent of its damage.
3. The target's flat resistance is subtracted, with a minimum of 1 damage.
4. An attack made with a weapon costs the weapon 1 durability (`Character.subtractDurabilityFromEquipment`). A weapon that breaks is unequipped.

Characters have no deflection or resistance of their own yet.

Every mob (player character or NPC) has a `Mode` and a `Target` (`backend/world/mob.go`). The mode is what the mob is doing:

| Mode | Started by | Notes |
| --- | --- | --- |
| `MOB_MODE_IDLE` | `SetModeIdle`, `Stop` | Doing nothing. |
| `MOB_MODE_ATTACK` | `SetModeAttack` (`attack` command, hostile NPCs) | Keeps going until the target is gone or the mob stops. |
| `MOB_MODE_CAST` | not yet implemented | Uses `spellTimer`. |
| `MOB_MODE_USE_ITEM` | `SetModeUseItem` (`use` command) | Resolves once in `Mob.update`, then goes idle. |

## Pending actions

Starting an action sets `HasPendingAction`. The client shows this as the status's `action`, which animates the prompt. `Mob.update` clears it on the mob's next turn, and that is when the action starts. An action that is still pending hasn't started yet.

## Starting and cancelling actions

Every mode a mob chooses to start must go through `Mob.beginAction`. It announces the cancellation of whatever the mob was already doing, then sets the new mode, sets the target and marks the action pending. `Mob.Stop` (the `stop` command) announces the cancellation in the same way and then goes idle.

`announceActionCancelled` messages the room only when all of these are true:

- The current action has started (`HasPendingAction` is false). A pending action is replaced without a message.
- The target is still valid (`isTargetValid`). If the target died or left the room, nothing is announced.
- The mode is attack ("stopped attacking") or cast ("cancelled their spell"). Cancelling an item use is never announced.

`SetModeIdle` is the silent reset. Use it when the action ends by itself: an item use resolving, a target becoming invalid, or the mob leaving the room. `Mob.Move` silently ends attacks and casts, because their targets are in the old room. A queued item use survives the move, and `useItem` checks its target when it resolves.

`attack` on the target the player is already attacking changes nothing and replies "You are already attacking X." This teaches players that attacks continue on their own.
