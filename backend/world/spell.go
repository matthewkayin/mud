package world

import (
	"github.com/mmcdole/lunar"
)

var SPELL_CAST_TIME_INSTANT int32 = 0

type SpellId int32

type SpellData struct {
	Name string
	Description string
	CastsToLearn int32

	ManaCost int32
	CastTime int32
	CanTargetPlayers bool

	onHit *lua.Function
}

var SPELL_DATA []*SpellData
