package world

import (
	"fmt"
	"log"
	"slices"
	"math/rand/v2"
	"encoding/json"
	"github.com/mmcdole/lunar"
)

// Rather than reset the NPC's sleepy timer after combat,
// we instead apply this adrenaline number to their sleepy timer,
// so sleepy time is delayed but not completely reset
const NPC_SLEEPY_ADRENALINE_DURATION int32 = (5 * 60) / WORLD_SECONDS_PER_UPDATE
const NPC_SURPRISE_DURATION int32 = 1

const NPC_RESPAWN_DURATION = 60 / WORLD_SECONDS_PER_UPDATE
const NPC_MOVEMENT_STEP_DURATION = 60 / WORLD_SECONDS_PER_UPDATE

const NPC_MOVEMENT_TYPE_OVERRIDE_NONE = NPC_MOVEMENT_TYPE_COUNT

const NPC_EVENT_PREVENT_DEFAULT = true

type NpcMode int
const (
	NPC_MODE_DEAD = iota
	NPC_MODE_IDLE
	NPC_MODE_SURPRISE
	NPC_MODE_AGGRO
	NPC_MODE_COUNT
)

func (mode NpcMode) String() string {
	switch mode {
		case NPC_MODE_DEAD:
			return "Dead"
		case NPC_MODE_IDLE:
			return "Idle"
		case NPC_MODE_SURPRISE:
			return "Surprise"
		case NPC_MODE_AGGRO:
			return "Aggro"
		default:
			return ""
	}
}

type NpcMovementType int
const (
	NPC_MOVEMENT_TYPE_SENTINEL = iota
	NPC_MOVEMENT_TYPE_WANDER
	NPC_MOVEMENT_TYPE_COUNT
)

func (movementType NpcMovementType) String() string {
	switch movementType {
		case NPC_MOVEMENT_TYPE_SENTINEL:
			return "Sentinel"
		case NPC_MOVEMENT_TYPE_WANDER:
			return "Wander"
		default:
			return ""
	}
}

// This could be replaced with a fine-grained number later
type NpcDisposition int
const (
	NPC_DISPOSITION_NEUTRAL = iota
	NPC_DISPOSITION_HOSTILE
	NPC_DISPOSITION_FRIENDLY
	NPC_DISPOSITION_COUNT
)

func (disposition NpcDisposition) String() string {
	switch disposition {
		case NPC_DISPOSITION_NEUTRAL:
			return "Neutral"
		case NPC_DISPOSITION_HOSTILE:
			return "Hostile"
		case NPC_DISPOSITION_FRIENDLY:
			return "Friendly"
		default:
			return ""
	}
}

type NpcEventType int
const (
	NPC_EVENT_TYPE_ATTACKED = iota
	NPC_EVENT_TYPE_PLAYER_ENTERED
	NPC_EVENT_TYPE_ITEM_GIVEN
)

type NpcEventAttacked struct {
	AttackerHandle MobHandle
}

type NpcEventPlayerEntered struct {
	PlayerHandle MobHandle
}

type NpcEventItemGiven struct {
	PlayerHandle MobHandle
	AddedToIndex int
	Amount int32
}

type NpcEvent struct {
	Type NpcEventType
	Data any
}

type Npc struct {
	// NPC "config" variables - tells us how to make a mob based on this NPC
	Id NpcId `ts_type:"string"`
	SpawnRoom int
	LevelRange Int32Range

	// Overrides - for things that would otherwise be specified by NPC data
	MovementTypeOverride NpcMovementType `ts_type:"string"`
	DropTableOverride DropTable

	// Maps each behavior param name declared by the NPC data to its value
	BehaviorParams map[string]any

	// NPC "instance" variables - keeps track of the NPC's current state
	mobHandle MobHandle
	mode NpcMode
	disposition NpcDisposition
	timer int32
	shouldReset bool
	// The table returned by the script's init() hook, or nil
	instance lua.Value
	events []NpcEvent
}

type NpcJson struct {
	Id string
	SpawnRoom int
	LevelRange Int32Range

	MovementTypeOverride string
	DropTableOverride DropTable

	BehaviorParams map[string]json.RawMessage
}

