# Mob actions

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
