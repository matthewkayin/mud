package world

import (
	"fmt"
	"github.com/mmcdole/lunar"
)

type deferredSignal struct {
	emitter lua.Value
	signal string
	args []lua.Value
}

func (world *World) subscribeToSignal(signal string, listener *lua.Function) {
	_, exists := world.signalListeners[signal]
	if !exists {
		world.signalListeners[signal] = make([]*lua.Function, 0, 1)
	}

	world.signalListeners[signal] = append(world.signalListeners[signal], listener)
}

// Calls every listener of the signal with the emitter followed by args. Lua can't be
// re-entered, so a signal emitted while a script is running is deferred until that script returns.
func (world *World) emitSignal(emitter lua.Value, signal string, args ...lua.Value) {
	if world.scriptIsRunning {
		world.deferredSignals = append(world.deferredSignals, deferredSignal {
			emitter: emitter,
			signal: signal,
			args: args,
		})
		return
	}

	listeners, exists := world.signalListeners[signal]
	if !exists {
		return
	}

	listenerArgs := make([]lua.Value, 0, len(args) + 1)
	listenerArgs = append(listenerArgs, emitter)
	listenerArgs = append(listenerArgs, args...)
	for _, listener := range listeners {
		world.scriptCall(listener.Value(), fmt.Sprintf("signal %s listener", signal), listenerArgs...)
	}
}

// Emits the signals that were emitted while a script was running
func (world *World) handleDeferredSignals() {
	if world.isHandlingDeferredSignals {
		return
	}
	world.isHandlingDeferredSignals = true
	defer func() {
		world.isHandlingDeferredSignals = false
	}()

	// Listeners can emit more signals, which are appended to the list while it is being handled
	for index := 0; index < len(world.deferredSignals); index++ {
		deferred := world.deferredSignals[index]
		world.emitSignal(deferred.emitter, deferred.signal, deferred.args...)
	}
	world.deferredSignals = world.deferredSignals[:0]
}
