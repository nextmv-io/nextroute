// © 2019-present nextmv.io inc

package nextroute

// ParallelSolveEvents is a struct that contains events that are fired during a
// solve invocation of the parallel solver.
type ParallelSolveEvents struct {
	// End is fired when the parallel solver is done. The first payload is the
	// solver, the second payload is the number of iterations, the third payload
	// is the best solution found.
	End *BaseEvent3[ParallelSolver, int, Solution]

	// NewSolution is fired when a new solution is found.
	NewSolution *BaseEvent2[ParallelSolveInformation, Solution]

	// Start is fired when the parallel solver is started. The first payload is
	// the parallel solver, the second payload is the options, the third payload
	// is the number of parallel runs will be invoked.
	Start *BaseEvent3[ParallelSolver, ParallelSolveOptions, int]
	// StartSolver is fired when one of the solver that will run in parallel is
	// started. The first payload is the parallel solve information, the second
	// payload is the solver, the third payload is the solve options, the fourth
	// payload is the start solution.
	StartSolver *BaseEvent4[ParallelSolveInformation, Solver, SolveOptions, Solution]
}

// NewParallelSolveEvents creates a new instance of ParallelSolveEvents.
func NewParallelSolveEvents() ParallelSolveEvents {
	return ParallelSolveEvents{
		End:         &BaseEvent3[ParallelSolver, int, Solution]{},
		NewSolution: &BaseEvent2[ParallelSolveInformation, Solution]{},
		Start:       &BaseEvent3[ParallelSolver, ParallelSolveOptions, int]{},
		StartSolver: &BaseEvent4[ParallelSolveInformation, Solver, SolveOptions, Solution]{},
	}
}
