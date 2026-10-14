---
name: text-color-audit
description: Audit player-facing messages for consistent use of text color tags (<y>, <r>, <g>, <b>) against the text color style guide. Use when a developer asks for a text color, message color or color style audit, of the whole repo or of the current git diff.
argument-hint: "[diff [base-ref]]"
disable-model-invocation: true
allowed-tools: Read, Grep, Glob, Bash(git diff:*), Bash(git status:*), Bash(git ls-files:*), Bash(git merge-base:*)
---

# Text color audit

Check that player-facing messages follow the text color style guide, then report what's wrong and offer to fix it.

Arguments: `$ARGUMENTS`

## 1. Load the rules

1. Read `docs/backend/text_colors.md`. It is the only source of rules. Don't invent rules that aren't in it.
2. Read `frontend/src/terminal/text_color.ts`. If its tags don't match the guide's tag table, report that before anything else.

## 2. Pick the scope

- **No arguments: whole repo.** Audit every non-test `.go` file and every `.lua` file under `backend/` (skip `*_test.go` and the generated `backend/data/world.d.lua`).
- **`diff`: uncommitted changes.** Run `git diff HEAD` for staged and unstaged changes, and `git status --porcelain` to find untracked files. Audit untracked `.go`/`.lua` files under `backend/` in full.
- **`diff <base-ref>`: changes since a ref**, for example `diff main` for the whole branch. Run `git merge-base <base-ref> HEAD`, then `git diff <merge-base>` (this includes uncommitted changes), plus untracked files as above.

In diff mode, report only lines the diff adds or changes. Read the code around them anyway so you know what each `%s`/`%d` holds. If a changed line is correct but now disagrees with an unchanged line nearby, mention that under the report's "Context" note, not as a finding.

## 3. Find the message sites

The guide's "Where messages are written" section lists the functions that send messages. Find:

1. Every string that already contains a color tag. Grep for `</?[a-z]>` in the scoped files.
2. Every call to a listed message function whose string includes something a rule covers: a mob, player or NPC name, a damage or healing amount, a death, an item name, a spell name, or an NPC dialogue subject. A string with no tags at all can still be a violation, because missing color counts too.

Strings are sometimes built in a variable or helper before they're sent. Follow them back to where they're formatted.

## 4. Check each site

For each format argument, work out what it holds (for example `mob.GetName()`, `targetMob.Data.Name` and `world.mob_get_name(...)` are names; `spellData.Name` is a spell name; `itemData.Name` is an item name). Then check every rule in the guide, including:

- the right tag for the kind of text
- what goes inside the tag (number and unit together, possessive `'s` inside, other punctuation outside)
- text that must not be colored (player-typed text, usage/help, errors)
- tags that are unknown, nested or left open

## 5. Report

Group findings by file. For each one give:

- `path:line`
- the rule it breaks (one short phrase)
- the current string
- the corrected string

End with a count of findings per rule. If there are none, say the scope is clean and what scope you checked.

## 6. Offer fixes

Don't edit anything yet. Ask whether to apply all the fixes, some of them, or none. If the developer approves:

1. Apply the edits.
2. Run `go test ./...` from `backend/` and report the result.
3. If a finding showed that the guide is missing a rule or has an unclear one, suggest the change to `docs/backend/text_colors.md`, but don't make it unless asked.
