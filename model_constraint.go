// © 2019-present nextmv.io inc

package nextroute

// Copier is the interface that all objects that can be copied must implement.
type Copier interface {
	// Copy returns a copy of the object.
	Copy() Copier
}

// ConstraintReporter is the interface that can be used by a constraint if it
// wants to report data about the constraint for a solution stop. This data
// is used in the basic formatter.
type ConstraintReporter interface {
	// ReportConstraint returns the data that should be reported for the given
	// solution stop.
	ReportConstraint(SolutionStop) map[string]any
}

// Locker is an interface for locking a constraint. This interface is called
// when the model is locked. The constraint can use this to initialize data
// structures that are used to check the constraint.
type Locker interface {
	// Lock locks the constraint on locking a model.
	Lock(model Model) error
}

// ConstraintStopDataUpdater is the interface than can be used by a constraint if
// it wants to store data with each stop in a solution.
type ConstraintStopDataUpdater interface {
	// UpdateConstraintStopData is called when a stop is added to a solution.
	// The solutionStop has all it's expression values set and this function
	// can use them to update the constraint data for the stop. The data
	// returned can be used by the estimate function and can be retrieved by the
	// SolutionStop.ConstraintData function.
	UpdateConstraintStopData(s SolutionStop) (Copier, error)
}

// ConstraintSolutionDataUpdater is the interface than can be used by a
// constraint if it wants to store data with each solution.
type ConstraintSolutionDataUpdater interface {
	// UpdateConstraintSolutionData is called when a solution has been modified.
	// The solution has all it's expression values set and this function
	// can use them to update the constraint data for the solution. The data
	// returned can be used by the estimate function and can be retrieved by the
	// Solution.ConstraintData function.
	UpdateConstraintSolutionData(s Solution) (Copier, error)
}

// RegisteredModelExpressions is the interface that exposes the expressions
// that should be registered with the model.
type RegisteredModelExpressions interface {
	// ModelExpressions registers the expressions that are registered with the
	// model.
	ModelExpressions() ModelExpressions
}

// ConstraintTemporal is the interface that is implemented by constraints that
// are temporal. This interface is used to determine if the constraint is
// using temporal expressions, specifically travel durations. Temporal
// constraints require special handling when adding initial stops to a
// new solution because of the triangular inequality.
type ConstraintTemporal interface {
	// IsTemporal returns true if the constraint is temporal.
	IsTemporal() bool
}

// SolutionStopViolationCheck is the interface that will be invoked on every
// update of a planned solution stop. The method DoesStopHaveViolations will be
// called to check if the stop violates the constraint. If the method returns
// true the solution will not be accepted. If un-planning can result in a
// violation of the constraint one of the violation checks must be implemented.
type SolutionStopViolationCheck interface {
	// DoesStopHaveViolations returns true if the stop violates the constraint.
	// The stop is not allowed to be nil. The stop must be part of the solution.
	// This method is only called if CheckedAt returns AtEachStop.
	DoesStopHaveViolations(stop SolutionStop) bool
}

// SolutionVehicleViolationCheck is the interface that will be invoked on every
// update of a last planned solution stop of a vehicle. The method
// DoesVehicleHaveViolations will be called to check if the vehicle violates
// the constraint. If the method returns true the solution will not be accepted.
// If un-planning can result in a violation of the constraint one of the
// violation checks must be implemented.
type SolutionVehicleViolationCheck interface {
	// DoesVehicleHaveViolations returns true if the vehicle violates the
	// constraint. The vehicle is not allowed to be nil. The vehicle must be
	// part of the solution. This method is only called if CheckedAt returns
	// AtEachVehicle.
	DoesVehicleHaveViolations(vehicle SolutionVehicle) bool
}

// SolutionViolationCheck is the interface that will be invoked once all
// updates on a solution have been executed. The method
// DoesSolutionHaveViolations will be called to check if the vehicle violates
// the constraint. If the method returns true the solution will not be accepted.
// If un-planning can result in a violation of the constraint one of the
// violation checks must be implemented.
type SolutionViolationCheck interface {
	// DoesSolutionHaveViolations returns true if the solution violates the
	// constraint. The solution is not allowed to be nil. This method is only
	// called if CheckedAt returns AtEachSolution.
	DoesSolutionHaveViolations(solution Solution) bool
}

// ModelConstraint is the interface that all constraints must implement.
// Constraints are used to estimate if a move is allowed and can be used to
// check if a solution is valid after a move is executed or plan units have
// been unplanned.
type ModelConstraint interface {
	// EstimateIsViolated estimates if the given solution, when changed by the
	// given move will be violated or not. The move is not allowed to be nil.
	// It should be a pure function, i.e. not change any state of the
	// constraint. The stopPositionsHint can be used to speed up the estimation
	// of the constraint violation.
	EstimateIsViolated(SolutionMoveStops) (isViolated bool, stopPositionsHint StopPositionsHint)
}

// ModelConstraints is a slice of ModelConstraint.
type ModelConstraints []ModelConstraint

// ConstraintDataUpdater is a deprecated interface. Please use
// ConstraintStopDataUpdater instead.
type ConstraintDataUpdater interface {
	// UpdateConstraintData is deprecated.
	UpdateConstraintData(s SolutionStop) (Copier, error)
}

type modelConstraintImpl struct {
	name        string
	expressions ModelExpressions
}

func (m *modelConstraintImpl) EstimateIsViolated(
	_ SolutionMoveStops,
	_ Solution,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	panic("implement me in derived class")
}

func newModelConstraintImpl(
	name string,
	expressions ModelExpressions,
) modelConstraintImpl {
	return modelConstraintImpl{
		expressions: expressions,
		name:        name,
	}
}

func (m *modelConstraintImpl) ModelExpressions() ModelExpressions {
	return m.expressions
}
