# Castle Recurse - Character Design (Disciplines)

Castle Recurse is a multi-user dungeon game where players create characters and go dungeoneering out in the wilds of Brooklyn and beyond. This document describes how characters are built and how they grow: what they can be good at, the gear they use, how they fight and cast spells, and how a party recovers in the wild. It is both the design rationale and the reference for implementing the system.

## Goals

The character system exists to:

1. **Fulfill fantasies.** A player can name what they want to be ("the tank", "the thief", "the healer") and build it.
2. **Let players say how they want to play.** Builds decide which parts of the game a character engages with.
3. **Make players depend on each other.** No character can be good at everything, so a party needs each other to succeed.
4. **Create variety in a game with permanent death.** Each new character should be a chance to play differently.
5. **Keep every choice legible.** A player reading their character sheet in a text terminal can tell what each point they spent does. A choice that can't be explained in one sentence isn't doing work.
6. **Keep every rule authorable.** Each mechanic is content the developers have to write, balance and test. A small set of rules that combine well beats a large set of rules that each need their own data.

Permanent death makes goal 5 matter more than usual. Players will make characters often, so creation and leveling have to be quick to understand.

### Design principles

These follow from the goals, and explain choices made throughout the document:

- **One kind of character choice.** Characters spend ranks in Disciplines, and that's the whole build. There are no attributes. A limited rank budget already stops a character from mastering everything, so a second layer of permanent choices at creation would only make players decide before they understand the game.
- **Broad Disciplines.** Each Discipline covers a whole fantasy, not one weapon or one action. Splitting weapons finely makes skills into gear gates and makes the "weapon master" fantasy cost most of a character's points.
- **A rank is one number.** Everything a Discipline does reads its rank directly. There are no perks, unlocks or per-skill formulas.
- **Items are a type and a tier.** Developers write a handful of type profiles instead of balancing every item.
- **Dependency through rules, not extra skills.** Where possible, interdependence comes from a rule (casters are interrupted when hit) rather than from adding a skill.

## Characters

A character is made of:

- A **Background**, chosen at creation, which gives starting ranks, starting items and an identity.
- **Ranks in Disciplines.** There are 10 Disciplines, each ranked 0 to 5.
- **A level**, from 1 to 12.
- **Gear**, each item described by a type and a tier (1 to 5).

There are no attributes. A character sheet is a name, a background, a level, HP/MP, and at most ten lines of Disciplines.

## Disciplines

A Discipline at rank N lets the character:

