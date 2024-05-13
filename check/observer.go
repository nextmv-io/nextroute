// © 2019-present nextmv.io inc

package check

import "github.com/nextmv-io/nextroute"

// Observer is an observer that is used to detect the constraints
// that are violated when evaluating moves.
type Observer interface {
	nextroute.SolutionObserver

	// Constraints returns the constraints that are estimated to be violated.
	Constraints() nextroute.ModelConstraints

	// OnPlanFailedConstraints returns the constraints that have failed on plan.
	OnPlanFailedConstraints() nextroute.ModelConstraints

	// Reset resets the constraints that have been violated.
	Reset()
}

// newObserver creates a new observer.
func newObserver() Observer {
	return &observerImpl{}
}

type observerImpl struct {
	estimateIsViolatedConstraints nextroute.ModelConstraints
	onPlanFailedConstraints       nextroute.ModelConstraints
}

// OnStopConstraintChecked implements Observer.
func (*observerImpl) OnStopConstraintChecked(
	_ nextroute.SolutionStop,
	_ nextroute.ModelConstraint,
	_ bool) {
}

// OnVehicleConstraintChecked implements Observer.
func (*observerImpl) OnVehicleConstraintChecked(
	_ nextroute.SolutionVehicle,
	_ nextroute.ModelConstraint,
	_ bool) {
}

func (o *observerImpl) Constraints() nextroute.ModelConstraints {
	return o.estimateIsViolatedConstraints
}

func (o *observerImpl) OnPlanFailedConstraints() nextroute.ModelConstraints {
	return o.onPlanFailedConstraints
}

func (o *observerImpl) Reset() {
	o.estimateIsViolatedConstraints = o.estimateIsViolatedConstraints[:0]
	o.onPlanFailedConstraints = o.onPlanFailedConstraints[:0]
}

func (o *observerImpl) OnNewSolution(_ nextroute.Model) {
}

func (o *observerImpl) OnNewSolutionCreated(_ nextroute.Solution) {
}

func (o *observerImpl) OnCopySolution(_ nextroute.Solution) {
}

func (o *observerImpl) OnCopiedSolution(_ nextroute.Solution) {
}

func (o *observerImpl) OnCheckConstraint(_ nextroute.ModelConstraint, _ nextroute.CheckedAt) {
}

func (o *observerImpl) OnSolutionConstraintChecked(_ nextroute.ModelConstraint, _ bool) {
}

func (o *observerImpl) OnEstimateIsViolated(_ nextroute.ModelConstraint) {
}

func (o *observerImpl) OnEstimatedIsViolated(
	_ nextroute.SolutionMove,
	constraint nextroute.ModelConstraint,
	violated bool,
	_ nextroute.StopPositionsHint,
) {
	if violated {
		o.estimateIsViolatedConstraints = append(o.estimateIsViolatedConstraints, constraint)
	}
}

func (o *observerImpl) OnEstimateDeltaObjectiveScore() {
}

func (o *observerImpl) OnEstimatedDeltaObjectiveScore(_ float64) {
}

func (o *observerImpl) OnBestMove(_ nextroute.Solution) {
}

func (o *observerImpl) OnBestMoveFound(_ nextroute.SolutionMove) {
}

func (o *observerImpl) OnPlan(_ nextroute.SolutionMove) {
}

func (o *observerImpl) OnPlanFailed(_ nextroute.SolutionMove, constraint nextroute.ModelConstraint) {
	o.onPlanFailedConstraints = append(o.onPlanFailedConstraints, constraint)
}

func (o *observerImpl) OnPlanSucceeded(_ nextroute.SolutionMove) {
}
