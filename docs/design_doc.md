# Castle Recurse - Design Proposal

## Motivation

Castle Recurse is a multi-user dungeon game where players create characters and go dungeoneering out in the wilds of Brooklyn and beyond. Players of the game create custom characters. Depending on how a character is built, some characters will be better at some things than other characters. The purpose of this system is:

1. It allows players to fulfill certain fantasies
2. It allows players to indicate how they would like to play the game
3. It creates dependencies among players, because a character cannot be good at everything, so they must rely on other players to succeed
4. It creates a variety of experience in a game in which death is permanent

## Experience and Levels

Characters earn experience (XP) by adventuring. A character's level is determined by the total amount of XP they have earned over their lifetime. Spending XP (on attributes and skills, see below) does not lower a character's level.

Going from level 1 to level 2 requires 756 XP. Each level after that requires 1.26 times the XP of the level before it.

| Level     | 1           | 2   | 3   | 4    | 5    | 6    |
| --------- | ----------- | --- | --- | ---- | ---- | ---- |
| XP to reach | n/a (start) | 756 | 952 | 1200 | 1512 | 1905 |

There is no hard level cap. Because every formula in this document scales off of level and rank, the system extends across any level range, and the level range of the game is decided by the content we build.

## Attributes

Each character has a set of Attributes which reflect their natural characteristics. These attributes are:

Vitality (VIT)
Strength (STR)
Agility (AGI)
Intelligence (INT)
Faith (FTH)

Attributes do nothing on their own. Their only purpose is to cap the rank a character can have in each of the skills that the attribute governs. All other character upgrades happen through skills.

An attribute cannot be higher than 4 + character level. At level 1 an attribute can be at most 5, and at level 10 it can be at most 14.

Attributes are bought with XP. Raising an attribute is expensive on purpose, and every time a character raises an attribute, the cost of raising each of their *other* attributes goes up. This discourages characters from raising every attribute, and pushes them towards a coherent cluster of attributes that fits the role they want to play.

## Skills

Skills represent a character's competencies. Each character has a rank in a skill which represents how good a character is in the action associated with that skill. Skills are where nearly all of a character's power comes from.

Some skills, such as Evade, are directly used in calculating how successful a character is at performing an action. Other skills, such as Swords, are associated with gear and determine what level of gear a character is able to equip.

### Progressing Skills

Skill ranks are bought with XP. Each skill has its own base cost `n`, which is the cost of going from rank 0 to rank 1. Each rank after that costs 1.26 times the rank before it:

```
cost of rank r = n * 1.26^(r - 1)
```

| Rank               | 1   | 2   | 3   | 4   | 5   | 6   |
| ------------------ | --- | --- | --- | --- | --- | --- |
| Cost (for n = 100) | 100 | 126 | 158 | 200 | 252 | 317 |

Every skill uses the same curve, but because each skill has a different `n`, we can balance skills individually. For example, Cantrips and Daggers are cheaper than other skills of their kind, and Marksmanship is more expensive.

A skill's rank cannot be higher than the character's score in the skill's governing attribute. This means a character's attributes decide what skills they can master, while still allowing them to dabble in anything.

Since the cost curve is non-linear, skills a few ranks below a character's cap are much cheaper than their main skills. This leaves room for supplemental skills. For example, a level 8 fighter with 12 STR who puts most of their XP into Swords might also have 8 ranks in Sneak. Sneak 8 isn't competitive with a dedicated thief, but it is good value: with the same `n`, rank 12 of Swords costs around 2.5 times as much as rank 8 of Sneak.

### Skill List

Vitality (VIT)
- Fortitude - Each rank adds 25 max HP. Higher ranks also make you better at resisting ailments like poison and disease (each rank reduces the chance of an ailment taking hold by 20%)
- Heavy Armor - Higher ranks allow you to wear higher-level heavy armor
- Survival - Lets you make campfires and effigies that restore the party outside of combat. How this scales is still open (see Open Questions)
- Taunt - Higher ranks make enemies more likely to give in to your taunts and attack you