func (npc *Npc) MarshalJSON() ([]byte, error) {
	if !npcIdIsValid(npc.Id) {
		return nil, fmt.Errorf("Cannot save NPC with invalid NPC id %d.", npc.Id)
	}
	npcData := NPC_DATA[npc.Id]

	movementTypeOverride := ""
	if npc.MovementTypeOverride != NPC_MOVEMENT_TYPE_OVERRIDE_NONE {
		movementTypeOverride = npc.MovementTypeOverride.String()
	}

	npcJson := NpcJson {
		Id: npcData.Id,
		SpawnRoom: npc.SpawnRoom,
		LevelRange: npc.LevelRange,

		MovementTypeOverride: movementTypeOverride,
		DropTableOverride: npc.DropTableOverride,

		BehaviorParams: make(map[string]json.RawMessage, len(npc.BehaviorParams)),
	}

	for name, value := range npc.BehaviorParams {
		paramType, isDeclared := npcData.BehaviorParams[name]
		if !isDeclared {
			return nil, fmt.Errorf("Cannot save NPC '%s' with undeclared behavior param '%s'.", npcData.Id, name)
		}

		jsonValue, err := npcBehaviorParamToJson(paramType, value)
		if err != nil {
			return nil, fmt.Errorf("Cannot save NPC '%s' behavior param '%s': %w", npcData.Id, name, err)
		}
		npcJson.BehaviorParams[name] = jsonValue
	}

	return json.Marshal(&npcJson)
}

func (npc *Npc) UnmarshalJSON(data []byte) error {
	var npcJson NpcJson
	err := json.Unmarshal(data, &npcJson)
	if err != nil {
		return err
	}

	var exists bool
	npc.Id, exists = NPC_ID_STR_TO_ID[npcJson.Id]
	if !exists {
		return fmt.Errorf("No NPC id matches '%s'.", npcJson.Id)
	}
	npcData := NPC_DATA[npc.Id]

	npc.SpawnRoom = npcJson.SpawnRoom
	npc.LevelRange = npcJson.LevelRange

	if npcJson.MovementTypeOverride == "" {
		npc.MovementTypeOverride = NPC_MOVEMENT_TYPE_OVERRIDE_NONE
	} else {
		npc.MovementTypeOverride, exists = EnumFromString(npcJson.MovementTypeOverride, NpcMovementType(NPC_MOVEMENT_TYPE_COUNT))
		if !exists {
			return fmt.Errorf("No NPC movement type matches '%s'.", npcJson.MovementTypeOverride)
		}
	}

	npc.DropTableOverride = npcJson.DropTableOverride

	// Missing params are reported by Validate()
	npc.BehaviorParams = make(map[string]any, len(npcJson.BehaviorParams))
	for name, jsonValue := range npcJson.BehaviorParams {
		paramType, isDeclared := npcData.BehaviorParams[name]
		if !isDeclared {
			return fmt.Errorf("NPC '%s' has no behavior param named '%s'.", npcData.Id, name)
		}

		value, err := npcBehaviorParamFromJson(paramType, jsonValue)
		if err != nil {
			return fmt.Errorf("NPC '%s' behavior param '%s': %w", npcData.Id, name, err)
		}
		npc.BehaviorParams[name] = value
	}

	return nil
}

func npcIdIsValid(npcId NpcId) bool {
	return npcId >= 0 && int(npcId) < len(NPC_DATA)
}

// Returns true if the value has the Go type that the param type expects
func npcBehaviorParamHasType(paramType NpcBehaviorParamType, value any) bool {
	switch paramType {
		case NPC_BEHAVIOR_PARAM_TYPE_STRING:
			_, ok := value.(string)
			return ok
		case NPC_BEHAVIOR_PARAM_TYPE_NUMBER:
			_, ok := value.(float64)
			return ok
		case NPC_BEHAVIOR_PARAM_TYPE_BOOLEAN:
			_, ok := value.(bool)
			return ok
		case NPC_BEHAVIOR_PARAM_TYPE_ITEM:
			_, ok := value.(Item)
			return ok
		case NPC_BEHAVIOR_PARAM_TYPE_DIRECTION:
			_, ok := value.(Direction)
			return ok
		default:
			return false
	}
}

