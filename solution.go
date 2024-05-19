// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/nextmv-io/nextroute/common"
)

// Solution is a solution to a model.
type Solution interface {
	// BestMove returns the best move for the given solution plan unit. The
	// best move is the move that has the lowest score. If there are no moves
	// available for the given solution plan unit, a move is returned which
	// is not executable, SolutionMoveStops.IsExecutable.
	BestMove(context.Context, SolutionPlanUnit) SolutionMove

	// ConstraintData returns the data of the constraint for the solution. The
	// constraint data of a solution is set by the
	// ConstraintSolutionDataUpdater.UpdateConstraintSolutionData method of the
	// constraint.
	ConstraintData(constraint ModelConstraint) any
	// Copy returns a deep copy of the solution.
	Copy() Solution

	// FixedPlanUnits returns the solution plan units that are fixed.
	// Fixed plan units are plan units that are not allowed to be planned or
	// unplanned. The union of fixed, planned and unplanned plan units
	// is the set of all plan units in the model.
	FixedPlanUnits() ImmutableSolutionPlanUnitCollection

	// Model returns the model of the solution.
	Model() Model

	// ObjectiveData returns the value of the objective for the solution. The
	// objective value of a solution is set by the
	// ObjectiveSolutionDataUpdater.UpdateObjectiveSolutionData method of the
	// objective. If the objective is not set on the solution, nil is returned.
	ObjectiveData(objective ModelObjective) any
	// ObjectiveValue returns the objective value for the objective in the
	// solution. Also returns 0.0 if the objective is not part of the solution.
	ObjectiveValue(objective ModelObjective) float64

	// PlannedPlanUnits returns the solution plan units that are planned as
	// a collection of solution plan units.
	PlannedPlanUnits() ImmutableSolutionPlanUnitCollection

	// Random returns the random number generator of the solution.
	Random() *rand.Rand

	// Score returns the score of the solution.
	Score() float64

	// SetRandom sets the random number generator of the solution. Returns an
	// error if the random number generator is nil.
	SetRandom(random *rand.Rand) error

	// SolutionPlanStopsUnit returns the [SolutionPlanStopsUnit] for the given
	// model plan unit.
	SolutionPlanStopsUnit(planUnit ModelPlanStopsUnit) SolutionPlanStopsUnit
	// SolutionPlanUnit returns the [SolutionPlanUnit] for the given
	// model plan unit.
	SolutionPlanUnit(planUnit ModelPlanUnit) SolutionPlanUnit
	// SolutionStop returns the solution stop for the given model stop.
	SolutionStop(stop ModelStop) SolutionStop
	// SolutionVehicle returns the solution vehicle for the given model vehicle.
	SolutionVehicle(vehicle ModelVehicle) SolutionVehicle

	// UnPlannedPlanUnits returns the solution plan units that are not
	// planned.
	UnPlannedPlanUnits() ImmutableSolutionPlanUnitCollection

	// Vehicles returns the vehicles of the solution.
	Vehicles() SolutionVehicles
}

// Solutions is a slice of solutions.
type Solutions []Solution

