// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"
	"math"
	"slices"
	"sync"
	"time"

	"github.com/nextmv-io/nextroute/common"
)

// SolutionVehicle is a vehicle in a solution.
type SolutionVehicle struct {
	solution *solutionImpl
	index    int
}

// SolutionVehicles is a slice of solution vehicles.
type SolutionVehicles []SolutionVehicle

func toSolutionVehicle(
	solution Solution,
	index int,
) SolutionVehicle {
	return SolutionVehicle{
		index:    index,
		solution: solution.(*solutionImpl),
	}
}

func (v SolutionVehicle) firstMovePlanStopsUnit(
	planUnit *solutionPlanStopsUnitImpl,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) (SolutionMove, error) {
	stop := false
	var bestMove SolutionMove = newNotExecutableSolutionMoveStops(planUnit)
	SolutionMoveStopsGenerator(
		v,
		planUnit,
		func(nextMove SolutionMoveStops) {
			value, allowed, hint := v.solution.checkConstraintsAndEstimateDeltaScore(nextMove)
			if hint.SkipVehicle() {
				stop = true
				return
			}
			nextMove.(*solutionMoveStopsImpl).value = value
			nextMove.(*solutionMoveStopsImpl).allowed = allowed
			nextMove.(*solutionMoveStopsImpl).valueSeen = nextMove.ValueSeen()
			if nextMove.IsExecutable() {
				bestMove = takeBestInPlace(bestMove, nextMove)
				stop = true
			}
		},
		planUnit.SolutionStops(),
		preAllocatedMoveContainer,
		func() bool {
			return stop
		},
	)
	return bestMove, nil
}

func (v SolutionVehicle) firstMovePlanUnitsUnit(
	planUnit *solutionPlanUnitsUnitImpl,
) (SolutionMove, error) {
	if planUnit.ModelPlanUnitsUnit().PlanOneOf() {
		return v.firstMovePlanOneOfUnit(planUnit)
	}
	return v.firstMovePlanAllUnit(planUnit)
}

func (v SolutionVehicle) firstMovePlanOneOfUnit(
	planUnit *solutionPlanUnitsUnitImpl,
) (SolutionMove, error) {
	planUnits := common.Shuffle(
		v.solution.Random(),
		planUnit.SolutionPlanUnits(),
	)
	for _, planUnit := range planUnits {
		move, err := v.FirstMove(planUnit)
		if err != nil {
			return nil, err
		}
		if move.IsExecutable() {
			return move, nil
		}
	}
	return NotExecutableMove, nil
}

func (v SolutionVehicle) firstMovePlanAllUnit(
	planUnit *solutionPlanUnitsUnitImpl,
) (SolutionMove, error) {
	planUnits := common.Shuffle(
		v.solution.model.Random(),
		planUnit.SolutionPlanUnits(),
	)
	moves := make(SolutionMoves, 0, len(planUnits))
	var err error
	for idx, propositionPlanUnit := range planUnits {
		var move SolutionMove

		if idx == 0 || planUnit.modelPlanUnitsUnit.SameVehicle() {
			move, err = v.FirstMove(propositionPlanUnit)
			if err != nil {
				return nil, err
			}
		} else {
			move = v.solution.BestMove(context.Background(), propositionPlanUnit)
		}

		if move.IsExecutable() {
			if planned, err := move.Execute(context.Background()); err != nil || !planned {
				if unplanned, err := revertMoves(moves); !unplanned || err != nil {
					return nil, fmt.Errorf("unplanning moves failed %w", err)
				}
				return NotExecutableMove, nil
			}
		} else {
			if unplanned, err := revertMoves(moves); !unplanned || err != nil {
				return nil, fmt.Errorf("unplanning moves failed %w", err)
			}
			return NotExecutableMove, nil
		}
		moves = append(moves, move)
	}
	if unplanned, err := revertMoves(moves); !unplanned || err != nil {
		return nil, fmt.Errorf("unplanning moves failed %w", err)
	}
	// This can happen in case of ctx.Done()
	if len(moves) != len(planUnits) {
		return NotExecutableMove, nil
	}

	return newSolutionMoveUnits(planUnit, moves), nil
}