func npcBehaviorParamToJson(paramType NpcBehaviorParamType, value any) (json.RawMessage, error) {
	if !npcBehaviorParamHasType(paramType, value) {
		return nil, fmt.Errorf("value %v is not of type %s", value, paramType.String())
	}

	switch paramType {
		case NPC_BEHAVIOR_PARAM_TYPE_ITEM: {
			// Item.MarshalJSON has a pointer receiver, so the Item must be addressable
			item := value.(Item)
			return json.Marshal(&item)
		}
		case NPC_BEHAVIOR_PARAM_TYPE_DIRECTION: {
			direction := value.(Direction)
			if direction < 0 || direction >= DIRECTION_COUNT {
				return nil, fmt.Errorf("%d is not a valid direction", direction)
			}
			return json.Marshal(direction.String())
		}
		default:
			return json.Marshal(value)
	}
}

func npcBehaviorParamFromJson(paramType NpcBehaviorParamType, data json.RawMessage) (any, error) {
	switch paramType {
		case NPC_BEHAVIOR_PARAM_TYPE_STRING: {
			var value string
			err := json.Unmarshal(data, &value)
			return value, err
		}
		case NPC_BEHAVIOR_PARAM_TYPE_NUMBER: {
			var value float64
			err := json.Unmarshal(data, &value)
			return value, err
		}
		case NPC_BEHAVIOR_PARAM_TYPE_BOOLEAN: {
			var value bool
			err := json.Unmarshal(data, &value)
			return value, err
		}
		case NPC_BEHAVIOR_PARAM_TYPE_ITEM: {
			var value Item
			err := json.Unmarshal(data, &value)
			return value, err
		}
		case NPC_BEHAVIOR_PARAM_TYPE_DIRECTION: {
			var directionString string
			err := json.Unmarshal(data, &directionString)
			if err != nil {
				return nil, err
			}
			direction, ok := EnumFromString(directionString, Direction(DIRECTION_COUNT))
			if !ok {
				return nil, fmt.Errorf("'%s' is not a valid direction", directionString)
			}
			return direction, nil
		}
		default:
			return nil, fmt.Errorf("unhandled behavior param type %d", paramType)
	}
}

// Converts a behavior param value into a value that lua.State.NewTableFrom accepts
func npcBehaviorParamToLua(paramType NpcBehaviorParamType, value any) any {
	switch paramType {
		case NPC_BEHAVIOR_PARAM_TYPE_ITEM: {
			item := value.(Item)
			return map[string]any {
				"name": ITEM_DATA[item.Id].Name,
				"amount": item.Amount,
				"durability": item.Durability,
			}
		}
		case NPC_BEHAVIOR_PARAM_TYPE_DIRECTION:
			return value.(Direction).String()
		default:
			return value
	}
}

func (npc *Npc) tryReset(world *World) {
	// No need to reset if the NPC is already dead,
	// it will just respawn after respawn timer is up
	if npc.mode == NPC_MODE_DEAD {
		npc.shouldReset = false
		return
	}

	// If NPC is not dead, try to despawn mob
	npcMob, npcMobExists := world.Mobs.GetIfExists(npc.mobHandle)
	if !npcMobExists {
		npc.timer = 0
		npc.shouldReset = false
		return
	}

	// If players are still in the room, then don't despawn
	npcRoom := &world.Rooms[npcMob.Data.Room]
	if npcRoom.hasPlayerOccupants(world) {
		return
	}

	// Otherwise, despawn
	npcRoom.RemoveOccupant(npc.mobHandle)
	npc.mode = NPC_MODE_DEAD
	npc.instance = lua.Nil()
	npc.timer = 0 // Trigger a respawn
	npc.shouldReset = false
}