// NewSolution returns a new Solution.
func NewSolution(
	m Model,
) (Solution, error) {
	model := m.(*modelImpl)

	err := model.lock()

	if err != nil {
		return nil, err
	}

	model.OnNewSolution(m)

	nrStops := 0
	nrFixedPlanUnits := 0
	nrPropositionPlanUnits := 0
	nrVehicles := len(model.vehicles)

	for _, planUnit := range model.PlanUnits() {
		if planStopsUnit, ok := planUnit.(ModelPlanStopsUnit); ok {
			nrStops += len(planStopsUnit.Stops())
		}
		if planUnit.IsFixed() {
			nrFixedPlanUnits++
		}
		if _, isElementOfPlanUnitsUnit := planUnit.PlanUnitsUnit(); isElementOfPlanUnitsUnit {
			nrPropositionPlanUnits++
		}
	}

	random := rand.New(rand.NewSource(m.Random().Int63()))

	nExpressions := len(model.expressions)

	solution := &solutionImpl{
		model: m,

		vehicleIndices:           make([]int, 0, nrVehicles),
		vehicles:                 make([]SolutionVehicle, 0, nrVehicles),
		solutionVehicles:         make([]SolutionVehicle, 0, nrVehicles),
		first:                    make([]int, 0, nrVehicles),
		last:                     make([]int, 0, nrVehicles),
		stop:                     make([]int, 0, nrStops),
		inVehicle:                make([]int, 0, nrStops),
		previous:                 make([]int, 0, nrStops),
		next:                     make([]int, 0, nrStops),
		stopPosition:             make([]int, 0, nrStops),
		cumulativeTravelDuration: make([]float64, 0, nrStops),
		arrival:                  make([]float64, 0, nrStops),
		slack:                    make([]float64, 0, nrStops),
		start:                    make([]float64, 0, nrStops),
		end:                      make([]float64, 0, nrStops),
		values:                   make(map[int][]float64, nExpressions),
		cumulativeValues:         make(map[int][]float64, nExpressions),
		stopToPlanUnit:           make([]*solutionPlanStopsUnitImpl, nrStops),
		constraintStopData:       make(map[ModelConstraint][]Copier),
		objectiveStopData:        make(map[ModelObjective][]Copier),
		constraintSolutionData:   make(map[ModelConstraint]Copier),
		objectiveSolutionData:    make(map[ModelObjective]Copier),
		random:                   random,
		fixedPlanUnits: newSolutionPlanUnitCollectionBaseImpl(
			random,
			len(model.planUnits),
		),
		plannedPlanUnits: newSolutionPlanUnitCollectionBaseImpl(
			random,
			len(model.planUnits),
		),
		unPlannedPlanUnits: newSolutionPlanUnitCollectionBaseImpl(
			random,
			len(model.planUnits),
		),
		propositionPlanUnits: newSolutionPlanUnitCollectionBaseImpl(
			random,
			len(model.planUnits),
		),
	}

	for _, expression := range model.expressions {
		solution.values[expression.Index()] = make([]float64, nrStops)
		solution.cumulativeValues[expression.Index()] = make([]float64, nrStops)
	}

	for _, constraint := range model.constraintsWithStopUpdater {
		solution.constraintStopData[constraint] = make([]Copier, nrStops)
	}

	for _, objective := range model.objectivesWithStopUpdater {
		solution.objectiveStopData[objective] = make([]Copier, nrStops)
	}

	stopsUsed := make(map[int]bool)

	stopIdx := 0
	idxToPlanUnit := make(map[int]SolutionPlanUnit)

	for _, planUnit := range model.PlanStopsUnits() {
		solutionPlanUnit := &solutionPlanStopsUnitImpl{
			modelPlanStopsUnit: planUnit,
			solutionStops: make(
				[]SolutionStop,
				len(planUnit.Stops()),
			),
		}
		idxToPlanUnit[planUnit.Index()] = solutionPlanUnit
		for idx, stop := range planUnit.Stops() {
			if _, ok := stopsUsed[stop.Index()]; ok {
				return nil, fmt.Errorf(
					"stop %v is used more than once in one or"+
						"more plan planUnit %v",
					stop,
					planUnit,
				)
			}
			solutionStop := toSolutionStop(
				solution,
				stopIdx,
			)

			solutionPlanUnit.solutionStops[idx] = solutionStop

			solution.stopToPlanUnit[solutionStop.Index()] = solutionPlanUnit

			stopIdx++

			solution.stop = append(
				solution.stop,
				stop.Index(),
			)
			solution.inVehicle = append(
				solution.inVehicle,
				-1,
			)
			solution.previous = append(
				solution.previous,
				solutionPlanUnit.solutionStops[idx].Index(),
			)
			solution.next = append(
				solution.next,
				solutionPlanUnit.solutionStops[idx].Index(),
			)
			solution.stopPosition = append(
				solution.stopPosition,
				-1,
			)
			solution.cumulativeTravelDuration = append(
				solution.cumulativeTravelDuration,
				0,
			)
			solution.arrival = append(
				solution.arrival,
				0,
			)
			solution.slack = append(
				solution.slack,
				math.MaxFloat64,
			)
			solution.start = append(
				solution.start,
				0,
			)
			solution.end = append(
				solution.end,
				0,
			)
		}

		if _, isElementOfPlanUnitsUnit := planUnit.PlanUnitsUnit(); isElementOfPlanUnitsUnit {
			solution.propositionPlanUnits.add(solutionPlanUnit)
		} else {
			solution.unPlannedPlanUnits.add(solutionPlanUnit)
		}
	}

	for _, planUnit := range model.PlanUnits() {
		if modelPlanUnitsUnit, ok := planUnit.(ModelPlanUnitsUnit); ok {
			solutionPlanUnitsUnit := &solutionPlanUnitsUnitImpl{
				modelPlanUnitsUnit: modelPlanUnitsUnit,
				solutionPlanUnits: make(
					SolutionPlanUnits,
					len(modelPlanUnitsUnit.PlanUnits()),
				),
			}
			for idx, modelPlanUnit := range modelPlanUnitsUnit.PlanUnits() {
				if _, ok := idxToPlanUnit[modelPlanUnit.Index()]; !ok {
					return nil, fmt.Errorf(
						"can not find the solution plan unit for mode plan unit %v,"+
							", should never happen, contact support",
						modelPlanUnit,
					)
				}
				solutionPlanUnitsUnit.solutionPlanUnits[idx] = idxToPlanUnit[modelPlanUnit.Index()]
			}

			idxToPlanUnit[modelPlanUnitsUnit.Index()] = solutionPlanUnitsUnit

			if _, isElementOfPlanUnitsUnit := modelPlanUnitsUnit.PlanUnitsUnit(); isElementOfPlanUnitsUnit {
				solution.propositionPlanUnits.add(solutionPlanUnitsUnit)
			} else {
				solution.unPlannedPlanUnits.add(solutionPlanUnitsUnit)
			}
		}
	}

	for _, vehicle := range model.vehicles {
		v, err := solution.newVehicle(vehicle)
		if err != nil {
			return nil, err
		}
		if v.Index() != vehicle.Index() {
			return nil, fmt.Errorf(
				"vehicle index %v does not match expected %v",
				v.Index(),
				vehicle.Index(),
			)
		}
	}

	if err := solution.addInitialSolution(m); err != nil {
		return nil, err
	}

	m.OnNewSolutionCreated(solution)

	return solution, nil
}

func (s *solutionImpl) unwrapRootPlanUnit(planUnit SolutionPlanUnit) SolutionPlanUnit {
	planUnitsUnit, isElementOfPlanUnitsUnit := planUnit.ModelPlanUnit().PlanUnitsUnit()
	for isElementOfPlanUnitsUnit {
		planUnit = s.solutionPlanUnitsUnit(planUnitsUnit)
		planUnitsUnit, isElementOfPlanUnitsUnit = planUnit.ModelPlanUnit().PlanUnitsUnit()
	}
	return planUnit
}

