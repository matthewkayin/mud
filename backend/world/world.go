package world

const WORLD_SECONDS_PER_UPDATE = 3

type WorldMessage struct {
	ToPlayers []int
	Message string
}

type World struct {
	Messages []WorldMessage

	Characters map[string]*Character
	PlayerCharacters map[int][]string

	Mobs MobArray
	Rooms []Room
}

func (world *World) Update() {
}

func (world *World) messagePlayer(playerId int, message string) {
	world.Messages = append(world.Messages, WorldMessage {
		ToPlayers: []int { playerId },
		Message: message,
	})
}

func (world *World) messageRoom(roomIndex int, message string) {
	room := &world.Rooms[roomIndex]

	toPlayers := make([]int, 0, len(room.occupants))
	for _, mobHandle := range room.occupants {
		mob := world.Mobs.Get(mobHandle)
		if mob.PlayerId == MOB_PLAYER_NONE {
			continue
		}

		toPlayers = append(toPlayers, mob.PlayerId)
	}

	if len(toPlayers) == 0 {
		return
	}

	world.Messages = append(world.Messages, WorldMessage {
		ToPlayers: toPlayers,
		Message: message,
	})
}