//spawns the mob associated with the npc
func (npc *Npc) spawnMob(world *World) {
	// Check if the room is empty of players before spawning
	npcRoom := &world.Rooms[npc.SpawnRoom]
	if npcRoom.hasPlayerOccupants(world) {
		return
	}

	npcData := NPC_DATA[npc.Id]

	dropTable := &npcData.dropTable
	if len(npc.DropTableOverride.Entries) != 0 {
		dropTable = &npc.DropTableOverride
	}

	// Create mob data
	level := npc.LevelRange.ChooseRandom()
	stats := calculateStatBlockAtLevel(&npcData.stats, &npcData.scaling, level)
	mobData := MobData {
		Name: npcData.Name,
		Room: npc.SpawnRoom,

		Level: level,
		Experience: npcData.experienceWorth + (npcData.experienceWorthScaling * (level - 1)),

		Stats: stats,
		Spells: []SpellId {},
		Inventory: dropTable.getLoot(),
		// Copy the equipment so that mobs don't share the NPC data's slices
		Equipment: Equipment {
			IsSlotInUse: slices.Clone(npcData.equipment.IsSlotInUse),
			SlotItem: slices.Clone(npcData.equipment.SlotItem),
			StatBonuses: npcData.equipment.StatBonuses,
		},
	}

	mobData.Health = mobData.MaxHealth()
	mobData.Mana = mobData.MaxMana()

	// Init mob
	npcMob := MobInit(&mobData)
	npcMob.Npc = npc

	// Add mob to world
	npc.mobHandle = world.Mobs.Push(npcMob)
	npcRoom.AddOccupant(world, npc.mobHandle)

	// Init behavior
	npc.setModeIdle()
	npc.disposition = npcData.startingDisposition

	npc.instance = lua.Nil()
	npc.events = make([]NpcEvent, 0, 1)
	if npcData.init != nil {
		npc.callInit(world)
	}
}

// Calls the init() hook with the behavior params and keeps the instance table it returns
func (npc *Npc) callInit(world *World) {
	npcData := NPC_DATA[npc.Id]

	paramsTree := make(map[string]any, len(npc.BehaviorParams))
	for name, value := range npc.BehaviorParams {
		paramsTree[name] = npcBehaviorParamToLua(npcData.BehaviorParams[name], value)
	}
	paramsTable, err := world.luaState.NewTableFrom(paramsTree)
	if err != nil {
		log.Printf("Warn - Error creating behavior params table for NPC '%s': %s", npcData.Id, err.Error())
		return
	}

	selfLuaHandle, err := world.getMobLuaHandle(npc.mobHandle)
	if err != nil {
		log.Printf("Warn - Error creating mob handle for NPC '%s': %s", npcData.Id, err.Error())
		return
	}

	result, ok := npc.callFunction(world, npcData.init, "init", selfLuaHandle, paramsTable.Value())
	if !ok || result.IsNil() {
		return
	}
	if result.Kind() != lua.TableKind {
		log.Printf("Warn - NPC '%s' init() returned a %s instead of a table.", npcData.Id, result.Kind().String())
		return
	}

	npc.instance = result
}

// Calls a script hook with the instance table and the NPC's mob handle, followed by args.
// Returns the hook's first result, or false if the hook is not defined or if it failed.
func (npc *Npc) callHook(world *World, hook *lua.Function, hookName string, args ...lua.Value) (lua.Value, bool) {
	if hook == nil {
		return lua.Nil(), false
	}

	selfLuaHandle, err := world.getMobLuaHandle(npc.mobHandle)
	if err != nil {
		log.Printf("Warn - Error creating mob handle for NPC '%s' %s(): %s", NPC_DATA[npc.Id].Id, hookName, err.Error())
		return lua.Nil(), false
	}

	luaArgs := make([]lua.Value, 0, len(args) + 2)
	luaArgs = append(luaArgs, npc.instance, selfLuaHandle)
	luaArgs = append(luaArgs, args...)
	return npc.callFunction(world, hook, hookName, luaArgs...)
}

// Calls a script function with exactly the given args and returns its first result.
// Returns false if the function is not defined or if it failed, in which case the error is logged.
func (npc *Npc) callFunction(world *World, function *lua.Function, functionName string, args ...lua.Value) (lua.Value, bool) {
	if function == nil {
		return lua.Nil(), false
	}

	result, err := world.luaState.CallOne(function.Value(), args...)
	if err != nil {
		log.Printf("Warn - NPC '%s' %s() failed: %s", NPC_DATA[npc.Id].Id, functionName, err.Error())
		return lua.Nil(), false
	}

	return result, true
}

// Calls a script event hook and returns true if the event's default should be prevented
func (npc *Npc) callEventHook(world *World, hook *lua.Function, hookName string, args ...lua.Value) bool {
	result, handled := npc.callHook(world, hook, hookName, args...)
	if !handled {
		return false
	}

	// The hook is allowed to not return a boolean, in which case preventDefault is false
	preventDefault, ok := result.AsBool()
	return ok && preventDefault
}