// moveContainer holds essential information to construct a move from in bestMovePlanSingleStop.
type moveContainer struct {
	nextIndex     int
	previousIndex int
	value         float64 // not allowed can be encoded as NaN
}

var moveContainerPool = sync.Pool{
	New: func() any {
		x := make([]moveContainer, 0, 128)
		return &x
	},
}

func returnToMoveContainerPool(movesPtr *[]moveContainer) {
	// this function is mainly used before each return in bestMovePlanSingleStop
	// since defer costs a bit of performance
	*movesPtr = (*movesPtr)[:0]
	moveContainerPool.Put(movesPtr)
}

func updateMoveInPlace(move SolutionMoveStops, moveContainer moveContainer) {
	move.(*solutionMoveStopsImpl).stopPositions[0].previousStopIndex = moveContainer.previousIndex
	move.(*solutionMoveStopsImpl).stopPositions[0].nextStopIndex = moveContainer.nextIndex
	move.(*solutionMoveStopsImpl).value = moveContainer.value
}

func (v SolutionVehicle) bestMovePlanSingleStop(
	_ context.Context,
	planUnit *solutionPlanStopsUnitImpl,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) SolutionMoveStops {
	candidateStop := planUnit.solutionStops[0]
	move := preAllocatedMoveContainer.singleStopPosSolutionMoveStop
	move.(*solutionMoveStopsImpl).reset()
	// ensure that stopPositions is a length 1 slice
	move.(*solutionMoveStopsImpl).stopPositions = append(
		move.(*solutionMoveStopsImpl).stopPositions,
		StopPosition{},
	)
	stop := v.First()

	movesPtr := moveContainerPool.Get().(*[]moveContainer)
	moves := *movesPtr

	first := true
	bestMoveContainer := moveContainer{
		value: math.Inf(1),
	}
	solution := planUnit.solution()
	rand := solution.random

	for !stop.IsLast() {
		stop = stop.Next()
		pos := newStopPosition(
			stop.Previous(),
			candidateStop,
			stop,
		)
		mc := moveContainer{
			previousIndex: pos.previousStopIndex,
			nextIndex:     pos.nextStopIndex,
		}
		move.(*solutionMoveStopsImpl).stopPositions[0] = pos
		if first {
			first = false
			allowed, hint := v.solution.checkConstraints(
				move,
			)
			if hint.SkipVehicle() {
				returnToMoveContainerPool(movesPtr)
				return move
			}
			if !allowed {
				continue
			}
		}
		value := v.solution.estimateDeltaScore(
			move,
		)
		mc.value = value
		if mc.value < bestMoveContainer.value {
			bestMoveContainer = mc
		} else if mc.value == bestMoveContainer.value {
			if rand.Float64() < 0.5 {
				bestMoveContainer = mc
			}
		}
		moves = append(moves, mc)
	}
	if len(moves) == 0 {
		returnToMoveContainerPool(movesPtr)
		move.(*solutionMoveStopsImpl).allowed = false
		return move
	}
	// we got the best move here in O(n)
	// we will check that here and maybe we are lucky, then we can skip
	// the sort step
	updateMoveInPlace(move, bestMoveContainer)
	allowed, _ := v.solution.checkConstraints(
		move,
	)
	if allowed {
		move.(*solutionMoveStopsImpl).allowed = true
		returnToMoveContainerPool(movesPtr)
		return move
	}
	// turned out that sorting is actually faster than using a minHeap (a specific implementation)
	// still worth to explore
	slices.SortFunc(moves, func(i, j moveContainer) int {
		if i.value < j.value {
			return -1
		}
		if i.value > j.value {
			return 1
		}
		if rand.Float64() < 0.5 {
			return -1
		}
		return 1
	})
	for _, mc := range moves {
		updateMoveInPlace(move, mc)
		allowed, _ := v.solution.checkConstraints(
			move,
		)
		if allowed {
			move.(*solutionMoveStopsImpl).allowed = true
			returnToMoveContainerPool(movesPtr)
			return move
		}
	}
	returnToMoveContainerPool(movesPtr)
	move.(*solutionMoveStopsImpl).allowed = false
	return move
}

