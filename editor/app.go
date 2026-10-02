package main

import (
	"context"
	"mud/world"
)

type EditorState struct {
	appContext context.Context

	world *world.World

	gridWidth int
	gridHeight int
	grid []int
}

func NewEditorState() *EditorState {
	gridWidth := 8
	gridHeight := 8
	grid := make([]int, gridWidth * gridHeight)
	for index := range gridWidth * gridHeight {
		grid[index] = world.ROOM_NONE
	}

	return &EditorState {
		world: &world.World{},

		gridWidth: gridWidth,
		gridHeight: gridHeight,
		grid: grid,
	}
}

func (state *EditorState) onStartup(ctx context.Context) {
	state.appContext = ctx
}

func (state *EditorState) GetWorld() *world.World {
	return state.world
}

func (state *EditorState) GetItemData() []*world.ItemData {
	return world.ITEM_DATA
}