func (npc *Npc) setModeIdle() {
	npc.mode = NPC_MODE_IDLE
	npc.timer = NPC_MOVEMENT_STEP_DURATION
}

func (npc *Npc) setModeSurprise(world *World) {
	npcMob := world.Mobs.Get(npc.mobHandle)

	npc.mode = NPC_MODE_SURPRISE
	npc.timer = NPC_SURPRISE_DURATION
	npcMob.alertness = MOB_ALERTNESS_MAX
}

func (npc *Npc) update(world *World) {
	// Try reset
	if npc.shouldReset {
		npc.tryReset(world)
	}

	// Respawn
	if npc.mode == NPC_MODE_DEAD {
		npc.timer--
		if npc.timer <= 0 {
			npc.spawnMob(world)
		}

		return
	}

	// Check for mob death
	npcMob, npcMobExists := world.Mobs.GetIfExists(npc.mobHandle)
	if !npcMobExists {
		npc.mode = NPC_MODE_DEAD
		npc.timer = NPC_RESPAWN_DURATION
		npc.instance = lua.Nil()
		return
	}

	// Handle events
	for index := range len(npc.events) {
		npc.onEvent(world, &npc.events[index])
	}
	npc.events = []NpcEvent{}

	npc.callHook(world, NPC_DATA[npc.Id].update, "update")

	switch npc.mode {
		case NPC_MODE_IDLE: {
			// Check if player in room
			if npc.disposition == NPC_DISPOSITION_HOSTILE {
				// Check if there is a player in the room
				roomHasPlayer := slices.ContainsFunc(world.Rooms[npcMob.Data.Room].Occupants, func (handle MobHandle) bool {
					mob := world.Mobs.Get(handle)
					if mob.CheckFlag(MOB_FLAG_HIDDEN) {
						return false
					}
					if !mob.IsPlayer() {
						return false
					}
					return true
				})

				// If room has player, get ready to fight
				if roomHasPlayer {
					npc.setModeSurprise(world)
					world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s is getting ready to fight!", npcMob.Data.Name))
					break
				}
			}

			// Update movement
			if npc.getMovementType() == NPC_MOVEMENT_TYPE_WANDER {
				npc.timer--

				if npc.timer <= 0 {
					npc.movementStep(world)
					npc.timer = NPC_MOVEMENT_STEP_DURATION
				}
			}
		}

		case NPC_MODE_SURPRISE: {
			// Countdown surpise timer, then switch to aggro
			npc.timer--
			if npc.timer <= 0 {
				npc.mode = NPC_MODE_AGGRO
			}
		}

		case NPC_MODE_AGGRO: {
			// If mob is doing something, then don't interrupt it
			if npcMob.Mode != MOB_MODE_IDLE {
				break
			}

			// Find a target to attack
			npcRoom := &world.Rooms[npcMob.Data.Room]
			for _, targetHandle := range npcRoom.Occupants {
				// Don't attack yourself
				if targetHandle == npc.mobHandle {
					continue
				}

				// For now, only attack players
				targetMob := world.Mobs.Get(targetHandle)
				if !targetMob.IsPlayer() {
					continue
				}

				// Don't attack hidden players
				if targetMob.CheckFlag(MOB_FLAG_HIDDEN) {
					continue
				}

				// Found target, set to attack
				npcMob.SetModeAttack(world, npc.mobHandle, targetHandle)
				break
			}

			// If mob is still idle at this point, it means no
			// target was found, so go back to idle
			if npcMob.Mode == MOB_MODE_IDLE {
				npc.mode = NPC_MODE_IDLE
			}
		}
	}
}

func (npc *Npc) getMovementType() NpcMovementType {
	if npc.MovementTypeOverride != NPC_MOVEMENT_TYPE_OVERRIDE_NONE {
		return npc.MovementTypeOverride
	}

	return NPC_DATA[npc.Id].MovementType
}