func reportInfeasibleInitialSolution(
	move SolutionMoveStops,
	constraint ModelConstraint,
) string {
	stopIDs := common.MapSlice(
		move.StopPositions(),
		func(stopPosition StopPosition) []string {
			return []string{stopPosition.Stop().ModelStop().ID()}
		})

	name := reflect.TypeOf(constraint).Name()
	stringer, ok := constraint.(fmt.Stringer)
	if ok {
		name = stringer.String()
	}
	identifier, ok := constraint.(Identifier)
	if ok {
		name = identifier.ID()
	}

	return fmt.Sprintf("infeasible initial solution: vehicle `%v` violates constraint `%v` for stops [%v]",
		move.Vehicle().ModelVehicle().ID(),
		name,
		strings.Join(stopIDs, ", "),
	)
}

func (s *solutionImpl) addInitialSolution(m Model) error {
	model := m.(*modelImpl)

	solutionObserver := newInitialSolutionObserver()

	model.AddSolutionObserver(solutionObserver)

	defer model.RemoveSolutionObserver(solutionObserver)

	for _, modelVehicle := range model.vehicles {
		solutionVehicle, ok := s.solutionVehicle(modelVehicle)
		if !ok {
			return fmt.Errorf(
				"vehicle %v not found in solution",
				modelVehicle.ID(),
			)
		}

		initialModelStops := modelVehicle.Stops()

		if len(initialModelStops) == 0 {
			continue
		}

		planUnits := common.UniqueDefined(
			common.Map(
				initialModelStops,
				func(modelStop ModelStop) SolutionPlanStopsUnit {
					return s.SolutionStop(modelStop).PlanStopsUnit()
				}),
			func(planUnit SolutionPlanStopsUnit) int {
				return planUnit.ModelPlanStopsUnit().Index()
			},
		)

		infeasiblePlanUnits := map[SolutionPlanUnit]bool{}
		allPlanUnits := map[SolutionPlanUnit]bool{}

	PlanUnitLoop:
		for _, planUnit := range planUnits {
			stopPositions := make(StopPositions, 0, len(planUnit.SolutionStops()))
			previousStop := solutionVehicle.First()

			solutionPlanUnit := s.unwrapRootPlanUnit(planUnit)
			allPlanUnits[solutionPlanUnit] = true

		ModelStopLoop:
			for modelStopIdx, modelStop := range initialModelStops {
				if len(stopPositions) == len(planUnit.SolutionStops()) {
					break
				}
				planUnitsUnit, hasPlanUnitsUnit := planUnit.ModelPlanUnit().PlanUnitsUnit()
				if hasPlanUnitsUnit && planUnitsUnit.PlanOneOf() {
					solutionPlanUnitsUnit := s.solutionPlanUnitsUnit(planUnitsUnit)
					if solutionPlanUnitsUnit.IsPlanned() {
						return fmt.Errorf(
							"infeasible initial solution: stop %v on vehicle %v is part of one-of plan unit [%v]"+
								" which is already planned",
							modelStop.ID(),
							modelVehicle.ID(),
							strings.Join(
								common.MapSlice(
									planUnitsUnit.PlanUnits(),
									func(stop ModelPlanUnit) []string {
										return []string{fmt.Sprintf("%v", stop)}
									}),
								", ",
							),
						)
					}
				}
				solutionStop := s.SolutionStop(modelStop)
				if solutionStop.IsPlanned() {
					previousStop = solutionStop
				}
				if modelStop.PlanStopsUnit().Index() == planUnit.ModelPlanStopsUnit().Index() {
					for nextIdx := modelStopIdx + 1; nextIdx < len(initialModelStops); nextIdx++ {
						nextModelStop := initialModelStops[nextIdx]
						nextSolutionStop := s.SolutionStop(nextModelStop)
						if nextSolutionStop.IsPlanned() ||
							nextModelStop.PlanStopsUnit().Index() == planUnit.ModelPlanStopsUnit().Index() {
							stopPositions = append(
								stopPositions,
								newStopPosition(
									previousStop,
									solutionStop,
									nextSolutionStop,
								),
							)
							previousStop = solutionStop
							continue ModelStopLoop
						}
						if nextSolutionStop.IsPlanned() {
							previousStop = solutionStop
						}
					}
					stopPositions = append(
						stopPositions,
						newStopPosition(
							previousStop,
							solutionStop,
							solutionVehicle.Last(),
						),
					)
				}
			}
			move, err := newMoveStops(planUnit, stopPositions, false)
			if err != nil {
				return err
			}

			for _, constraint := range model.constraints {
				if filterConstraint(constraint, false) {
					continue
				}

				isViolated, hint := constraint.EstimateIsViolated(move)

				if hint == nil {
					return newErrorOnNilHint(constraint)
				}

				s.Model().OnEstimatedIsViolated(move, constraint, isViolated, hint)

				if isViolated {
					if solutionPlanUnit.IsFixed() {
						return fmt.Errorf(
							reportInfeasibleInitialSolution(
								move,
								constraint,
							),
						)
					}
					infeasiblePlanUnits[solutionPlanUnit] = true
					continue PlanUnitLoop
				}
			}

			index, err := move.(*solutionMoveStopsImpl).attach()
			if err != nil {
				return err
			}
			constraint, _, err := s.isFeasible(index, false)
			if err != nil {
				return err
			}
			if constraint != nil {
				if planUnit.IsFixed() {
					return fmt.Errorf(
						reportInfeasibleInitialSolution(
							move,
							solutionObserver.Constraint(),
						),
					)
				}
				for _, position := range move.(*solutionMoveStopsImpl).stopPositions {
					position.Stop().detach()
				}
				infeasiblePlanUnits[solutionPlanUnit] = true
				continue
			}
		}

		constraint, index, err := s.isFeasible(solutionVehicle.First().Index(), true)
		for ; constraint != nil; constraint, index, err = s.isFeasible(solutionVehicle.First().Index(), true) {
			if err != nil {
				return err
			}

			if index == solutionVehicle.First().Index() {
				return fmt.Errorf("infeasible initial solution at start of vehicle: %v", constraint)
			}

			for index == solutionVehicle.Last().Index() ||
				s.unwrapRootPlanUnit(s.stopToPlanUnit[index]).IsFixed() {
				index = s.previous[index]
				if index == solutionVehicle.First().Index() {
					return fmt.Errorf(
						"no feasible route from start to end found for vehicle %v"+
							" due to constraint %v, no further stops to remove",
						solutionVehicle.ModelVehicle().ID(),
						constraint)
				}
			}

			for _, solutionPlanUnit := range s.unwrapRootPlanUnit(s.stopToPlanUnit[index]).PlannedPlanStopsUnits() {
				if solutionPlanUnit.IsPlanned() {
					for _, solutionStop := range solutionPlanUnit.SolutionStops() {
						solutionStop.detach()
					}
				}
			}

			infeasiblePlanUnits[s.unwrapRootPlanUnit(s.stopToPlanUnit[index])] = true
		}

		for solutionPlanUnit := range allPlanUnits {
			if _, ok := infeasiblePlanUnits[solutionPlanUnit]; ok {
				continue
			}

			s.unPlannedPlanUnits.remove(solutionPlanUnit)

			if solutionPlanUnit.IsFixed() {
				s.fixedPlanUnits.add(solutionPlanUnit)
			} else {
				s.plannedPlanUnits.add(solutionPlanUnit)
			}
		}

		// Make sure all constraints and objectives are up-to-date
		_, _, err = s.isFeasible(solutionVehicle.First().Index(), true)
		if err != nil {
			return err
		}
	}

	return nil
}

