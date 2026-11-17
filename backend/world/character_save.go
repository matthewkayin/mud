package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mud/database"
	"mud/bitset"
	"mud/util"
	"os"
	"path/filepath"
	"strings"
)

type CharacterJson struct {
	PlayerKey string

	Name string
	Room RoomPosition
	RoomsDiscovered []RoomPosition

	Race string
	Class string

	Experience int32
	ExpTotal int32

	Skills []string

	Health int32
	Mana int32

	Inventory []ItemJson
	Equipment []EquipmentSlotJson
}

type ItemJson struct {
	Name string
	Amount int32
	Durability int32
}

type EquipmentSlotJson struct {
	Slot string
	Item ItemJson
}

var CHARACTER_SAVES_FOLDER = filepath.Join(database.SAVES_FOLDER, "characters")

func characterSavePath(character *Character) string {
	filename := fmt.Sprintf("%s.json", strings.ToLower(strings.ReplaceAll(character.Name, " ", "_")))
	return filepath.Join(CHARACTER_SAVES_FOLDER, filename)
}

func (world *World) SaveCharacter(character *Character) {
	// Convert character to json struct
	path := characterSavePath(character)
	characterJson := world.characterToJson(character)

	// Write the JSON
	err := database.WriteJsonAtomically(path, characterJson, true)
	if err != nil {
		log.Printf("Failed to write character JSON: %s.", err.Error())
		return
	}

	log.Printf("Saved character file %s.", path)
}

func (world *World) LoadCharacters() error {
	world.Characters = make(map[string]*Character)
	world.PlayerCharacters = make(map[string][]string)

	files, err := os.ReadDir(CHARACTER_SAVES_FOLDER)
	if err != nil {
		return fmt.Errorf("Failed to read character save folder: %s", err.Error())
	}

	var problems []error
	for _, fileEntry := range files {
		path := filepath.Join(CHARACTER_SAVES_FOLDER, fileEntry.Name())

		if !strings.HasSuffix(path, ".json") {
			log.Printf("Ignoring non-JSON file %s in character saves folder.", path)
			continue
		}

		character, err := world.loadCharacter(path)
		if err != nil {
			problems = append(problems, fmt.Errorf("%s: %s", path, err.Error()))
			continue
		}

		world.AddCharacter(character.PlayerKey, character)
	}

	return errors.Join(problems...)
}

func (world *World) loadCharacter(path string) (*Character, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
	var characterJson CharacterJson
	err = json.Unmarshal(data, &characterJson)
	if err != nil {
		return nil, err
	}

	// Convert to character
	character, err := world.characterFromJson(&characterJson)
	return character, err
}

func deleteCharacter(character *Character) {
	path := characterSavePath(character)
	err := os.Remove(path)
	if err != nil {
		log.Printf("Error deleting character save %s: %s", path, err.Error())
		return
	}

	log.Printf("Deleted character save file %s.", path)
}

func (world *World) characterToJson(character *Character) CharacterJson {
	characterJson := CharacterJson {
		PlayerKey: character.PlayerKey,
		Name: character.Name,
		Room: world.Rooms[character.Room].Position,
		RoomsDiscovered: make([]RoomPosition, 0, len(world.Rooms)),

		// We save the lowercase name because world.CharacterData and world.RaceData use lowercase names for keys
		Race: strings.ToLower(character.Race.Name),
		Class: strings.ToLower(character.Class.Name),

		Experience: character.Experience,

		Skills: character.Skills.ToStringArray(),

		Health: character.Health,
		Mana: character.Mana,

		Inventory: make([]ItemJson, 0, len(character.Inventory.Items)),
		Equipment: make([]EquipmentSlotJson, 0, EQUIPMENT_SLOT_COUNT),
	}

	// Rooms discovered
	for index := range len(world.Rooms) {
		if bitset.Check(character.RoomsDiscovered, index) {
			characterJson.RoomsDiscovered = append(characterJson.RoomsDiscovered, world.Rooms[index].Position)
		}
	}

	// Inventory
	for index := range len(character.Inventory.Items) {
		characterJson.Inventory = append(characterJson.Inventory, itemToJson(&character.Inventory.Items[index]))
	}

	// Equipment
	for index := range EQUIPMENT_SLOT_COUNT {
		if !character.Equipment.IsSlotInUse[index] {
			continue
		}

		// A two-handed weapon marks the off hand as in use, but the item is saved with the main hand
		if index == EQUIPMENT_SLOT_OFF_HAND && character.Equipment.IsTwoHandedWeaponEquipped() {
			continue
		}

		characterJson.Equipment = append(characterJson.Equipment, EquipmentSlotJson {
			Slot: EquipmentSlot(index).String(),
			Item: itemToJson(&character.Equipment.SlotItem[index]),
		})
	}

	return characterJson
}

