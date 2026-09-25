package world

import "fmt"

type NpcBehaviorId int
const (
	NPC_BEHAVIOR_GOBLIN = iota
)

type NpcBehavior interface {
	update(world *World, npc *Npc)
	onAttacked(world *World, npc *Npc)
	GetDescription(world *World, npc *Npc) (string, bool)
}

func NpcBehaviorInit(id NpcBehaviorId) NpcBehavior {
	switch id {
		case NPC_BEHAVIOR_GOBLIN:
			return behaviorGoblinInit()
		default:
			panic(fmt.Sprintf("Npc behavior %d has no init function!", id))
	}
}
