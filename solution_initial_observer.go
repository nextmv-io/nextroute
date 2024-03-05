// © 2019-present nextmv.io inc

package nextroute

// InitialSolutionObserver is an observer that is used to detect the constraint
// that is violated by the initial solution.
type InitialSolutionObserver interface {
	SolutionObserver

	// Constraint returns the constraint that is violated by the initial
	// solution. Can be nil if no constraint is violated.
	Constraint() ModelConstraint
}

// newInitialSolutionObserver creates a new initial solution observer.
func newInitialSolutionObserver() InitialSolutionObserver {
	return &initialSolutionObserver{}
}

type initialSolutionObserver struct {
	constraint ModelConstraint
}

// OnSolutionConstraintChecked implements InitialSolutionObserver.
func (i *initialSolutionObserver) OnSolutionConstraintChecked(
	constraint ModelConstraint,
	feasible bool,
) {
	if !feasible {
		i.constraint = constraint
	}
}

// OnStopConstraintChecked implements InitialSolutionObserver.
func (i *initialSolutionObserver) OnStopConstraintChecked(
	_ SolutionStop,
	constraint ModelConstraint,
	feasible bool,
) {
	if !feasible {
		i.constraint = constraint
	}
}

// OnVehicleConstraintChecked implements InitialSolutionObserver.
func (i *initialSolutionObserver) OnVehicleConstraintChecked(
	_ SolutionVehicle,
	constraint ModelConstraint,
	feasible bool,
) {
	if !feasible {
		i.constraint = constraint
	}
}

func (i *initialSolutionObserver) Constraint() ModelConstraint {
	return i.constraint
}

func (i *initialSolutionObserver) OnNewSolution(_ Model) {
}

func (i *initialSolutionObserver) OnNewSolutionCreated(_ Solution) {
}

func (i *initialSolutionObserver) OnCopySolution(_ Solution) {
}

func (i *initialSolutionObserver) OnCopiedSolution(_ Solution) {
}

func (i *initialSolutionObserver) OnCheckConstraint(_ ModelConstraint, _ CheckedAt) {

}

func (i *initialSolutionObserver) OnEstimateIsViolated(constraint ModelConstraint) {
	i.constraint = constraint
}

func (i *initialSolutionObserver) OnEstimatedIsViolated(
	_ SolutionMove,
	_ ModelConstraint,
	_ bool,
	_ StopPositionsHint,
) {
}

func (i *initialSolutionObserver) OnEstimateDeltaObjectiveScore() {
}

func (i *initialSolutionObserver) OnEstimatedDeltaObjectiveScore(_ float64) {
}

func (i *initialSolutionObserver) OnBestMove(_ Solution) {
}

func (i *initialSolutionObserver) OnBestMoveFound(_ SolutionMove) {
}

func (i *initialSolutionObserver) OnPlan(_ SolutionMove) {
}

func (i *initialSolutionObserver) OnPlanFailed(_ SolutionMove, _ ModelConstraint) {
}

func (i *initialSolutionObserver) OnPlanSucceeded(_ SolutionMove) {
}
