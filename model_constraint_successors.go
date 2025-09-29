// © 2019-present nextmv.io inc

package nextroute

import "fmt"

// SuccessorConstraint is a constraint that disallows certain stops to be
// planned after other stops.
type SuccessorConstraint struct {
	disallowedSuccessors map[*ModelStop]ModelStops
}

// NewSuccessorConstraint returns a new SuccessorConstraint.
func NewSuccessorConstraint() (*SuccessorConstraint, error) {
	return &SuccessorConstraint{
		disallowedSuccessors: make(map[*ModelStop]ModelStops),
	}, nil
}

// Lock locks the constraint to the model.
func (l *SuccessorConstraint) Lock(model *Model) error {
	// initialize disallowedSuccessors
	model.disallowedSuccessors = make([][]bool, model.NumberOfStops())
	for i := range model.disallowedSuccessors {
		model.disallowedSuccessors[i] = make([]bool, model.NumberOfStops())
	}

	// copy the information from disallowedSuccessors to the model
	for stop, successors := range l.disallowedSuccessors {
		for _, successor := range successors {
			model.disallowedSuccessors[stop.Index()][successor.Index()] = true
		}
	}
	return nil
}

// DisallowSuccessors disallows the given successors to be planned after the
// given stop.
func (l *SuccessorConstraint) DisallowSuccessors(
	stop *ModelStop,
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

// String returns the name of the constraint.
func (l *SuccessorConstraint) String() string {
	return "successor"
}

// EstimationCost returns the cost of estimating whether a move violates the
// constraint.
func (l *SuccessorConstraint) EstimationCost() Cost {
	return LinearStop
}

// EstimateIsViolated estimates whether the given move violates the constraint.
func (l *SuccessorConstraint) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	model := move.PlanStopsUnit().Solution().Model()
	stopPositions := move.StopPositions()
	for _, stopPosition := range stopPositions {
		stop := stopPosition.Stop().ModelStop()
		nextModelStop := stopPosition.Next().ModelStop()
		if disallowed := model.disallowedSuccessors[stop.Index()][nextModelStop.Index()]; disallowed {
			return true, NoPositionsHint()
		}
	}
	return false, NoPositionsHint()
}

// DoesStopHaveViolations returns true if the stop has violations of the
// constraint.
func (l *SuccessorConstraint) DoesStopHaveViolations(
	stop SolutionStop,
) bool {
	model := stop.Solution().Model()
	stopImpl := stop
	previousModelStop := stopImpl.Previous().ModelStop()
	if disallowed := model.disallowedSuccessors[previousModelStop.Index()][stop.ModelStop().Index()]; disallowed {
		return true
	}
	return false
}
