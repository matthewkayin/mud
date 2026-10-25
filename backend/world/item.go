package world

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mmcdole/lunar"
)

const ITEM_INFUSION_USES_UNLIMITED int32 = -1
const ITEM_USABLE_BY_JOB_ANY JobId = -1
const ITEM_REINFORCED_DURABILITY_PERCENT int32 = 120
const ITEM_REINFORCED_DAMAGE_PERCENT int32 = 120
const ITEM_REINFORCED_PENETRATION_PERCENT int32 = 120
const ITEM_REINFORCED_DEFLECTION_PERCENT int32 = 120

type ItemId int32

type ItemInfusion struct {
	ConditionId ConditionId
	ApplyChance float32
	ApplyDuration int32
	UsesRemaining int32
}

type Item struct {
	Id ItemId `ts_type:"string"`
	Amount int32
	Durability int32
	Infusion ItemInfusion
}

type ItemInfusionJson struct {
	ConditionId string
	ApplyChance float32
	ApplyDuration int32
	UsesRemaining int32
}

type ItemJson struct {
	Id string
	Amount int32
	Durability int32
	Infusion *ItemInfusionJson `json:",omitempty"`
}

func (item *Item) MarshalJSON() ([]byte, error) {
	if !itemIdIsValid(item.Id) {
		return nil, fmt.Errorf("Cannot save item with invalid item id %d.", item.Id)
	}

	itemJson := ItemJson {
		Id: ITEM_DATA[item.Id].Name,
		Amount: item.Amount,
		Durability: item.Durability,
	}

	// Only write the infusion if the item has one
	if item.HasInfusion() {
		itemJson.Infusion = &ItemInfusionJson {
			ConditionId: CONDITION_DATA[item.Infusion.ConditionId].Name,
			ApplyChance: item.Infusion.ApplyChance,
			ApplyDuration: item.Infusion.ApplyDuration,
			UsesRemaining: item.Infusion.UsesRemaining,
		}
	}

	return json.Marshal(&itemJson)
}

func (item *Item) UnmarshalJSON(data []byte) error {
	var itemJson ItemJson
	err := json.Unmarshal(data, &itemJson)
	if err != nil {
		return err
	}

	var exists bool
	item.Id, exists = ITEM_NAME_TO_ID[itemJson.Id]
	if !exists {
		return fmt.Errorf("No item ID matches '%s'.", itemJson.Id)
	}
	item.Amount = itemJson.Amount
	item.Durability = itemJson.Durability

	item.Infusion = ItemInfusion{}
	if itemJson.Infusion != nil {
		var conditionExists bool
		item.Infusion.ConditionId, conditionExists = CONDITION_NAME_TO_ID[itemJson.Infusion.ConditionId]
		if !conditionExists {
			return fmt.Errorf("Item infusion has condition '%s', which does not exist.", itemJson.Infusion.ConditionId)
		}
		item.Infusion.ApplyChance = itemJson.Infusion.ApplyChance
		item.Infusion.ApplyDuration = itemJson.Infusion.ApplyDuration
		item.Infusion.UsesRemaining = itemJson.Infusion.UsesRemaining
	}

	return nil
}

// ITEM DATA

type ItemKind int
const (
	ITEM_KIND_CONSUMABLE = iota
	ITEM_KIND_EQUIPMENT_ONE_HANDED
	ITEM_KIND_EQUIPMENT_TWO_HANDED
	ITEM_KIND_EQUIPMENT_SHIELD
	ITEM_KIND_EQUIPMENT_OUTFIT
	ITEM_KIND_EQUIPMENT_ACCESSORY
	ITEM_KIND_EQUIPMENT_SPELLBOOK
	ITEM_KIND_SPELL_SCROLL
	ITEM_KIND_RECIPE
	ITEM_KIND_MISC // Indicates an item which has no special properties, like gold or a material
	ITEM_KIND_COUNT
)

type ItemDataConsumable struct {
	UsableByJob JobId
	TargetType TargetType
	ValidateTarget *lua.Function
	onUse *lua.Function
}

