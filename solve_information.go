package nextroute

import (
	"time"
)

type solveInformationImpl struct {
	start          time.Time
	solver         Solver
	solveOperators SolveOperators
	deltaScore     float64
	iteration      int
}

func (s *solveInformationImpl) Iteration() int {
	return s.iteration
}

func (s *solveInformationImpl) Solver() Solver {
	return s.solver
}

func (s *solveInformationImpl) SolveOperators() SolveOperators {
	return s.solveOperators
}

func (s *solveInformationImpl) Start() time.Time {
	return s.start
}

func (s *solveInformationImpl) DeltaScore() float64 {
	return s.deltaScore
}
