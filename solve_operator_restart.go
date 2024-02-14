package nextroute

import (
	"context"
)

// NewSolveOperatorRestart creates a new solve-operator that restarts the solver
// after a certain number of iterations without improvement.
// SolveOperatorRestart is a solve-operator that restarts the solver after a
// certain number of iterations without improvement. The restart is done by
// invoking the Restart method on the solver and replaces the current work
// solution with the best solution found so far.
func NewSolveOperatorRestart(
	maximumIterations SolveParameter,
) (SolveOperatorRestart, error) {
	return &solveOperatorRestartImpl{
		SolveOperator: NewSolveOperator(
			1.0,
			true,
			SolveParameters{maximumIterations},
		),
	}, nil
}

type solveOperatorRestartImpl struct {
	SolveOperator
	lastImprovement int
}

func (d *solveOperatorRestartImpl) MaximumIterations() SolveParameter {
	return d.Parameters()[0]
}

func (d *solveOperatorRestartImpl) OnStartSolve(_ SolveInformation) {
	d.lastImprovement = 0
}

func (d *solveOperatorRestartImpl) OnBetterSolution(
	solveRunInformation SolveInformation,
) {
	d.lastImprovement = solveRunInformation.Iteration()
}

func (d *solveOperatorRestartImpl) Execute(
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