Strength (STR)
- Swords - Higher ranks allow you to equip higher-level swords and increase your hit and crit chance with them
- Axes - Higher ranks allow you to equip higher-level axes and increase your hit and crit chance with them
- Bludgeoning - Higher ranks allow you to equip higher-level hammers, maces and staves and increase your hit and crit chance with them
- Shields - Higher ranks allow you to equip higher-level shields
- Blacksmithing - Higher ranks allow you to craft and repair higher-level gear. Blacksmithing does not cover ranged weapons
- Power Attack - Higher ranks improve your power attacks

Agility (AGI)
- Marksmanship - Higher ranks allow you to equip higher-level bows and crossbows and increase your hit and crit chance with them. Also covers fletching (crafting bows and ammunition)
- Daggers - Higher ranks allow you to equip higher-level daggers and increase your hit and crit chance with them
- Dual Wield - Higher ranks reduce the penalties for using a weapon in your off hand
- Sneak - Higher ranks means sneaking is more successful
- Evade - Higher ranks increase your dodge chance and your escape chance during combat
- Light Armor - Higher ranks allow you to wear higher-level light armor
- Security - Higher ranks increase your chance of unlocking doors and disarming traps

Intelligence (INT)
- Keen - Higher ranks make you better at spotting hidden enemies. Keen also lets you evaluate an opponent: you can see a monster's info if your Keen is greater than or equal to the monster's level
- Spellcraft - Each rank adds 10 max mana and 1 spell slot. Metamagic spells (dispel, doublecast, etc.) have Spellcraft requirements
- Scribe - Higher ranks allow you to craft spellbooks and scrolls of higher-level spells
- Alchemy - Higher ranks allow you to craft higher-level potions and poisons
- Enchanting - Higher ranks allow you to craft higher-level magical accessories, like rings and amulets
- Psychic Magic - Higher ranks let you learn higher-level spells related to crowd control and utility
- Cantrips - Higher ranks let you learn higher-level quick, quality-of-life spells

Faith (FTH)
- Meditation - Higher ranks increase the mana you recover while meditating
- Willpower - Higher ranks increase your chance to deflect spells (see Spell Deflection) and make you better at maintaining concentration
- Evocation - Higher ranks let you learn higher-level spells related to blasting things with magic
- Druidry - Higher ranks let you learn higher-level spells related to healing and nature magic
- Conjuration - Higher ranks let you learn higher-level spells related to summoning friends and monsters

## Character Creation

When a player creates a character, they choose a race and a starting class.

- Race gives +1 to two attributes, and +1 rank in several unrelated skills that suit different builds.
- Starting class gives between +1 and +4 to each attribute, and +3 or +4 ranks in two or three skills, so that the character is workable straight away.

The system is freeform: any character can build towards any role. But a character whose race and starting class don't line up with the build they want pays an XP tax, which can be significant, to catch up on the attributes and skills they didn't start with.

## Guilds

Characters can join a guild. Each guild has an NPC who inducts new members. To join, a character usually has to complete a quest, pay a large amount of gold or an expensive item, and have enough ranks in a cluster of skills that shows they know the guild's core skills. Guilds are mutually exclusive, so a character can only be in one guild.

Joining a guild gives access to a guild skill that only members of that guild can rank up. For example:

- Necromancer guild members can learn Black Magic.
- Bard guild members can learn Performance, which lets them perform music and poetry. Performances are out-of-combat rituals that give the party boons before a fight.

Guild skills are not capped by an attribute. They are capped by the character's rank in the guild, which starts at a value like 8. As a character meets harder and more painful skill requirements, they can rise in the guild, and a Master of the guild can rank their guild skill up to 16. Like everything else, this lets guilds extend across any level range.

Guild skills give large bonuses to specific party roles. Most percentage-based increases to output (for example, +X% damage) come from guild skills, and stack on top of what a character gets from their core skills. Some guild skills can also grant part of their bonus to allies in the same room. These shared bonuses do not stack with copies of the same bonus.

Guilds serve four purposes:

