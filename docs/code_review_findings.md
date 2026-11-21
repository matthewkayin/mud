# Code review findings: item targets

Findings from a review of the uncommitted item target changes on `kayin/changes` (`fuzzy.go`, `menu_world_look.go`, `mob.go`, `item_target.go`). Ordered from most to least severe.

## Important

### 1. Items on the floor make the player's own item targets ambiguous

`backend/game/fuzzy.go:376`

`fuzzyFindTargetItem` searches the inventory, equipment and room floor as one list, and `fuzzyFind` reports `AMBIGUOUS` on a tie. A floor item therefore ties with an item the player holds, and commands that used to work now fail with an ambiguity error.

**Example:** The player holds a Sword and another Sword lies on the floor. `use whetstone on sword` used to target the held sword. Now `findTarget` returns the ambiguous match error. Any player can cause this just by dropping a duplicate item.

**Fix:** Search the player's own items first and only fall back to the floor when nothing matches there, or break ties in favour of the player's items.

**Human Response:** Don't fix it this way. It is impossible to know the user's intent in this situation. They might have just searched the room and seen the sword on the ground or in the chest, in which case that sword is at the top of their mind and they may not be thinking of the one in their inventory, and so using the whetstone on the sword in their inventory might cause them to waste the item in a way that they didn't intend. Instead, we should just make sure that the AMBIGUOUS error message makes it clear to the user that the cause of the ambiguity is that there are multiple locations containing the same item. The user should resolve the ambiguity by specifying "use whetsone on sword in inventory" or "use whetsone on sword in room".

### 2. `look at` hides item search errors

`backend/game/menu_world_look.go:66`

The handler discards the `err` from `fuzzyFindTargetItem` and only checks `err == nil`. An ambiguous match falls through to the remaining checks and ends with a "don't see" message, even though the player can see several matching items.

**Example:** A Potion is in the inventory and another is on the floor. `look at potion` makes `fuzzyFind` return `AMBIGUOUS`, no room detail matches, and the player gets "You don't see any 'potion' in the room."

**Fix:** When the item search fails because the match is ambiguous or the number is out of range, send that error instead of falling through.

**Human Response:** I agree, send that error, and also send different error messages for ambiguous location vs ambiguous item. Ambiguous item means that there are 2 swords in the player's inventory. Ambiguous location means there is one sword in their inventory and one on the ground.

### 3. Fuzzy item matches hide room details

`backend/game/menu_world_look.go:65`

Items are now checked before room details and the room description. Fuzzy matching accepts word and substring matches, so any carried item whose name contains the word hides a room detail. The code already checks chests before items for this reason ("chest key"), but room details still come after items.

**Example:** A room has the detail "door" and the player carries a "Door Key". `look at door` matches the key as a substring (score 99) and describes the key, so the door detail can no longer be looked at.

**Fix:** Check `room.Details` (exact key) before the fuzzy item search, or only let an item win when it is an exact match or scores higher.

**Human Response:** Check room.Details before fuzzy item search is the approach we should take. But room description fuzzy matching should come *after* item search, since description matching is mostly just for flavor.

### 4. Any `look at` text containing "in" goes only to the item search

`backend/game/menu_world_look.go:30`

`splitArgsBy` splits on the first literal "in", and the `userSpecifiedIn` branch returns early. Mobs, chests, room details and description sentences whose names contain "in" can no longer be looked at.

**Example:** A room has the detail "light in window". `look at light in window` gives "'window' is not 'room', 'inventory', an equipment slot, or a container in the room." instead of the detail. An item named "Jack in the Box" also can't be found unless the player adds a location.

**Fix:** Only treat "in <location>" as a location when `parseItemLocation` recognises the words. Otherwise fall through to the normal search using all the words.

**Human Response:** I agree with this fix.

### 5. Targeting an item in a chest rolls the chest's loot and can delete items

`backend/game/fuzzy.go:322`

Looking at or targeting an item `in <chest>` calls `chest.GetInventory`, which rolls loot and opens the chest. For a `CHEST_MODE_REPLENISH` chest it also replaces the existing inventory. So just looking at an item in a chest (`look at x in chest`), or a `use` whose target later fails validation, changes world state and can delete items that were in the chest.

**Example:** A chest's replenish timer has run out and it still holds items a player dropped. `look at sword in chest` rerolls the loot, overwrites `chest.inventory` and destroys the dropped items. It also restarts `CHEST_REPLENISH_DURATION`.

**Fix:** Use a read-only accessor for looking and for validation, and only roll loot when a player explicitly searches or opens the chest. If rolling here is intended, do it only once the action is actually carried out.

**Human Response:** This is actually a result of a bug in the chest code. The chest code should have been calling loot.replenish when the chest was in replenish mode. I have made the change to fix that part already. Regarding the roll itself, I think we should continue to call chest GetInventory() here, otherwise it creates a discontinuity. If someone says "look at potion in chest" and we respond "there is no potion in chest", and then they search chest and see a potion, that doesn't match up. However, you're right that we should be wary of deleting dropped items just from a "look at". I've given it some thought and realized that we can just use replenish always instead of getLoot() on chests, so I've changed the chest code accordingly.

## Nits

### 6. Searches with a number never match an equipment slot item

`backend/game/fuzzy.go:332`

The new name check calls `scoreFuzzyMatch(item.NameWithCondition(), itemWords)` on the raw words. These still include any trailing number, so a search with a number always scores 0. The old code accepted the item in the named slot without checking its name.

**Example:** With a Ring equipped, `look at ring 1 in accessory` scores 0 because the word "1" matches nothing, and the player gets "There is no item called 'ring 1' in your accessory."

**Fix:** Strip the number with `getFuzzyNumberFromArgs` before scoring and check that it is 1 or not given, or score only the words that aren't numbers.

**Human Response:** This is ok to leave as-is. Add a comment to the code indicating that this issue exists but that we've decided to leave the behavior as-is.

### 7. Lua item hooks get the item ID but not where the item is

`backend/world/mob.go:612`

`buildTargetHookArgs` appends only `lua.Number(itemTarget.Id)`. Targets can now be on the floor or in a chest, but `on_use` and `validate_target` can't tell where the item is. Any script that looks the ID up in the user's inventory or equipment fails for items held anywhere else.

**Example:** An item with `target_type` ITEM looks up its target in the user's inventory by ID. `use kit on sword in room` passes the floor sword's ID, the lookup finds nothing, and the hook errors with "But it failed!"

**Fix:** Pass the holder (or a resolved item handle) to Lua along with the ID, or document that hooks must not assume the item is in the user's inventory.

### 8. The "in" separator is case-sensitive

`backend/game/menu_world_look.go:30`

`splitArgsBy` only matches "in" in lowercase, while `parseItemLocation` accepts "inventory" and "room" in any case. So location words work in any case but "In" and "IN" don't work as the separator.

**Example:** `look at Ring In Inventory` searches for the whole phrase "Ring In Inventory" as a mob, chest or item, finds nothing, and replies "You don't see any 'ring in inventory' in the room."

**Fix:** Compare case-insensitively in `splitArgsBy`, or lowercase the words first.

**Human Response:** Compare case-insensitively in splitArgsBy, because that will fix the issue in all cases beyond just "in".
