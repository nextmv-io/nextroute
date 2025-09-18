// © 2019-present nextmv.io inc

package common

import (
	"fmt"
	"math"
)

// Haversine calculates the distance between two locations using the
// Haversine formula. Haversine is a good approximation for short
// distances (up to a few hundred kilometers).
func Haversine(from, to Location) (Distance, error) {
	if !from.IsValid() || !to.IsValid() {
		return 0,
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

	return HaversineUnsafe(from, to), nil
}

// HaversineUnsafe calculates the distance between two locations using the
// Haversine formula. This function does not check
// if the locations are valid.
func HaversineUnsafe(from, to Location) Distance {
	x1 := from.longitudeRadians
	y1 := from.latitudeRadians
	x2 := to.longitudeRadians
	y2 := to.latitudeRadians

	dx := x1 - x2
	dy := y1 - y2

	sdy := math.Sin(dy * 0.5)
	sdx := math.Sin(dx * 0.5)
	a := (sdy * sdy) + math.Cos(y1)*math.Cos(y2)*sdx*sdx

	return Distance(
		2 * radius * math.Atan2(math.Sqrt(a), math.Sqrt(1-a)),
	)
}

const (
	radius = 6371 * 1000
)
