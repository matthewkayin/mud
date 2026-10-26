package world

import (
	"fmt"
	"github.com/mmcdole/lunar"
)

func (world *World) subscribeToSignal(signal string, listener *lua.Function) {
	_, exists := world.signalListeners[signal]
	if !exists {
		world.signalListeners[signal] = make([]*lua.Function, 0, 1)
	}

	world.signalListeners[signal] = append(world.signalListeners[signal], listener)
}

func (world *World) emitSignal(signal string, args ...lua.Value) {
	listeners, exists := world.signalListeners[signal]
	if !exists {
		return
	}

	for _, listener := range listeners {
		world.scriptCall(listener.Value(), fmt.Sprintf("signal %s listener", signal), args...)
	}
}
