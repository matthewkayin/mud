package world

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/mmcdole/lunar"
)

const FIXTURE_RESET_DURATION int = (5 * 60) / WORLD_SECONDS_PER_UPDATE

type FixtureId int

type Fixture struct {
	Id FixtureId `ts_type:"string"`
	BehaviorParams map[string]any

	instance lua.Value
	timer int
}

type FixtureJson struct {
	Id string
	BehaviorParams map[string]json.RawMessage
}

func (fixture *Fixture) MarshalJSON() ([]byte, error) {
	fixtureData := FIXTURE_DATA[fixture.Id]

	fixtureJson := FixtureJson {
		Id: fixtureData.Name,
	}

	var err error
	fixtureJson.BehaviorParams, err = behaviorParamsToJson(fixture.BehaviorParams, fixtureData.BehaviorParams)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&fixtureJson)
}

func (fixture *Fixture) UnmarshalJSON(data []byte) error {
	var fixtureJson FixtureJson
	err := json.Unmarshal(data, &fixtureJson)
	if err != nil {
		return err
	}

	var exists bool
	fixture.Id, exists = FIXTURE_NAME_TO_STRING[fixtureJson.Id]
	if !exists {
		return fmt.Errorf("No Fixture exists named '%s'.", fixtureJson.Id)
	}

	fixtureData := FIXTURE_DATA[fixture.Id]
	fixture.BehaviorParams, err = behaviorParamsFromJson(fixtureJson.BehaviorParams, fixtureData.BehaviorParams)
	if err != nil {
		return err
	}

	return nil
}

type FixtureData struct {
	Name string
	BehaviorParams map[string]BehaviorParamType `ts_type:"Record<string, BehaviorParamType>"`

	description lua.Value

	init *lua.Function
	reset *lua.Function
	interactions map[string]*lua.Function
}

// LOAD

var FIXTURE_DATA []*FixtureData
var FIXTURE_NAME_TO_STRING map[string]FixtureId

const FIXTURE_DATA_FOLDER = "fixtures"

func (world *World) loadFixtureData() {
	// Read fixture folder
	log.Printf("Loading fixture data...")
	paths, err := scriptGetFilesFrom(world.dataFolder, FIXTURE_DATA_FOLDER)
	if err != nil {
		log.Fatalf("Error opening fixture data folder: %s", err.Error())
	}

	FIXTURE_DATA = make([]*FixtureData, 0, len(paths))
	FIXTURE_NAME_TO_STRING = make(map[string]FixtureId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			log.Fatalf("%s: %s", path, err.Error())
		}

		// Parse fixture data
		parser := ScriptParser{}
		fixtureData := parser.parseFixture(table)
		if fixtureData == nil {
			log.Fatalf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, isDuplicate := FIXTURE_NAME_TO_STRING[fixtureData.Name]
		if isDuplicate {
			log.Fatalf("Fixture %s has name '%s' which is a duplicate of another fixture.", path, fixtureData.Name)
		}

		// Store fixture in FIXUTRE_DATA
		FIXTURE_NAME_TO_STRING[fixtureData.Name] = FixtureId(len(FIXTURE_DATA))
		FIXTURE_DATA = append(FIXTURE_DATA, fixtureData)
		log.Printf("Loaded Fixture '%s'.", fixtureData.Name)
	}

	log.Printf("All fixtures have been loaded.")
}

