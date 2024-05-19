// © 2019-present nextmv.io inc

package nextroute

import "fmt"

// SuccessorConstraint is a constraint that disallows certain stops to be
// planned after other stops.
type SuccessorConstraint interface {
	ModelConstraint
	DisallowSuccessors(ModelStop, ModelStops) error
}

// NewSuccessorConstraint returns a new SuccessorConstraint.
func NewSuccessorConstraint() (SuccessorConstraint, error) {
	return &successorConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"successor",
			ModelExpressions{},
		),
		disallowedSuccessors: make(map[ModelStop]ModelStops),
	}, nil
}

type successorConstraintImpl struct {
	modelConstraintImpl
	disallowedSuccessors map[ModelStop]ModelStops
}

func (l *successorConstraintImpl) Lock(model Model) error {
	modelImpl := model.(*modelImpl)

	// initialize disallowedSuccessors
	modelImpl.disallowedSuccessors = make([][]bool, modelImpl.NumberOfStops())
	for i := range modelImpl.disallowedSuccessors {
		modelImpl.disallowedSuccessors[i] = make([]bool, modelImpl.NumberOfStops())
	}

	// copy the information from disallowedSuccessors to the model
	for stop, successors := range l.disallowedSuccessors {
		for _, successor := range successors {
			modelImpl.disallowedSuccessors[stop.Index()][successor.Index()] = true
		}
	}
	return nil
}

func (l *successorConstraintImpl) DisallowSuccessors(
	stop ModelStop,
	successors ModelStops,
) error {
	if stop == nil {
		return fmt.Errorf("stop cannot be nil")
	}
	if stop.Model().IsLocked() {
		return fmt.Errorf(lockErrorMessage, "disallow successors")
	}
	if successors == nil {
		return fmt.Errorf("successors cannot be nil")
	}
	if _, ok := l.disallowedSuccessors[stop]; !ok {
		l.disallowedSuccessors[stop] = ModelStops{}
	}
	l.disallowedSuccessors[stop] = append(l.disallowedSuccessors[stop], successors...)
	return nil
}

func (l *successorConstraintImpl) String() string {
	return l.name
}

func (l *successorConstraintImpl) EstimationCost() Cost {
	return LinearStop
}

func (l *successorConstraintImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	modelImpl := move.PlanStopsUnit().Solution().Model().(*modelImpl)
	stopPositions := move.StopPositions()
	for _, stopPosition := range stopPositions {
		stop := stopPosition.Stop().ModelStop()
		nextModelStop := stopPosition.Next().ModelStop()
		if disallowed := modelImpl.disallowedSuccessors[stop.Index()][nextModelStop.Index()]; disallowed {
			return true, noPositionsHint()
		}
	}
	return false, noPositionsHint()
}

func (l *successorConstraintImpl) DoesStopHaveViolations(
	stop SolutionStop,
) bool {
	modelImpl := stop.Solution().Model().(*modelImpl)
	stopImpl := stop
	previousModelStop := stopImpl.Previous().modelStop()
	if disallowed := modelImpl.disallowedSuccessors[previousModelStop.Index()][stop.ModelStop().Index()]; disallowed {
		return true
	}
	return false
}
