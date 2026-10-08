package world

import (
	"fmt"
	"log"
	"github.com/mmcdole/lunar"
)

type ItemId int32
const (
	ITEM_GOLD = iota
	ITEM_DUMMY_MATERIAL
	ITEM_SWORD
	ITEM_AXE
	ITEM_ARMOR_CHAINMAIL
	ITEM_SPELLBOOK_FIREBOLT
	ITEM_SPELLBOOK_CURE
	ITEM_POTION_HEALTH
	ITEM_POTION_MANA
	ITEM_RECIPE_HEALTH_POT
	ITEM_RECIPE_MANA_POT
	ITEM_RECIPE_SWORD
	ITEM_RECIPE_AXE
)

type Item struct {
	Id ItemId
	Amount int32
	Durability int32
}

// ITEM DATA

type ItemKind int
const (
	ITEM_KIND_CONSUMABLE = iota
	ITEM_KIND_EQUIPMENT_ONE_HANDED
	ITEM_KIND_EQUIPMENT_TWO_HANDED
	ITEM_KIND_EQUIPMENT_OUTFIT
	ITEM_KIND_EQUIPMENT_ACCESSORY
	ITEM_KIND_EQUIPMENT_SPELLBOOK
	ITEM_KIND_SPELL_SCROLL
	ITEM_KIND_RECIPE
	ITEM_KIND_MISC // Indicates an item which has no special properties, like gold or a material
	ITEM_KIND_COUNT
)

type ItemDataConsumable struct {
	onUse *lua.Function
}

type ItemDataSpellScroll struct {
	Spell SpellId
}

type ItemDataWeapon struct {
	Damage int32
	MaxDurability int32
	StatBonuses StatBlock
	StatRequirements StatBlock
}

type ItemDataOutfit struct {
	Armor int32
	MaxDurability int32
	StealthPenality float32
	StatBonuses StatBlock
	StatRequirements StatBlock
}

type ItemDataAccessory struct {
	StatBonuses StatBlock
	StatRequirements StatBlock
}

type ItemDataSpellbook struct {
	Spell SpellId
	StatRequirements StatBlock
}

type ItemDataRecipe struct {
	Recipe RecipeId
}

type ItemData struct {
	Name string
	Description string
	Kind ItemKind
	Size int32
	Data any
}

// LOAD

var ITEM_DATA []*ItemData
var ITEM_NAME_TO_ID map[string]ItemId

const ITEM_DATA_FOLDER = WORLD_DATA_FOLDER + "/items"

func (world *World) loadItemData() {
	// Read item folder
	log.Printf("Loading item data...")
	paths, err := scriptGetFilesFrom(ITEM_DATA_FOLDER)
	if err != nil {
		log.Fatalf("Error opening item data folder: %s", err.Error())
	}

	ITEM_DATA = make([]*ItemData, 0, len(paths))
	ITEM_NAME_TO_ID = make(map[string]ItemId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			log.Fatalf("%s: %s", path, err.Error())
		}

		// Parse item data
		parser := ScriptParser{}
		itemData := parser.parseItem(table)
		if itemData == nil {
			log.Fatalf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateItemName := ITEM_NAME_TO_ID[itemData.Name]
		if duplicateItemName {
			log.Fatalf("Item %s has name '%s' which is a duplicate of another item.", path, itemData.Name)
		}

		// Store item in ITEM_DATA
		ITEM_NAME_TO_ID[itemData.Name] = ItemId(len(ITEM_DATA))
		ITEM_DATA = append(ITEM_DATA, itemData)
		log.Printf("Loaded item '%s'.", itemData.Name)
	}

	log.Printf("All item data has been loaded.")
}

