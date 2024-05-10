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
type SolutionVehicle interface {
	// FirstMove creates a move that adds the given plan unit to the
	// vehicle after the first solution stop of the vehicle. The move is
	// first feasible move after the first solution stop based on the
	// estimates of the constraint, this move is not necessarily executable.
	FirstMove(SolutionPlanUnit) (SolutionMove, error)

	// BestMove returns the best move for the given solution plan unit on
	// the invoking vehicle. The best move is the move that has the lowest
	// score. If there are no moves available for the given solution plan
	// unit, a move is returned which is not executable, SolutionMoveStops.IsExecutable.
	BestMove(context.Context, SolutionPlanUnit) SolutionMove

	// Duration returns the duration of the vehicle. The duration is the
	// time the vehicle is on the road. The duration is the time between
	// the start time and the end time.
	Duration() time.Duration
	// DurationValue returns the duration value of the vehicle. The duration
	// value is the value of the duration of the vehicle. The duration value
	// is the value in model duration units.
	DurationValue() float64

	// End returns the end time of the vehicle. The end time is the time
	// the vehicle ends at the end stop.
	End() time.Time
	// EndValue returns the end value of the vehicle. The end value is the
	// value of the end of the last stop. The end value is the value in
	// model duration units since the model epoch.
	EndValue() float64

	// First returns the first stop of the vehicle. The first stop is the
	// start stop.
	First() SolutionStop

	// Index returns the index of the vehicle in the solution.
	Index() int
	// IsEmpty returns true if the vehicle is empty, false otherwise. A
	// vehicle is empty if it does not have any stops. The start and end
	// stops are not considered.
	IsEmpty() bool

	// Last returns the last stop of the vehicle. The last stop is the end
	// stop.
	Last() SolutionStop

	// ModelVehicle returns the modeled vehicle type of the vehicle.
	ModelVehicle() ModelVehicle

	// NumberOfStops returns the number of stops in the vehicle. The start
	// and end stops are not considered.
	NumberOfStops() int

	// SolutionStops returns the stops in the vehicle. The start and end
	// stops are included in the returned stops.
	SolutionStops() SolutionStops
	// Start returns the start time of the vehicle. The start time is
	// the time the vehicle starts at the start stop, it has been set
	// in the factory method of the vehicle Solution.NewVehicle.
	Start() time.Time
	// StartValue returns the start value of the vehicle. The start value
	// is the value of the start of the first stop. The start value is
	// the value in model duration units since the model epoch.
	StartValue() float64

	// Unplan removes all stops from the vehicle. The start and end stops
	// are not removed. Fixed stops are not removed.
	Unplan() (bool, error)
}

// SolutionVehicles is a slice of solution vehicles.
type SolutionVehicles []SolutionVehicle

type solutionVehicleImpl struct {
	solution *solutionImpl
	index    int
}

func toSolutionVehicle(
	solution Solution,
	index int,
) SolutionVehicle {
	return solutionVehicleImpl{
		index:    index,
		solution: solution.(*solutionImpl),
	}
}

