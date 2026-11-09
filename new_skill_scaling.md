welcome to our new document, where life is better without the the google corporation. we can breathe easy here, 
and we can control our own formatting without any issue!

SCALING

some initial thought on our new scaling environment:

magic has 4 dimensions, which are (1) more mana per rest, (2) mana discounts when using skills into which you have
specialized, and (3) unlocking new spells, (4) QoL effects like concentration bonuses for battlemages of slots for 
toolbox casters. all 4 of these dimensions will be levers we want to manipulate. 

in the system that seems obvious to me, we make it so 
-schools of magic give you new spell unlocks and each rank also gives discounts in mana costs of spells from that school
-we can have core utility skills like spellcraft or a concentration skill that can be used for QoL that is less important.
-max mana is a skill

this system has only one downside, which is that giving mana discounts to a schookl o fmagic necessarily also unlocks new spells
from that school, but I don't see how that limits our design space.

the reason we want these levers is because some spells do damage/healing, and some provide utility, and we want to allow
moderate ranks in a skill to gain utility or heavy ranks so that skill is a main output of the character where that character
prefers to spend all their mana. for example, illusion magic might give CC effects, it might have weak psychic damage effects,
weak support, &c. with moderate illusion magic you can get some CC if your group needs it, but without a heavy mana discount
you prefer to spend you mana elsewhere unless necessary. with deep illusion magic it becomes more full-featured as a spell school
and then you feel positive about casting your CC at a high discount, while other spell schools are relegated to obscurity.

this all also means that we want to make it so that high level magic can cast more core utility per hexcrawl ~or~ put out more
damage per hexcrawl, while keeping mid-rank unlocks like sleep relevant erven in the endgame. my basic napkin math goes as follows:

