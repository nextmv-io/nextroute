// © 2019-present nextmv.io inc

package common

import (
	"fmt"
	"time"
)

// NewSpeed creates a new speed.
func NewSpeed(
	distance float64,
	unit SpeedUnit,
) Speed {
	meters := distance
	switch unit.distanceUnit {
	case Kilometers:
		meters *= factorKilometersToMeters
	case Miles:
		meters *= factorMilesToMeters
	}
	meters *= 1.0 / unit.duration.Hours()
	return Speed(meters)
}

// KilometersPerHour is a speed unit of kilometers per hour.
var KilometersPerHour = SpeedUnit{
	distanceUnit: Kilometers,
	duration:     time.Hour,
}

// MilesPerHour is a speed unit of miles per hour.
var MilesPerHour = SpeedUnit{
	distanceUnit: Miles,
	duration:     time.Hour,
}

// MetersPerSecond is a speed unit of meters per second.
var MetersPerSecond = SpeedUnit{
	distanceUnit: Meters,
	duration:     time.Second,
}

// NewSpeedUnit returns a new speed unit.
func NewSpeedUnit(
	distanceUnit DistanceUnit,
	duration time.Duration,
) SpeedUnit {
	return SpeedUnit{
		distanceUnit: distanceUnit,
		duration:     duration,
	}
}

// SpeedUnit represents a unit of speed.
type SpeedUnit struct {
	distanceUnit DistanceUnit
	duration     time.Duration
}

// Speed stands for distance over time. It is represented in meters per hour.
type Speed float64

func (s Speed) String() string {
	return fmt.Sprintf("%v meters/hour", float64(s))
}

// Value returns the speed in the specified unit.
func (s Speed) Value(unit SpeedUnit) float64 {
	return float64(s) * unit.distanceUnit.convertFromMeters() * unit.duration.Hours()
}
