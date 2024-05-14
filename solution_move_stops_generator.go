// © 2019-present nextmv.io inc

package nextroute

import (
	"github.com/nextmv-io/nextroute/common"
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
	stops SolutionStops,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) chan SolutionMoveStops {
	ch := make(chan SolutionMoveStops)
	go func() {
		defer close(ch)
		SolutionMoveStopsGenerator(
			vehicle,
			planUnit,
			func(move SolutionMoveStops) {
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
	vehicle SolutionVehicle,
	planUnit SolutionPlanUnit,
	quit <-chan struct{},
	stops SolutionStops,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) chan SolutionMoveStops {
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
	vehicle SolutionVehicle,
	planUnit SolutionPlanUnit,
	yield func(move SolutionMoveStops),
	stops SolutionStops,
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
	yield func(move SolutionMoveStops),
	stops SolutionStops,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
	shouldStop func() bool,
) {
	source := common.Map(stops, func(stop SolutionStop) solutionStopImpl {
		return stop.(solutionStopImpl)
	})
	target := common.Map(vehicle.SolutionStops(), func(stop SolutionStop) solutionStopImpl {
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

	generate(positions, locations, source, target, func() {
		m.(*solutionMoveStopsImpl).reset()
		m.(*solutionMoveStopsImpl).planUnit = planUnit
		m.(*solutionMoveStopsImpl).stopPositions = positions
		m.(*solutionMoveStopsImpl).allowed = false
		m.(*solutionMoveStopsImpl).valueSeen = 1
		yield(m)
	}, shouldStop)
}

func isNotAllowed(from, to solutionStopImpl) bool {
	fromModelStop := from.modelStop()
	toModelStop := to.modelStop()
	model := fromModelStop.model

	return model.disallowedSuccessors[fromModelStop.index][toModelStop.index]
}

func mustBeNeighbours(from, to solutionStopImpl) bool {
	fromModelStop := from.modelStop()
	if !fromModelStop.HasPlanStopsUnit() {
		return false
	}
	toModelStop := to.modelStop()

	return fromModelStop.
		planUnit.(*planMultipleStopsImpl).
		dag.(*directedAcyclicGraphImpl).
		hasDirectArc(fromModelStop.index, toModelStop.index)
}

func generate(
	stopPositions []stopPositionImpl,
	combination []int,
	source []solutionStopImpl,
	target []solutionStopImpl,
	yield func(),
	shouldStop func() bool,
) {
	if shouldStop() {
		return
	}

	if len(combination) == len(source) {
		yield()
		return
	}

	start := 0
	if len(combination) > 0 {
		start = combination[len(combination)-1] - 1
	}

	for i := start; i < len(target)-1; i++ {
		if i > 0 && mustBeNeighbours(target[i], target[i+1]) {
			continue
		}
		combination = append(combination, i+1)

		positionIdx := len(combination) - 1

		stopPositions[positionIdx].previousStopIndex = target[combination[positionIdx]-1].index
		stopPositions[positionIdx].nextStopIndex = target[combination[positionIdx]].index

		if positionIdx > 0 {
			if combination[positionIdx] == combination[positionIdx-1] {
				stopPositions[positionIdx].previousStopIndex = stopPositions[positionIdx-1].stopIndex
				stopPositions[positionIdx-1].nextStopIndex = stopPositions[positionIdx].stopIndex
			} else {
				stopPositions[positionIdx-1].nextStopIndex = target[combination[positionIdx-1]].index
				if mustBeNeighbours(stopPositions[positionIdx-1].stop(), stopPositions[positionIdx].stop()) {
					break
				}
			}

			if isNotAllowed(stopPositions[positionIdx-1].stop(), stopPositions[positionIdx-1].next()) {
				combination = combination[:positionIdx]
				if stopPositions[positionIdx-1].nextStopIndex != stopPositions[positionIdx].previousStopIndex {
					break
				}
				continue
			}
		}

		if isNotAllowed(stopPositions[positionIdx].previous(), stopPositions[positionIdx].stop()) {
			combination = combination[:positionIdx]
			continue
		}

		generate(stopPositions, combination, source, target, yield, shouldStop)

		combination = combination[:len(combination)-1]
	}
}
