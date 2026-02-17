// © 2019-present nextmv.io inc

package common

import (
	"fmt"
	"math"
)

// NewFastHaversine returns a new FastHaversine.
func NewFastHaversine(lat float64) FastHaversine {
	const RE = 6378.137          // equatorial radius
	const FE = 1 / 298.257223563 // flattening

	const E2 = FE * (2 - FE)
	const RAD = math.Pi / 180
	const m = RAD * RE * 1000
	coslat := math.Cos(lat * RAD)
	w2 := 1 / (1 - E2*(1-coslat*coslat))
	w := math.Sqrt(w2)

	return FastHaversine{
		kx: m * w * coslat,        // based on normal radius of curvature
		ky: m * w * w2 * (1 - E2), // based on meridonal radius of curvature
	}
}

// FastHaversine is a fast approximation of the haversine distance.
type FastHaversine struct {
	kx float64
	ky float64
}

func wrap(deg float64) float64 {
	for deg < -180 {
		deg += 360
	}
	for deg > 180 {
		deg -= 360
	}
	return deg
}

// Distance returns the distance between two locations in meters.
func (f FastHaversine) Distance(from, to Location) (float64, error) {
	if !from.IsValid() || !to.IsValid() {
		return 0.0,
			fmt.Errorf(
				"from (lon: %f, lat: %f) (valid = %t) or "+
					"to (lon: %f, lat: %f) (valid = %t) are invalid",
				from.Longitude(),
				from.Latitude(),
				from.IsValid(),
				to.Longitude(),
				to.Latitude(),
				to.IsValid(),
			)
	}
	dx := wrap(from.Longitude()-to.Longitude()) * f.kx
	dy := (from.Latitude() - to.Latitude()) * f.ky
	return math.Sqrt(dx*dx + dy*dy), nil
}
