package script

import (
	"time"

	"github.com/btcsuite/btcd/wire"
)

// SequenceNumber represents the nSequence field
type SequenceNumber int64

// Disable the lock
func (s SequenceNumber) Disable() SequenceNumber { return s | wire.SequenceLockTimeDisabled }

// IsSeconds sets the seconds bit (use blocks otherwise)
func (s SequenceNumber) IsSeconds() SequenceNumber { return s | wire.SequenceLockTimeIsSeconds }

// Set the sequence value from an int32
func (s SequenceNumber) Set(ns int32) SequenceNumber {
	return SequenceNumber(ns&wire.SequenceLockTimeMask) | s
}

// Set the sequence value from a time.Duration
func (s SequenceNumber) SetDuration(d time.Duration) SequenceNumber {
	u := d / time.Second
	var i int32
	if u%512 != 0 {
		i = 1
	}
	return s.Set(int32(u/512) + i).IsSeconds()
}