type solutionImpl struct {
	model                  Model
	scores                 map[ModelObjective]float64
	values                 map[int][]float64
	objectiveStopData      map[ModelObjective][]Copier
	constraintStopData     map[ModelConstraint][]Copier
	objectiveSolutionData  map[ModelObjective]Copier
	constraintSolutionData map[ModelConstraint]Copier
	cumulativeValues       map[int][]float64

	// TODO: explore if stopToPlanUnit should rather contain interfaces
	stopToPlanUnit       []*solutionPlanStopsUnitImpl
	random               *rand.Rand
	plannedPlanUnits     solutionPlanUnitCollectionBaseImpl
	fixedPlanUnits       solutionPlanUnitCollectionBaseImpl
	unPlannedPlanUnits   solutionPlanUnitCollectionBaseImpl
	propositionPlanUnits solutionPlanUnitCollectionBaseImpl
	vehicleIndices       []int

	// TODO: explore if vehicles should rather be interfaces, then we can avoid creating new vehicles on the fly
	vehicles                 []SolutionVehicle
	solutionVehicles         []SolutionVehicle
	start                    []float64
	slack                    []float64
	arrival                  []float64
	next                     []int
	stopPosition             []int
	first                    []int
	stop                     []int
	cumulativeTravelDuration []float64
	end                      []float64
	previous                 []int
	inVehicle                []int
	last                     []int
	randomMutex              sync.Mutex
}

func (s *solutionImpl) SolutionPlanStopsUnit(planUnit ModelPlanStopsUnit) SolutionPlanStopsUnit {
	if planUnit == nil {
		return nil
	}
	return s.solutionPlanStopsUnit(planUnit)
}

func (s *solutionImpl) SolutionPlanUnit(planUnit ModelPlanUnit) SolutionPlanUnit {
	if planUnit == nil {
		return nil
	}
	return s.solutionPlanUnit(planUnit)
}

func (s *solutionImpl) solutionPlanUnit(planUnit ModelPlanUnit) SolutionPlanUnit {
	solutionPlanUnit := s.plannedPlanUnits.SolutionPlanUnit(planUnit)
	if solutionPlanUnit != nil {
		return solutionPlanUnit
	}
	solutionPlanUnit = s.unPlannedPlanUnits.SolutionPlanUnit(planUnit)
	if solutionPlanUnit != nil {
		return solutionPlanUnit
	}
	solutionPlanUnit = s.fixedPlanUnits.SolutionPlanUnit(planUnit)
	if solutionPlanUnit != nil {
		return solutionPlanUnit
	}
	solutionPlanUnit = s.propositionPlanUnits.SolutionPlanUnit(planUnit)
	if solutionPlanUnit != nil {
		return solutionPlanUnit
	}
	return nil
}

func (s *solutionImpl) solutionPlanStopsUnit(planUnit ModelPlanStopsUnit) *solutionPlanStopsUnitImpl {
	return s.solutionPlanUnit(planUnit).(*solutionPlanStopsUnitImpl)
}

func (s *solutionImpl) solutionPlanUnitsUnit(planUnit ModelPlanUnitsUnit) *solutionPlanUnitsUnitImpl {
	return s.solutionPlanUnit(planUnit).(*solutionPlanUnitsUnitImpl)
}

func (s *solutionImpl) SolutionStop(stop ModelStop) SolutionStop {
	if stop != nil && stop.HasPlanStopsUnit() {
		return s.SolutionPlanStopsUnit(stop.PlanStopsUnit()).SolutionStop(stop)
	}
	return SolutionStop{}
}

func (s *solutionImpl) SolutionVehicle(vehicle ModelVehicle) SolutionVehicle {
	if solutionVehicle, ok := s.solutionVehicle(vehicle); ok {
		return solutionVehicle
	}
	return SolutionVehicle{}
}

