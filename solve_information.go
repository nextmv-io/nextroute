package nextroute

import (
	"time"

	"github.com/nextmv-io/sdk/nextroute"
)

type solveInformationImpl struct {
	start          time.Time
	solver         nextroute.Solver
	solveOperators nextroute.SolveOperators
	deltaScore     float64
	iteration      int
}

func (s *solveInformationImpl) Iteration() int {
	return s.iteration
}

func (s *solveInformationImpl) Solver() nextroute.Solver {
	return s.solver
}

func (s *solveInformationImpl) SolveOperators() nextroute.SolveOperators {
	return s.solveOperators
}

func (s *solveInformationImpl) Start() time.Time {
	return s.start
}

func (s *solveInformationImpl) DeltaScore() float64 {
	return s.deltaScore
}
