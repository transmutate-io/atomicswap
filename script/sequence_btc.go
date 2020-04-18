package script

import (
	"time"

	"github.com/btcsuite/btcd/wire"
)

// SequenceNumberBTC represents the nSequence field in bitcoin
type SequenceNumberBTC int64

// Disable the lock
func (s SequenceNumberBTC) Disable() SequenceNumberBTC { return s | wire.SequenceLockTimeDisabled }

// IsSeconds sets the seconds bit (use blocks otherwise)
func (s SequenceNumberBTC) IsSeconds() SequenceNumberBTC { return s | wire.SequenceLockTimeIsSeconds }

// Set the sequence value from an int32
func (s SequenceNumberBTC) Set(ns int32) SequenceNumberBTC {
	return SequenceNumberBTC(ns&wire.SequenceLockTimeMask) | s
}

// Set the sequence value from a time.Duration
func (s SequenceNumberBTC) SetDuration(d time.Duration) SequenceNumberBTC {
	u := d / time.Second
	var i int32
	if u%512 != 0 {
		i = 1
	}
	return s.Set(int32(u/512) + i).IsSeconds()
}
