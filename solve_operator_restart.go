package nextroute

import (
	"context"

	"github.com/nextmv-io/sdk/nextroute"
)

// NewSolveOperatorRestart creates a new solve-operator that restarts the solver
// after a certain number of iterations without improvement.
// SolveOperatorRestart is a solve-operator that restarts the solver after a
// certain number of iterations without improvement. The restart is done by
// invoking the Restart method on the solver and replaces the current work
// solution with the best solution found so far.
func NewSolveOperatorRestart(
	maximumIterations nextroute.SolveParameter,
) (nextroute.SolveOperatorRestart, error) {
	return &solveOperatorRestartImpl{
		SolveOperator: NewSolveOperator(
			1.0,
			true,
			nextroute.SolveParameters{maximumIterations},
		),
	}, nil
}

type solveOperatorRestartImpl struct {
	nextroute.SolveOperator
	lastImprovement int
}

func (d *solveOperatorRestartImpl) MaximumIterations() nextroute.SolveParameter {
	return d.Parameters()[0]
}

func (d *solveOperatorRestartImpl) OnStartSolve(_ nextroute.SolveInformation) {
	d.lastImprovement = 0
}

func (d *solveOperatorRestartImpl) OnBetterSolution(
	solveRunInformation nextroute.SolveInformation,
) {
	d.lastImprovement = solveRunInformation.Iteration()
}

func (d *solveOperatorRestartImpl) Execute(
	_ context.Context,
	solveRunInformation nextroute.SolveInformation,
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
