// © 2019-present nextmv.io inc

package nextroute

// SolutionObserver is an interface that can be implemented to observe the
// solution manipulation process.
type SolutionObserver interface {
	// OnNewSolution is called when a new solution is going to be created.
	OnNewSolution(model Model)
	// OnNewSolutionCreated is called when a new solution has been created.
	OnNewSolutionCreated(solution Solution)

	// OnCopySolution is called when a solution is going to be copied.
	OnCopySolution(solution Solution)
	// OnCopiedSolution is called when a solution has been copied.
	OnCopiedSolution(solution Solution)

	// OnCheckConstraint is called when a constraint is going to be checked.
	OnCheckConstraint(
		constraint ModelConstraint,
		violation CheckedAt,
	)

	// OnSolutionConstraintChecked is called when a constraint has been checked.
	OnSolutionConstraintChecked(
		constraint ModelConstraint,
		feasible bool,
	)

	// OnStopConstraintChecked is called when a stop constraint has been checked.
	OnStopConstraintChecked(
		stop SolutionStop,
		constraint ModelConstraint,
		feasible bool,
	)

	// OnVehicleConstraintChecked is called when a vehicle constraint has been checked.
	OnVehicleConstraintChecked(
		vehicle SolutionVehicle,
		constraint ModelConstraint,
		feasible bool,
	)

	// OnEstimateIsViolated is called when the delta constraint is going to be
	// estimated if it will be violated
	OnEstimateIsViolated(
		constraint ModelConstraint,
	)
	// OnEstimatedIsViolated is called when the delta constraint score
	// has been estimated.
	OnEstimatedIsViolated(
		move SolutionMove,
		constraint ModelConstraint,
		isViolated bool,
		planPositionsHint StopPositionsHint,
	)
	// OnEstimateDeltaObjectiveScore is called when the delta objective score is
	// going to be estimated.
	OnEstimateDeltaObjectiveScore()
	// OnEstimatedDeltaObjectiveScore is called when the delta objective score
	// has been estimated.
	OnEstimatedDeltaObjectiveScore(
		estimate float64,
	)
	// OnBestMove is called when the solution is asked for it's best move.
	OnBestMove(solution Solution)
	// OnBestMoveFound is called when the solution has found it's best move.
	OnBestMoveFound(move SolutionMove)

	// OnPlan is called when a move is going to be planned.
	OnPlan(move SolutionMove)
	// OnPlanFailed is called when a move has failed to be planned.
	OnPlanFailed(move SolutionMove, constraint ModelConstraint)
	// OnPlanSucceeded is called when a move has succeeded to be planned.
	OnPlanSucceeded(move SolutionMove)
}

// SolutionObservers is a slice of SolutionObserver.
type SolutionObservers []SolutionObserver

// SolutionUnPlanObserver is an interface that can be implemented to observe the
// plan units un-planning process.
type SolutionUnPlanObserver interface {
	// OnUnPlan is called when a planUnit is going to be un-planned.
	OnUnPlan(planUnit SolutionPlanStopsUnit)
	// OnUnPlanFailed is called when a planUnit has failed to be un-planned.
	OnUnPlanFailed(planUnit SolutionPlanStopsUnit)
	// OnUnPlanSucceeded is called when a planUnit has succeeded to be un-planned.
	OnUnPlanSucceeded(planUnit SolutionPlanStopsUnit)
}

// SolutionUnPlanObservers is a slice of SolutionUnPlanObserver.
type SolutionUnPlanObservers []SolutionUnPlanObserver

// SolutionObserved is an interface that can be implemented to observe the
// solution manipulation process.
type SolutionObserved interface {
	SolutionObserver
	SolutionUnPlanObserver

	// AddSolutionObserver adds the given solution observer to the solution
	// observed.
	AddSolutionObserver(observer SolutionObserver)

	// AddSolutionUnPlanObserver adds the given solution un-plan observer to the
	// solution observed.
	AddSolutionUnPlanObserver(observer SolutionUnPlanObserver)

	// RemoveSolutionObserver remove the given solution observer from the
	// solution observed.
	RemoveSolutionObserver(observer SolutionObserver)

	// RemoveSolutionUnPlanObserver remove the given solution un-plan observer
	// from the solution observed.
	RemoveSolutionUnPlanObserver(observer SolutionUnPlanObserver)

	// SolutionObservers returns the solution observers.
	SolutionObservers() SolutionObservers

	// SolutionUnPlanObservers returns the solution un-plan observers.
	SolutionUnPlanObservers() SolutionUnPlanObservers
}

type solutionObservedImpl struct {
	observers       SolutionObservers
	unplanObservers SolutionUnPlanObservers
}

func (s *solutionObservedImpl) AddSolutionObserver(observer SolutionObserver) {
	s.observers = append(s.observers, observer)
}

