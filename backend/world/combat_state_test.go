package world

import (
	"strings"
	"testing"
)

type combatStateTestState struct {
	world *World
	player *Mob
	messages []string
}

func newCombatStateTestState() *combatStateTestState {
	state := &combatStateTestState{}
	state.world = &World {
		Mobs: NewMobArray(),
		Rooms: []Room {
			{ Inventory: NewInventory() },
			{ Inventory: NewInventory() },
		},
		MessagePlayer: func(playerKey string, message string) {
			state.messages = append(state.messages, message)
		},
	}

	character := &Character {
		Equipment: NewEquipment(),
		Inventory: NewInventory(),
	}
	character.Name = "Warrior"
	character.Health = 10
	handle := state.world.Mobs.Push(NewCharacterMob(character))
	state.world.Rooms[0].AddOccupant(handle)
	state.player = state.world.Mobs.Get(handle)

	return state
}

func (state *combatStateTestState) spawnNpc(disposition NpcDisposition, room int) *Mob {
	npc := &Npc {
		data: &NpcData{},
		MobData: MobData {
			Name: "Cow",
			Room: room,
			Health: 10,
		},
		disposition: disposition,
		threats: make([]NpcThreat, 0, 1),
	}
	handle := state.world.Mobs.Push(NewNpcMob(npc))
	npc.mobHandle = handle
	state.world.Rooms[room].AddOccupant(handle)
	return state.world.Mobs.Get(handle)
}

func TestHostileNpcPutsBothInCombat(t *testing.T) {
	state := newCombatStateTestState()
	npc := state.spawnNpc(NPC_DISPOSITION_HOSTILE, 0)

	if !state.player.IsInCombat(state.world) {
		t.Errorf("expected player to be in combat with a hostile NPC")
	}
	if !npc.IsInCombat(state.world) {
		t.Errorf("expected hostile NPC to be in combat with a player")
	}
}

func TestNeutralNpcWithoutThreatIsNotCombat(t *testing.T) {
	state := newCombatStateTestState()
	npc := state.spawnNpc(NPC_DISPOSITION_NEUTRAL, 0)

	if state.player.IsInCombat(state.world) {
		t.Errorf("expected player not to be in combat with a calm neutral NPC")
	}
	if npc.IsInCombat(state.world) {
		t.Errorf("expected calm neutral NPC not to be in combat")
	}
}

func TestNeutralNpcWithThreatPutsBothInCombat(t *testing.T) {
	state := newCombatStateTestState()
	npc := state.spawnNpc(NPC_DISPOSITION_NEUTRAL, 0)
	npc.Npc.setThreat(state.player.Handle, 5)

	if !state.player.IsInCombat(state.world) {
		t.Errorf("expected player to be in combat with a neutral NPC that has threat on them")
	}
	if !npc.IsInCombat(state.world) {
		t.Errorf("expected neutral NPC with threat to be in combat")
	}
}

func TestThreatInAnotherRoomIsNotCombat(t *testing.T) {
	state := newCombatStateTestState()
	npc := state.spawnNpc(NPC_DISPOSITION_NEUTRAL, 1)
	npc.Npc.setThreat(state.player.Handle, 5)

	if state.player.IsInCombat(state.world) {
		t.Errorf("expected threat from another room not to put the player in combat")
	}
	if npc.IsInCombat(state.world) {
		t.Errorf("expected NPC whose threat is in another room not to be in combat")
	}
}

func TestDeadHostileNpcIsNotCombat(t *testing.T) {
	state := newCombatStateTestState()
	npc := state.spawnNpc(NPC_DISPOSITION_HOSTILE, 0)
	npc.Health = 0

	if state.player.IsInCombat(state.world) {
		t.Errorf("expected a dead hostile NPC not to put the player in combat")
	}
}

func TestNeutralNpcStopsAttackingWhenThreatIsGone(t *testing.T) {
	state := newCombatStateTestState()
	npc := state.spawnNpc(NPC_DISPOSITION_NEUTRAL, 0)

	// Threat of 1 decays to 0 on the next update, as a taunt runs out
	npc.Npc.setThreat(state.player.Handle, 1)
	npc.SetModeAttack(state.world, state.player.Handle)
	npc.HasPendingAction = false

	npc.Npc.update(state.world)

	if npc.Mode != MOB_MODE_IDLE {
		t.Errorf("expected neutral NPC to stop attacking once its threat decayed, mode is %d", npc.Mode)
	}
	announced := false
	for _, message := range state.messages {
		if strings.Contains(message, "stopped attacking") {
			announced = true
		}
	}
	if !announced {
		t.Errorf("expected the NPC to announce that it stopped attacking, got %v", state.messages)
	}
}

func TestHostileNpcKeepsAttacking(t *testing.T) {
	state := newCombatStateTestState()
	npc := state.spawnNpc(NPC_DISPOSITION_HOSTILE, 0)
	npc.SetModeAttack(state.world, state.player.Handle)
	npc.HasPendingAction = false

	npc.Npc.update(state.world)

	if !npc.IsAttacking(state.player.Handle) {
		t.Errorf("expected hostile NPC to keep attacking the player without threat")
	}
}
