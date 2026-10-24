# God Commands

Gods are admins, listed by user ID in the `GOD_USER_IDS` env var. A God runs a God command by prefixing it with `/` (e.g. `/shutdown now`, `/help`) from any menu. `GameState.handleCommand` (`backend/game/game.go`) sends these commands to `MENU_GOD` (`backend/game/menu_god.go`). God commands are developer tools, so their code stays in `menu_god.go`. They use the existing `world` API and don't add code to `world`.

`handleCommand` doesn't send a heartbeat after a God command. A command that changes HP or MP calls `godSendHeartbeats` so that the God and the target player see the change right away.

Every command except `/shutdown` acts on the God's own mob, so it only works while the God is logged in (`godRequireLoggedIn`).

## Commands

### /shutdown [now] [in \<minutes\>] [cancel] [status]
Puts the server into maintenance mode: saves and kicks all players, then refuses new connections until the server restarts. `in` schedules the shutdown, `cancel` cancels a scheduled one, and `status` shows the time left.

### /kill \<target\>
Sets the HP of a mob in the room to 0 and tells the room. The target is fuzzy-found (`fuzzyFindTarget`), so `self` works. The mob is removed on the next world update by the normal death handling: EXP is granted, a corpse is left and NPCs respawn. Killing a player, the God included, is permadeath.

### /heal \<target\>
Restores a mob in the room to full HP and MP and removes all of its conditions, including beneficial ones, so that combat tests start from a clean slate.

### /create \<item\> [durability \<durability-percent\>]
Adds one of the item to the room's inventory. The item name is fuzzy-found over all items. The durability percent is a percent of the item's max durability. It defaults to 100, can't be negative and can be over 100.

### /set-level \<target\> \<level\>
Sets a player's level (1 to `MOB_MAX_LEVEL`) and resets their experience. Their stats are recalculated for the new level. All their class unlocks (abilities and class spells) are removed, and then the unlocks for levels 1 through the new level are granted again. HP and MP are capped at the new maximums. The command doesn't grant starting items. It uses the same stat formula as `calculateStatAtLevel` in `backend/world/character.go`, so the two must be changed together.

### /teleport \<room\>
Moves the God to the first room whose name matches exactly, ignoring case. The move works like a normal move (`Room.MoveOccupant`): the leave and enter messages are announced and NPCs in the new room get the player-entered event.
