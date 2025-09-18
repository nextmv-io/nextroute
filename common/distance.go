// © 2019-present nextmv.io inc

// Package common contains common types and functions.
package common

import "fmt"

// DistanceUnit is the unit of distance.
type DistanceUnit float64

// NewDistance returns a new distance.
func NewDistance(
	value float64,
	unit DistanceUnit,
) Distance {
	switch unit {
	case Kilometers:
		value *= factorKilometersToMeters
	case Miles:
		value *= factorMilesToMeters
	}

	return Distance(value)
}

// The constants are encoded by their most common conversion factor to meters.
const (
	// Kilometers is 1000 meters.
	Kilometers DistanceUnit = factorMetersToKilometers
	// Meters is the distance travelled by light in a vacuum in
	// 1/299,792,458 seconds.
	Meters = 1.0
	// Miles is 1609.34 meters.
	Miles = factorMetersToMiles
)

const (
	factorMetersToKilometers = 0.001
	factorMetersToMiles      = 0.000621371
)

const (
	factorKilometersToMeters = 1000
	factorMilesToMeters      = 1609.34
)

func (d DistanceUnit) convertFromMeters() float64 {
	return float64(d)
}

// String returns the string representation of the distance unit.
func (d DistanceUnit) String() string {
	switch d {
	case Kilometers:
		return "kilometers"
	case Meters:
		return "meters"
	case Miles:
		return "miles"
	}
	return fmt.Sprintf("unknown distance unit %v", int(d))
}

// Distance is a distance in a given unit. It is stored as meters internally.
type Distance float64

// Value returns the distance in the specified unit.
func (d Distance) Value(unit DistanceUnit) float64 {
	return float64(d) * unit.convertFromMeters()
}
