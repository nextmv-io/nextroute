// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
)

// SolveOperatorRestart is a solve operator that restarts the solver.
// The operator will set the working solution to the best solution found so far
// after MaximumIterations number of iterations without finding a better
// solution.
type SolveOperatorRestart struct {
	solveOperator
	lastImprovement int
}

// NewSolveOperatorRestart creates a new solve-operator that restarts the solver
// after a certain number of iterations without improvement.
// SolveOperatorRestart is a solve-operator that restarts the solver after a
// certain number of iterations without improvement. The restart is done by
// invoking the Restart method on the solver and replaces the current work
// solution with the best solution found so far.
func NewSolveOperatorRestart(
	maximumIterations SolveParameter,
) (*SolveOperatorRestart, error) {
	return &SolveOperatorRestart{
		solveOperator: solveOperator{
			probability:            1.0,
			canResultInImprovement: true,
			parameters:             SolveParameters{maximumIterations},
		},
	}, nil
}

// MaximumIterations returns the maximum iterations of the solve operator.
func (d *SolveOperatorRestart) MaximumIterations() SolveParameter {
	return d.Parameters()[0]
}

// OnStartSolve implements the InterestedInStartSolve interface.
func (d *SolveOperatorRestart) OnStartSolve(_ SolveInformation) {
	d.lastImprovement = 0
}

// OnBetterSolution implements the InterestedInBetterSolution interface.
func (d *SolveOperatorRestart) OnBetterSolution(
	solveRunInformation SolveInformation,
) {
	d.lastImprovement = solveRunInformation.Iteration()
}

// Execute implements the SolveOperator interface.
func (d *SolveOperatorRestart) Execute(
	_ context.Context,
	solveRunInformation SolveInformation,
) error {
	if solveRunInformation.Solver().WorkSolution().Score() == solveRunInformation.Solver().BestSolution().Score() {
		d.lastImprovement = solveRunInformation.Iteration()
	}
	if solveRunInformation.Iteration()-d.lastImprovement >
		d.MaximumIterations().Value() {
		solveRunInformation.Solver().Reset(solveRunInformation.Solver().BestSolution(), solveRunInformation)
		d.lastImprovement = solveRunInformation.Iteration()
	}
	return nil
}
