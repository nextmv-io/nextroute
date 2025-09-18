// © 2019-present nextmv.io inc

package nextroute

import "github.com/nextmv-io/nextroute/common"

// haversineDistance processes the locations to make sure that they are valid
// to return the corresponding distance.
func haversineDistance(from, to common.Location) common.Distance {
	if !from.IsValid() || !to.IsValid() {
		// this is important as some functionality depends on invalid locations having zero distance
		return common.NewDistance(0., common.Meters)
	}
	return common.HaversineUnsafe(from, to)
}