func (v solutionVehicleImpl) firstMovePlanStopsUnit(
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

func (v solutionVehicleImpl) firstMovePlanUnitsUnit(
	planUnit *solutionPlanUnitsUnitImpl,
) (SolutionMove, error) {
	if planUnit.ModelPlanUnitsUnit().PlanOneOf() {
		return v.firstMovePlanOneOfUnit(planUnit)
	}
	return v.firstMovePlanAllUnit(planUnit)
}

func (v solutionVehicleImpl) firstMovePlanOneOfUnit(
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

func (v solutionVehicleImpl) firstMovePlanAllUnit(
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

func (v solutionVehicleImpl) bestMovePlanSingleStop(
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
		stopPositionImpl{},
	)
	stop := v.first()

	movesPtr := moveContainerPool.Get().(*[]moveContainer)
	moves := *movesPtr

	first := true
	bestMoveContainer := moveContainer{
		value: math.Inf(1),
	}
	solution := planUnit.solution()
	rand := solution.random

	for !stop.IsLast() {
		stop = stop.next()
		pos := newStopPosition(
			stop.previous(),
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

func (v solutionVehicleImpl) bestMoveSequence(
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

func (v solutionVehicleImpl) bestMovePlanMultipleStops(
	ctx context.Context,
	planUnit *solutionPlanStopsUnitImpl,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) SolutionMove {
	var bestMove SolutionMove = newNotExecutableSolutionMoveStops(planUnit)
	sequenceGeneratorSync(planUnit, func(sequence SolutionStops) {
		newMove := v.bestMoveSequence(ctx, planUnit, sequence, preAllocatedMoveContainer)
		bestMove = takeBestInPlace(bestMove, newMove)
	})
	return bestMove
}

func (v solutionVehicleImpl) bestMovePlanStopsUnit(
	ctx context.Context,
	planUnit *solutionPlanStopsUnitImpl,
	preAllocatedMoveContainer *PreAllocatedMoveContainer,
) SolutionMove {
	if planUnit.ModelPlanStopsUnit().NumberOfStops() == 1 {
		return v.bestMovePlanSingleStop(ctx, planUnit, preAllocatedMoveContainer)
	}

	return v.bestMovePlanMultipleStops(ctx, planUnit, preAllocatedMoveContainer)
}

func (v solutionVehicleImpl) bestMovePlanUnitsUnit(
	ctx context.Context,
	planUnit *solutionPlanUnitsUnitImpl,
) SolutionMove {
	if planUnit.ModelPlanUnitsUnit().PlanOneOf() {
		return v.bestMovePlanOneOfUnit(ctx, planUnit)
	}
	return v.bestMovePlanAllUnit(ctx, planUnit)
}

func (v solutionVehicleImpl) bestMovePlanOneOfUnit(
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

func (v solutionVehicleImpl) bestMovePlanAllUnit(
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

func (v solutionVehicleImpl) FirstMove(
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

func (v solutionVehicleImpl) BestMove(
	ctx context.Context,
	planUnit SolutionPlanUnit,
) SolutionMove {
	var allocations *PreAllocatedMoveContainer
	if _, ok := planUnit.(SolutionPlanStopsUnit); ok {
		allocations = NewPreAllocatedMoveContainer(planUnit)
	}
	return v.bestMove(ctx, planUnit, allocations)
}

func (v solutionVehicleImpl) bestMove(
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

func (v solutionVehicleImpl) IsEmpty() bool {
	return v.last().Position() == 1
}

func (v solutionVehicleImpl) NumberOfStops() int {
	return v.last().Position() - 1
}

func (v solutionVehicleImpl) Index() int {
	return v.index
}

func (v solutionVehicleImpl) First() SolutionStop {
	return v.first()
}

func (v solutionVehicleImpl) Last() SolutionStop {
	return v.last()
}

func (v solutionVehicleImpl) first() solutionStopImpl {
	return solutionStopImpl{
		index:    v.solution.first[v.index],
		solution: v.solution,
	}
}

func (v solutionVehicleImpl) last() solutionStopImpl {
	return solutionStopImpl{
		index:    v.solution.last[v.index],
		solution: v.solution,
	}
}

func (v solutionVehicleImpl) DurationValue() float64 {
	return v.EndValue() - v.StartValue()
}

func (v solutionVehicleImpl) Duration() time.Duration {
	return v.End().Sub(v.Start())
}

func (v solutionVehicleImpl) StartValue() float64 {
	return v.first().StartValue()
}

func (v solutionVehicleImpl) Start() time.Time {
	return v.first().Start()
}

func (v solutionVehicleImpl) EndValue() float64 {
	return v.last().EndValue()
}

func (v solutionVehicleImpl) End() time.Time {
	return v.last().End()
}

func (v solutionVehicleImpl) Next() SolutionStop {
	return solutionStopImpl{
		index:    v.solution.model.NumberOfStops() + v.index*2 + 1,
		solution: v.solution,
	}
}

func (v solutionVehicleImpl) SolutionStops() SolutionStops {
	solutionStops := make(SolutionStops, 0, v.NumberOfStops()+2)
	solutionStop := v.First()
	for !solutionStop.IsLast() {
		solutionStops = append(solutionStops, solutionStop)
		solutionStop = solutionStop.Next()
	}
	solutionStops = append(solutionStops, solutionStop)
	return solutionStops
}

func (v solutionVehicleImpl) solutionStops() []solutionStopImpl {
	solutionStops := make([]solutionStopImpl, 0, v.NumberOfStops()+2)
	solutionStop := v.first()
	for !solutionStop.IsLast() {
		solutionStops = append(solutionStops, solutionStop)
		solutionStop = solutionStop.next()
	}
	solutionStops = append(solutionStops, solutionStop)
	return solutionStops
}

func (v solutionVehicleImpl) ModelVehicle() ModelVehicle {
	return v.solution.model.Vehicle(v.solution.vehicleIndices[v.index])
}

func (v solutionVehicleImpl) Unplan() (bool, error) {
	// TODO notify observers
	solutionStops := common.Filter(v.solutionStops(), func(solutionStop solutionStopImpl) bool {
		return !solutionStop.IsFixed()
	})
	if len(solutionStops) == 0 {
		return false, nil
	}

	solution := solutionStops[0].solution

	planUnits := common.Map(solutionStops, func(solutionStop solutionStopImpl) *solutionPlanStopsUnitImpl {
		return solutionStop.planStopsUnit()
	})
	for _, planUnit := range planUnits {
		solution.unPlannedPlanUnits.add(planUnit)
		solution.plannedPlanUnits.remove(planUnit)
	}
	stopPositions := common.Map(solutionStops, func(solutionStop solutionStopImpl) StopPosition {
		return newStopPosition(
			solutionStop.previous(),
			solutionStop,
			solutionStop.next(),
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
			stopPosition := stopPositions[i].(stopPositionImpl)
			beforeStop := stopPosition.next()
			stopPosition.stop().attach(
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
