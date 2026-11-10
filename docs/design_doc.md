# Castle Recurse - Design Proposal A (Skill-Buy System)

## Motivation

Castle Recurse is a multi-user dungeon game where players create characters and go dungeoneering out in the wilds of Brooklyn and beyond. Players of the game create custom characters. Depending on how a character is built, some characters will be better at some things than other characters. The purpose of this system is:

1. It allows players to fulfill certain fantasies
2. It allows players to indicate how they would like to play the game
3. It creates dependencies among players, because a character cannot be good at everything, so they must rely on other players to succeed
4. It creates a variety of experience in a game in which death is permanent

## Attributes

Each character has a set of Attributes which reflect their natural characteristics. These attributes are:

Vitality (VIT)
Strength (STR)
Agility (AGI)
Intelligence (INT)
Faith (FTH)

Attributes values range from 1 - 5. They are decided once at character creation and cannot be changed afterwards.

## Skills

Skills represent a character's competencies. Each character has a rank in a skill which represents how good a character is in the action associated with that skill. 

Some skills, such as Evasion, are directly used in calculating how successful a character is at performing an action. Other skills, such as Blades, are associated with gear and are used to determine what level of gear a character is able to equip. 

### Progressing Skills

Characters have a max level of 12.

Each character starts with X skill ranks (either chosen at character creation, or determined by race / archetype). 
When a character levels up, they get Y skill point(s) that they can use to level up a skill of their choice. 

I was originally thinking that characters should get 1 skill point per level, but after looking at the number of skills, I think we might need to do 2 to 3 points.
Also, Claude suggested that we split skill points into Combat and Utility pools, so that some of your points are forced to go towards Utility skills. I don't love it from an aesthetic POV, but it might be good.

The max rank of a skill is determined by two rules.
1. A skill's rank cannot be higher than the character's associated Attribute score. This rule means that a character's attributes regulate what skills they can be masters at, while still allowing them to dabble. 
2. The max rank of a skill cannot be higher than 1 + floor[level / 3]. Since many skills control gear-level, this rule prevents characters from getting access to higher-tier gear until they are at the appropriate level.

| Level    | 1-2 | 3-5 | 6-8 | 9-11 | 12 |
| -------- | --- | --- | --- | ---- | -- |
| Max Rank | 1   | 2   | 3   | 4    | 5  |

### Skill List

Vitality (VIT)
- Fortitude - Higher ranks means you are better at resisting conditions like poison
- Armor - Higher ranks allow you to wear higher-level heavy armor.
- Survival - Lets you make campfires and do first aid? I haven't yet figured out how this scales, or what this mechanic looks like

Strength (STR)
- Smith - Higher ranks allow you to craft and repair higher-level gear
- Block/Shields - Higher ranks allow you to equip higher level shields
- Sword - Higher ranks allow you to equip higher level swords
- Hammers - Higher ranks allow you to equip higher level hammers
- Axes - Higher ranks allow you to equip higher level axes

Agility (AGI)
- Sneak - Higher ranks means sneaking is more successful
- Evade - Higher ranks increase dodge chance during combat. Higher ranks also allow you to wear higher-level light armor (eh oh this, but I don't like the light armor/med armor/heavy armor skills).
- Escape - Higher ranks increase escape chance during combat
- Security - Higher ranks increase your chance of unlocking doors and disarming traps
- Bows - Higher ranks allow you to equip higher level bows
- Knives - Higher ranks allow you to equip higher level knives
- Dual Wielding - Higher ranks reduce the penalties for using a weapon in your off hand

Intelligence (INT)
- Keen - Higher ranks means you are better at "evaluate opponent" (possibly this reveals weaknesses in addition to telling how strong the opponent is relative to you) and you are better at spotting hidden enemies
- Scribe - Higher ranks allow you to craft spellbooks and scrolls of higher level spells
- Alchemy - Higher ranks allow you to craft higher level potions and poisons
- Spellcraft/Memory - Higher ranks increase your spell slots
- Illusion - Higher ranks let you learn higher level spells related to crowd control and good utility
- Necromancy - Higher ranks let you learn higher level necromancy spells

Faith (FTH)
- Music - Higher ranks let you learn higher level songs? And songs are an out-of-combat ritual that can provide boons to parties before a fight?
- Enchant - Higher ranks allow you to enchant higher level gear / make higher level enchantments?
- Concentration - Higher ranks make you better at maintaining concentration
- Evocation - Higher ranks let your learn higher level spells related to blasting things with magic
- Conjuration - Higher ranks let you learn higher level spells related to summoning friends and monsters. And maybe food?
- Holy/Prayer/Life Magic - Higher ranks let you learn higher level spells related to healing and divine magic

These skills and their placements are still up in the air, just like many things in this document. Examples of things that I'm not sure about: should Survival go under VIT or INT? Should Music go under FTH or should it go under VIT? Should Conjuration go under INT instead of FTH? Is tying Light Armor to Evasion a good decision (I don't really like the idea of Light Armor / Med Armor / Heavy Armor skills)?
Also should we have a Leatherworking/Bow-making/Fletching skill/skills if we're adding bows?