func (s *solutionImpl) solutionVehicle(vehicle ModelVehicle) (SolutionVehicle, bool) {
	if vehicle != nil {
		return SolutionVehicle{
			index:    vehicle.Index(),
			solution: s,
		}, true
	}
	return SolutionVehicle{}, false
}

func (s *solutionImpl) Copy() Solution {
	model := s.model.(*modelImpl)

	model.OnCopySolution(s)
	s.randomMutex.Lock()
	random := rand.New(rand.NewSource(s.random.Int63()))
	s.randomMutex.Unlock()

	// in order to reduce the number of allocations, we allocate
	// larger chunks of memory for the slices and then slice them
	// to the correct size
	nrStops := len(s.stop)
	nrVehicles := len(s.vehicles)
	nrExpressions := len(model.expressions)
	ints := make([]int, 5*nrStops+3*nrVehicles)
	first, ints := common.CopySliceFrom(ints, s.first)
	vehicleIndices, ints := common.CopySliceFrom(ints, s.vehicleIndices)
	last, ints := common.CopySliceFrom(ints, s.last)
	inVehicle, ints := common.CopySliceFrom(ints, s.inVehicle)
	previous, ints := common.CopySliceFrom(ints, s.previous)
	next, ints := common.CopySliceFrom(ints, s.next)
	stopPosition, ints := common.CopySliceFrom(ints, s.stopPosition)
	stop, _ := common.CopySliceFrom(ints, s.stop)
	floats := make([]float64, (5+2*nrExpressions)*nrStops)
	start, floats := common.CopySliceFrom(floats, s.start)
	end, floats := common.CopySliceFrom(floats, s.end)
	arrival, floats := common.CopySliceFrom(floats, s.arrival)
	slack, floats := common.CopySliceFrom(floats, s.slack)
	cumulativeTravelDuration, floats := common.CopySliceFrom(floats, s.cumulativeTravelDuration)
	solution := &solutionImpl{
		arrival:                  arrival,
		slack:                    slack,
		constraintStopData:       make(map[ModelConstraint][]Copier, len(s.constraintStopData)),
		objectiveStopData:        make(map[ModelObjective][]Copier, len(s.objectiveStopData)),
		constraintSolutionData:   make(map[ModelConstraint]Copier, len(s.constraintSolutionData)),
		objectiveSolutionData:    make(map[ModelObjective]Copier, len(s.objectiveSolutionData)),
		cumulativeTravelDuration: cumulativeTravelDuration,
		cumulativeValues:         make(map[int][]float64, len(s.cumulativeValues)),
		stopToPlanUnit:           make([]*solutionPlanStopsUnitImpl, len(s.stopToPlanUnit)),
		end:                      end,
		first:                    first,
		inVehicle:                inVehicle,
		last:                     last,
		model:                    model,
		next:                     next,
		previous:                 previous,
		start:                    start,
		stop:                     stop,
		stopPosition:             stopPosition,
		values:                   make(map[int][]float64, len(s.values)),
		vehicleIndices:           vehicleIndices,
		random:                   random,
		fixedPlanUnits: newSolutionPlanUnitCollectionBaseImpl(
			random, s.fixedPlanUnits.Size(),
		),
		plannedPlanUnits: newSolutionPlanUnitCollectionBaseImpl(
			random, s.plannedPlanUnits.Size(),
		),
		unPlannedPlanUnits: newSolutionPlanUnitCollectionBaseImpl(
			random, s.unPlannedPlanUnits.Size(),
		),
		propositionPlanUnits: newSolutionPlanUnitCollectionBaseImpl(
			random, s.propositionPlanUnits.Size(),
		),
		scores: make(map[ModelObjective]float64, len(s.scores)),
	}

	solution.vehicles = slices.Clone(s.vehicles)
	solution.solutionVehicles = slices.Clone(s.solutionVehicles)
	// update solution reference
	for idx := range solution.vehicles {
		solution.vehicles[idx].solution = solution
		solution.solutionVehicles[idx] = solution.vehicles[idx]
	}

	for _, expression := range model.expressions {
		eIndex := expression.Index()
		solution.cumulativeValues[eIndex], floats = common.CopySliceFrom(floats, s.cumulativeValues[eIndex])
		solution.values[eIndex], floats = common.CopySliceFrom(floats, s.values[eIndex])
	}

	for _, constraint := range model.constraintsWithStopUpdater {
		solution.constraintStopData[constraint] = make(
			[]Copier,
			len(s.constraintStopData[constraint]),
		)
		for idx, data := range s.constraintStopData[constraint] {
			if data == nil {
				solution.constraintStopData[constraint][idx] = nil
			} else {
				solution.constraintStopData[constraint][idx] = data.Copy()
			}
		}
	}

	for _, constraint := range model.constraintsWithSolutionUpdater {
		if s.constraintSolutionData[constraint] != nil {
			solution.constraintSolutionData[constraint] = s.constraintSolutionData[constraint].Copy()
		}
	}

	for _, objective := range model.objectivesWithStopUpdater {
		solution.objectiveStopData[objective] = make(
			[]Copier,
			len(s.objectiveStopData[objective]),
		)
		for idx, data := range s.objectiveStopData[objective] {
			if data == nil {
				solution.objectiveStopData[objective][idx] = nil
			} else {
				solution.objectiveStopData[objective][idx] = data.Copy()
			}
		}
	}

	for _, objective := range model.objectivesWithSolutionUpdater {
		if s.objectiveSolutionData[objective] != nil {
			solution.objectiveSolutionData[objective] = s.objectiveSolutionData[objective].Copy()
		}
	}

	for _, solutionPlanUnit := range s.fixedPlanUnits.solutionPlanUnits {
		solution.fixedPlanUnits.add(copySolutionPlanUnit(solutionPlanUnit, solution))
	}

	for _, solutionPlanUnit := range s.plannedPlanUnits.solutionPlanUnits {
		solution.plannedPlanUnits.add(copySolutionPlanUnit(solutionPlanUnit, solution))
	}

	for _, solutionPlanUnit := range s.unPlannedPlanUnits.solutionPlanUnits {
		solution.unPlannedPlanUnits.add(copySolutionPlanUnit(solutionPlanUnit, solution))
	}

	for idx, score := range s.scores {
		solution.scores[idx] = score
	}

	model.OnCopiedSolution(solution)

	return solution
}

