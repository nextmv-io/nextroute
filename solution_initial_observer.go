package nextroute

import "github.com/nextmv-io/sdk/nextroute"

// InitialSolutionObserver is an observer that is used to detect the constraint
// that is violated by the initial solution.
type InitialSolutionObserver interface {
	nextroute.SolutionObserver

	// Constraint returns the constraint that is violated by the initial
	// solution. Can be nil if no constraint is violated.
	Constraint() nextroute.ModelConstraint
}

// newInitialSolutionObserver creates a new initial solution observer.
func newInitialSolutionObserver() InitialSolutionObserver {
	return &initialSolutionObserver{}
}

type initialSolutionObserver struct {
	constraint nextroute.ModelConstraint
}

// OnSolutionConstraintChecked implements InitialSolutionObserver.
func (i *initialSolutionObserver) OnSolutionConstraintChecked(
	constraint nextroute.ModelConstraint,
	feasible bool,
) {
	if !feasible {
		i.constraint = constraint
	}
}

// OnStopConstraintChecked implements InitialSolutionObserver.
func (i *initialSolutionObserver) OnStopConstraintChecked(
	_ nextroute.SolutionStop,
	constraint nextroute.ModelConstraint,
	feasible bool,
) {
	if !feasible {
		i.constraint = constraint
	}
}

// OnVehicleConstraintChecked implements InitialSolutionObserver.
func (i *initialSolutionObserver) OnVehicleConstraintChecked(
	_ nextroute.SolutionVehicle,
	constraint nextroute.ModelConstraint,
	feasible bool,
) {
	if !feasible {
		i.constraint = constraint
	}
}

func (i *initialSolutionObserver) Constraint() nextroute.ModelConstraint {
	return i.constraint
}

func (i *initialSolutionObserver) OnNewSolution(_ nextroute.Model) {
}

func (i *initialSolutionObserver) OnNewSolutionCreated(_ nextroute.Solution) {
}

func (i *initialSolutionObserver) OnCopySolution(_ nextroute.Solution) {
}

func (i *initialSolutionObserver) OnCopiedSolution(_ nextroute.Solution) {
}

func (i *initialSolutionObserver) OnCheckConstraint(_ nextroute.ModelConstraint, _ nextroute.CheckedAt) {

}

func (i *initialSolutionObserver) OnEstimateIsViolated(constraint nextroute.ModelConstraint) {
	i.constraint = constraint
}

func (i *initialSolutionObserver) OnEstimatedIsViolated(
	_ nextroute.SolutionMove,
	_ nextroute.ModelConstraint,
	_ bool,
	_ nextroute.StopPositionsHint,
) {
}

func (i *initialSolutionObserver) OnEstimateDeltaObjectiveScore() {
}

func (i *initialSolutionObserver) OnEstimatedDeltaObjectiveScore(_ float64) {
}

func (i *initialSolutionObserver) OnBestMove(_ nextroute.Solution) {
}

func (i *initialSolutionObserver) OnBestMoveFound(_ nextroute.SolutionMove) {
}

func (i *initialSolutionObserver) OnPlan(_ nextroute.SolutionMove) {
}

func (i *initialSolutionObserver) OnPlanFailed(_ nextroute.SolutionMove, _ nextroute.ModelConstraint) {
}

func (i *initialSolutionObserver) OnPlanSucceeded(_ nextroute.SolutionMove) {
}
