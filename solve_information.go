// © 2019-present nextmv.io inc

package nextroute

import (
	"time"
)

// SolveInformation contains information about the current solve.
type SolveInformation interface {
	// DeltaScore returns the delta score of the last executed solve operator.
	DeltaScore() float64

	// Iteration returns the current iteration.
	Iteration() int

	// Solver returns the solver.
	Solver() Solver
	// SolveOperators returns the solve-operators that has been executed in
	// the current iteration.
	SolveOperators() SolveOperators
	// Start returns the start time of the solver.
	Start() time.Time
}

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