1. A good party includes members of different guilds, to get a good mix of guild skills.
2. Players have meaningful choices in how they build their character around their guild, and in how they spend their extra XP across the rest of the skill list.
3. Guild requirements can make players take skills they otherwise would skip.
4. Guild requirements cut across attributes, which lets us balance hybrid builds. The core skill system gives non-magic characters the best damage per round, gives spellcasters the best utility, and makes hybrid characters like battlemages inefficient. Guilds then nudge players into a spread of builds, including some that require hybrids.

## Core Resources

```
HP    = 40 + (level * 10) + (Fortitude * 25)
Mana  = 40 + (Spellcraft * 10)
Spell slots = 2 + Spellcraft
```

## Combat

### Attacks

Each attack rolls in this order:

1. The target rolls to evade the attack (see Checks).
2. The attacker rolls to crit. Crit chance is based on the attacker's rank in the skill of the weapon they are using. A crit multiplies damage by 1.5.
3. If the attack did not crit, the target rolls to deflect it (see Deflection).
4. The target's damage resistance is subtracted from the damage.

Weapon skills replace agility in all of the to-hit and crit math. An attacker's to-hit value is:

```
to_hit = level + (weapon_skill * 2)
```

The damage formula is:

```
damage = max(floor, (weapon_damage * crit_multiplier * deflection_multiplier) - damage_resistance)
floor  = max(1, level / 2)
```

Strength no longer adds damage. A character's damage comes from the weapons their skills let them equip.

Since a character with low rank in a weapon skill is stuck using lower-level weapons, they face two penalties against high-level enemies: their weapons deal less damage, and their attacks are deflected more often. The damage floor means they still deal some damage. For example, a wizard with a few ranks in Bludgeoning to use a staff won't deal 0 damage with their basic attack in the late game.

### Deflection

Armor and shields have a deflection value, and weapons have a penetration value. When a character is hit, they roll to deflect:

```
deflection_chance = (deflection - penetration) * 5%
```

Penetration runs from around 2 to around 22 across the level range. Light armor deflection runs from around 2 to around 22, and heavy armor deflection runs from around 12 to around 32. Armor deflection grows linearly with the armor's level, so we have full control over how player deflection scales across the level range.

A successful deflection reduces the damage of the attack by 40%. Armor and shields can add to this:

- Light armor adds up to +5%.
- Heavy armor adds up to +10%.
- Shields add up to +5%.

For example, a character wearing high-level heavy armor and a shield would reduce the damage of a deflected attack by 55%.

Spells use a similar roll against the target's Willpower (see Spell Deflection).

Heavy armor comes with penalties: -2 to Evade for dodging, and -4 to Sneak and to Evade for escaping. Penalties cannot reduce a skill below 0. Light armor has no penalties and some high-level light armor gives bonuses to Sneak or Evade, but it has lower damage resistance than heavy armor.

Bosses and most dungeon mobs have around 10 more deflection than other mobs of their level. This means a dedicated martial is needed to deal real damage in a dungeon, while a pure caster can still damage regular mobs with their meagar weapons in the overworld.

### Power Attacks

A power attack is a one-off action. It takes a turn to charge, just like a spell with a cast time. If the attacker is hit while charging, they must pass a concentration check (see Checks), or the power attack is interrupted.

To be worth the turn spent charging, a power attack should hit for more than twice as much as a regular attack. A successful power attack also triggers the weapon's technique (see Weapon Techniques). The Power Attack skill improves power attacks, but exactly how is still open (see Open Questions).

## Checks

Every check in the game uses the same formula, so that all checks are easy to reason about and tune:

```
chance = clamp(K * (rank / (rank + difficulty)), min_chance, max_chance)
```

`rank` is the value of the character making the check. `difficulty` is either a fixed value (like the difficulty of a lock) or the opposing character's value in a contested check. `K` is a constant set per check, which lets us tune how hard each kind of check is.

