# Castle Recurse - Design Proposal B (Disciplines)

## Motivation

Castle Recurse is a multi-user dungeon game where players create characters and go dungeoneering out in the wilds of Brooklyn and beyond. Players create custom characters, and how a character is built decides what they are good at. Proposal A states four goals for this system, and this proposal keeps all of them:

1. It allows players to fulfill certain fantasies
2. It allows players to indicate how they would like to play the game
3. It creates dependencies among players, because a character cannot be good at everything, so they must rely on other players to succeed
4. It creates a variety of experience in a game in which death is permanent

It adds two more, which Proposal A leaves unstated:

5. **Every choice is legible.** A player reading their character sheet in a text terminal should be able to tell what each point they spent does for them. If a choice can't be explained in one sentence, players won't feel it, and it isn't doing work.
6. **Every rule is authorable.** Each mechanic is content the developers have to write, balance and test. A small set of rules that combine well beats a large set of rules that each need their own data.

Goal 4 makes goal 5 matter more than usual. With permanent death, players will make characters often. Character creation and leveling have to be quick to understand, or every death becomes a chore of re-reading the rules.

### Why change Proposal A

Proposal A has five attributes, about 27 skills, two rules that cap skill ranks, three damage types, four stats per weapon, spell slots, a spell-forgetting NPC, and several skills whose mechanics are not designed yet. Looked at from first principles, much of this is doing the same job twice:

- **Attributes and skill points are two layers of the same choice.** An attribute's only job in A is to cap skills. A limited budget of skill points already stops a character from mastering everything, so the attribute layer adds a second, permanent decision at creation that players have to make before they understand the game.
- **Most skills are gear gates, not abilities.** Sword, Axe, Hammer, Knives, Bows, Block and Armor each say "you may equip better X." Seven skills for one rule is a lot of character-sheet lines for the player and a lot of tuning for the developer, and splitting weapons this finely makes the "weapon master" fantasy cost a huge share of a character's points.
- **Numeric skills are hard to feel.** "+5% dodge" from a rank of Evade is invisible in play. A rank that unlocks a named thing ("you can now Taunt") is felt immediately.
- **Three armor values triple the data.** Every mob and every armor piece needs Slash, Pierce and Crush values, and players need to know which damage type they deal and what each enemy resists, all through text.

Proposal B keeps the parts of A that work (the rank cap by level, weapon techniques, spellbooks, magic focuses, gear tiers) and collapses the rest.

## Overview

A character is made of three things:

- A **Background**, chosen at creation, which gives a few starting ranks and some flavor.
- **Ranks in Disciplines**. There are 10 Disciplines, each with ranks 0 to 5.
- **Gear**, which is described by a type and a tier (1 to 5).

There are no attributes. The whole character sheet is a name, a background, a level, HP/MP, and at most ten lines of Disciplines.

## Disciplines

A Discipline is a broad area of competence. Each one represents a fantasy a player can name ("I'm the tank", "I'm the thief", "I'm the healer").

