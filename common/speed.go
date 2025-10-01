// © 2019-present nextmv.io inc

package common

import (
	"fmt"
	"time"
)

// Speed is the interface for a speed.
type Speed struct {
	metersPerHour float64
}

// NewSpeed creates a new speed.
func NewSpeed(
	distance float64,
	unit SpeedUnit,
) Speed {
	meters := distance
	switch unit.DistanceUnit() {
	case Kilometers:
		meters *= factorKilometersToMeters
	case Miles:
		meters *= factorMilesToMeters
	}
	meters *= 1.0 / unit.Duration().Hours()
	return Speed{
		metersPerHour: meters,
	}
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

// String returns the string representation of the speed.
func (s Speed) String() string {
	return fmt.Sprintf("%v meters/hour", s.metersPerHour)
}

// Value returns the speed in the specified unit.
func (s Speed) Value(unit SpeedUnit) float64 {
	distancePerHour := s.metersPerHour
	switch unit.DistanceUnit() {
	case Kilometers:
		distancePerHour *= factorMetersToKilometers
	case Miles:
		distancePerHour *= factorMetersToMiles
	}
	return distancePerHour * unit.duration.Hours()
}

// SpeedUnit represents a unit of speed.
type SpeedUnit struct {
	distanceUnit DistanceUnit
	duration     time.Duration
}

// DistanceUnit returns the distance unit of the speed unit.
func (s SpeedUnit) DistanceUnit() DistanceUnit {
	return s.distanceUnit
}

// Duration returns the duration of the speed unit.
func (s SpeedUnit) Duration() time.Duration {
	return s.duration
}