## Weapons and Armor

Weapons are varied in their form and function, making them better in different circumstances. This allows fighters who specialize in different weapons to feel like their characters are different. It also allows fighters who specialize in a variety of weapon forms to be true weapon-masters who have a tool for different situations (just like how spellcasters have a variety of schools of magic to train in).

### Weapon properties

1. Damage - How hard does it hit
2. Damage Type - Slash, Pierce, or Crush
3. Speed - How many attack rolls does the weapon make per round
4. Penetration - How much armor is ignored by this weapon

And the damage formula is

damage = max(floor, weapon_damage - (armor[damage_type] - weapon_penetration))

The floor chosen for this formula could be the character's level (as it was before), or it could be some other number like 1, or 10% of the weapon_damage, or something. We could also consider adding strength to the damage as a bonus, since strength is always a fairly small number that does not grow, and this makes high-strength characters stronger straight away.

The rationale behind these properties is that fast-attacking weapons like daggers have less penetration so they are less effective against armor. Since they are fast-attacking, they are better for applying poisons and trying to get crits. Meanwhile large weapons like hammers have more penetration so they don't care so much about the target's armor, but they swing less often.

Damage Type is to help make items like Axe feel unique as well, but if we think having three different armor values to juggle is too much we can simplify things and go back to just armor. Under the damage type system, every enemy mob and every piece of player armor would have a value for their Slash, Pierce, and Crush resitance.

### Weapon Techniques

Each weapon type also has an associated technique. These can be activated abilities or passives that have a chance to proc. I'm in favor of making them passives for most things, but I am open to making them triggered skills (with cooldowns? or which cost MP?) if we think that would make things more fun for martials.

The purpose of techniques is to make martial combat more interesting and to add unique mechanics and feel to weapon types.

### The Weapon Table

| Weapon | Damage Type  | Speed       | Penetration | Technique                                                          | Notes |
| ------ | ------------ | ----------- | ----------- | ------------------------------------------------------------------ | -- |
| Sword  | Slash/Pierce | Medium      | Medium      | Riposte: Has a chance to counterattack after being attacked        | Since some swords are slash and some swords are pierce, this makes the skill a more versatile choice for a fighter to invest in, even if swords are never the high-damage kings. |
| Axe    | Slash        | Slow        | Low         | Sunder: Has a chance to reduce the target's armor for N rounds     | Sunder could be activated, percent chance, or an automatic stacking ability that just sunders more armor at higher ranks. The idea is that Sunder exposes targets for the whole party. |
| Hammer | Crush        | Slow        | High        | Stagger: Has a chance to stun the target, skipping their next turn |  |
| Dagger | Pierce       | Fast        | Low         | Backstab: Ignores armor when attacking from stealth or against a distracted or staggared target | Allows for thieves to deal increased damage when the party is distracting foes for them. |
| Bows   | Pierce       | Very Slow   | High        | Aim: Allows archer to charge shots for higher damage, provided they can maintain concentration. | The longer an archer can aim, the better their DPS, because they take the subtractive armor penalty once instead of over multiple rounds. As an archer ranks up, they can aim for longer and potentially pull off better shots from the longer aim. |

There's room to throw polearm and/or crossbow in the mix here if you think that'd be cool, but idk what their technique would be. But no guns.

## Spells

Spellcasters have a certain number of spell slots. This should be determined by a skill probably, but it could also be based on level or INT somehow. 

Spells are learned by using a spellbook. Once you use a spellbook, you instantly know the spell, and it occupies a spell slot. 
You can forget a spell by speaking to a spell forgetter. This will allow you to learn other spells, but if you want to learn the spell again, you will need to find another spellbook of the same kind (which means that scribes are now more helpful, even to just themselves).

Each spellbook has a skill rank and a skill associated with it. To learn a spell from a book, you must have the rank and skill required. For example, if Firebolt is a Level 2 Evocation Spell, then you must have at least 2 in Evocation to learn it.

Spellcasters must equip a magic focus item in one of their hands in order to be able to cast, so they are not constantly holding a spellbook, but the focus requirement maintains the equipment tax on spellcasting.

Possibly we could add a FTH skill that lets spellcasters upcast their spells for extra damage (but longer charge time, similar to 5e's rules for upcasting spells for an extra 1d8 dice or whatever).