| Group   | Discipline | Covers                                                        | Gear it gates                  |
| ------- | ---------- | ------------------------------------------------------------- | ------------------------------ |
| Martial | Guard      | Heavy armor, shields, toughness, protecting allies           | Heavy armor, shields           |
| Martial | Heavy      | Big weapons that punch through armor                          | Axes, hammers, greatswords     |
| Martial | Finesse    | Fast weapons, dodging, dual wielding                          | Swords, knives                 |
| Martial | Marksman   | Ranged attacks                                                | Bows, crossbows                |
| Magic   | Arcane     | Blasting and crowd control (A's Evocation and Illusion)      | Wands, staves                  |
| Magic   | Divine     | Healing, protection, blessings (A's Holy magic)              | Holy symbols                   |
| Magic   | Occult     | Summons, curses, draining (A's Necromancy and Conjuration)   | Skulls, ritual knives          |
| Utility | Shadow     | Sneaking, locks and traps, escaping                           | none                           |
| Utility | Scout      | Spotting, evaluating enemies, campcraft, first aid            | none                           |
| Utility | Craft      | Smithing, alchemy, scribing, enchanting                       | none (rank sets recipe tier)   |

### What a rank does

A rank in a Discipline does up to three things, and every Discipline uses the same three:

1. **Gear tier.** You can equip gear of that Discipline with a tier up to your rank. Guard 3 lets you wear tier-3 heavy armor.
2. **Checks.** When you attempt something the Discipline covers (picking a lock, spotting a hidden enemy, resisting poison), your rank is what gets rolled. See [Checks](#checks).
3. **A perk.** Each rank unlocks one named ability, listed in the [perk tables](#perk-tables). The rank-1 perk is the Discipline's core mechanic. Higher ranks add to it.

So a Discipline at rank N always means "tier-N gear, rank-N checks, and the first N perks in its table." There is no other per-rank math.

### Getting ranks

- At creation, a character picks a Background (2 ranks) and spends 1 more rank anywhere. They start with 3 ranks.
- Each level from 2 to 12 grants 1 rank. A level-12 character has 14 ranks.
- A Discipline's rank can't be higher than the cap for the character's level. This is Proposal A's rule 2, unchanged:

| Level    | 1-2 | 3-5 | 6-8 | 9-11 | 12 |
| -------- | --- | --- | --- | ---- | -- |
| Max Rank | 1   | 2   | 3   | 4    | 5  |

With 14 ranks and a cap of 5, a finished character can master two Disciplines and be strong in a third, or spread out into a generalist with four or five Disciplines at rank 3. Nobody can cover all ten, which is goal 3. That budget replaces Proposal A's attribute caps.

Ranks are permanent. Death is permanent too, so the way to try a new build is a new character.

### Backgrounds

A Background is the answer to "what did you do before you went into the dungeon?" It grants 2 ranks, at least one of them in a Utility Discipline, plus a starting item or two.

Backgrounds solve the problem Proposal A raises about Utility points: combat Disciplines are more tempting, so without a nudge every character would be all-combat. Proposal A's idea of a separate Utility point pool fixes this with a new rule. Backgrounds fix it with a choice the player already wants to make, and they give a new character an identity before they've spent a single level-up.

Examples:

| Background       | Ranks             | Starting items          |
| ---------------- | ----------------- | ----------------------- |
| Locksmith        | Shadow 1, Craft 1 | Lockpicks               |
| Night Guard      | Guard 1, Scout 1  | Lantern, tier-1 shield  |
| Seminary Dropout | Divine 1, Craft 1 | Holy symbol, a scroll   |
| Bodega Cat Sitter| Finesse 1, Scout 1| Knife, bag of treats    |
| Hedge Witch      | Occult 1, Scout 1 | Ritual knife, herbs     |

Backgrounds are pure data (a name, a description, two ranks, some items), so servers running their own MUD can write their own.

## Checks

Every non-combat roll in the game uses one formula:

```
chance = clamp(50% + 20% × (rank − difficulty), 5%, 95%)
```

`difficulty` is the tier of the thing being attempted: a tier-3 lock, a tier-2 trap, a hidden tier-4 monster, a tier-1 poison. A character with Shadow 2 picking a tier-3 lock has a 30% chance. An untrained character (rank 0) facing a tier-1 lock has a 30% chance.

| Rank − difficulty | −3 or less | −2  | −1  | 0   | +1  | +2 or more |
| ----------------- | ---------- | --- | --- | --- | --- | ---------- |
| Chance            | 5%         | 10% | 30% | 50% | 70% | 90%-95%    |

This replaces separate formulas for Sneak, Security, Escape, Keen and Fortitude. Developers add a new kind of check by saying which Discipline it uses and giving the target a tier. Players learn one table.

## Gear

Every item is described by two fields: a **type** (which says which Discipline gates it and how it behaves) and a **tier** (1 to 5). Damage, armor, speed and penetration come from the type's profile scaled by tier. A unique item can override any of these, but most items are just a name, a type and a tier.

This means the developer writes a handful of type profiles instead of balancing every item, and a player can look at "tier-3 hammer" and know what it does.

### Weapons

The Proposal A weapon table carries over, with the weapon's technique belonging to the weapon type and the Discipline deciding how good you are with it. An axe and a hammer both use Heavy, but an axe Sunders and a hammer Staggers, so the choice of weapon still matters to a Heavy fighter.

| Type       | Discipline | Speed            | Penetration | Technique                                                          |
| ---------- | ---------- | ---------------- | ----------- | ------------------------------------------------------------------ |
| Sword      | Finesse    | 1 attack/round   | Low         | Riposte: chance to counterattack after being attacked              |
| Knife      | Finesse    | 2 attacks/round  | None        | Backstab: ignores armor from stealth or against a staggered target |
| Axe        | Heavy      | 1 attack/round   | High        | Sunder: chance to reduce the target's armor for a few rounds       |
| Hammer     | Heavy      | 1 attack/round   | High        | Stagger: chance to make the target skip its next turn              |
| Greatsword | Heavy      | 1 attack/round   | Medium      | Sweep: hits a second enemy                                         |
| Bow        | Marksman   | Aims, then fires | High        | Aim: charge for more damage, broken if you take damage             |
| Crossbow   | Marksman   | 1 attack/2 rounds| Very high   | Bolt: no aim needed, but slow to reload                            |

### Armor

There is a single armor value. Damage is:

```
damage = max(1, weapon_damage − max(0, armor − penetration))
```

Proposal A's trade-off (fast weapons lose to armor, slow weapons ignore it) is still there through penetration and speed. Dropping damage types removes two thirds of the armor data and the "which damage type am I dealing?" question. Axes still feel distinct through Sunder rather than through a damage type.

Armor comes in two weights, so there is no Light/Medium/Heavy Armor skill:

- **Light armor** needs no Discipline. Anyone can wear it at any tier their level allows. It gives less armor.
- **Heavy armor** needs Guard at its tier. It gives more armor, but it's loud: you can't sneak in it, and Finesse's dodge doesn't work in it.

That's the whole armor system. The thief wears light armor because heavy armor would stop them sneaking, not because of a skill that sets their armor tier.

### Focuses

Spellcasters must hold a focus in one hand to cast, as in Proposal A. A focus is gear like any other: a tier-3 wand needs Arcane 3. This keeps A's equipment tax on casting, and it means magic and weapons follow the same gear rule.

## Spells

Spells keep Proposal A's spellbook model, minus the slot system:

- Each spell has a school (Arcane, Divine or Occult) and a tier.
- You learn a spell by using its spellbook, which needs your rank in the school to be at least the spell's tier. Learning is permanent.
- There are no spell slots and no spell forgetter. MP is the limit: you can know any number of spells but can only cast so many before you rest.
- MP comes from magic ranks: each rank in Arcane, Divine or Occult adds to your MP pool. A character who dabbles in magic has a little MP. A dedicated caster has a lot.

Spell slots in Proposal A needed their own skill (Spellcraft/Memory), a forgetting NPC, and a "which spells should I keep" decision that players face out of combat, away from the fun. Making MP the limit moves that decision into combat, where it's "which spell is worth casting now?"

Spellbooks are still loot, and Craft can copy them (see Craft below), so a scribe is still valuable to the party and to themselves.

### Concentration

Proposal A has a Concentration skill. Proposal B replaces it with a rule: **taking damage interrupts any channeled spell or aimed shot.** This creates a dependency without a skill. Casters and archers are fragile while they charge, so they want a Guard character who can Taunt and Intercept. That is goal 3, delivered by one sentence of rules.

## Perk Tables

Every rank-1 perk is the Discipline's core mechanic. These are a first draft meant to show the shape. The exact perks should change in playtesting.

### Guard

All ranks: +10 max HP per rank.

| Rank | Perk                                                                      |
| ---- | ------------------------------------------------------------------------- |
| 1    | Wear heavy armor and shields.                                             |
| 2    | **Taunt**: force an enemy to target you for a few rounds.                 |
| 3    | **Fortitude**: Guard checks resist poison, bleeding and other conditions. |
| 4    | **Intercept**: chance to take a hit meant for an ally.                    |
| 5    | **Bulwark**: with a shield, reduce all incoming damage by its tier.       |

### Heavy

| Rank | Perk                                                                  |
| ---- | --------------------------------------------------------------------- |
| 1    | Wield heavy weapons and use their techniques.                         |
| 2    | **Cleave**: when you kill an enemy, make a free attack on another.    |
| 3    | Technique chance is doubled.                                          |
| 4    | **Momentum**: each consecutive hit on the same target deals more.     |
| 5    | **Execute**: double damage against sundered or staggered targets.     |

### Finesse

| Rank | Perk                                                                       |
| ---- | -------------------------------------------------------------------------- |
| 1    | Wield swords and knives and use their techniques.                          |
| 2    | **Dodge**: chance to evade attacks while not in heavy armor.               |
| 3    | **Dual Wield**: attack with your off-hand weapon too.                      |
| 4    | **Envenom**: poisons applied to your blades last longer and stack.         |
| 5    | **Flurry**: your attacks can crit for double damage.                       |

### Marksman

| Rank | Perk                                                                          |
| ---- | ----------------------------------------------------------------------------- |
| 1    | Wield bows and crossbows and use their techniques. Aim for up to 1 round.     |
| 2    | Aim for up to 2 rounds.                                                       |
| 3    | **Called Shot**: aimed shots against a Scout-marked target always crit.       |
| 4    | Aim for up to 3 rounds.                                                       |
| 5    | **Deadeye**: a fully aimed shot ignores armor.                                |

### Arcane, Divine, Occult

Magic Disciplines unlock spells through spellbooks, so their perk tables are shorter. Each rank lets you learn spells of that tier, use focuses of that tier, and adds MP.

| Rank | Arcane                                    | Divine                                                    | Occult                                         |
| ---- | ----------------------------------------- | --------------------------------------------------------- | ---------------------------------------------- |
| 1    | Use wands and staves. Learn tier-1 spells.| Use holy symbols. Learn tier-1 spells.                    | Use occult focuses. Learn tier-1 spells.        |
| 3    | **Quicken**: one spell per fight casts instantly. | **Sanctuary**: your heals also cleanse a condition. | **Second Servant**: keep two summons at once.  |
| 5    | **Overchannel**: spend double MP to cast a spell as if it were one tier higher. | **Deathward**: once per rest, an ally who would die drops to 1 HP instead. | **Raise**: a slain enemy can become your summon. |

Overchannel is Proposal A's upcasting idea, folded into a perk instead of a separate Faith skill. Deathward is a deliberate exception to permanent death, which makes the healer the most valued party member. See [Open Questions](#open-questions).

### Shadow

| Rank | Perk                                                                         |
| ---- | ---------------------------------------------------------------------------- |
| 1    | Sneak, pick locks and escape combat with Shadow checks.                      |
| 2    | **Disarm**: Shadow checks find and disarm traps.                             |
| 3    | **Melt Away**: re-enter stealth mid-combat if no enemy is targeting you.     |
| 4    | **Lift**: steal an item from a mob that hasn't noticed you.                  |
| 5    | **Vanish**: your whole party escapes with your Shadow check.                 |

### Scout

| Rank | Perk                                                                                    |
| ---- | --------------------------------------------------------------------------------------- |
| 1    | Spot hidden enemies and evaluate opponents with Scout checks.                           |
| 2    | **Campfire**: build a fire where the party rests and recovers faster. First aid.        |
| 3    | **Mark**: evaluating reveals a weakness, and the party deals extra damage to the target.|
| 4    | **Track**: find where nearby mobs are and which way they went.                          |
| 5    | **Watchful**: your party can't be ambushed.                                             |

### Craft

Craft covers everything Proposal A split into Smith, Alchemy, Scribe and Enchant. **Craft rank N lets you make any recipe of tier N or lower**, given its materials. What you make depends on the recipe (a sword, a potion, a scroll, an enchantment), not on a separate skill. Copying a spellbook needs the original book and Craft at the spell's tier, but no rank in the spell's school.

| Rank | Perk                                                                 |
| ---- | -------------------------------------------------------------------- |
| 1    | Craft and repair tier-1 recipes.                                     |
| 3    | **Salvage**: break items down for some of their materials.           |
| 5    | **Masterwork**: crafted items have a chance to come out one tier up. |

Lumping crafting into one Discipline means a crafter is always useful to their party, whatever the party needs that day, and it keeps new crafting types (fletching, leatherworking) to new recipes rather than new skills.

## What the Party Looks Like

The rules don't define classes, but the Disciplines point to familiar roles, each of which needs the others:

- **The tank** (Guard + Heavy or Finesse) protects casters and archers, who get interrupted when hit.
- **The striker** (Heavy, Finesse or Marksman) does the damage, and does more when a Scout Marks a target or a Heavy ally staggers it for a Knife's Backstab.
- **The caster** (one or two magic schools) heals, blasts or summons, and needs a tank to keep channeling.
- **The specialist** (Shadow, Scout, Craft) gets the party through locked doors, spots ambushes and keeps gear repaired.

Every Background pushes at least one Utility rank onto each character, so most parties have some coverage of locks, spotting and crafting even when nobody built for them.

## Compared to Proposal A

| | Proposal A | Proposal B |
| - | - | - |
| Attributes | 5, set at creation | None (Backgrounds instead) |
| Skills | ~27 | 10 Disciplines |
| Rank cap rules | 2 (attribute and level) | 1 (level) |
| Check formulas | One per skill | One for all checks |
| Armor values | 3 (Slash, Pierce, Crush) | 1 |
| Per-item fields | Damage, type, speed, penetration | Type and tier |
| Spell limits | Spell slots, Spellcraft skill, forgetter NPC | MP |
| Concentration | A skill | A rule (damage interrupts) |
| Utility points | Possibly a separate pool | Backgrounds |

### What this gives up

- **Fine-grained weapon specialization.** A sword-only fighter can't be cheaper than a sword-and-knife fighter. Weapon types still differ in play through techniques, but the "master of every weapon" fantasy is now 3 Disciplines rather than 5 skills.
- **Natural talent.** Without attributes, nothing says one character was born stronger than another. Backgrounds carry some of that identity, but they don't cap anything.
- **Music.** Proposal A's songs have no home here. They could come back as Divine spells cast out of combat, or as their own Discipline if bards are a fantasy worth the extra line.
- **Crafting depth.** A single Craft Discipline is less of a career than four separate crafting skills.

## Open Questions

- Is 1 rank per level enough to feel progression, or should levels alternate between a rank and a perk-free bonus (such as HP)?
- Should Deathward exist? It softens permanent death, which could undercut goal 4, but it also makes a Divine character the person everyone wants in their party.
- Does Craft need to be split (for example Forge for gear and Brew for potions and scrolls) to make crafters feel like a real build?
- Should magic Disciplines get perks at ranks 2 and 4 too, or is learning a new tier of spells enough?
- Is 50%/20% the right check curve? It makes a one-rank gap a big deal, which keeps tiers meaningful but may feel swingy.