func (s *solutionObservedImpl) AddSolutionUnPlanObserver(observer SolutionUnPlanObserver) {
	s.unplanObservers = append(s.unplanObservers, observer)
}

func (s *solutionObservedImpl) RemoveSolutionUnPlanObserver(observer SolutionUnPlanObserver) {
	for i := 0; i < len(s.unplanObservers); i++ {
		if s.unplanObservers[i] == observer {
			s.unplanObservers = append(s.unplanObservers[:i], s.unplanObservers[i+1:]...)
			break
		}
	}
}

func (s *solutionObservedImpl) SolutionUnPlanObservers() SolutionUnPlanObservers {
	observers := make(SolutionUnPlanObservers, len(s.unplanObservers))
	copy(observers, s.unplanObservers)
	return observers
}

func (s *solutionObservedImpl) RemoveSolutionObserver(observer SolutionObserver) {
	for i := 0; i < len(s.observers); i++ {
		if s.observers[i] == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

func (s *solutionObservedImpl) OnUnPlan(
	planUnit SolutionPlanStopsUnit,
) {
	if len(s.unplanObservers) == 0 {
		return
	}
	for _, observer := range s.unplanObservers {
		observer.OnUnPlan(planUnit)
	}
}

func (s *solutionObservedImpl) OnUnPlanFailed(
	planUnit SolutionPlanStopsUnit,
) {
	if len(s.unplanObservers) == 0 {
		return
	}
	for _, observer := range s.unplanObservers {
		observer.OnUnPlanFailed(planUnit)
	}
}

func (s *solutionObservedImpl) OnUnPlanSucceeded(planUnit SolutionPlanStopsUnit) {
	if len(s.unplanObservers) == 0 {
		return
	}
	for _, observer := range s.unplanObservers {
		observer.OnUnPlanSucceeded(planUnit)
	}
}

func (s *solutionObservedImpl) SolutionObservers() SolutionObservers {
	observers := make(SolutionObservers, len(s.observers))
	copy(observers, s.observers)
	return observers
}

func (s *solutionObservedImpl) OnBestMove(solution Solution) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnBestMove(solution)
	}
}

func (s *solutionObservedImpl) OnBestMoveFound(move SolutionMove) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnBestMoveFound(move)
	}
}

func (s *solutionObservedImpl) OnNewSolution(model Model) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnNewSolution(model)
	}
}

func (s *solutionObservedImpl) OnNewSolutionCreated(solution Solution) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnNewSolutionCreated(solution)
	}
}

func (s *solutionObservedImpl) OnCopySolution(solution Solution) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnCopySolution(solution)
	}
}

func (s *solutionObservedImpl) OnCopiedSolution(solution Solution) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnCopiedSolution(solution)
	}
}

func (s *solutionObservedImpl) OnCheckConstraint(
	constraint ModelConstraint,
	checkViolationAt CheckedAt,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnCheckConstraint(constraint, checkViolationAt)
	}
}

func (s *solutionObservedImpl) OnStopConstraintChecked(
	stop SolutionStop,
	constraint ModelConstraint,
	violated bool,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnStopConstraintChecked(stop, constraint, violated)
	}
}

func (s *solutionObservedImpl) OnVehicleConstraintChecked(
	vehicle SolutionVehicle,
	constraint ModelConstraint,
	violated bool,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnVehicleConstraintChecked(vehicle, constraint, violated)
	}
}

func (s *solutionObservedImpl) OnSolutionConstraintChecked(
	constraint ModelConstraint,
	violated bool,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnSolutionConstraintChecked(constraint, violated)
	}
}

func (s *solutionObservedImpl) OnEstimateIsViolated(
	constraint ModelConstraint,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnEstimateIsViolated(constraint)
	}
}

func (s *solutionObservedImpl) OnEstimatedIsViolated(
	move SolutionMove,
	constraint ModelConstraint,
	isViolated bool,
	planPositionsHint StopPositionsHint,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnEstimatedIsViolated(
			move,
			constraint,
			isViolated,
			planPositionsHint,
		)
	}
}

func (s *solutionObservedImpl) OnEstimateDeltaObjectiveScore() {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnEstimateDeltaObjectiveScore()
	}
}

func (s *solutionObservedImpl) OnEstimatedDeltaObjectiveScore(
	estimate float64,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnEstimatedDeltaObjectiveScore(estimate)
	}
}

func (s *solutionObservedImpl) OnPlan(move SolutionMove) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnPlan(move)
	}
}

func (s *solutionObservedImpl) OnPlanFailed(
	move SolutionMove,
	constraint ModelConstraint,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnPlanFailed(move, constraint)
	}
}

func (s *solutionObservedImpl) OnPlanSucceeded(move SolutionMove) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnPlanSucceeded(move)
	}
}
