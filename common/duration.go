// © 2019-present nextmv.io inc

package common

import (
	"fmt"
	"time"
)

// DurationUnit is the unit of duration.
type DurationUnit int

const (
	// NanoSecond is 1/1,000,000,000 of a second.
	NanoSecond DurationUnit = iota
	// MicroSecond is 1/1,000,000 of a second.
	MicroSecond
	// MilliSecond is 1/1,000 of a second.
	MilliSecond
	// Second is the SI unit of time.
	Second
	// Minute is 60 seconds.
	Minute
	// Hour is 60 minutes.
	Hour
	// Day is 24 hours.
	Day
)

// String returns the string representation of the duration unit.
func (d DurationUnit) String() string {
	switch d {
	case NanoSecond:
		return "nanoseconds"
	case MicroSecond:
		return "microseconds"
	case MilliSecond:
		return "milliseconds"
	case Second:
		return "seconds"
	case Minute:
		return "minutes"
	case Hour:
		return "hours"
	case Day:
		return "days"
	}
	return fmt.Sprintf("unknown duration unit %v", int(d))
}

// NewDuration returns a new duration by unit.
func NewDuration(unit DurationUnit) time.Duration {
	switch unit {
	case NanoSecond:
		return time.Nanosecond
	case MicroSecond:
		return time.Microsecond
	case MilliSecond:
		return time.Millisecond
	case Second:
		return time.Second
	case Minute:
		return 60 * time.Second
	case Hour:
		return time.Hour
	case Day:
		return 24 * time.Hour
	default:
		panic(fmt.Sprintf("unknown duration unit %v", int(unit)))
	}
}