func (s *solutionImpl) SetRandom(random *rand.Rand) error {
	if random == nil {
		return fmt.Errorf("random is nil")
	}
	s.random = random

	return nil
}

func (s *solutionImpl) Random() *rand.Rand {
	return s.random
}

func (s *solutionImpl) newVehicle(
	modelVehicle ModelVehicle,
) (SolutionVehicle, error) {
	if modelVehicle == nil {
		return SolutionVehicle{}, fmt.Errorf("modelVehicle is nil")
	}

	model := s.model.(*modelImpl)

	start := modelVehicle.Start().Sub(model.Epoch()) / model.DurationUnit()

	s.arrival = append(s.arrival, float64(start), 0)
	s.slack = append(s.slack, math.MaxFloat64, math.MaxFloat64)
	s.cumulativeTravelDuration = append(s.cumulativeTravelDuration, 0, 0)
	s.end = append(s.end, float64(start), 0)
	s.first = append(s.first, len(s.stop))
	s.inVehicle = append(s.inVehicle, len(s.vehicles), len(s.vehicles))
	s.last = append(s.last, len(s.stop)+1)
	s.next = append(s.next, len(s.next)+1, len(s.next)+1)
	s.previous = append(s.previous, len(s.previous), len(s.previous))
	s.stopPosition = append(s.stopPosition, 0, 1)
	s.start = append(s.start, float64(start), 0)
	s.stop = append(
		s.stop,
		modelVehicle.First().Index(),
		modelVehicle.Last().Index(),
	)
	s.vehicleIndices = append(s.vehicleIndices, modelVehicle.Index())
	s.vehicles = append(s.vehicles, SolutionVehicle{
		index:    modelVehicle.Index(),
		solution: s,
	})
	s.solutionVehicles = append(s.solutionVehicles, SolutionVehicle{
		index:    modelVehicle.Index(),
		solution: s,
	})

	for _, expression := range model.expressions {
		value := expression.Value(
			modelVehicle.VehicleType(),
			modelVehicle.First(),
			modelVehicle.First(),
		)
		s.values[expression.Index()] = append(
			s.values[expression.Index()],
			value,
			0,
		)
		s.cumulativeValues[expression.Index()] = append(
			s.cumulativeValues[expression.Index()],
			value,
			value,
		)
	}

	for _, constraint := range model.constraintsWithStopUpdater {
		s.constraintStopData[constraint] = append(
			s.constraintStopData[constraint],
			nil,
			nil,
		)
	}

	for _, objective := range model.objectivesWithStopUpdater {
		s.objectiveStopData[objective] = append(
			s.objectiveStopData[objective],
			nil,
			nil,
		)
	}

	constraint, _, err := s.isFeasible(len(s.stop)-2, true)
	if err != nil {
		return SolutionVehicle{}, err
	}
	if constraint != nil {
		return SolutionVehicle{}, fmt.Errorf("failed creating new vehicle: %v", constraint)
	}

	return toSolutionVehicle(s, len(s.vehicles)-1), nil
}

func (s *solutionImpl) checkConstraintsAndEstimateDeltaScore(
	m SolutionMoveStops,
) (deltaScore float64, feasible bool, planPositionsHint StopPositionsHint) {
	model := s.model.(*modelImpl)
	for _, constraint := range model.constraints {
		s.model.OnEstimateIsViolated(
			constraint,
		)

		var isViolated bool
		var hint *stopPositionHintImpl

		isViolatedTemp, hintTemp := constraint.EstimateIsViolated(m)

		if hintTemp == nil {
			panic(newErrorOnNilHint(constraint))
		}

		hint = hintTemp.(*stopPositionHintImpl)
		isViolated = isViolatedTemp

		s.model.OnEstimatedIsViolated(
			m,
			constraint,
			isViolated,
			hint,
		)

		if isViolated {
			return 0.0, false, hint
		}
	}

	s.model.OnEstimateDeltaObjectiveScore()

	objectiveEstimate := 0.0
	objectiveEstimate = s.Model().Objective().EstimateDeltaValue(m)

	s.model.OnEstimatedDeltaObjectiveScore(objectiveEstimate)

	return objectiveEstimate,
		true,
		constNoPositionsHint
}

var constNoPositionsHintImpl = noPositionsHint()

func (s *solutionImpl) checkConstraints(
	m SolutionMoveStops,
) (feasible bool, planPositionsHint *stopPositionHintImpl) {
	model := s.model.(*modelImpl)
	for _, constraint := range model.constraints {
		s.model.OnEstimateIsViolated(
			constraint,
		)

		var isViolated bool
		var hint *stopPositionHintImpl

		isViolatedTemp, hintTemp := constraint.EstimateIsViolated(m)

		if hintTemp == nil {
			panic(newErrorOnNilHint(constraint))
		}

		hint = hintTemp.(*stopPositionHintImpl)

		isViolated = isViolatedTemp

		s.model.OnEstimatedIsViolated(
			m,
			constraint,
			isViolated,
			hint,
		)

		if isViolated {
			return false, hint
		}
	}

	return true, constNoPositionsHintImpl
}