func (parser *ScriptParser) parseItem(table *lua.Table) *ItemData {
	itemData := &ItemData{}

	itemData.Name = parser.getString(table, "name")
	itemData.Description = parser.getString(table, "description")
	itemData.Size = parser.getInt32(table, "size")
	itemData.Kind = ItemKind(parser.getInt(table, "kind"))

	switch itemData.Kind {
		case ITEM_KIND_CONSUMABLE: {
			data := &ItemDataConsumable{}
			data.onUse = parser.getFunction(table, "on_use")
			itemData.Data = data
		}

		case ITEM_KIND_EQUIPMENT_ONE_HANDED, ITEM_KIND_EQUIPMENT_TWO_HANDED: {
			data := &ItemDataWeapon{}
			data.Damage = parser.getInt32(table, "damage")
			data.MaxDurability = parser.getInt32(table, "max_durability")
			data.StatBonuses = parser.getStatBlock(table, "stat_bonuses")
			data.StatRequirements = parser.getStatBlock(table, "stat_requirements")
			itemData.Data = data
		}

		case ITEM_KIND_EQUIPMENT_OUTFIT: {
			data := &ItemDataOutfit{}
			data.Armor = parser.getInt32(table, "armor")
			data.MaxDurability = parser.getInt32(table, "max_durability")
			data.StealthPenality = parser.getFloat32(table, "stealth_penalty")
			data.StatBonuses = parser.getStatBlock(table, "stat_bonuses")
			data.StatRequirements = parser.getStatBlock(table, "stat_requirements")
			itemData.Data = data
		}

		case ITEM_KIND_EQUIPMENT_ACCESSORY: {
			data := &ItemDataAccessory{}
			data.StatBonuses = parser.getStatBlock(table, "stat_bonuses")
			data.StatRequirements = parser.getStatBlock(table, "stat_requirements")
			itemData.Data = data
		}

		case ITEM_KIND_EQUIPMENT_SPELLBOOK: {
			data := &ItemDataSpellbook{}

			spellName := parser.getString(table, "spell")

			var exists bool
			data.Spell, exists = SPELL_NAME_TO_ID[spellName]
			if !exists {
				parser.addProblem(fmt.Errorf("Spellbook spell '%s' does not exist.", spellName))
			}

			data.StatRequirements = parser.getStatBlock(table, "stat_requirements")
			itemData.Data = data
		}

		case ITEM_KIND_SPELL_SCROLL: {
			data := &ItemDataSpellScroll{}

			spellName := parser.getString(table, "spell")

			var exists bool
			data.Spell, exists = SPELL_NAME_TO_ID[spellName]
			if !exists {
				parser.addProblem(fmt.Errorf("Spellbook spell '%s' does not exist.", spellName))
			}

			itemData.Data = data
		}

		case ITEM_KIND_RECIPE: {
			data := &ItemDataRecipe{}

			recipeName := parser.getString(table, "recipe")

			var exists bool
			data.Recipe, exists = RECIPE_NAME_TO_ID[recipeName]
			if !exists {
				parser.addProblem(fmt.Errorf("Recipe '%s' does not exist.", recipeName))
			}

			itemData.Data = data
		}

		case ITEM_KIND_MISC: {
			// Misc items have no item-specific data
		}

		default: {
			parser.addProblem(fmt.Errorf("Unrecognized item kind %d", itemData.Kind))
		}
	}

	if len(parser.problems) != 0 {
		return nil
	}

	return itemData
}

// HELPERS

func (itemData *ItemData) ItemIsOneHanded() bool {
	return itemData.Kind == ITEM_KIND_EQUIPMENT_ONE_HANDED ||
		itemData.Kind == ITEM_KIND_EQUIPMENT_SPELLBOOK
}

func (itemData *ItemData) ItemCanStack() bool {
	return itemData.Kind == ITEM_KIND_CONSUMABLE ||
		itemData.Kind == ITEM_KIND_SPELL_SCROLL ||
		itemData.Kind == ITEM_KIND_MISC
}

func ItemKindToString(kind ItemKind) string {
	switch kind {
		case ITEM_KIND_CONSUMABLE:
			return "Consumable"
		case ITEM_KIND_EQUIPMENT_ONE_HANDED:
			return "One-Handed Weapon"
		case ITEM_KIND_EQUIPMENT_TWO_HANDED:
			return "Two-handed Weapon"
		case ITEM_KIND_EQUIPMENT_OUTFIT:
			return "Outfit"
		case ITEM_KIND_EQUIPMENT_SPELLBOOK:
			return "Spellbook"
		case ITEM_KIND_SPELL_SCROLL:
			return "Spell Scroll"
		case ITEM_KIND_RECIPE:
			return "Recipe"
		case ITEM_KIND_MISC:
			return "Misc"
		default:
			panic(fmt.Sprintf("Item kind %d not handled", kind))
	}
}

