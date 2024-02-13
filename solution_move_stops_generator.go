package nextroute

import (
	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

// SolutionMoveStopsGeneratorChannel generates all possible moves for a given
// vehicle and plan unit.
//
// Example:
//
//	 quit := make(chan struct{})
//		defer close(quit)
//
//		for solutionMoveStopsImpl := range SolutionMoveStopsGeneratorChannel(
//			vehicle,
//			planUnit,
//			quit,
//		) {
//	 }
func SolutionMoveStopsGeneratorChannel(
	vehicle solutionVehicleImpl,
	planUnit *solutionPlanStopsUnitImpl,
	quit <-chan struct{},
	stops nextroute.SolutionStops,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) chan nextroute.SolutionMoveStops {
	ch := make(chan nextroute.SolutionMoveStops)
	go func() {
		defer close(ch)
		SolutionMoveStopsGenerator(
			vehicle,
			planUnit,
			func(move nextroute.SolutionMoveStops) {
				select {
				case <-quit:
					return
				case ch <- move:
				}
			},
			stops,
			preAllocatedMoveContainer,
			func() bool {
				select {
				case <-quit:
					return true
				default:
					return false
				}
			},
		)
	}()
	return ch
}

// SolutionMoveStopsGeneratorChannelTest is here only for testing purposes.
func SolutionMoveStopsGeneratorChannelTest(
	vehicle nextroute.SolutionVehicle,
	planUnit nextroute.SolutionPlanUnit,
	quit <-chan struct{},
	stops nextroute.SolutionStops,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) chan nextroute.SolutionMoveStops {
	return SolutionMoveStopsGeneratorChannel(
		vehicle.(solutionVehicleImpl),
		planUnit.(*solutionPlanStopsUnitImpl),
		quit,
		stops,
		preAllocatedMoveContainer,
	)
}

// SolutionMoveStopsGeneratorTest is here only for testing purposes.
func SolutionMoveStopsGeneratorTest(
	vehicle nextroute.SolutionVehicle,
	planUnit nextroute.SolutionPlanUnit,
	yield func(move nextroute.SolutionMoveStops),
	stops nextroute.SolutionStops,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
	shouldStop func() bool,
) {
	SolutionMoveStopsGenerator(
		vehicle.(solutionVehicleImpl),
		planUnit.(*solutionPlanStopsUnitImpl),
		yield,
		stops,
		preAllocatedMoveContainer,
		shouldStop,
	)
}

// SolutionMoveStopsGenerator generates all possible moves for a given vehicle and
// plan unit. The function yield is called for each solutionMoveStopsImpl.
func SolutionMoveStopsGenerator(
	vehicle solutionVehicleImpl,
	planUnit *solutionPlanStopsUnitImpl,
	yield func(move nextroute.SolutionMoveStops),
	stops nextroute.SolutionStops,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
	shouldStop func() bool,
) {
	source := common.Map(stops, func(stop nextroute.SolutionStop) solutionStopImpl {
		return stop.(solutionStopImpl)
	})
	target := common.Map(vehicle.SolutionStops(), func(stop nextroute.SolutionStop) solutionStopImpl {
		return stop.(solutionStopImpl)
	})
	m := preAllocatedMoveContainer.singleStopPosSolutionMoveStop
	m.(*solutionMoveStopsImpl).reset()
	m.(*solutionMoveStopsImpl).planUnit = planUnit
	m.(*solutionMoveStopsImpl).allowed = false
	if len(source) == 0 {
		yield(m)
		return
	}

	// TODO: we can reuse the stopPositions slice from m
	positions := make([]stopPositionImpl, len(source))
	for idx := range source {
		positions[idx].stopIndex = source[idx].index
		positions[idx].solution = source[idx].solution
	}

	locations := make([]int, 0, len(source))

	combineAscending(locations, len(source), len(target)-1, func(locations []int) {
		for idx, location := range locations {
			positions[idx].previousStopIndex = target[location-1].index
			positions[idx].nextStopIndex = target[location].index
			if idx > 0 && locations[idx-1] == location {
				positions[idx].previousStopIndex = source[idx-1].index
			}
			if idx < len(locations)-1 && locations[idx+1] == location {
				positions[idx].nextStopIndex = source[idx+1].index
			}
		}
		m.(*solutionMoveStopsImpl).reset()
		m.(*solutionMoveStopsImpl).planUnit = planUnit
		m.(*solutionMoveStopsImpl).stopPositions = positions
		m.(*solutionMoveStopsImpl).allowed = false
		m.(*solutionMoveStopsImpl).valueSeen = 1
		yield(m)
	}, shouldStop)
}

// combineAscending generates all combinations of n elements from m where
// the elements are in ascending order.
// 2,3 (2 elements can be at 3 locations) will generate:
// [1 1] first at first location, second at first location
// [1 2] first at first location, second at second location
// [1 3] first at first location, second at third location
// [2 2] first at second location, second at second location
// [2 3] first at second location, second at third location
// [3 3] first at third location, second at third location
// 3, 2 (3 elements can be at 2 locations) will generate:
// [1 1 1] first at first location, second at first location, third at first location
// [1 1 2] first at first location, second at first location, third at second location
// [1 2 2] first at first location, second at second location, third at second location
// [2 2 2] first at second location, second at second location, third at second location
// Being at the same location means element are next to each other.
// The function yield is called for each combination of size n.
func combineAscending(combination []int, n int, m int, yield func([]int), shouldStop func() bool) {
	if shouldStop() {
		return
	}
	if len(combination) == n {
		yield(combination)
		return
	}
	start := 0
	if len(combination) > 0 {
		start = combination[len(combination)-1] - 1
	}
	for i := start; i < m; i++ {
		combination = append(combination, i+1)
		combineAscending(combination, n, m, yield, shouldStop)
		combination = combination[:len(combination)-1]
	}
}