func (v SolutionVehicle) bestMoveSequence(
	_ context.Context,
	planUnit *solutionPlanStopsUnitImpl,
	sequence SolutionStops,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) SolutionMove {
	var bestMove SolutionMove = newNotExecutableSolutionMoveStops(planUnit)
	stop := false
	SolutionMoveStopsGenerator(
		v,
		planUnit,
		func(nextMove SolutionMoveStops) {
			value, allowed, hint := v.solution.checkConstraintsAndEstimateDeltaScore(nextMove)
			if hint.SkipVehicle() {
				stop = true
				return
			}
			nextMove.(*solutionMoveStopsImpl).value = value
			nextMove.(*solutionMoveStopsImpl).allowed = allowed
			nextMove.(*solutionMoveStopsImpl).valueSeen = nextMove.ValueSeen()
			if allowed {
				bestMove = takeBestInPlace(bestMove, nextMove)
			}
		},
		sequence,
		preAllocatedMoveContainer,
		func() bool {
			return stop
		},
	)

	return bestMove
}

func (v SolutionVehicle) bestMovePlanMultipleStops(
	ctx context.Context,
	planUnit *solutionPlanStopsUnitImpl,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) SolutionMove {
	var bestMove SolutionMove = newNotExecutableSolutionMoveStops(planUnit)
	quitSequenceGenerator := make(chan struct{})
	defer close(quitSequenceGenerator)
	for sequence := range SequenceGeneratorChannel(planUnit, quitSequenceGenerator) {
		newMove := v.bestMoveSequence(ctx, planUnit, sequence, preAllocatedMoveContainer)
		bestMove = takeBestInPlace(bestMove, newMove)
	}
	return bestMove
}

func (v SolutionVehicle) bestMovePlanStopsUnit(
	ctx context.Context,
	planUnit *solutionPlanStopsUnitImpl,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) SolutionMove {
	if planUnit.ModelPlanStopsUnit().NumberOfStops() == 1 {
		return v.bestMovePlanSingleStop(ctx, planUnit, preAllocatedMoveContainer)
	}

	return v.bestMovePlanMultipleStops(ctx, planUnit, preAllocatedMoveContainer)
}

func (v SolutionVehicle) bestMovePlanUnitsUnit(
	ctx context.Context,
	planUnit *solutionPlanUnitsUnitImpl,
) SolutionMove {
	if planUnit.ModelPlanUnitsUnit().PlanOneOf() {
		return v.bestMovePlanOneOfUnit(ctx, planUnit)
	}
	return v.bestMovePlanAllUnit(ctx, planUnit)
}

func (v SolutionVehicle) bestMovePlanOneOfUnit(
	ctx context.Context,
	planUnit *solutionPlanUnitsUnitImpl,
) SolutionMove {
	move := NotExecutableMove

	for _, planUnit := range planUnit.solutionPlanUnits {
		move = move.TakeBest(
			v.BestMove(ctx, planUnit),
		)
	}
	return move
}

func revertMoves(moves SolutionMoves) (bool, error) {
	for i := len(moves) - 1; i >= 0; i-- {
		if unplanned, err := moves[i].PlanUnit().UnPlan(); err != nil || !unplanned {
			return false, err
		}
	}
	return true, nil
}

func (v SolutionVehicle) bestMovePlanAllUnit(
	ctx context.Context,
	planUnit *solutionPlanUnitsUnitImpl,
) SolutionMove {
	planUnits := common.Shuffle(
		v.solution.Random(),
		planUnit.SolutionPlanUnits(),
	)

	moves := make(SolutionMoves, 0, len(planUnits))
	for idx, propositionPlanUnit := range planUnits {
		var move SolutionMove

		if idx == 0 || planUnit.modelPlanUnitsUnit.SameVehicle() {
			move = v.BestMove(ctx, propositionPlanUnit)
		} else {
			move = v.solution.BestMove(ctx, propositionPlanUnit)
		}

		if move.IsExecutable() {
			if planned, err := move.Execute(ctx); err != nil || !planned {
				if unplanned, err := revertMoves(moves); !unplanned || err != nil {
					panic(fmt.Errorf("unplanning moves failed %w", err))
				}
				return NotExecutableMove
			}
		} else {
			if unplanned, err := revertMoves(moves); !unplanned || err != nil {
				panic(fmt.Errorf("unplanning moves failed %w", err))
			}
			return NotExecutableMove
		}
		moves = append(moves, move)
	}
	if unplanned, err := revertMoves(moves); !unplanned || err != nil {
		panic(fmt.Errorf("unplanning moves failed %w", err))
	}
	// This can happen in case of ctx.Done()
	if len(moves) != len(planUnits) {
		return NotExecutableMove
	}

	return newSolutionMoveUnits(planUnit, moves)
}

