package world

import (
	"fmt"
	"math/rand/v2"
)

const BEHAVIOR_GOBLIN_SLEEP_DURATION int32 = (10 * 60) / WORLD_SECONDS_PER_UPDATE
const BEHAVIOR_GOBLIN_AWAKE_TIMER int32 = (50 * 60) / WORLD_SECONDS_PER_UPDATE
const BEHAVIOR_GOBLIN_SURPRISE_TIMER int32 = 3

type BehaviorGoblinMode int
const (
	BEHAVIOR_GOBLIN_MODE_IDLE = iota
	BEHAVIOR_GOBLIN_MODE_SURPRISED
	BEHAVIOR_GOBLIN_MODE_AGGRO
	BEHAVIOR_GOBLIN_MODE_SLEEPY
)

type BehaviorGoblin struct {
	Mode BehaviorGoblinMode
	SleepyTimer int32
	SurprisedTimer int32
}

func behaviorGoblinInit() *BehaviorGoblin {
	return &BehaviorGoblin{
		Mode: BEHAVIOR_GOBLIN_MODE_IDLE,
		SleepyTimer: 1 + rand.Int32N(BEHAVIOR_GOBLIN_AWAKE_TIMER),
	}
}

func (behavior *BehaviorGoblin) update(world *World, npc *Npc) {
	npcMob := world.Mobs.Get(npc.mobHandle)

	switch behavior.Mode {
		case BEHAVIOR_GOBLIN_MODE_IDLE: {
			//check if player is in room
			var playerInRoom bool
			for _, mobHandle := range world.Rooms[npcMob.Data.Room].Occupants {
				mob := world.Mobs.Get(mobHandle)
				if mob.PlayerCharacter != nil {
					playerInRoom = true
					break
				}
			}

			//they get ready to fight if there is someone in room through surprised mode
			if playerInRoom {
				world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s is getting ready to fight.", npcMob.Data.Name))
				behavior.Mode = BEHAVIOR_GOBLIN_MODE_SURPRISED
				behavior.SurprisedTimer = BEHAVIOR_GOBLIN_SURPRISE_TIMER
				break
			}

			//sometimes they take a nap for self-healing or for leisure
			behavior.SleepyTimer--
			if behavior.SleepyTimer < 0 {
				world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s lies down and goes to sleep", npcMob.Data.Name))
				behavior.Mode = BEHAVIOR_GOBLIN_MODE_SLEEPY
				behavior.SleepyTimer = max(BEHAVIOR_GOBLIN_SLEEP_DURATION, npcMob.Data.MaxHealth() - npcMob.Data.Health)
			}
		}

		case BEHAVIOR_GOBLIN_MODE_SURPRISED: {
			behavior.SurprisedTimer--
			if behavior.SurprisedTimer <= 0 {
				behavior.Mode = BEHAVIOR_GOBLIN_MODE_AGGRO
			}
		}

		case BEHAVIOR_GOBLIN_MODE_AGGRO: {
			if npcMob.Mode != MOB_MODE_IDLE {
				break
			}

			// If the mob is not doing anything, then find a target
			npcRoom := &world.Rooms[npcMob.Data.Room]
			for _, targetHandle := range npcRoom.Occupants {
				// Don't attack yourself
				if targetHandle == npc.mobHandle {
					continue
				}

				// For now, only attack players
				targetMob := world.Mobs.Get(targetHandle)
				if targetMob.PlayerCharacter == nil {
					continue
				}

				// Found target, set to attack
				npcMob.SetModeAttack(world, npc.mobHandle, targetHandle)
				break
			}

			// if no one is in the room then we try to go back to idle
			if npcMob.Mode == MOB_MODE_IDLE {
				behavior.Mode = BEHAVIOR_GOBLIN_MODE_IDLE
				behavior.SleepyTimer = BEHAVIOR_GOBLIN_AWAKE_TIMER
			}
		}

		case BEHAVIOR_GOBLIN_MODE_SLEEPY: {
			behavior.SleepyTimer--
			if npcMob.Data.Health < npcMob.Data.MaxHealth() {
				npcMob.Data.Health++
			}
			if behavior.SleepyTimer	< 0 {
				world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s has woken up!", npcMob.Data.Name))
				behavior.Mode = BEHAVIOR_GOBLIN_MODE_IDLE
				behavior.SleepyTimer = BEHAVIOR_GOBLIN_AWAKE_TIMER
			}
		}
	}
}


func (behavior *BehaviorGoblin) onAttacked(world *World, npc *Npc) {
	npcMob := world.Mobs.Get(npc.mobHandle)
	if behavior.Mode == BEHAVIOR_GOBLIN_MODE_SLEEPY {
		world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s is disgruntled that you have attacked them in their sleep!", npcMob.Data.Name))
		behavior.Mode = BEHAVIOR_GOBLIN_MODE_SURPRISED
		behavior.SurprisedTimer = BEHAVIOR_GOBLIN_SURPRISE_TIMER * 2
	}
}

// true and false indicate whether the description needs to be printed when a room containing this NPC is entered by a player
func (behavior *BehaviorGoblin) GetDescription(world *World, npc *Npc) (string, bool) {
	npcMob := world.Mobs.Get(npc.mobHandle)

	switch behavior.Mode {
		case BEHAVIOR_GOBLIN_MODE_SLEEPY: {
			return fmt.Sprintf("%s is currently taking a nap", npcMob.Data.Name), true
		}

		default: {
			return fmt.Sprintf("%s is a repulsive monster that wants to eat you.", npcMob.Data.Name), false
		}
	}
}
