# Text Colors

Messages sent to players can color parts of their text with inline tags. The web client's terminal (`frontend/src/terminal/parser.ts`) turns each tag into a colored span. This guide says which tag each kind of text gets, so that the same kind of text always looks the same.

## Tags

| Tag | Color | Used for |
| --- | --- | --- |
| `<y>...</y>` | Yellow | Names of mobs, player characters and NPCs |
| `<r>...</r>` | Red | Damage, critical hits, kills |
| `<g>...</g>` | Green | Healing, item names, spell names |
| `<b>...</b>` | Blue | Keywords in NPC dialogue |
| `<w>...</w>` | White | Don't use it. Untagged text is already white |

The tag set and the hex values are defined in `frontend/src/terminal/text_color.ts`. If you add a color there, add a rule for it here.

How the parser reads tags:

- Tags can't be nested. The parser ends a tag at the first matching close tag and doesn't look for tags inside it.
- A tag must open and close within one message. A tag that isn't closed, or that isn't one of the tags above, is printed as plain text.
- Player input goes through `sanitizeString` (`backend/game/game.go`), which strips `<` and `>`, so players can't add colors to their own text.

Apart from a possessive `'s` on a name, punctuation that follows colored text goes outside the tag: `<b>Castle Recurse</b>!`, not `<b>Castle Recurse!</b>`.

## Rules

### Names are yellow

Color the name of every mob, player character and NPC, wherever it shows up: in combat messages, in spell messages, as the speaker of dialogue, and as the subject of an action.

- A possessive `'s` goes inside the tag.
- Other punctuation after the name goes outside the tag.

```go
// Correct
fmt.Sprintf("<y>%s</y> dodged <y>%s's</y> attack!", targetMob.GetName(), mob.GetName())
fmt.Sprintf("<y>%s</y> has <r>slain</r> <y>%s</y>.", mob.GetName(), targetMob.GetName())

// Incorrect
fmt.Sprintf("%s dodged %s's attack!", targetMob.GetName(), mob.GetName())   // names not colored
fmt.Sprintf("<y>%s</y>'s %s broke!", mob.GetName(), itemData.Name)          // possessive outside the tag
fmt.Sprintf("has <r>slain</r> <y>%s.</y>", targetMob.GetName())             // period inside the tag
```

### Damage and kills are red

- A damage amount colors the number and its unit together: `<r>12 damage</r>`.
- The critical hit prefix is red.
- In a death message, the verb that says the mob died is red (`<r>slain</r>`, `<r>burnt to a crisp</r>`).

```lua
-- Correct
string.format("<y>%s</y> took <r>%d damage</r> from the firebolt.", target_name, damage)
string.format("<y>%s</y> has <r>burnt to a crisp</r>.", target_name)

-- Incorrect
string.format("<y>%s</y> took <r>%d</r> damage from the firebolt.", target_name, damage)   -- unit outside the tag
string.format("<y>%s</y> has burnt to a crisp.", target_name)                              -- death verb not colored
```

### Healing, items and spells are green

- A healing amount colors the number and its unit together: `<g>12 HP</g>`.
- Item names are green.
- Spell names are green.

```go
// Correct
fmt.Sprintf("<y>%s</y> cast <g>%s</g>!", mob.GetName(), spellData.Name)
fmt.Sprintf("<y>%s's</y> <g>%s</g> broke!", mob.GetName(), itemData.Name)

// Incorrect
fmt.Sprintf("<y>%s</y> cast %s!", mob.GetName(), spellData.Name)        // spell name not colored
fmt.Sprintf("<y>%s</y> regained <g>%d</g> HP.", targetName, healing)    // unit outside the tag
```

### NPC dialogue keywords are blue

In NPC dialogue, color the words a player can ask the NPC about next. These are the subjects that the NPC script's `on_dialog` function responds to.

```lua
world.say_to_player(self, player, "There is a <b>bridge</b> not far from here.")
```

### Text that is never colored

- Text the player typed, such as the message in `say` or `tell`. The speaker's name is still yellow.
- Command usage, help text and menu prompts.
- Error messages and other system messages, such as login and character creation messages.

## Where messages are written

Check new or changed strings in these places against the rules above:

- Go, in `backend/world/`: `world.messageRoom`, `world.messagePlayer` and `world.messageRoomWithOptions` (`event.go`).
- Go, in `backend/game/`: `gamestate.messageRoom`, `gamestate.broadcast` and `player.send(...)`.
- Lua scripts in `backend/data/`: `world.message_room`, `world.message_player`, `world.say_to_room` and `world.say_to_player`. The `say_to_*` functions color the speaker's name themselves, so the script only colors the message.

If a message needs to color a kind of text that no rule covers, add the rule to this file first.
