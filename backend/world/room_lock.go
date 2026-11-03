package world

import (
	"fmt"
)

const LOCK_UNLOCK_DURATION int32 = (5 * 60) / WORLD_SECONDS_PER_UPDATE

type LockType int
const (
	LOCK_TYPE_PHYSICAL LockType = iota
	LOCK_TYPE_MAGICAL
	LOCK_TYPE_COUNT
)

func (lockType LockType) String() string {
	switch lockType {
		case LOCK_TYPE_PHYSICAL:
			return "physical"
		case LOCK_TYPE_MAGICAL:
			return "magical"
		default:
			panic(fmt.Sprintf("%d is not a valid lock type.", lockType))
	}
}

// Locks are one-directional, they only block leaving the room through the lock's direction.
// A lock is active while its unlock timer is 0. Unlocking it starts the timer,
// and the lock becomes active again when the timer runs out.
type Lock struct {
	Type LockType
	Direction Direction
	unlockTimer int32
}

func (lock *Lock) IsActive() bool {
	return lock.unlockTimer == 0
}

func (lock *Lock) Unlock() {
	lock.unlockTimer = LOCK_UNLOCK_DURATION
}

func (lock *Lock) Relock() {
	lock.unlockTimer = 0
}

func (lock *Lock) update() {
	if lock.unlockTimer > 0 {
		lock.unlockTimer--
	}
}