func (item *Item) Size() int32 {
	return ITEM_DATA[item.Id].Size * item.Amount
}

func (item *Item) GetStatBonuses() *StatBlock {
	itemData := ITEM_DATA[item.Id]
	switch itemData.Kind {
		case ITEM_KIND_EQUIPMENT_ONE_HANDED, ITEM_KIND_EQUIPMENT_TWO_HANDED:
			weaponData := itemData.Data.(*ItemDataWeapon)
			return &weaponData.StatBonuses
		case ITEM_KIND_EQUIPMENT_OUTFIT:
			outfitData := itemData.Data.(*ItemDataOutfit)
			return &outfitData.StatBonuses
		case ITEM_KIND_EQUIPMENT_ACCESSORY:
			accessoryData := itemData.Data.(*ItemDataAccessory)
			return &accessoryData.StatBonuses
		default:
			return nil
	}
}

func (item *Item) GetStatRequirements() *StatBlock {
	itemData := ITEM_DATA[item.Id]
	switch itemData.Kind {
		case ITEM_KIND_EQUIPMENT_ONE_HANDED, ITEM_KIND_EQUIPMENT_TWO_HANDED:
			weaponData := itemData.Data.(*ItemDataWeapon)
			return &weaponData.StatRequirements
		case ITEM_KIND_EQUIPMENT_OUTFIT:
			outfitData := itemData.Data.(*ItemDataOutfit)
			return &outfitData.StatRequirements
		case ITEM_KIND_EQUIPMENT_ACCESSORY:
			accessoryData := itemData.Data.(*ItemDataAccessory)
			return &accessoryData.StatRequirements
		case ITEM_KIND_EQUIPMENT_SPELLBOOK:
			spellbookData := itemData.Data.(*ItemDataSpellbook)
			return &spellbookData.StatRequirements
		default:
			return nil
	}
}

func (item *Item) GetNameWithCondition() string {
	itemData := ITEM_DATA[item.Id]

	maxDurability := itemData.GetMaxDurability()
	if maxDurability == 0 {
		return itemData.Name
	}

	itemIsWeapon := itemData.Kind == ITEM_KIND_EQUIPMENT_ONE_HANDED || itemData.Kind == ITEM_KIND_EQUIPMENT_TWO_HANDED
	if item.Durability < maxDurability / 2 {
		return "Damaged " + itemData.Name
	} else if item.Durability > maxDurability && itemIsWeapon {
		return "Sharpened " + itemData.Name
	} else if item.Durability > maxDurability && !itemIsWeapon {
		return "Fortified " + itemData.Name
	} else {
		return itemData.Name
	}
}

func (item *Item) GetNameWithAmount() string {
	itemName := item.GetNameWithCondition()

	if item.Amount == 1 {
		return itemName
	}
	return fmt.Sprintf("%d %s", item.Amount, itemName)
}

func (itemData *ItemData) GetMaxDurability() int32 {
	switch itemData.Kind {
		case ITEM_KIND_EQUIPMENT_ONE_HANDED, ITEM_KIND_EQUIPMENT_TWO_HANDED:
			weaponData := itemData.Data.(*ItemDataWeapon)
			return weaponData.MaxDurability
		case ITEM_KIND_EQUIPMENT_OUTFIT:
			outfitData := itemData.Data.(*ItemDataOutfit)
			return outfitData.MaxDurability
		case ITEM_KIND_EQUIPMENT_SPELLBOOK:
			spellbookData := itemData.Data.(*ItemDataSpellbook)
			spellData := SPELL_DATA[spellbookData.Spell]
			return spellData.CastsToLearn
		default:
			return 0
	}
}
