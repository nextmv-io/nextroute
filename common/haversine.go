// © 2019-present nextmv.io inc

package common

import (
	"fmt"
	"math"
)

// Haversine calculates the distance between two locations using the
// Haversine formula. Haversine is a good approximation for short
// distances (up to a few hundred kilometers).
// It returns an error if either of the locations are invalid.
func Haversine(from, to Location) (Distance, error) {
	if !from.IsValid() || !to.IsValid() {
		return Distance{},
			fmt.Errorf(
				"from (lon: %f, lat: %f) (valid = %t) or "+
					"to (lon: %f, lat: %f) (valid = %v) are invalid",
				from.Longitude(),
				from.Latitude(),
				from.IsValid(),
				to.Longitude(),
				to.Latitude(),
				to.IsValid(),
			)
	}

	return Haversine0(from, to), nil
}

// Haversine0 calculates the distance between two locations using the
// Haversine formula. Distance is 0 if either location is invalid.
func Haversine0(from, to Location) Distance {
	if !from.IsValid() || !to.IsValid() {
		return Distance{meters: 0, unit: Meters}
	}
	x1 := from.longitudeRadians
	y1 := from.latitudeRadians
	x2 := to.longitudeRadians
	y2 := to.latitudeRadians

	dx := x1 - x2
	dy := y1 - y2

	sdy := math.Sin(dy / 2)
	sdx := math.Sin(dx / 2)

	a := (sdy * sdy) + math.Cos(y1)*math.Cos(y2)*sdx*sdx
	return Distance{
		meters: 2 * radius * math.Atan2(math.Sqrt(a), math.Sqrt(1-a)),
		unit:   Meters,
	}
}

const radius = 6371 * 1000