type ItemDataSpellScroll struct {
	Spell SpellId
}

type ItemDataWeapon struct {
	Damage int32
	ArmorPenetration int32
	MaxDurability int32
	StatBonuses StatBlock
	StatRequirements StatBlock
	Prefix string
}

type ItemDataArmor struct {
	Deflection int32
	DamageResistance int32
	MaxDurability int32
	StealthPenalty float32
	StatBonuses StatBlock
	StatRequirements StatBlock
	Prefix string
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
	Price int32
	Size int32
	Notes string
	Data any
}

// LOAD

var ITEM_DATA []*ItemData
var ITEM_NAME_TO_ID map[string]ItemId
var ITEM_GOLD ItemId

const ITEM_DATA_FOLDER = "items"

func (world *World) loadItemData() {
	// Read item folder
	log.Printf("Loading item data...")
	paths, err := scriptGetFilesFrom(world.dataFolder, ITEM_DATA_FOLDER)
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
		itemData := parser.parseItemData(table)
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

	// Set the gold item ID
	var exists bool
	ITEM_GOLD, exists = ITEM_NAME_TO_ID["Gold"]
	if !exists {
		log.Fatalf("There must be an item called 'Gold'.")
	}

	log.Printf("All item data has been loaded.")
}

func (parser *ScriptParser) parseItemData(table *lua.Table) *ItemData {
	itemData := &ItemData{}

	itemData.Name = parser.getString(table, "name")
	itemData.Description = parser.getString(table, "description")
	itemData.Price = parser.getInt32(table, "price")
	parser.checkInt32NonNegative(itemData.Price, "price")
	itemData.Size = parser.getInt32(table, "size")
	itemData.Notes = parser.getString(table, "notes")

	kindString := parser.getString(table, "kind")

	var ok bool
	itemData.Kind, ok = EnumFromString[ItemKind](kindString, ITEM_KIND_COUNT)
	if !ok {
		parser.addProblem(fmt.Errorf("'%s' is not a valid item kind.", kindString))
	}

	switch itemData.Kind {
		case ITEM_KIND_CONSUMABLE: {
			data := &ItemDataConsumable{}

			jobName := parser.getString(table, "usable_by_job")
			if jobName == "any" {
				data.UsableByJob = ITEM_USABLE_BY_JOB_ANY
			} else {
				var exists bool
				data.UsableByJob, exists = JOB_NAME_TO_ID[jobName]
				if !exists {
					parser.addProblem(fmt.Errorf("No job exists called '%s'.", jobName))
				}
			}

			data.TargetType = parser.getTargetType(table, "target_type")
			if data.TargetType == TARGET_TYPE_INVENTORY_ITEM || data.TargetType == TARGET_TYPE_ARGUMENT {
				data.ValidateTarget = parser.getFunction(table, "validate_target")
			}

			data.onUse = parser.getFunction(table, "on_use")
			itemData.Data = data
		}

		case ITEM_KIND_EQUIPMENT_ONE_HANDED, ITEM_KIND_EQUIPMENT_TWO_HANDED: {
			data := &ItemDataWeapon{}
			data.Damage = parser.getInt32(table, "damage")
			parser.checkInt32NonNegative(data.Damage, "damage")
			data.ArmorPenetration = parser.getInt32(table, "armor_penetration")
			parser.checkInt32NonNegative(data.ArmorPenetration, "armor_penetration")
			data.MaxDurability = parser.getInt32(table, "max_durability")
			parser.checkInt32NonNegative(data.MaxDurability, "max_durability")
			data.StatBonuses = parser.getStatBlock(table, "stat_bonuses", false)
			data.StatRequirements = parser.getStatBlock(table, "stat_requirements", false)
			data.Prefix = parser.getString(table, "prefix")
			itemData.Data = data
		}

		case ITEM_KIND_EQUIPMENT_SHIELD, ITEM_KIND_EQUIPMENT_OUTFIT: {
			data := &ItemDataArmor{}
			data.Deflection = parser.getInt32(table, "deflection")
			parser.checkInt32NonNegative(data.Deflection, "deflection")
			data.DamageResistance = parser.getInt32(table, "damage_resistance")
			parser.checkInt32NonNegative(data.DamageResistance, "damage_resistance")
			data.MaxDurability = parser.getInt32(table, "max_durability")
			parser.checkInt32NonNegative(data.MaxDurability, "max_durability")
			data.StealthPenalty = parser.getFloat32(table, "stealth_penalty")
			data.StatBonuses = parser.getStatBlock(table, "stat_bonuses", false)
			data.StatRequirements = parser.getStatBlock(table, "stat_requirements", false)
			data.Prefix = parser.getString(table, "prefix")
			itemData.Data = data
		}

		case ITEM_KIND_EQUIPMENT_ACCESSORY: {
			data := &ItemDataAccessory{}
			data.StatBonuses = parser.getStatBlock(table, "stat_bonuses", false)
			data.StatRequirements = parser.getStatBlock(table, "stat_requirements", false)
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

			data.StatRequirements = parser.getStatBlock(table, "stat_requirements", false)
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

func (itemData *ItemData) ItemIsWeapon() bool {
	return itemData.Kind == ITEM_KIND_EQUIPMENT_ONE_HANDED ||
		itemData.Kind == ITEM_KIND_EQUIPMENT_TWO_HANDED
}

func (itemData *ItemData) ItemIsArmor() bool {
	return itemData.Kind == ITEM_KIND_EQUIPMENT_OUTFIT ||
		itemData.Kind == ITEM_KIND_EQUIPMENT_SHIELD
}

func (itemData *ItemData) GetItemNamePrefix() string {
	if itemData.ItemIsWeapon() {
		return itemData.Data.(*ItemDataWeapon).Prefix
	} else if itemData.ItemIsArmor() {
		return itemData.Data.(*ItemDataArmor).Prefix
	} else {
		return ""
	}
}

func (itemData *ItemData) ItemCanStack() bool {
	return itemData.Kind == ITEM_KIND_CONSUMABLE ||
		itemData.Kind == ITEM_KIND_SPELL_SCROLL ||
		itemData.Kind == ITEM_KIND_MISC
}

func itemKindToScriptEnum(kind ItemKind) string {
	switch kind {
		case ITEM_KIND_CONSUMABLE:
			return "CONSUMABLE"
		case ITEM_KIND_EQUIPMENT_ONE_HANDED:
			return "EQUIPMENT_ONE_HANDED"
		case ITEM_KIND_EQUIPMENT_TWO_HANDED:
			return "EQUIPMENT_TWO_HANDED"
		case ITEM_KIND_EQUIPMENT_SHIELD:
			return "EQUIPMENT_SHIELD"
		case ITEM_KIND_EQUIPMENT_OUTFIT:
			return "EQUIPMENT_OUTFIT"
		case ITEM_KIND_EQUIPMENT_ACCESSORY:
			return "EQUIPMENT_ACCESSORY"
		case ITEM_KIND_EQUIPMENT_SPELLBOOK:
			return "EQUIPMENT_SPELLBOOK"
		case ITEM_KIND_SPELL_SCROLL:
			return "SPELL_SCROLL"
		case ITEM_KIND_RECIPE:
			return "RECIPE"
		case ITEM_KIND_MISC:
			return "MISC"
		default:
			panic(fmt.Sprintf("Item kind %d not handled", kind))
	}
}

func (kind ItemKind) String() string {
	switch kind {
		case ITEM_KIND_CONSUMABLE:
			return "Consumable"
		case ITEM_KIND_EQUIPMENT_ONE_HANDED:
			return "One-Handed Weapon"
		case ITEM_KIND_EQUIPMENT_TWO_HANDED:
			return "Two-Handed Weapon"
		case ITEM_KIND_EQUIPMENT_SHIELD:
			return "Shield"
		case ITEM_KIND_EQUIPMENT_OUTFIT:
			return "Outfit"
		case ITEM_KIND_EQUIPMENT_ACCESSORY:
			return "Accessory"
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
		case ITEM_KIND_EQUIPMENT_OUTFIT, ITEM_KIND_EQUIPMENT_SHIELD:
			outfitData := itemData.Data.(*ItemDataArmor)
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
		case ITEM_KIND_EQUIPMENT_OUTFIT, ITEM_KIND_EQUIPMENT_SHIELD:
			outfitData := itemData.Data.(*ItemDataArmor)
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

	if item.Durability < maxDurability / 2 {
		return "Damaged " + itemData.Name
	} else if item.Durability > maxDurability {
		return itemData.GetItemNamePrefix() + " " + itemData.Name
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
		case ITEM_KIND_EQUIPMENT_OUTFIT, ITEM_KIND_EQUIPMENT_SHIELD:
			outfitData := itemData.Data.(*ItemDataArmor)
			return outfitData.MaxDurability
		case ITEM_KIND_EQUIPMENT_SPELLBOOK:
			spellbookData := itemData.Data.(*ItemDataSpellbook)
			spellData := SPELL_DATA[spellbookData.Spell]
			return spellData.CastsToLearn
		default:
			return 0
	}
}

func (item *Item) GetDurabilityPercent() int32 {
	itemData := ITEM_DATA[item.Id]
	maxDurability := itemData.GetMaxDurability()
	if maxDurability == 0 {
		return 0
	}

	return (item.Durability * 100) / itemData.GetMaxDurability()
}

func (item *Item) GetDamage() int32 {
	itemData := ITEM_DATA[item.Id]
	if !itemData.ItemIsWeapon() {
		return -1
	}

	weaponData := itemData.Data.(*ItemDataWeapon)
	damage := weaponData.Damage

	//give damage bonus for reinforcement
	if item.Durability > weaponData.MaxDurability {
		damage = (damage * ITEM_REINFORCED_DAMAGE_PERCENT) / 100
	}

	return damage
}

func (item *Item) GetArmorPenetration() int32 {
	itemData := ITEM_DATA[item.Id]
	if !itemData.ItemIsWeapon() {
		return 0
	}

	weaponData := itemData.Data.(*ItemDataWeapon)
	penetration := weaponData.ArmorPenetration

	//give damage bonus for reinforcement
	if item.Durability > weaponData.MaxDurability {
		penetration = (penetration * ITEM_REINFORCED_PENETRATION_PERCENT) / 100
	}

	return penetration
}

func (item *Item) GetDeflection() int32 {
	itemData := ITEM_DATA[item.Id]
	if !itemData.ItemIsArmor() {
		return -1
	}
	armorData := itemData.Data.(*ItemDataArmor)
	deflection := armorData.Deflection

	if item.Durability > armorData.MaxDurability {
		deflection = (deflection * ITEM_REINFORCED_DEFLECTION_PERCENT) / 100
	}

	return deflection
}

func (item *Item) GetSellPrice() int32 {
	itemData := ITEM_DATA[item.Id]

	// The less durability an item has, the less it sells for
	durabilityPercent := float32(1.0)
	maxDurability := itemData.GetMaxDurability()
	if maxDurability != 0 {
		durabilityPercent = float32(item.Durability) / float32(maxDurability)
	}

	// Items sell at 50% of their canonical price
	sellRate := float32(0.5)
	price := float32(itemData.Price) * sellRate * durabilityPercent

	return int32(price)
}

func (item *Item) HasInfusion() bool {
	return item.Infusion.UsesRemaining != 0
}

func (item *Item) isRepairable() bool {
	itemData := ITEM_DATA[item.Id]
	return itemData.ItemIsWeapon() || itemData.ItemIsArmor()
}

func (item *Item) repair() {
	itemData := ITEM_DATA[item.Id]
	item.Durability = itemData.GetMaxDurability()
}

func (item *Item) reinforce() {
	itemData := ITEM_DATA[item.Id]
	item.Durability = (itemData.GetMaxDurability() * ITEM_REINFORCED_DURABILITY_PERCENT) / 100
}
