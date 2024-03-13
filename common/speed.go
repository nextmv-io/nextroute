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
	switch unit.DistanceUnit() {
	case Kilometers:
		meters *= factorKilometersToMeters
	case Miles:
		meters *= factorMilesToMeters
	}
	meters *= 1.0 / unit.Duration().Hours()
	return &speedImpl{
		metersPerHour: meters,
	}
}

// KilometersPerHour is a speed unit of kilometers per hour.
var KilometersPerHour = &speedUnitImpl{
	distanceUnit: Kilometers,
	duration:     time.Hour,
}

// MilesPerHour is a speed unit of miles per hour.
var MilesPerHour = &speedUnitImpl{
	distanceUnit: Miles,
	duration:     time.Hour,
}

// MetersPerSecond is a speed unit of meters per second.
var MetersPerSecond = &speedUnitImpl{
	distanceUnit: Meters,
	duration:     time.Second,
}

// NewSpeedUnit returns a new speed unit.
func NewSpeedUnit(
	distanceUnit DistanceUnit,
	duration time.Duration,
) SpeedUnit {
	return &speedUnitImpl{
		distanceUnit: distanceUnit,
		duration:     duration,
	}
}

// Speed is the interface for a speed.
type Speed interface {
	// Value returns the speed in the specified unit.
	Value(unit SpeedUnit) float64
}

// SpeedUnit represents a unit of speed.
type SpeedUnit interface {
	// DistanceUnit returns the distance unit of the speed unit.
	DistanceUnit() DistanceUnit
	// Duration returns the duration of the speed unit.
	Duration() time.Duration
}

type speedUnitImpl struct {
	distanceUnit DistanceUnit
	duration     time.Duration
}

func (s *speedUnitImpl) DistanceUnit() DistanceUnit {
	return s.distanceUnit
}

func (s *speedUnitImpl) Duration() time.Duration {
	return s.duration
}

type speedImpl struct {
	metersPerHour float64
}

func (s *speedImpl) String() string {
	return fmt.Sprintf("%v meters/hour", s.metersPerHour)
}

func (s *speedImpl) Value(unit SpeedUnit) float64 {
	distancePerHour := s.metersPerHour
	switch unit.DistanceUnit() {
	case Kilometers:
		distancePerHour *= factorMetersToKilometers
	case Miles:
		distancePerHour *= factorMetersToMiles
	}
	return distancePerHour * unit.Duration().Hours()
}
