// © 2019-present nextmv.io inc

package nextroute

import "fmt"

// DirectPrecedencesConstraint is a constraint that limits the vehicles a plan unit
// can be added to. The Attribute constraint configures compatibility
// attributes for stops and vehicles separately. This is done by specifying
// a list of attributes for stops and vehicles, respectively. Stops that
// have configured attributes are only compatible with vehicles that match
// at least one of them. Stops that do not have any specified attributes are
// compatible with any vehicle. Vehicles that do not have any specified
// attributes are only compatible with stops without attributes.
type DirectPrecedencesConstraint interface {
	ModelConstraint
	DisallowSuccessors(ModelStop, ModelStops) error
}

// NewDirectPrecedencesConstraint returns a new DirectPrecedencesConstraint.
func NewDirectPrecedencesConstraint() (DirectPrecedencesConstraint, error) {
	return &directPrecedencesConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"direct_precedences",
			ModelExpressions{},
		),
		disallowedSuccessors: make(map[ModelStop]ModelStops),
	}, nil
}

type directPrecedencesConstraintImpl struct {
	modelConstraintImpl
	disallowedSuccessors map[ModelStop]ModelStops
}

func (l *directPrecedencesConstraintImpl) Lock(model Model) error {
	modelImpl := model.(*modelImpl)
	// copy the information from disallowedSuccessors to the model
	for stop, successors := range l.disallowedSuccessors {
		for _, successor := range successors {
			modelImpl.disallowedSuccessors[stop.Index()][successor.Index()] = true
		}
	}
	return nil
}

func (l *directPrecedencesConstraintImpl) DisallowSuccessors(
	stop ModelStop,
	successors ModelStops,
) error {
	if stop == nil {
		return fmt.Errorf("stop cannot be nil")
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

func (l *directPrecedencesConstraintImpl) String() string {
	return l.name
}

func (l *directPrecedencesConstraintImpl) EstimationCost() Cost {
	return LinearStop
}

func (l *directPrecedencesConstraintImpl) EstimateIsViolated(
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

func (l *directPrecedencesConstraintImpl) DoesStopHaveViolations(
	stop SolutionStop,
) bool {
	modelImpl := stop.Solution().Model().(*modelImpl)
	stopImpl := stop.(solutionStopImpl)
	nextModelStop := stopImpl.next().modelStop()
	if disallowed := modelImpl.disallowedSuccessors[stop.Index()][nextModelStop.Index()]; disallowed {
		return true
	}
	return false
}
