package game


type Npc struct {
	isActive bool
	behavior Behavior
	mobHandle MobHandle
}

type Behavior struct {
	description string
	handler func (*GameState, *Room, *Mob)
}

var aggro = Behavior {
	description: "Just attacks whomever is in the room.",
	handler: func (gameState *GameState, room *Room, mob *Mob) {
		for _, handle := range room.occupants {
			inRoom := gameState.world.Mobs.Get(handle)
			if inRoom.player != nil {
				mob.Mode = MobMode(ACTION_TYPE_ATTACK)
				mob.Target = handle
				break
			}
		}
	},
}

func (npc *Npc) onDeath () {
	npc.isActive = false
}

func (world *World) SpawnNpc (roomIndex uint) {
	newMob := Mob {
		Data: MobData {
			Name: "Goblin",

			Vitality: 10,
			Strength: 10,
			Agility: 10,
			Intelligence: 10,
			Faith: 10,

			Health: 20,
			Mana: 20,
		},
	}

	newMob.Data.Room = roomIndex

	newNpc := Npc {
		isActive: true,
		behavior: aggro,
		mobHandle: world.Mobs.Push(newMob),
		}

	spawnedMob := world.Mobs.Get(newNpc.mobHandle)
	spawnedMob.npc = &newNpc

	world.Rooms[roomIndex].occupants = append(world.Rooms[roomIndex].occupants, newNpc.mobHandle)
	world.Rooms[roomIndex].npcs = append(world.Rooms[roomIndex].npcs, newNpc)
}