a starting full-progression caster, eg a wizard gets a max mana of 50, and a max leve wizard gets a max mana of 100
low rank (let's say a 2 out of 10) in a skill like illusion results in a 10% mana discount on illusion spell at 5% per rank
and gets you a utility spell called "beguile", while at high level let's say it gets you roughly 7 out of 10 for a 35% ish discount.

this means that if beguile costs 12 ish mana we can cast 5 ish of them per hexcrawl at low level and 12 ish of them per hexcrawl
at high level, making a high level caster more powerful but keeping beguile relevant across the level range.

with this math, a theoretical fireball can deal 50 damage at 12 mana at low level (as a benchmark) to get us 250 damage per hexcrawl
if you fgocus on being a blaster class at low levels, and then at a high level we can make more powerful nukes for faster throughput.
at high level we can target some arbitrary integer, let's say it's 4000 damage acrross the hexcrawl, and that gets us a high level
flamestrike that costs 36 mana and does something in the neighborhood of 170 damage.

in summary, this means something like:
-spell schools that simultaneously unlock new skills and take you from something like a 5% or 10% mana discount for a low level wiz
   up to something like a 35% or 40% mana discount at high level for spells from that school
-core utility skills that do things like interrupt resist, max mana, and so on
-low level wiz max mana at something like 50 ish and 100 ish at the high level
-6 mana for a twiddly/bad utility effec that barely matters and mostly taxes the memory slot, 12 man for most ytility effects
   and post-testing we can figure out casting utility effects up to as much as 18 or 24 for high level sick utility spells 
-benchmark firebolt at a ratio of something like 12 mana and 50 damage and at a high level flamestrike at something like 36 mana
   and 170 damage

only question: do we want to be able to apply bonuses like +fire damage or +healing regardless of that school the spell comes from,
eg from accessories? ditto for spell resist

we can use this to work back to the costs of all other spells, eg buffs, debuffs, utility effects, and with testing we can
find the appropriate mana cost for monster summons once we see how much damage per round a monster does post-testing, which just
means that we add necro and minion summoning once we get a good gameflow with warriors and wizards vs monsters at any given level.

let's talk about how martials relatively scale. defensively they get mitigation * hp for ehp or mitigation * recieved healing. 
that means offensively they get damage * some bonus from things like crit rate. defensively a martial scales mitigation and hp,
or they scale mitigation and their healer scales healing, while offensively they hit for more damage from a better weapon and get some
multiplicative scaling.

that means tenatively we can make them hit for some arbitrary integer per round, let's say ~10 per round with some misses
and some crits and some deflects and so on if they build and orc warrior, so the weapon comes out to 8 - 10 damage and we 
get maybe a 20% increase to dps from some crit rate or some such thing, and caster does less damage. the is where our first 
real complexity creeps in, which is that casters are going to be squishy and will burn through scarce resources (hp and mana)
while leveling, so i think we should keep armor vs penetration and make it so you need cutting-edge top-flight weapons to get past
armor in dungeons, but this much mroe relaxed outside of dungeons. that means we might allow 6 damage per round out of dungeons
for a wimpy necro with a walking stick but deflects crush that down to 2 or 3 in a dungeon where they cast spells for the group.

okay, what is the damage per round we want on a martial at a high level to keep up with casters. let's say we target something nutty 
like 180 damage per round at the highest level so that across a large and evil dungeon you eventually outdamage a caster who is 
profligate with flamestrikes, we might want to make it so that a cutting-edge high-level weapon does ~90 damage per hit, has penetration
to get through high level dungeon/boss armor, and gets on average a roughly 100% increase damage from the passive effects of cranked 
supplementary skills

in summary:
-weapons have equipment weights that don't scale much across level range, damage and penetration
-we can start with 8 being high end and 5 being low end pre-armor and pre-bonuses on weapon damage, and work up to a 
   whopping 80 high end weapon damage
-supplementary stats can give you up to a 20% ish dps increase at low level on top of weapon damage assuming orc warrior 
   and up to 100% at max level

ehp ranges tbd, and i think we work back from monster designs during testing, because ehp values don't just depend on how much 
bosses hit for, they depend on how much you lose per solo hexcrawl leveling excursion, and depend on how long monsters survive 
to deal more damage per round. if monsters go down quick they can hit hard, but if we slow combat down and give them lots of 
hp then they need to hit for ~low~ damage outside of high level/group dungeons. 

also, because there is no pvp we need no mechanical symmetry which we had vestigially in our old system from its earliest days,
so we can just design player EHP against monster gameflow, and monster stats to set fight duration at the right legnth and dungeon
difficulty at the right difficulty.

SKILL BLOCKS

oh boy, making this skill list is so fun it makes me wish we had some kind of skillbuy system in this game!

MAGIC:
(we need some core functionality skills for things like mana and concentration -- we also need to figure out how spell slots work)
-spellcraft (cool in theme -- unclear what it does from the above list)
-faith (cool in theme -- unclear what it does from the above list)
-arcana/esoterica (cool in theme -- unclear what it does from the above list)
(now the schools in magic, i think we should make some schools a bit eclectic with some substandard/niche stuff it can do
both for coolness purposes and also to make one-school spellcaster classes cool)
-elemental magic (best for blaster casters)
-life magic (best for healing -- also has buffs)
-conjuration (best for summons -- also has utility, eg summoning friends to you or your to your friends? banishment?)
-illusion magic/psychic magic (best for crowd control, good utility, maybe has mediocre instant pyshic attacks?)
-UNDECIDED NAME (we need a school for some form of reality mangipulation aimed at utility and battlemage so we can put 
mage armor, mage weapon and unlocks all in one?) 
(the class-only schools)
-black magic (necromancer, everything here has evil downsides like sacrificing life)
-holy magic (priest only, primarily buffs and utility? -- also could call it "miracles?" or something?)

COMBAT:
(where does equip weight come from if there is no skill scaling for it?)
toughness (max hp)
-light armor
-medium armor
-heavy armor
-shields (block?)

(i think there should be skills that give you all the multiplicaively scaling stuff like hit rate and crit and our math is super simple)
one option is to just have an attack rating skill, but alternately we can break it up as:
-one-handed combat
-two-handed combat
-marksmanship

(and then the skills that unlock the weapons with the better stats that can simple and easily muiltiplicatively scaled 
by the fighting style)
-swords (blades? include knives in it?)
-axes
-bludgeoning
-polearms?
-bows
-crossbows?
-throwing weapons?
-guns?

EXPLORATION:
-inspect (needs new name?)
-sneak
-breaking and entering (traps and locks in one skill -- security? burgling?)
-escape (roll into another skill? give another name?)
-first aid (from survivalist)
-camping (from survivalist -- i seperated survivalist functionality and put the two skills in explioration for coolness)

TRADE:
-blacksmithing (from blacksmith -- do we break out repair into a seperate skill for coolness?)
-enchanting (from enchanter -- do we split out scrolls crafter into scribing for coolness?)
-alchemy (from alchemist -- if alchemists get the philosopher's stone do they get a seperate transmutaton skill for coolness?)
-music (for bards! we imply its like an out of combat job for coolness purposes!)

(should there be core trade options for all players regardless of jobs? eh scribe and repair are core and enchanters and bsmiths
just get big bonuses to them?)


