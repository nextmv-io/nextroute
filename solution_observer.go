package nextroute

import "github.com/nextmv-io/sdk/nextroute"

type solutionObservedImpl struct {
	observers       nextroute.SolutionObservers
	unplanObservers nextroute.SolutionUnPlanObservers
}

func (s *solutionObservedImpl) AddSolutionObserver(observer nextroute.SolutionObserver) {
	s.observers = append(s.observers, observer)
}

func (s *solutionObservedImpl) AddSolutionUnPlanObserver(observer nextroute.SolutionUnPlanObserver) {
	s.unplanObservers = append(s.unplanObservers, observer)
}

func (s *solutionObservedImpl) RemoveSolutionUnPlanObserver(observer nextroute.SolutionUnPlanObserver) {
	for i := 0; i < len(s.unplanObservers); i++ {
		if s.unplanObservers[i] == observer {
			s.unplanObservers = append(s.unplanObservers[:i], s.unplanObservers[i+1:]...)
			break
		}
	}
}

func (s *solutionObservedImpl) SolutionUnPlanObservers() nextroute.SolutionUnPlanObservers {
	observers := make(nextroute.SolutionUnPlanObservers, len(s.unplanObservers))
	copy(observers, s.unplanObservers)
	return observers
}

func (s *solutionObservedImpl) RemoveSolutionObserver(observer nextroute.SolutionObserver) {
	for i := 0; i < len(s.observers); i++ {
		if s.observers[i] == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

func (s *solutionObservedImpl) OnUnPlan(
	planUnit nextroute.SolutionPlanStopsUnit,
) {
	if len(s.unplanObservers) == 0 {
		return
	}
	for _, observer := range s.unplanObservers {
		observer.OnUnPlan(planUnit)
	}
}

func (s *solutionObservedImpl) OnUnPlanFailed(
	planUnit nextroute.SolutionPlanStopsUnit,
) {
	if len(s.unplanObservers) == 0 {
		return
	}
	for _, observer := range s.unplanObservers {
		observer.OnUnPlanFailed(planUnit)
	}
}

func (s *solutionObservedImpl) OnUnPlanSucceeded(planUnit nextroute.SolutionPlanStopsUnit) {
	if len(s.unplanObservers) == 0 {
		return
	}
	for _, observer := range s.unplanObservers {
		observer.OnUnPlanSucceeded(planUnit)
	}
}

func (s *solutionObservedImpl) SolutionObservers() nextroute.SolutionObservers {
	observers := make(nextroute.SolutionObservers, len(s.observers))
	copy(observers, s.observers)
	return observers
}

func (s *solutionObservedImpl) OnBestMove(solution nextroute.Solution) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnBestMove(solution)
	}
}

func (s *solutionObservedImpl) OnBestMoveFound(move nextroute.SolutionMove) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnBestMoveFound(move)
	}
}

func (s *solutionObservedImpl) OnNewSolution(model nextroute.Model) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnNewSolution(model)
	}
}

func (s *solutionObservedImpl) OnNewSolutionCreated(solution nextroute.Solution) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnNewSolutionCreated(solution)
	}
}

func (s *solutionObservedImpl) OnCopySolution(solution nextroute.Solution) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnCopySolution(solution)
	}
}

func (s *solutionObservedImpl) OnCopiedSolution(solution nextroute.Solution) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnCopiedSolution(solution)
	}
}

func (s *solutionObservedImpl) OnCheckConstraint(
	constraint nextroute.ModelConstraint,
	checkViolationAt nextroute.CheckedAt,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnCheckConstraint(constraint, checkViolationAt)
	}
}

func (s *solutionObservedImpl) OnStopConstraintChecked(
	stop nextroute.SolutionStop,
	constraint nextroute.ModelConstraint,
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
	vehicle nextroute.SolutionVehicle,
	constraint nextroute.ModelConstraint,
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
	constraint nextroute.ModelConstraint,
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
	constraint nextroute.ModelConstraint,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnEstimateIsViolated(constraint)
	}
}

func (s *solutionObservedImpl) OnEstimatedIsViolated(
	move nextroute.SolutionMove,
	constraint nextroute.ModelConstraint,
	isViolated bool,
	planPositionsHint nextroute.StopPositionsHint,
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

func (s *solutionObservedImpl) OnPlan(move nextroute.SolutionMove) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnPlan(move)
	}
}

func (s *solutionObservedImpl) OnPlanFailed(
	move nextroute.SolutionMove,
	constraint nextroute.ModelConstraint,
) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnPlanFailed(move, constraint)
	}
}

func (s *solutionObservedImpl) OnPlanSucceeded(move nextroute.SolutionMove) {
	if len(s.observers) == 0 {
		return
	}
	for _, observer := range s.observers {
		observer.OnPlanSucceeded(move)
	}
}
