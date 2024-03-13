// © 2019-present nextmv.io inc

package nextroute

import "github.com/nextmv-io/nextroute/common"

// haversineDistance processes the locations to make sure that they are valid
// to return the corresponding distance.
func haversineDistance(from, to common.Location) common.Distance {
	// this check is redudant here, as it's already done in the
	// in the Haversine function.
	// However we have to check it here to return a 0 distance without a heap allocation.
	// If common.Haversine returns an error then this will cause a heap allocation.
	// TODO: room for optimization here.
	if !from.IsValid() || !to.IsValid() {
		return common.NewDistance(0., common.Meters)
	}
	v, err := common.Haversine(from, to)
	if err != nil {
		panic(err)
	}
	return v
}
