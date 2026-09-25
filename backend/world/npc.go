package world

import (
	"fmt"
	"math/rand/v2"
)

type Npc struct {
	Id NpcId
	MinLevel int32
	MaxLevel int32
	SpawnRoom int
	RespawnDuration int

	Behavior NpcBehavior `json:"-"`
	mobHandle MobHandle
	respawnTimer int
}

//spawns the mob associated with the npc
func (npc *Npc) spawnMob(world *World) {
	npcData := NPC_DATA[npc.Id]

	// Create mob data
	level := npc.MinLevel + rand.Int32N(npc.MaxLevel - npc.MinLevel)
	stats := calculateStatBlockAtLevel(&npcData.baseStats, &npcData.scaling, level)
	mobData := MobData {
		Name: npcData.name,
		Room: npc.SpawnRoom,

		Level: level,
		Experience: 0, // TODO

		Stats: stats,
		Spells: []Spell {},
		Inventory: npc.determineDrops(),
		Equipment: npcData.equipment,
	}

	mobData.Health = mobData.MaxHealth()
	mobData.Mana = mobData.MaxMana()

	// Init mob
	npcMob := MobInit(&mobData)
	npcMob.Npc = npc

	// Add mob to world
	npc.mobHandle = world.Mobs.Push(npcMob)
	npcRoom := &world.Rooms[npcMob.Data.Room]
	npcRoom.AddOccupant(world, npc.mobHandle)

	// Create Npc behavior
	npc.Behavior = NpcBehaviorInit(npcData.behaviorId)

	world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s has spawned into this room.", npcMob.GetName()))
}

func (npc *Npc) determineDrops() Inventory {
	npcData := NPC_DATA[npc.Id]

	// Determine total drop chance among all NPC drops
	var dropChanceTotal int32 = 0
	for _, drop := range npcData.drops {
		dropChanceTotal += drop.dropChance
	}

	inventory := Inventory { Items: []Item {} }

	// Roll for an item and add it to the inventory
	for _ = range npcData.dropCount {
		roll := rand.Int32N(dropChanceTotal)
		var dropChance int32 = 0
		var dropIndex int = 0

		for dropIndex < len(npcData.drops) {
			dropChance += npcData.drops[dropIndex].dropChance
			if roll <= dropChance {
				break
			}
			dropIndex++
		}

		if dropIndex == len(npcData.drops) {
			panic("Something bad happened during Npc drop determination")
		}

		drop := &npcData.drops[dropIndex]

		// Determine amount
		var amount int32 = 1
		if drop.maxAmount - drop.minAmount > 0 {
			amount = drop.minAmount + rand.Int32N(drop.maxAmount - drop.minAmount)
		}

		// Determine durability
		var durability int32 = 0
		if drop.maxDurability - drop.minDurability > 0 {
			durability = drop.minDurability + rand.Int32N(drop.maxDurability - drop.minDurability)
		}

		inventory.AddItem(Item {
			Id: drop.itemId,
			Amount: amount,
			Durability: durability,
		})
	}

	return inventory
}

func (npc *Npc) update(world *World) {
	// If respawning, just update timer and don't do anything else
	if npc.respawnTimer != 0 {
		npc.respawnTimer--
		if npc.respawnTimer == 0 {
			npc.spawnMob(world)
		}
	}
	if npc.respawnTimer != 0 {
		return
	}

	// Check for mob death
	_, npcMobExists := world.Mobs.GetIfExists(npc.mobHandle)
	if !npcMobExists {
		npc.respawnTimer = npc.RespawnDuration
		return
	}

	// Behavior update
	npc.Behavior.update(world, npc)
}