func (world *World) characterFromJson(characterJson *CharacterJson) (*Character, error) {
	character := &Character{}

	character.PlayerKey = characterJson.PlayerKey
	character.Name = characterJson.Name

	// Room
	var exists bool
	character.Room, exists = world.RoomPositionToIndex[characterJson.Room]
	if !exists {
		return nil, fmt.Errorf("No room exists with position %s", characterJson.Room.String())
	}

	// Rooms discovered
	character.RoomsDiscovered = bitset.New(len(world.Rooms))
	for _, roomPosition := range characterJson.RoomsDiscovered {
		index, exists := world.RoomPositionToIndex[roomPosition]
		if !exists {
			return nil, fmt.Errorf("Character has discovered room with position %s, but no room at that position exists.", roomPosition.String())
		}

		bitset.Set(character.RoomsDiscovered, index, true)
	}

	// Race
	character.Race, exists = world.RaceData[characterJson.Race]
	if !exists {
		return nil, fmt.Errorf("Unrecognized race '%s'.", characterJson.Race)
	}

	// Class
	character.Class, exists = world.ClassData[characterJson.Class]
	if !exists {
		return nil, fmt.Errorf("Unrecognized class '%s'.", characterJson.Class)
	}

	// Character level is cached on the Character struct but is not saved in Json
	// We re-calculate their level when we load their character file
	// If we change the level scaling, character levels will automatically update to match
	character.Level = LevelFromExpTotal(characterJson.ExpTotal)
	character.Experience = characterJson.Experience
	character.calculateStats()

	character.Health = characterJson.Health
	character.Mana = characterJson.Mana

	// Skills
	var err error
	character.Skills, err = skillFlagsFromStringArray(characterJson.Skills)
	if err != nil {
		return nil, err
	}

	// Inventory
	character.Inventory = NewInventory()
	for index := range characterJson.Inventory {
		item, err := world.itemFromJson(&characterJson.Inventory[index])
		if err != nil {
			return nil, err
		}

		character.Inventory.AddItem(item)
	}

	// Equipment
	character.Equipment = NewEquipment()
	for index := range characterJson.Equipment {
		equipmentSlotJson := &characterJson.Equipment[index]

		slot, ok := util.EnumFromString[EquipmentSlot](equipmentSlotJson.Slot, EQUIPMENT_SLOT_COUNT)
		if !ok {
			return nil, fmt.Errorf("Character has equipment in slot '%s' which does not exist.", equipmentSlotJson.Slot)
		}

		item, err := world.itemFromJson(&equipmentSlotJson.Item)
		if err != nil {
			return nil, err
		}

		_, wasEquipped := character.Equipment.Equip(slot, item)
		if !wasEquipped {
			return nil, fmt.Errorf("Character has item %s equipped in slot %s, but it cannot be equipped there.", item.Name, slot.String())
		}
	}

	return character, nil
}

func itemToJson(item *Item) ItemJson {
	return ItemJson {
		Name: item.Name,
		Amount: item.Amount,
		Durability: item.Durability,
	}
}

func (world *World) itemFromJson(itemJson *ItemJson) (Item, error) {
	itemType, exists := world.ItemTypes[itemJson.Name]
	if !exists {
		return Item{}, fmt.Errorf("Character has item %s which does not exist.", itemJson.Name)
	}

	return world.createItem(CreateItemParams {
		itemType: itemType,
		amount: itemJson.Amount,
		durability: itemJson.Durability,
	}), nil
}
