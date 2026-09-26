package world

type BehaviorId int

type BehaviorEventType int
const (
	BEHAVIOR_EVENT_TYPE_ATTACKED = iota
	BEHAVIOR_EVENT_TYPE_PLAYER_ENTERED
	BEHAVIOR_EVENT_TYPE_ITEM_GIVEN
)

type BehaviorEventAttacked struct {
	AttackerHandle MobHandle
}

type BehaviorEventPlayerEntered struct {
	PlayerHandle MobHandle
}

type BehaviorEventItemGiven struct {
	PlayerHandle MobHandle
	AddedToIndex int
	Amount int32
}

type BehaviorEvent struct {
	Type BehaviorEventType
	Data any
}

type Behavior interface {
	init(npc *Npc, world *World)
	update(npc *Npc, world *World)
	onEvent(npc *Npc, world *World, event BehaviorEvent) bool
	getDescription(npc *Npc, world *World) (string, bool)
}
