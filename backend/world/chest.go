package world

import (
	"fmt"
)

const CHEST_REPLENISH_DURATION = (5 * 60) / WORLD_SECONDS_PER_UPDATE
const CHEST_DECAY_DURATION = (5 * 60) / WORLD_SECONDS_PER_UPDATE

type ChestMode int
const (
	// Unboxed means the chest has never been opened and we should roll fresh loot when they open it
	CHEST_MODE_UNBOXED ChestMode = iota
	// Replenish means the chest reset timer has gone off and the chest should replenish its loot on the next open
	CHEST_MODE_REPLENISH
	// Open means the chest has been opened
	CHEST_MODE_OPENED
)

type ChestKind int
const (
	CHEST_KIND_CHEST ChestKind = iota
	CHEST_KIND_CORPSE
)

type Chest struct {
	Name string
	Description string
	kind ChestKind
	mode ChestMode
	loot *LootTable
	inventory Inventory
	timer int
}

func NewChest(name string, description string, loot *LootTable) Chest {
	return Chest {
		Name: name,
		Description: description,
		kind: CHEST_KIND_CHEST,
		mode: CHEST_MODE_UNBOXED,
		loot: loot,
		inventory: NewInventory(),
		timer: 0,
	}
}

func NewCorpse(name string, loot *LootTable) Chest {
	return Chest {
		Name: fmt.Sprintf("%s's Corpse", name),
		Description: "The corpse is maimed beyond recognition.",
		kind: CHEST_KIND_CORPSE,
		mode: CHEST_MODE_UNBOXED,
		loot: loot,
		inventory: NewInventory(),
		timer: CHEST_DECAY_DURATION,
	}
}

func NewPlayerCorpse(character *Character) Chest {
	return Chest {
		Name: fmt.Sprintf("%s's Corpse", character.Name),
		Description: fmt.Sprintf("You see the remains of %s the %s. May they rest in peace.", character.Name, character.Race.Name),
		kind: CHEST_KIND_CORPSE,
		mode: CHEST_MODE_OPENED,
		loot: nil,
		inventory: character.Inventory,
		timer: CHEST_DECAY_DURATION,
	}
}

// Returns the inventory of the chest
// If the chest hasn't been opened yet, then the loot table will be rolled on based on the looter's Luck
func (chest *Chest) GetInventory(world *World, looter *Character) *Inventory {
	if chest.mode == CHEST_MODE_UNBOXED {
		chest.inventory = chest.loot.getLoot(world, looter.Stats.Values[STAT_LCK])
		chest.setOpen()
	}
	if chest.mode == CHEST_MODE_REPLENISH {
		chest.inventory = chest.loot.getLoot(world, looter.Stats.Values[STAT_LCK])
		chest.setOpen()
	}

	return &chest.inventory
}

func (chest *Chest) setOpen() {
	chest.mode = CHEST_MODE_OPENED
	if chest.kind == CHEST_KIND_CHEST {
		chest.timer = CHEST_REPLENISH_DURATION
	}
}

func (chest *Chest) isRottedCorpse() bool {
	return chest.kind == CHEST_KIND_CORPSE && chest.timer <= 0
}

func (chest *Chest) update() {
	if chest.timer > 0 {
		chest.timer--
		if chest.timer == 0 && chest.kind == CHEST_KIND_CHEST {
			chest.mode = CHEST_MODE_REPLENISH
		}
	}
}
