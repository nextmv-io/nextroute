// © 2019-present nextmv.io inc

package nextroute

import (
	"math"

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
	solution := vehicle.solution
	nrStops := len(stops)

	m := preAllocatedMoveContainer.solutionMoveStops
	m.(*solutionMoveStopsImpl).planUnit = planUnit
	m.(*solutionMoveStopsImpl).reset()

	if nrStops == 0 {
		yield(m)
		return
	}

	if cap(m.(*solutionMoveStopsImpl).stopPositions) < nrStops {
		m.(*solutionMoveStopsImpl).stopPositions = make([]stopPositionImpl, nrStops)
	}
	m.(*solutionMoveStopsImpl).stopPositions = m.(*solutionMoveStopsImpl).stopPositions[:nrStops]

	for idx, stop := range stops {
		m.(*solutionMoveStopsImpl).stopPositions[idx].stopIndex = stop.(solutionStopImpl).index
		m.(*solutionMoveStopsImpl).stopPositions[idx].solution = solution
	}

	target := common.Map(vehicle.SolutionStops(), func(stop SolutionStop) solutionStopImpl {
		return stop.(solutionStopImpl)
	})

	combination := make([]int, 0, nrStops)

	generate(
		vehicle,
		planUnit,
		m.(*solutionMoveStopsImpl).stopPositions, combination, target, func() {
			m.(*solutionMoveStopsImpl).allowed = false
			m.(*solutionMoveStopsImpl).valueSeen = 1
			yield(m)
		},
		shouldStop,
	)
}

func isNotAllowed(from, to solutionStopImpl) bool {
	fromModelStop := from.modelStop()
	toModelStop := to.modelStop()
	model := fromModelStop.Model()

	return model.(*modelImpl).disallowedSuccessors[fromModelStop.Index()][toModelStop.Index()]
}

func mustBeNeighbours(from, to solutionStopImpl) bool {
	if !from.modelStop().HasPlanStopsUnit() {
		return false
	}

	return from.modelStop().
		PlanStopsUnit().
		DirectedAcyclicGraph().
		HasDirectArc(from.ModelStop(), to.ModelStop())
}

