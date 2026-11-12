package world

type MobData struct {
	Name string
	Room int

	Health int32
	Mana int32

	Inventory Inventory
}