func (parser *ScriptParser) parseFixture(table *lua.Table) *FixtureData {
	fixtureData := &FixtureData{}

	fixtureData.Name = parser.getString(table, "name")
	fixtureData.description = table.RawGetString("description")
	if fixtureData.description.IsNil() {
		parser.addProblem(fmt.Errorf("missing required field 'description'"))
	} else if !(fixtureData.description.Kind() == lua.StringKind || fixtureData.description.Kind() == lua.FunctionKind) {
		parser.addProblem(fmt.Errorf("Fixture description must be either a string or a function returning a string"))
	}
	fixtureData.BehaviorParams = parser.getBehaviorParams(table, "behavior_params")

	fixtureData.init = parser.getOptionalFunction(table, "init")
	fixtureData.reset = parser.getOptionalFunction(table, "reset")

	// Interactions table
	fixtureData.interactions = make(map[string]*lua.Function)
	interactionsTable := parser.getTable(table, "interactions")
	if interactionsTable == nil {
		parser.addProblem(fmt.Errorf("Fixture must have an interactions table."))
		return nil
	}

	// Interactions map each interaction name to the function that handles it
	key := lua.Nil()
	for {
		var value lua.Value
		var ok bool
		var err error
		key, value, ok, err = interactionsTable.Next(key)
		if err != nil {
			parser.addProblem(fmt.Errorf("error reading 'interactions': %w", err))
			break
		}
		if !ok {
			break
		}

		name, isString := key.AsString()
		if !isString {
			parser.addProblem(fmt.Errorf("invalid key type in 'interactions'. expected %s, got %s", lua.StringKind, key.Kind().String()))
			continue
		}
		if name != strings.ToLower(name) {
			parser.addProblem(fmt.Errorf("interaction name '%s' must be lowercase", name))
		}
		if strings.ContainsFunc(name, unicode.IsSpace) {
			parser.addProblem(fmt.Errorf("interaction name '%s' must not contain spaces", name))
		}

		function, isFunction := value.AsFunction()
		if !isFunction {
			parser.addProblem(fmt.Errorf("invalid type for 'interactions.%s'. expected %s, got %s", name, lua.FunctionKind, value.Kind().String()))
			continue
		}

		fixtureData.interactions[name] = function
	}

	if len(fixtureData.interactions) == 0 {
		parser.addProblem(fmt.Errorf("A fixture must have at least one interaction."))
	}

	if parser.getError() != nil {
		return nil
	}

	return fixtureData
}

func (fixture *Fixture) init(world *World) {
	fixtureData := FIXTURE_DATA[fixture.Id]

	if fixtureData.init == nil {
		fixture.instance = lua.Nil()
		return
	}

	behaviorParams, err := behaviorParamsToLua(world, fixture.BehaviorParams, fixtureData.BehaviorParams)
	if err != nil {
		log.Printf("Warn - Error creating behavior params table for fixture '%s': %s", fixtureData.Name, err.Error())
	}

	result, ok := world.scriptCall(fixtureData.init.Value(), fmt.Sprintf("fixture %s init()", fixtureData.Name), behaviorParams.Value())
	if !ok || result.IsNil() {
		return
	}
	if result.Kind() != lua.TableKind {
		log.Printf("Warn - Fixture '%s' init() returned a %s instead of a table.", fixtureData.Name, result.Kind().String())
		return
	}

	fixture.instance = result
}

func (fixture *Fixture) reset(world *World) {
	fixtureData := FIXTURE_DATA[fixture.Id]

	if fixtureData.reset == nil {
		return
	}

	world.scriptCall(fixtureData.reset.Value(), fmt.Sprintf("fixture %s reset()", fixtureData.Name), fixture.instance)
}

func (fixture *Fixture) GetVerbs() []string {
	fixtureData := FIXTURE_DATA[fixture.Id]
	verbs := make([]string, 0, len(fixtureData.interactions))
	for verb, _ := range fixtureData.interactions {
		verbs = append(verbs, verb)
	}

	return verbs
}

func (fixture *Fixture) Interact(world *World, playerHandle MobHandle, verb string) bool {
	fixtureData := FIXTURE_DATA[fixture.Id]
	hook, exists := fixtureData.interactions[verb]
	if !exists {
		return false
	}

	playerLuaHandle, err := world.getMobLuaHandle(playerHandle)
	if err != nil {
		log.Print(err.Error())
		return false
	}

	world.scriptCall(hook.Value(), fmt.Sprintf("fixture %s interact %s", fixtureData.Name, verb), fixture.instance, playerLuaHandle)
	return true
}

func (fixture *Fixture) update(world *World) {
	fixture.timer--
	if fixture.timer <= 0 {
		fixture.reset(world)
		fixture.timer = FIXTURE_RESET_DURATION
	}
}


func (fixture *Fixture) GetName() string {
	return FIXTURE_DATA[fixture.Id].Name
}

func (fixture *Fixture) GetDescription(world *World) (string, bool) {
	fixtureData := FIXTURE_DATA[fixture.Id]

	if fixtureData.description.Kind() == lua.StringKind {
		description, _ := fixtureData.description.AsString()
		return description, true
	}

	if fixtureData.description.Kind() == lua.FunctionKind {
		getDescription, _ := fixtureData.description.AsFunction()
		result, ok := world.scriptCall(getDescription.Value(), fmt.Sprintf("fixture %s description()", fixtureData.Name), fixture.instance)
		if !ok {
			return "", false
		}
		description, ok := result.AsString()
		if !ok {
			log.Printf("Warn - Fixture description() did not return a string.")
			return "", false
		}

		return description, true
	}

	return "", false
}