func (npc *Npc) movementStep(world *World) {
	npcMob := world.Mobs.Get(npc.mobHandle)
	npcRoom := &world.Rooms[npcMob.Data.Room]

	switch npc.getMovementType() {
		case NPC_MOVEMENT_TYPE_SENTINEL: {
			log.Printf("Warn - movementStep() called on a sentinel NPC with key %s, mob name %s, and mob handle %d:%d.",
				NPC_DATA[npc.Id].Id, npcMob.Data.Name, npc.mobHandle.Id, npc.mobHandle.Generation)
		}

		case NPC_MOVEMENT_TYPE_WANDER: {
			// Collect an array of all possible exits
			exitRoomIndices := make([]int, 0, 4)
			for directionIndex := range DIRECTION_COUNT {
				// Don't walk into non-existing or locked rooms
				adjacentRoomIndex := npcRoom.Exits[directionIndex]
				if adjacentRoomIndex == ROOM_NONE || npcRoom.ExitIsLocked[directionIndex] {
					continue
				}

				// Don't walk into safe rooms
				adjacentRoom := &world.Rooms[adjacentRoomIndex]
				if adjacentRoom.IsSafeZone {
					continue
				}

				exitRoomIndices = append(exitRoomIndices, adjacentRoomIndex)
			}

			// If there are no exits, don't move
			if len(exitRoomIndices) == 0 {
				return
			}

			// Otherwise, choose a random exit
			index := rand.IntN(len(exitRoomIndices))
			newRoomIndex := exitRoomIndices[index]

			npcRoom.MoveOccupant(world, npc.mobHandle, newRoomIndex)
		}
	}
}

func (npc *Npc) PushEvent(event NpcEvent) {
	npc.events = append(npc.events, event)
}

func (npc *Npc) onEvent(world* World, event *NpcEvent) {
	switch event.Type {
		case NPC_EVENT_TYPE_ATTACKED: {
			data := event.Data.(NpcEventAttacked)

			attackerLuaHandle, err := world.getMobLuaHandle(data.AttackerHandle)
			if err != nil {
				log.Print(err.Error())
				return
			}

			npc.callHook(world, NPC_DATA[npc.Id].onAttacked, "on_attacked", attackerLuaHandle)
		}

		case NPC_EVENT_TYPE_PLAYER_ENTERED: {
			data := event.Data.(NpcEventPlayerEntered)

			playerLuaHandle, err := world.getMobLuaHandle(data.PlayerHandle)
			if err != nil {
				log.Print(err.Error())
				return
			}

			preventDefault := npc.callEventHook(world, NPC_DATA[npc.Id].onPlayerEntered, "on_player_entered", playerLuaHandle)
			if preventDefault {
				return
			}

			// Default
			if npc.disposition == NPC_DISPOSITION_NEUTRAL {
				npc.disposition = NPC_DISPOSITION_HOSTILE
			}
		}

		case NPC_EVENT_TYPE_ITEM_GIVEN: {
			data := event.Data.(NpcEventItemGiven)

			playerLuaHandle, err := world.getMobLuaHandle(data.PlayerHandle)
			if err != nil {
				log.Print(err.Error())
				return
			}

			preventDefault := npc.callEventHook(world, NPC_DATA[npc.Id].onItemGiven, "on_item_given",
				playerLuaHandle,
				lua.Number(float64(data.AddedToIndex)),
				lua.Number(float64(data.Amount)),
			)

			if preventDefault {
				return
			}

			// Default - give the item back
			npcMob := world.Mobs.Get(npc.mobHandle)
			playerMob := world.Mobs.Get(data.PlayerHandle)
			item := npcMob.Data.Inventory.RemoveItems(data.AddedToIndex, data.Amount)
			playerMob.Data.Inventory.AddItem(item)

			world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s is uninterested in this item. They returned it to %s.", npcMob.Data.Name, playerMob.Data.Name))
		}
	}
}

func (npc *Npc) GetDescription() string {
	return NPC_DATA[npc.Id].description
}

func (npc *Npc) GetStatusDescription(world *World) (string, bool) {
	npcData := NPC_DATA[npc.Id]
	result, ok := npc.callHook(world, npcData.getStatusDescription, "get_status_description")
	if !ok || result.IsNil() {
		return "", false
	}

	resultString, isString := result.AsString()
	if !isString {
		log.Printf("Warn - NPC '%s' get_status_description() returned a %s instead of a string.", npcData.Id, result.Kind().String())
		return "", false
	}

	return resultString, true
}