func (s *solutionImpl) estimateDeltaScore(
	m SolutionMoveStops,
) (deltaScore float64) {
	s.model.OnEstimateDeltaObjectiveScore()

	objectiveEstimate := s.model.(*modelImpl).objective.EstimateDeltaValue(m)

	s.model.OnEstimatedDeltaObjectiveScore(objectiveEstimate)

	return objectiveEstimate
}

func (s *solutionImpl) ConstraintData(constraint ModelConstraint) any {
	return s.constraintSolutionData[constraint]
}

func (s *solutionImpl) ObjectiveData(objective ModelObjective) any {
	return s.objectiveSolutionData[objective]
}

func (s *solutionImpl) ObjectiveValue(objective ModelObjective) float64 {
	if value, ok := s.scores[objective]; ok {
		return value
	}
	return 0.0
}

func (s *solutionImpl) Score() float64 {
	return s.scores[s.model.Objective()]
}

func (s *solutionImpl) FixedPlanUnits() ImmutableSolutionPlanUnitCollection {
	return &s.fixedPlanUnits
}

func (s *solutionImpl) PlannedPlanUnits() ImmutableSolutionPlanUnitCollection {
	return &s.plannedPlanUnits
}

func (s *solutionImpl) UnPlannedPlanUnits() ImmutableSolutionPlanUnitCollection {
	return &s.unPlannedPlanUnits
}

// PreAllocatedMoveContainer is used to reduce allocations.
// It contains objects that can be used instead of allocating new ones.
type PreAllocatedMoveContainer struct {
	// singleStopPosSolutionMoveStop has the underlying type *solutionMoveStopsImpl.
	// and has a length 1 stopPositions slice.
	singleStopPosSolutionMoveStop SolutionMoveStops
}

// NewPreAllocatedMoveContainer creates a new PreAllocatedMoveContainer.
// The PreAllocatedMoveContainer is initialized with concrete values depending on the planUnit type at runtime.
func NewPreAllocatedMoveContainer(planUnit SolutionPlanUnit) *PreAllocatedMoveContainer {
	allocations := PreAllocatedMoveContainer{}
	switch planUnit.(type) {
	case SolutionPlanStopsUnit:
		m := newNotExecutableSolutionMoveStops(planUnit.(*solutionPlanStopsUnitImpl))
		m.stopPositions = make([]StopPosition, 1, 2)
		allocations.singleStopPosSolutionMoveStop = m
	case SolutionPlanUnitsUnit:
	}
	return &allocations
}

func (s *solutionImpl) BestMove(ctx context.Context, planUnit SolutionPlanUnit) SolutionMove {
	if planUnit.Solution().(*solutionImpl) != s {
		panic("plan planUnit does not belong to this solution")
	}

	var bestMove SolutionMove
	// we initialize bestMove with the most likely type the moves will have
	switch planUnit.(type) {
	case SolutionPlanStopsUnit:
		bestMove = newNotExecutableSolutionMoveStops(planUnit.(*solutionPlanStopsUnitImpl))
	case SolutionPlanUnitsUnit:
		bestMove = newNotExecutableSolutionMoveUnits(planUnit.(*solutionPlanUnitsUnitImpl))
	default:
		bestMove = NotExecutableMove
	}

	s.model.OnBestMove(s)

	if planUnit.IsPlanned() {
		return bestMove
	}

	solutionVehicle := SolutionVehicle{
		index:    -1,
		solution: s,
	}

	// Depending on the type of planUnit we'll have to pre-allocate
	// some data structure that can be reused for all moves.
	// This is done to reduce allocations.
	sharedMoveContainer := NewPreAllocatedMoveContainer(planUnit)

	for vehicleIdx := 0; vehicleIdx < len(s.vehicles); vehicleIdx++ {
		solutionVehicle.index = vehicleIdx
		select {
		case <-ctx.Done():
			return bestMove
		default:
			newMove := solutionVehicle.bestMove(ctx, planUnit, sharedMoveContainer)
			bestMove = takeBestInPlace(bestMove, newMove)
		}
	}

	s.model.OnBestMoveFound(bestMove)
	return bestMove
}

func (s *solutionImpl) Model() Model {
	return s.model
}

func (s *solutionImpl) Vehicles() SolutionVehicles {
	return slices.Clone(s.solutionVehicles)
}

func (s *solutionImpl) vehiclesMutable() SolutionVehicles {
	return s.solutionVehicles
}

func (s *solutionImpl) value(
	expression ModelExpression,
	index int,
) float64 {
	return s.values[expression.Index()][index]
}

func (s *solutionImpl) cumulativeValue(
	expression ModelExpression,
	index int,
) float64 {
	return s.cumulativeValues[expression.Index()][index]
}

func (s *solutionImpl) constraintValue(
	constraint ModelConstraint,
	index int,
) any {
	if data, ok := s.constraintStopData[constraint]; ok {
		return data[index]
	}
	return nil
}

func (s *solutionImpl) objectiveValue(
	objective ModelObjective,
	index int,
) any {
	if data, ok := s.objectiveStopData[objective]; ok {
		return data[index]
	}
	return nil
}

func filterConstraint(constraint ModelConstraint, includeTemporal bool) bool {
	if includeTemporal {
		return false
	}
	if constraintTemporal, ok := constraint.(ConstraintTemporal); ok && constraintTemporal.IsTemporal() {
		return true
	}
	return false
}