// FirstMove creates a move that adds the given plan unit to the
// vehicle after the first solution stop of the vehicle. The move is
// first feasible move after the first solution stop based on the
// estimates of the constraint, this move is not necessarily executable.
func (v SolutionVehicle) FirstMove(
	planUnit SolutionPlanUnit,
) (SolutionMove, error) {
	switch planUnit.(type) {
	case SolutionPlanStopsUnit:
		allocations := NewPreAllocatedMoveContainer(planUnit)
		return v.firstMovePlanStopsUnit(planUnit.(*solutionPlanStopsUnitImpl), allocations)
	case SolutionPlanUnitsUnit:
		return v.firstMovePlanUnitsUnit(planUnit.(*solutionPlanUnitsUnitImpl))
	}
	return NotExecutableMove, nil
}

// BestMove returns the best move for the given solution plan unit on
// the invoking vehicle. The best move is the move that has the lowest
// score. If there are no moves available for the given solution plan
// unit, a move is returned which is not executable, SolutionMoveStops.IsExecutable.
func (v SolutionVehicle) BestMove(
	ctx context.Context,
	planUnit SolutionPlanUnit,
) SolutionMove {
	var allocations *PreAllocatedMoveContainer
	if _, ok := planUnit.(SolutionPlanStopsUnit); ok {
		allocations = NewPreAllocatedMoveContainer(planUnit)
	}
	return v.bestMove(ctx, planUnit, allocations)
}

func (v SolutionVehicle) bestMove(
	ctx context.Context,
	planUnit SolutionPlanUnit,
	sharedMoveContainer *PreAllocatedMoveContainer,
) SolutionMove {
	select {
	case <-ctx.Done():
		return NotExecutableMove
	default:
		if planUnit.IsPlanned() {
			return NotExecutableMove
		}
		switch planUnit.(type) {
		case SolutionPlanStopsUnit:
			return v.bestMovePlanStopsUnit(ctx, planUnit.(*solutionPlanStopsUnitImpl), sharedMoveContainer)
		case SolutionPlanUnitsUnit:
			return v.bestMovePlanUnitsUnit(ctx, planUnit.(*solutionPlanUnitsUnitImpl))
		}
		return NotExecutableMove
	}
}

// IsEmpty returns true if the vehicle is empty, false otherwise. A
// vehicle is empty if it does not have any stops. The start and end
// stops are not considered.
func (v SolutionVehicle) IsEmpty() bool {
	return v.Last().Position() == 1
}

// NumberOfStops returns the number of stops in the vehicle. The start
// and end stops are not considered.
func (v SolutionVehicle) NumberOfStops() int {
	return v.Last().Position() - 1
}

// Index returns the index of the vehicle in the solution.
func (v SolutionVehicle) Index() int {
	return v.index
}

// First returns the first stop of the vehicle. The first stop is the
// start stop.
func (v SolutionVehicle) First() SolutionStop {
	return SolutionStop{
		index:    v.solution.first[v.index],
		solution: v.solution,
	}
}

// Last returns the last stop of the vehicle. The last stop is the end
// stop.
func (v SolutionVehicle) Last() SolutionStop {
	return SolutionStop{
		index:    v.solution.last[v.index],
		solution: v.solution,
	}
}

// DurationValue returns the duration value of the vehicle. The duration
// value is the value of the duration of the vehicle. The duration value
// is the value in model duration units.
func (v SolutionVehicle) DurationValue() float64 {
	return v.EndValue() - v.StartValue()
}

