// © 2019-present nextmv.io inc

package nextroute

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
	vehicle SolutionVehicle,
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
		vehicle,
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
		vehicle,
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
	vehicle SolutionVehicle,
	planUnit *solutionPlanStopsUnitImpl,
	yield func(move SolutionMoveStops),
	stops SolutionStops,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
	shouldStop func() bool,
) {
	source := stops
	target := vehicle.SolutionStops()
	m := preAllocatedMoveContainer.singleStopPosSolutionMoveStop
	m.(*solutionMoveStopsImpl).reset()
	m.(*solutionMoveStopsImpl).planUnit = planUnit
	m.(*solutionMoveStopsImpl).allowed = false
	if len(source) == 0 {
		yield(m)
		return
	}

	// TODO: we can reuse the stopPositions slice from m
	positions := make([]StopPosition, len(source))
	for idx := range source {
		positions[idx].stopIndex = source[idx].index
		positions[idx].solution = source[idx].solution
	}

	locations := make([]int, 0, len(source))

	model := vehicle.solution.model.(*modelImpl)
	if model.hasDisallowedSuccessors() || model.hasDirectSuccessors {
		generate(positions, locations, source, target, func() {
			m.(*solutionMoveStopsImpl).reset()
			m.(*solutionMoveStopsImpl).planUnit = planUnit
			m.(*solutionMoveStopsImpl).stopPositions = positions
			m.(*solutionMoveStopsImpl).allowed = false
			m.(*solutionMoveStopsImpl).valueSeen = 1
			yield(m)
		}, shouldStop)
	} else {
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
}

func isNotAllowed(model *modelImpl, from, to SolutionStop) bool {
	if !model.hasDisallowedSuccessors() {
		return false
	}
	return model.disallowedSuccessors[from.ModelStopIndex()][to.ModelStopIndex()]
}

func mustBeNeighbours(model *modelImpl, from, to SolutionStop) bool {
	if !model.hasDirectSuccessors {
		return false
	}

	fromModelStop := from.modelStop()
	if !fromModelStop.HasPlanStopsUnit() {
		return false
	}

	return fromModelStop.
		planUnit.(*planMultipleStopsImpl).
		dag.(*directedAcyclicGraphImpl).
		hasDirectArc(fromModelStop.index, to.ModelStopIndex())
}

func generate(
	stopPositions []StopPosition,
	combination []int,
	source []SolutionStop,
	target []SolutionStop,
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

	model := target[0].modelStop().model

	for i := start; i < len(target)-1; i++ {
		if i > 0 && mustBeNeighbours(model, target[i], target[i+1]) {
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
				if mustBeNeighbours(model, stopPositions[positionIdx-1].Stop(), stopPositions[positionIdx].Stop()) {
					break
				}
			}

			if isNotAllowed(model, stopPositions[positionIdx-1].Stop(), stopPositions[positionIdx-1].Next()) {
				combination = combination[:positionIdx]
				if stopPositions[positionIdx-1].nextStopIndex != stopPositions[positionIdx].previousStopIndex {
					break
				}
				continue
			}
		}

		if isNotAllowed(model, stopPositions[positionIdx].Previous(), stopPositions[positionIdx].Stop()) {
			combination = combination[:positionIdx]
			continue
		}

		generate(stopPositions, combination, source, target, yield, shouldStop)

		combination = combination[:len(combination)-1]
	}
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