// isFeasible returns the first constraint that is not feasible or nil if all
// constraints are feasible. Furthermore, it returns the index of the stop
// causing the violation.
func (s *solutionImpl) isFeasible(index int, includeTemporal bool) (
	violatedConstraint ModelConstraint,
	violatedIndex int,
	err error,
) {
	model := s.model.(*modelImpl)
	vehicle := s.model.Vehicle(s.vehicleIndices[s.inVehicle[index]]).(*modelVehicleImpl)
	vehicleType := vehicle.VehicleType()

	solutionStop := SolutionStop{
		solution: s,
		index:    index,
	}

	for _, constraint := range model.constraintsWithStopUpdater {
		value, err := constraint.(ConstraintStopDataUpdater).
			UpdateConstraintStopData(
				solutionStop,
			)
		if err != nil {
			return nil, -1, err
		}
		s.constraintStopData[constraint][index] = value
	}

	for _, objective := range model.objectivesWithStopUpdater {
		value, err := objective.(ObjectiveStopDataUpdater).
			UpdateObjectiveStopData(
				solutionStop,
			)
		if err != nil {
			return nil, -1, err
		}
		s.objectiveStopData[objective][index] = value
	}

	for s.next[index] != index {
		end := s.end[index]
		next := s.next[index]

		for _, expression := range model.expressions {
			value := expression.Value(
				vehicleType,
				model.stops[s.stop[index]],
				model.stops[s.stop[next]],
			)
			s.values[expression.Index()][next] = value
			s.cumulativeValues[expression.Index()][next] = s.cumulativeValues[expression.Index()][index] + value
		}

		travelDuration, arrival, start, end := vehicleType.TemporalValues(
			end,
			model.stops[s.stop[index]],
			model.stops[s.stop[next]],
		)

		s.cumulativeTravelDuration[next] = s.cumulativeTravelDuration[index] + travelDuration
		s.arrival[next] = arrival
		s.start[next] = start
		s.end[next] = end

		s.stopPosition[next] = s.stopPosition[index] + 1

		solutionStop = SolutionStop{
			solution: s,
			index:    next,
		}

		for _, constraint := range model.constraintsWithStopUpdater {
			value, err := constraint.(ConstraintStopDataUpdater).
				UpdateConstraintStopData(
					solutionStop,
				)
			if err != nil {
				return nil, -1, err
			}
			s.constraintStopData[constraint][next] = value
		}

		for _, objective := range model.objectivesWithStopUpdater {
			value, err := objective.(ObjectiveStopDataUpdater).
				UpdateObjectiveStopData(
					solutionStop,
				)
			if err != nil {
				return nil, -1, err
			}
			s.objectiveStopData[objective][next] = value
		}

		index = next

		for _, constraint := range model.constraintMap[AtEachStop] {
			if filterConstraint(constraint, includeTemporal) {
				continue
			}
			if s.isStopNotFeasible(constraint, solutionStop) {
				return constraint, index, nil
			}
		}
		if s.next[index] == index {
			for _, constraint := range model.constraintMap[AtEachVehicle] {
				if filterConstraint(constraint, includeTemporal) {
					continue
				}
				if s.isVehicleNotFeasible(constraint, s.inVehicle[index]) {
					return constraint, index, nil
				}
			}
		}
	}
	for _, constraint := range model.constraintMap[AtEachSolution] {
		if filterConstraint(constraint, includeTemporal) {
			continue
		}
		if s.isSolutionNotFeasible(constraint) {
			return constraint, index, nil
		}
	}

	for _, constraint := range model.constraintsWithSolutionUpdater {
		value, err := constraint.(ConstraintSolutionDataUpdater).
			UpdateConstraintSolutionData(s)
		if err != nil {
			return nil, -1, err
		}
		s.constraintSolutionData[constraint] = value
	}

	for _, objective := range model.objectivesWithSolutionUpdater {
		value, err := objective.(ObjectiveSolutionDataUpdater).
			UpdateObjectiveSolutionData(s)
		if err != nil {
			return nil, -1, err
		}
		s.objectiveSolutionData[objective] = value
	}

	slack := 0.0
	for s.previous[index] != index {
		previous := s.previous[index]

		slack += s.start[index] - s.arrival[index]

		s.slack[index] = slack

		index = previous
	}

	terms := model.objective.Terms()
	// TODO: do we always have to init the map?
	if s.scores == nil {
		s.scores = make(map[ModelObjective]float64, len(terms)+1)
	}
	totalScore := 0.0
	for _, term := range terms {
		score := term.Objective().Value(s) * term.Factor()
		s.scores[term.Objective()] = score
		totalScore += score
	}
	s.scores[model.objective] = totalScore

	return nil, -1, nil
}

func (s *solutionImpl) isStopNotFeasible(
	constraint ModelConstraint,
	stop SolutionStop,
) bool {
	s.model.OnCheckConstraint(constraint, AtEachStop)
	violated := constraint.(SolutionStopViolationCheck).DoesStopHaveViolations(stop)
	s.model.OnStopConstraintChecked(stop, constraint, !violated)
	return violated
}

func (s *solutionImpl) isVehicleNotFeasible(
	constraint ModelConstraint,
	vehicleIndex int,
) bool {
	s.model.OnCheckConstraint(constraint, AtEachVehicle)
	violated := constraint.(SolutionVehicleViolationCheck).
		DoesVehicleHaveViolations(
			toSolutionVehicle(s, vehicleIndex),
		)
	s.model.OnVehicleConstraintChecked(toSolutionVehicle(s, vehicleIndex), constraint, !violated)
	return violated
}

func (s *solutionImpl) isSolutionNotFeasible(
	constraint ModelConstraint,
) bool {
	s.model.OnCheckConstraint(constraint, AtEachSolution)
	violated := constraint.(SolutionViolationCheck).
		DoesSolutionHaveViolations(s)
	s.model.OnSolutionConstraintChecked(constraint, !violated)
	return violated
}