// Duration returns the duration of the vehicle. The duration is the
// time the vehicle is on the road. The duration is the time between
// the start time and the end time.
func (v SolutionVehicle) Duration() time.Duration {
	return v.End().Sub(v.Start())
}

// StartValue returns the start value of the vehicle. The start value
// is the value of the start of the first stop. The start value is
// the value in model duration units since the model epoch.
func (v SolutionVehicle) StartValue() float64 {
	return v.First().StartValue()
}

// Start returns the start time of the vehicle. The start time is
// the time the vehicle starts at the start stop, it has been set
// in the factory method of the vehicle Solution.NewVehicle.
func (v SolutionVehicle) Start() time.Time {
	return v.First().Start()
}

// EndValue returns the end value of the vehicle. The end value is the
// value of the end of the last stop. The end value is the value in
// model duration units since the model epoch.
func (v SolutionVehicle) EndValue() float64 {
	return v.Last().EndValue()
}

// End returns the end time of the vehicle. The end time is the time
// the vehicle ends at the end stop.
func (v SolutionVehicle) End() time.Time {
	return v.Last().End()
}

// SolutionStops returns the stops in the vehicle. The start and end
// stops are included in the returned stops.
func (v SolutionVehicle) SolutionStops() SolutionStops {
	solutionStops := make(SolutionStops, 0, v.NumberOfStops()+2)
	solutionStop := v.First()
	for !solutionStop.IsLast() {
		solutionStops = append(solutionStops, solutionStop)
		solutionStop = solutionStop.Next()
	}
	solutionStops = append(solutionStops, solutionStop)
	return solutionStops
}

// ModelVehicle returns the modeled vehicle type of the vehicle.
func (v SolutionVehicle) ModelVehicle() ModelVehicle {
	return v.solution.model.Vehicle(v.solution.vehicleIndices[v.index])
}

// Unplan removes all stops from the vehicle. The start and end stops
// are not removed. Fixed stops are not removed.
func (v SolutionVehicle) Unplan() (bool, error) {
	// TODO notify observers
	solutionStops := common.Filter(v.SolutionStops(), func(solutionStop SolutionStop) bool {
		return !solutionStop.IsFixed()
	})
	if len(solutionStops) == 0 {
		return false, nil
	}

	solution := solutionStops[0].solution

	planUnits := common.Map(solutionStops, func(solutionStop SolutionStop) *solutionPlanStopsUnitImpl {
		return solutionStop.planStopsUnit()
	})
	for _, planUnit := range planUnits {
		solution.unPlannedPlanUnits.add(planUnit)
		solution.plannedPlanUnits.remove(planUnit)
	}
	stopPositions := common.Map(solutionStops, func(solutionStop SolutionStop) StopPosition {
		return newStopPosition(
			solutionStop.Previous(),
			solutionStop,
			solutionStop.Next(),
		)
	})

	index := solutionStops[0].PreviousIndex()

	for _, solutionStop := range solutionStops {
		solutionStop.detach()
	}
	constraint, _, err := solution.isFeasible(index, true)
	if err != nil {
		return false, err
	}
	if constraint != nil {
		for i := len(stopPositions) - 1; i >= 0; i-- {
			stopPosition := stopPositions[i]
			beforeStop := stopPosition.Next()
			stopPosition.Stop().attach(
				beforeStop.PreviousIndex(),
			)
		}
		for _, planUnit := range planUnits {
			solution.unPlannedPlanUnits.remove(planUnit)
			solution.plannedPlanUnits.add(planUnit)
		}
		constraint, _, err := solution.isFeasible(index, true)
		if err != nil {
			return false, err
		}
		if constraint != nil {
			return false, fmt.Errorf(
				"undoing failed unplan vehicle failed: %v", constraint,
			)
		}
	}

	return true, nil
}

// IsZero returns true if the solution vehicle is the zero value.
// In this case it is not safe to use the solution vehicle.
func (v SolutionVehicle) IsZero() bool {
	return v.solution == nil && v.index == 0
}
