# Castle Recurse - Design Proposal

## Motivation

Keep is simple, stupid.

## Experience and Levels

Characters earn experience (XP) by adventuring. A character's level is determined by the total amount of XP they have earned over their lifetime.

Going from level 1 to level 2 requires 756 XP. Each level after that requires 1.26 times the XP of the level before it.

| Level     | 1           | 2   | 3   | 4    | 5    | 6    |
| --------- | ----------- | --- | --- | ---- | ---- | ---- |
| XP to reach | n/a (start) | 756 | 952 | 1200 | 1512 | 1905 |

**Question**: Do we have a hard level cap? Or do we just leave the level cap open?
**Question**: Where does 756 come from as the XP number at level 1? Just curious lol.

## Attributes / Primary Stats

Each character has a set of attributes which reflect their capabilities. These are also called "primary stats" or just "stats", because it's less typing that way. The attributes are:

Vitality (VIT) 
  - Determines how much health a character has
  - Determines a character's chance to resist conditions like poison
Strength (STR) 
  - Determines how hard a character hits things
Agility (AGI) 
  - Determines a character's evasion, critical hit chance, and stealth chance
Intelligence (INT) 
  - Determines a character's max mana
  - Determines how many spell slots a character has
Faith (FTH) 
  - Determines a character's spell damage and resistance

A character's attributes are determined by their race and class. The attributes increase as they level up.

**Question**: What formula governs attribute scaling on level up?
**Question**: Do we want to re-think these attributes in any way? i.e. have different attributes / make the attributes do different things. 
  - For example, should AGI directly affect damage with certain weapon types like daggers? Or should we keep the distinct split that currently exists between them?

## Secondary stats

Secondary stats are the stats that actually affect game rules. These are calculated using a combination of a characters Primary Stats, Equipment, and Conditions.

max_health = 10 * VIT
max_mana = 10 * INT
spell_slots = INT / 4
attack_damage = weapon_damge + (STR / 2)
to_hit_chance = level + (agi * 2)

**TODO:** fill in the rest of these

## Skills and Gear-Level

Skills, which could also be called Proficiencies, determine what your character can do. If you are skilled in Heavy Armor, then you can wear Heavy Armor. If you are skilled in Evocation magic, then you can learn spells from Evocation books. If you are skilled in Stealth, then you can attempt to sneak.

There are no skill ranks. You either know a skill, or you don't (again, they could be called proficiencies).

A character is granted skills by either their class or their race. This means hobbits can have the sneak skill, and dwarfs can have the axe skill even if they are not a warrior.

Gear is further limited by level. So an Outfit might require the Heavy Armor skill and also a level 3 character. This is our primary way to limit gear both by class and character level.

There is a design-space here for us to make skills something that players can collect as they go, as little additions to their character. For example, jobs could be implemented by an NPC who simply grants the skill of Blacksmithing or Alchemy. 

This does not prevent us from limiting these skills either: the master blacksmith might refuse to train someone who is already skilled in Alchemy or Enchanting.

Some other examples of where this might be interesting: 
- Music. Who said being a bard has to be a whole dedicated class? I mean, it could be, but also anyone can pick up an instrument, right? So imagine if there are different instruments. You could find a guitar, but that doesn't mean you know how to play it. But then you find someone who teaches you guitar, and now you have the guitar skill and you can play it to provide the party a minor boon (or even just for fun, nobody said guitar has to be useful). We could then repeat this for a variety of instruments and let players learn some or all of them if they so choose.
- Camping. Why not make camping a skill that some characters know, rather than a crafting job? It could be something that some characters know already and/or it could be a skill that characters learn from a quest. The quest could double as a tutorial, teaching them about camping and its importance.
- Weapon types. Maybe the warrior doesn't know all the weapon types right away. Might they can find a master of a certain weapon type later on who will teach them the weapon they want to know. Again, we could limit this as much or as little as we want, so maybe the spear-master won't teach his renowned technique to just anyone, you have to have proven your mettle (so either you need to be a certain class + level and/or you need to have a certain skill pre-requisite)
- Schools of Magic. Maybe a wizard doesn't know all the magic schools right away. Maybe they know illusion, evocation, and cantrips, but if they are interested in necromancy, they need to seek out someone to teach them. Or maybe there's a quest that leads them to a crypt of an ancient lich whose secrets have been sealed away in an ancient tome. If you read the tome, you learn the secrets of necromancy and the skill is unlocked to you. But only characters with the arcane pre-requisite skills can understand what the tome has to say.
- Prayer - Maybe players interested in roleplaying a religous character could get access to a prayer skill that grants minor boons. But characters who know cantrips, illusion, or necromancy have dabbled in dark magics and are thus not taught this skill by the church.

You could also limit the number of skills a character knows based upon their INT, so if the game world is littered with different skills-to-be-learned, then that doesn't mean a character can just become a master in everything, they must choose which ones they want to learn and accept that they will never know the rest.

But we're also trying to limit scope so we could keep the skills system very simple, with skills being given only from race and class and that's it.

## Jobs and Crafting

I think we should just have Blacksmith, Alchemist, and Enchanter. To limit scope we should ditch Scribe, and I think we should make camping just a skill that lets you make a campfire and that's it.

Learning recipes is limited based on character level.

I said that jobs should be a skill, but actually jobs could be things that provide one or more skills.

For example, the Blacksmith job gives you the Smithing Skill and the Repair skill, but maybe the Repair skill isn't tied only to smiths. Maybe dwarfs and warriors can repair as well, and Blacksmith therefore becomes a way that characters can pick up repair if they didn't get it from their race / class.