| Check         | Rank                       | Difficulty                                      | Modifiers                                                    |
| ------------- | -------------------------- | ----------------------------------------------- | ------------------------------------------------------------ |
| Evade         | 5 + level + (Evade * 2)    | Attacker's to-hit                               | Heavy armor penalty                                          |
| Escape        | 5 + level + (Evade * 2)    | Highest to-hit of the hostile mobs in the room  | Heavy armor penalty, escape cooldown                         |
| Stealth       | Sneak                      | Highest Keen of the hostile mobs in the room    | Armor penalty, escape cooldown, enemy alertness              |
| Taunt         | Taunt                      | Target's level                                  |                                                              |
| Concentration | Willpower                  | Damage taken                                    | Weapon interrupt factor (for power attacks)                  |
| Security      | Security                   | Lock or trap difficulty                         |                                                              |

Concentration is rolled whenever a character is hit while charging a spell or a power attack. On a failure, the spell or power attack is interrupted.

### Alertness

Each mob has an alertness value between 0 and 1. A mob's alertness goes to its max whenever the mob attacks, casts a spell, or takes damage, and it cools down by a fixed amount each tick.

When a character tries to hide, the stealth chance is multiplied by `(1 - alertness)` using the highest alertness of the hostile mobs in the room. This means that a character can't slip into stealth in the middle of a fight. They have to break away and wait for the enemies to calm down.

### Escape Cooldown

Each mob has an escape chance between 0 and 1, which starts at 1. Escape checks and stealth checks are multiplied by the escape chance.

Whenever a character attempts to escape, or fails an attempt to hide, their escape chance drops to 0. It recovers by 0.1 each tick. This means characters can escape easily from a first encounter, but they can't flit from room to room without waiting for their escape chance to build back up.

## Weapons and Armor

Weapons are varied in their form and function, making them better in different circumstances. This allows fighters who specialize in different weapons to feel like their characters are different. It also allows fighters who specialize in a variety of weapon forms to be true weapon-masters who have a tool for different situations (just like how spellcasters have a variety of schools of magic to train in).

### Weapon properties

1. Damage - How hard does it hit
2. Speed - How many attack rolls does the weapon make per round
3. Penetration - How much the weapon reduces the target's deflection chance

Magical weapons, especially swords, can also give stat bonuses to the wielder, such as Willpower, evade chance, Fortitude, or mana.

### Weapon Techniques

Each weapon type has an associated technique. Techniques trigger when the wielder lands a successful power attack. Techniques are what make power attacks worth charging, and they are what make each weapon type feel different from the others.

Two-handed weapons are far less than twice as powerful as one-handed weapons with their regular attacks, but they have considerably stronger techniques than their one-handed siblings.

### The Weapon Table

| Weapon                  | Skill        | Speed     | Penetration | Technique                                                          | Notes |
| ----------------------- | ------------ | --------- | ----------- | ------------------------------------------------------------------ | -- |
| Sword                   | Swords       | Medium    | Medium      | | A versatile all-rounder. |
| Axe                     | Axes         | Slow      | Low         | Sunder: Reduces the target's deflection and damage resistance for N rounds | Sunder exposes targets for the whole party. |
| Hammer                  | Bludgeoning  | Slow      | High        | Stagger: Stuns the target, skipping their next turn                |  |
| Dagger                  | Daggers      | Fast      | Low         | Backstab: When used from stealth or against a distracted or staggered target, the attack can't be deflected and ignores damage resistance | Slightly weaker than swords, axes, and hammers. No two-handed daggers. Weak power attacks outside of Backstab. Smaller off-hand penalty than other weapons. Allows thieves to deal increased damage when the party is distracting foes for them. |
| Short bow / Crossbow    | Marksmanship | Slow      | High        | Power Shot: A charged shot that deals increased damage             | Easy to keep a power shot charged while being hit. Regular attacks are competitive with other weapons. |
| Longbow / Greatbow      | Marksmanship | Very Slow | High        | Power Shot: A charged shot that deals greatly increased damage     | Much harder to keep a power shot charged while being hit (a high interrupt factor), but power shots deal far more damage. |

### Weapon Builds

Swords, axes, and hammers are stronger than daggers and bows, but they all sit in STR. Daggers, bows, and Dual Wield all sit in AGI. This creates a spread of builds with different costs:

