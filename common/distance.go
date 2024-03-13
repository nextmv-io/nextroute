// © 2019-present nextmv.io inc

// Package common contains common types and functions.
package common

import "fmt"

// DistanceUnit is the unit of distance.
type DistanceUnit int

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

	return Distance{
		meters: value,
		unit:   unit,
	}
}

const (
	// Kilometers is 1000 meters.
	Kilometers DistanceUnit = iota
	// Meters is the distance travelled by light in a vacuum in
	// 1/299,792,458 seconds.
	Meters
	// Miles is 1609.34 meters.
	Miles
)

const (
	factorMetersToKilometers = 0.001
	factorMetersToMiles      = 0.000621371
)

const (
	factorKilometersToMeters = 1000
	factorMilesToMeters      = 1609.34
)

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

// Distance is a distance in a given unit.
type Distance struct {
	meters float64
	unit   DistanceUnit
}

// Value returns the distance in the specified unit.
func (d Distance) Value(unit DistanceUnit) float64 {
	returnValue := d.meters
	switch unit {
	case Kilometers:
		returnValue *= factorMetersToKilometers
	case Miles:
		returnValue *= factorMetersToMiles
	}
	return returnValue
}