- **Use gear:** equip that Discipline's gear up to tier N.
- **Roll checks:** attempt anything the Discipline covers at rank N (see [Checks](#checks)).
- **Scale:** any number the Discipline controls is a simple multiple of N.

| Discipline | Group   | Gear (up to tier N)        | What rank N does                                                                                   |
| ---------- | ------- | -------------------------- | -------------------------------------------------------------------------------------------------- |
| Guard      | Martial | Heavy armor, shields       | +10 max HP per rank. Guard checks to [taunt](#taunt) and to resist poison and other conditions.     |
| Heavy      | Martial | Axes, hammers, greatswords | 10% technique chance per rank. `mend` gear of tier N or lower at [camp](#camp).                    |
| Finesse    | Martial | Knives                     | 10% technique chance per rank. 5% dodge chance per rank when not in heavy armor.                   |
| Arms       | Martial | Swords, bows, crossbows    | 10% technique chance per rank. [Aim](#weapons) for up to N rounds.                                  |
| Arcane     | Magic   | Wands, staves              | Learn Arcane spells up to tier N: blasting and crowd control. +10 max MP per rank.                 |
| Divine     | Magic   | Holy symbols               | Learn Divine spells up to tier N: healing, protection and blessings. +10 max MP per rank.          |
| Occult     | Magic   | Skulls, ritual knives      | Learn Occult spells up to tier N: summons, curses and draining. +10 max MP per rank.               |
| Shadow     | Utility | none                       | Shadow checks to sneak, pick locks, disarm traps and escape combat.                                |
| Scout      | Utility | none                       | Scout checks to spot hidden enemies and evaluate opponents. Build a tier-N campfire at [camp](#camp). |
| Herbalism  | Utility | none                       | Herbalism checks for first aid. `tend` conditions and `brew` potions of tier N or lower at [camp](#camp). |

A player can read "Finesse 3" on their sheet and know exactly what they have: tier-3 knives, a 30% technique chance and 15% dodge. A developer adds a new use for a Discipline by saying which number it reads.

Heavy is the one Martial Discipline with a utility job: the person swinging a hammer is the obvious party smith. Arms has no utility job, but covers both melee and ranged weapons.

### Getting ranks

- At creation, a character picks a Background (2 ranks) and spends 1 more rank on any Discipline. They start with 3 ranks.
- Each level from 2 to 12 grants 1 rank. A level-12 character has 14 ranks.
- A Discipline's rank can't be higher than the cap for the character's level:

| Level    | 1-2 | 3-5 | 6-8 | 9-11 | 12 |
| -------- | --- | --- | --- | ---- | -- |
| Max Rank | 1   | 2   | 3   | 4    | 5  |

Many Disciplines gate gear, so the cap stops characters from using high-tier gear before they reach the matching level.

With 14 ranks and a cap of 5, a finished character can master two Disciplines and be strong in a third, or spread into a generalist with four or five Disciplines at rank 3. Nobody can cover all ten.

Ranks are permanent. There is no respec. Death is permanent too, so the way to try a new build is a new character.

### Backgrounds

A Background is the answer to "what did you do before you went into the dungeon?" It grants 2 ranks, **at least one of them in a Utility Discipline**, plus a starting item or two.

The Utility requirement exists because combat Disciplines are more tempting. Without it, most characters would be all-combat and nobody could open a lock. Backgrounds spread utility ranks across characters through a choice players want to make anyway, and give a new character an identity before their first level-up.

| Background        | Ranks                 | Starting items         |
| ----------------- | --------------------- | ---------------------- |
| Locksmith         | Shadow 1, Heavy 1     | Lockpicks              |
| Night Guard       | Guard 1, Scout 1      | Lantern, tier-1 shield |
| Seminary Dropout  | Divine 1, Herbalism 1 | Holy symbol, bandages  |
| Bodega Cat Sitter | Finesse 1, Scout 1    | Knife, bag of treats   |
| Fencing Captain   | Arms 1, Scout 1       | Tier-1 sword           |
| Hedge Witch       | Occult 1, Herbalism 1 | Ritual knife, herbs    |
| Bike Mechanic     | Heavy 1, Scout 1      | Wrench (tier-1 hammer) |

Backgrounds are pure data (a name, a description, two ranks, some items), so each server running its own MUD can write its own.

## Checks

Every non-combat roll uses one formula:

```
chance = clamp(50% + 20% × (rank − difficulty), 5%, 95%)
```

`rank` is the character's rank in the Discipline that covers the action. `difficulty` is the tier of the thing being attempted: a tier-3 lock, a tier-2 trap, a hidden tier-4 monster, a tier-1 poison.

| Rank − difficulty | −3 or less | −2  | −1  | 0   | +1  | +2  | +3 or more |
| ----------------- | ---------- | --- | --- | --- | --- | --- | ---------- |
| Chance            | 5%         | 10% | 30% | 50% | 70% | 90% | 95%        |

A character with Shadow 2 picking a tier-3 lock has a 30% chance. An untrained character (rank 0) facing a tier-1 lock also has a 30% chance.

To add a new kind of check, say which Discipline it uses and give its target a tier.

## Gear

Every item has two fields:

- **Type** says which Discipline gates the item and gives its profile (damage, armor, speed, penetration, technique).
- **Tier** (1 to 5) scales the profile.

A unique item can override any of its profile values, but most items are just a name, a type and a tier. A player who sees "tier-3 hammer" knows what it does.

### Durability

Gear wears down with use. Worn gear is less effective, and broken gear does nothing until it's repaired. A smith in town fully repairs anything for a fee. A Heavy character can partly repair gear in the field with `mend` (see [Camp](#camp)).

### Weapons

| Type       | Discipline | Speed             | Penetration | Technique                                                          |
| ---------- | ---------- | ----------------- | ----------- | ------------------------------------------------------------------ |
| Axe        | Heavy      | 1 attack/round    | High        | Sunder: chance to reduce the target's armor for a few rounds       |
| Hammer     | Heavy      | 1 attack/round    | High        | Stagger: chance to make the target skip its next turn              |
| Greatsword | Heavy      | 1 attack/round    | Medium      | Sweep: chance to hit a second enemy                                |
| Knife      | Finesse    | 2 attacks/round   | None        | Backstab: ignores armor from stealth or against a staggered target |
| Sword      | Arms       | 1 attack/round    | Low         | Riposte: chance to counterattack after being attacked              |
| Bow        | Arms       | Aims, then fires  | High        | Aim: charge for more damage, broken if you take damage             |
| Crossbow   | Arms       | 1 attack/2 rounds | Very high   | Bolt: no aim needed, but slow to reload                            |

- Techniques with a chance (Sunder, Stagger, Sweep, Riposte) use the wielder's technique chance from the weapon's Discipline: 10% per rank.
- Backstab and Bolt always apply when their condition is met.
- Aim can charge for up to the wielder's Arms rank in rounds.
- Only knives can be dual wielded. A knife in each hand doubles the attacks, and so the Backstabs.

Techniques belong to the weapon type, and the Discipline decides how often they trigger. An axe and a hammer both use Heavy, but an axe Sunders and a hammer Staggers, so the choice of weapon still matters within a Discipline.

#### Weapon roles

- **Heavy weapons** ignore armor through penetration and set up the party: Sunder weakens a target for everyone, and Stagger opens it to Backstab.
- **Knives** are openers. Each hit is weak and has no penetration, so a knife is the worst weapon in a straight fight against armor. But Backstab ignores armor entirely, so from stealth, or against a target a Heavy ally has staggered, a knife does more than any other weapon. Two fast hits also make knives the best way to deliver poisons.
- **Arms** is the soldier's kit. The bow is the plan: aim from the back while the tank holds the line. The sword is the fallback when an enemy reaches you and interrupts your aim, and its off hand can hold a shield.

### Armor

Each piece of armor, and each mob, has a single armor value. Damage per hit is:

```
damage = max(1, weapon_damage − max(0, armor − penetration))
```

Fast weapons lose more to armor because they pay the subtraction on every hit. Slow, high-penetration weapons mostly ignore it.

Armor comes in two weights:

- **Light armor** needs no Discipline. Anyone can wear it at any tier their level allows. It gives less armor.
- **Heavy armor** needs Guard at its tier. It gives more armor, but it's loud: you can't sneak in it, and Finesse's dodge doesn't work in it.

There is no separate armor skill for light armor. The thief wears light armor because heavy armor would stop them sneaking.

### Focuses

Spellcasters must hold a focus in one hand to cast. A focus is gear like any other: a tier-3 wand needs Arcane 3. The focus is an equipment cost on casting, the same way a weapon is for fighting.

## Combat Rules

### Interruption

**Taking damage interrupts any channeled spell or aimed shot.** Casters and archers are fragile while they charge, so they want a tank between them and the enemy. This one rule does the job a concentration skill would otherwise do, and creates the dependency directly.

### Taunt

A Guard character can taunt an enemy with a Guard check against the enemy's tier. On a success, the enemy attacks the taunter for the next few rounds.

## Spells

- Each spell has a school (Arcane, Divine or Occult) and a tier.
- A character learns a spell by using its spellbook, which needs their rank in the school to be at least the spell's tier. Learning is permanent.
- There are no spell slots. A character can know any number of spells. **MP is the limit:** each rank in a magic Discipline adds 10 max MP, and spells cost MP to cast.
- Spellbooks are loot. They can't be copied, so a character's spells are a record of where they've been.

Spell slots would need their own skill, a way to forget spells, and a "which spells do I keep?" decision made out of combat. With MP as the limit, the decision happens in combat, where it's "which spell is worth casting now?"

## Camp

There is no crafting. Nobody forges swords or writes scrolls, and there are no recipe lists. Towns cover the economy: a smith repairs gear and an apothecary sells potions.

Camp keeps the part of crafting that makes players need each other: **recovering together around a fire, far from town.** Camp is how a party stays out on a long expedition instead of walking back to town after every fight.

### How camp works

1. **A Scout builds the fire.** `camp` takes a few minutes and makes a campfire whose tier is the builder's Scout rank. Without a fire, a party recovers slowly and anything nearby can find them.
2. **The party rests at the fire.** Resting restores HP and MP faster than resting without a fire, and faster still at higher-tier fires. A wandering mob of a higher tier than the fire may still find the camp.
3. **Everyone else contributes while resting.** Any character at a fire can use their camp commands:

| Command | Discipline | Effect                                                                                                     |
| ------- | ---------- | ---------------------------------------------------------------------------------------------------------- |
| `tend`  | Herbalism  | Cure conditions (poison, bleeding, disease) of tier N or lower on party members.                           |
| `brew`  | Herbalism  | Turn herbs into a potion. The potion's tier is the lower of the herbs' tier and N.                         |
| `mend`  | Heavy      | Restore some durability to party members' gear of tier N or lower. It can't restore gear to full. That still needs a town smith. |

Camp commands always work if the character's rank is high enough. There's no roll, which keeps the fire a calm moment between fights.

### Outside of camp

- **First aid** is Herbalism's one action away from a fire: bandage an ally to stop bleeding or restore a little HP, with a Herbalism check against the wound's tier. It's weaker than Divine healing, but it doesn't cost MP.
- **Herbs** are found while exploring, like other loot. Higher-tier areas have higher-tier herbs. There's one kind of herb per tier, so there are no recipes to learn.
- **Potions** from the apothecary and from `brew` are the same items. Brewing doesn't make anything the town can't sell. It lets a party resupply without going back.

### Making camp matter

Scout and Herbalism are only as valuable as the problems they solve. If conditions are rare or resting in the wild is easy, they become dead ranks. Expeditions need to be long enough, and conditions harsh enough, that a party without a herbalist or a fire feels it. Mending is part of Heavy, so it never costs a fighter a wasted rank.

## Party Roles

The rules don't define classes, but the Disciplines point to familiar roles, each of which needs the others:

- **The tank** (Guard with Heavy or Arms) taunts enemies away from casters and archers, who are interrupted when hit. A Heavy tank is also the party's smith.
- **The striker** (Heavy, Finesse or Arms) does the damage. A knife fighter does far more when a Heavy ally staggers the target, and everyone hits harder against a sundered one.
- **The caster** (one or two magic schools) heals, blasts or summons, and needs a tank to keep channeling.
- **The specialist** (Shadow, Scout, Herbalism) gets the party through locked doors, spots ambushes, and keeps the party going at camp.

Every Background includes a Utility rank, so most parties cover some locks, spotting and camp even when nobody built for them.

## Tuning Values

Starting values for implementation. They're expected to change in playtesting.

| Value                    | Starting value                        |
| ------------------------ | ------------------------------------- |
| Check chance             | 50% + 20% × (rank − difficulty), 5%-95% |
| Technique chance         | 10% per rank                          |
| Dodge chance             | 5% per rank (Finesse, no heavy armor) |
| Guard HP bonus           | +10 max HP per rank                   |
| Magic MP                 | +10 max MP per rank in each magic Discipline |
| Max Aim                  | Arms rank, in rounds                  |
| Damage floor             | 1                                     |
| Starting ranks           | 3 (2 from Background, 1 free)         |
| Ranks per level          | 1 (levels 2-12)                       |

## Deliberately Left Out

These were considered and left out to keep the system small. Each could come back if playtesting shows it's missed.

- **Attributes.** A rank budget already limits builds. Backgrounds carry the "who you were" identity that attributes might.
- **Damage types.** Slash/Pierce/Crush would triple armor data for every mob and item, and players would need to track it through text. Weapon techniques make weapon types distinct instead.
- **Perks.** A rank that unlocks a named ability is exciting, but 50 perks is a lot to write and balance. Leveling gets its excitement from new gear tiers and new spellbooks instead.
- **Crafting.** Smithing gear, alchemy recipes, scribing and enchanting all need a recipe system and a materials economy. Scribing would also hand high-level spells to low-level characters. Camp keeps the cooperative part.
- **Spell slots.** MP limits casting instead.
- **Upcasting.** Casting a spell at a higher tier for more effect has no home without perks.
- **Dual wielding swords.** Only knives can be dual wielded, so knives have a clear advantage over swords.
- **Music.** Songs could come back as Divine spells, or as a camp command that gives the party a boon for the next fight.

## Open Questions

- What is a character's base HP and MP, and how much does level add?
- Is 1 rank per level enough to feel progression without perks, or should levels also give a flat bonus (such as HP)?
- Should Scout's fire and Herbalism's first aid stay in different Disciplines, or should a single Survival Discipline cover both, at the cost of the three-person campfire?
- How harsh should gear wear and conditions be? Scout and Herbalism are only worth ranks if a party without them suffers.
- Finesse's only extra is dodge, while Heavy mends and Arms covers two weapon styles. Is Finesse strong enough?
- Is 50%/20% the right check curve? It makes a one-rank gap a big deal, which keeps tiers meaningful but may feel swingy.