- Two-handed STR weapons only need one attribute, but they are weaker with regular attacks.
- Twin daggers and bows only need one attribute (AGI), but the weapons are weaker.
- Dual-wielding a pair of STR weapons is the most powerful, but it is expensive because it pushes two attributes (STR for the weapons and AGI for Dual Wield).

## Spells

Spellcasters have a number of spell slots determined by their Spellcraft rank. Some powerful high-level spells, like strong nukes and summons, occupy 2 slots.

Spells are learned by using a spellbook. Once you use a spellbook, you instantly know the spell, and it occupies a spell slot.
You can forget a spell by speaking to a spell forgetter. This will allow you to learn other spells, but if you want to learn the spell again, you will need to find another spellbook of the same kind (which means that scribes are now more helpful, even to just themselves).

Each spellbook has a skill rank and a skill associated with it. To learn a spell from a book, you must have the rank and skill required. For example, if Firebolt is a Level 2 Evocation Spell, then you must have at least 2 in Evocation to learn it.

Spellcasters must equip a magic focus item in one of their hands in order to be able to cast, so they are not constantly holding a spellbook, but the focus requirement maintains the equipment tax on spellcasting.

### Schools of Magic

Each school of magic has a spread of things it is better and worse at. A caster who learns one school gets a lot of new options. Each additional school gives diminishing returns, because the caster ends up using only the best spells from each school.

- Psychic Magic (INT) - Mostly crowd control, with some odd damage spells and out-of-combat utility.
- Evocation (FTH) - The best mana-to-damage conversion of any school, but poor utility.
- Druidry (FTH) - Much better at healing than other schools, and much worse at dealing damage or crowd control. It might have the odd attack spell, like a sunbeam or grasping roots.
- Conjuration (FTH) - Summoning friends and monsters.
- Cantrips (INT) - Eclectic quality-of-life spells with fast cast times, like Mage Weapon, Mage Armor, and Unlock.

Casters reach their peak by mixing INT and FTH schools. Guild requirements are what break this up, so that not every caster can have it all.

Metamagic spells, like dispel and doublecast, require ranks in Spellcraft.

### Spell Deflection

Damaging spells have a spell penetration value, which is rolled against the target's Willpower:

```
spell_deflection_chance = (Willpower - spell_penetration) * 5%
```

Each spell that targets a mob decides in its implementation whether it rolls for deflection, and what happens when it is deflected. Some spells, like Mind Control, fizzle and fail to take effect. Others, like Firebolt, deal reduced damage. By default, a deflected damaging spell has its damage reduced by 60% (the `SPELL_DEFLECTION` constant).

### Meditation

A character can use the Meditate command to enter a meditative state, in which they recover mana each tick and earn a small amount of XP. Being attacked or damaged forces a character out of meditation. Higher Meditation ranks increase the mana recovered each tick.

Like the escape cooldown, meditation has a hidden failure chance. Each time a character's meditation ends for any reason, their chance of failing to enter meditation next time goes up. Resting in town resets it to 0.

## Open Questions

- Power attacks: How much damage should a power attack deal compared to a regular attack? Should a successful power attack always trigger the weapon's technique, or only have a chance to? What does the Power Attack skill scale (damage, technique strength, charge reliability)?
- Survival: How does Survival scale? A campfire that slowly heals HP isn't very useful to a high-level party. One option is to give Survival cooking, so that survivalists cook meals at their campfire that restore the party and give buffs, and the meals get better at higher ranks.
- Should poisons belong to Alchemy or Survival?
- Should metamagic be reserved for a Wizard guild?
- Checks: What should `K`, the difficulty values, and the min/max chance be for each check? Should `K` multiply the chance, or weight the difficulty (the old evasion formula was `T / (T + 6 * A)`, which weights the difficulty)?
- Spell potency: Spell damage and condition durations used to scale with FTH. Now that attributes do nothing on their own, what scales them? The rank in the spell's school is the obvious candidate.
- Attribute offset: Should max attributes be offset higher (for example, 10 higher than 4 + level), with the first 10 ranks of many skills being weak and starting classes giving rank 10 in their skills?
- What is the base value of each attribute before race and starting class bonuses?
- Which races, starting classes, and guilds should the game have?