//   - combination[i] will define before which stop in target the stop at
//     stopPositions[i].stop should be inserted in target.
//   - if len(combination) == len(stopPositions) then we have a complete
//     definition of where all the stop positions should be inserted.
//   - combination[i] >= combination[i-1] should be true for all i > 0, this is
//     because we want to preserve the order of the stops associated with the
//     stop positions.
func generate(
	vehicle SolutionVehicle,
	solutionPlanStopsUnit SolutionPlanStopsUnit,
	stopPositions []stopPositionImpl,
	combination []int,
	target []solutionStopImpl,
	yield func(),
	shouldStop func() bool,
) {
	if shouldStop() {
		return
	}

	if len(combination) == len(stopPositions) {
		yield()
		return
	}

	start := 0

	if len(combination) != 0 {
		start = combination[len(combination)-1] - 1
	}

	// can we be more precise on where we can end?
	end := len(target) - 1

	excludePositions := map[int]struct{}{}

	interleaveConstraint := vehicle.ModelVehicle().Model().InterleaveConstraint()

	if interleaveConstraint != nil {
		solution := vehicle.SolutionStops()[0].Solution()

		modelPlanUnit := solutionPlanStopsUnit.ModelPlanUnit()

		// what is the first planned target position for this plan unit if
		// it is a composition of plan units (plan units unit)?
		firstPlannedTargetPosition := math.MaxInt64

		if planUnitsUnit, hasPlanUnitsUnit := modelPlanUnit.PlanUnitsUnit(); hasPlanUnitsUnit {
			data := vehicle.Last().ConstraintData(interleaveConstraint).(*interleaveConstraintData)
			if solutionStopSpan, ok := data.solutionPlanStopUnits[planUnitsUnit.Index()]; ok {
				firstPlannedTargetPosition = solutionStopSpan.first
			}
		}

		sourceDisallowedInterleaves := interleaveConstraint.SourceDisallowedInterleaves(modelPlanUnit)

		for _, sourceDisallowedInterleave := range sourceDisallowedInterleaves {
			targetSolutionPlanUnit := solution.SolutionPlanUnit(sourceDisallowedInterleave.Target())
			if targetSolutionPlanUnit.IsPlanned() {
				var first SolutionStop
				var last SolutionStop
				for _, plannedPlanStopsUnit := range targetSolutionPlanUnit.PlannedPlanStopsUnits() {
					if plannedPlanStopsUnit.SolutionStops()[0].Vehicle() == vehicle {
						// the source is not allowed to be interleaved into these positions
						for _, solutionStop := range plannedPlanStopsUnit.SolutionStops() {
							if first == nil || solutionStop.Position() < first.Position() {
								first = solutionStop
							}
							if last == nil || solutionStop.Position() > last.Position() {
								last = solutionStop
							}
						}
					}
				}
				for s := first.Next(); s != last; s = s.Next() {
					excludePositions[s.Position()-1] = struct{}{}
				}
				excludePositions[last.Position()-1] = struct{}{}
			}
		}

		targetDisallowedInterleaves := interleaveConstraint.TargetDisallowedInterleaves(modelPlanUnit)

	TargetDisallowedInterleavesLoop:
		for _, targetDisallowedInterleave := range targetDisallowedInterleaves {
			for _, source := range targetDisallowedInterleave.Sources() {
				sourceSolutionPlanUnit := solution.SolutionPlanUnit(source)
				if sourceSolutionPlanUnit.IsPlanned() {
					// Each source can not be interleaved with the to be planned
					// plan unit (target).
					for _, plannedPlanStopsUnit := range sourceSolutionPlanUnit.PlannedPlanStopsUnits() {
						if plannedPlanStopsUnit.SolutionStops()[0].Vehicle() == vehicle {
							firstStopOfPlanUnit := plannedPlanStopsUnit.SolutionStops()[0]

							// If we already decided to be before or after this planned source all position must be
							// before or after. We force before by setting end, we do not have to force after as the
							// sequence of stops already does that.
							if len(combination) > 0 &&
								combination[len(combination)-1] <= firstStopOfPlanUnit.Position() {
								end = firstStopOfPlanUnit.Position() + 1
								break TargetDisallowedInterleavesLoop
							}

							// If we already planned part of the plan units unit target we must plan this target also
							// before or after the planned source.
							if firstPlannedTargetPosition != math.MaxInt64 {
								lastStopOfPlanUnit :=
									plannedPlanStopsUnit.SolutionStops()[len(plannedPlanStopsUnit.SolutionStops())-1]

								if firstPlannedTargetPosition > lastStopOfPlanUnit.Position() {
									start = lastStopOfPlanUnit.Position() + 1
								}

								if firstPlannedTargetPosition < firstStopOfPlanUnit.Position() {
									end = firstStopOfPlanUnit.Position() + 1
								}
							}
						}
					}
				}
			}
		}
	}

	positions := make([]int, 0, end-start)
	for i := start; i < end; i++ {
		if _, excludePosition := excludePositions[i]; !excludePosition {
			positions = append(positions, i)
		}
	}

	for _, i := range positions {
		// if the stops on the existing vehicle must be neighbours, we can only
		// explore combinations where the stops are neighbours. In other words
		// we can not add something between these two stops.
		if i > 0 && mustBeNeighbours(target[i], target[i+1]) {
			continue
		}

		stopPositionIdx := len(combination)

		combination = append(combination, i+1)

		stopPositions[stopPositionIdx].previousStopIndex = target[combination[stopPositionIdx]-1].index
		stopPositions[stopPositionIdx].nextStopIndex = target[combination[stopPositionIdx]].index

		if stopPositionIdx > 0 {
			previousStopPositionIdx := stopPositionIdx - 1
			if combination[stopPositionIdx] == combination[stopPositionIdx-1] {
				stopPositions[stopPositionIdx].previousStopIndex = stopPositions[previousStopPositionIdx].stopIndex
				stopPositions[previousStopPositionIdx].nextStopIndex = stopPositions[stopPositionIdx].stopIndex
			} else {
				stopPositions[previousStopPositionIdx].nextStopIndex =
					target[combination[previousStopPositionIdx]].index
				if mustBeNeighbours(
					stopPositions[previousStopPositionIdx].stop(),
					stopPositions[stopPositionIdx].stop(),
				) {
					// if the previous stop and this stop must be neighbours, we can break immediately and only explore
					// them as neighbours (no need to try to position positionIdx at any other position).
					break
				}
			}

			if isNotAllowed(
				stopPositions[previousStopPositionIdx].stop(),
				stopPositions[previousStopPositionIdx].next(),
			) {
				// undo the last combination (undo where we position stop at stopPositionIdx)
				combination = combination[:stopPositionIdx]
				// if selecting a new position for the stop at stopPositionIdx does not influence the previous stop next
				// stop it does not matter where we position the stop at stopPositionIdx, therefor we trigger a back
				// track immediately and go for the next position for the previous stop position.
				if stopPositions[previousStopPositionIdx].nextStopIndex !=
					stopPositions[stopPositionIdx].previousStopIndex {
					break
				}
				continue
			}
		}

		if isNotAllowed(stopPositions[stopPositionIdx].previous(), stopPositions[stopPositionIdx].stop()) {
			combination = combination[:stopPositionIdx]
			continue
		}

		generate(
			vehicle,
			solutionPlanStopsUnit,
			stopPositions,
			combination,
			target,
			yield,
			shouldStop,
		)

		combination = combination[:len(combination)-1]
	}
}
