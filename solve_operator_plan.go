package nextroute

import (
	"context"

	"github.com/nextmv-io/sdk/nextroute"
)

// NewSolveOperatorPlan creates a new SolveOperatorPlan.
func NewSolveOperatorPlan(
	groupSize nextroute.SolveParameter,
) (nextroute.SolveOperatorPlan, error) {
	return &solveOperatorPlanImpl{
		SolveOperator: NewSolveOperator(
			1.0,
			true,
			nextroute.SolveParameters{groupSize},
		),
	}, nil
}

type solveOperatorPlanImpl struct {
	nextroute.SolveOperator
}

func (d *solveOperatorPlanImpl) GroupSize() nextroute.SolveParameter {
	return d.Parameters()[0]
}

func (d *solveOperatorPlanImpl) Execute(
	ctx context.Context,
	runTimeInformation nextroute.SolveInformation,
) error {
	workSolution := runTimeInformation.
		Solver().
		WorkSolution()

	unplannedPlanUnits := NewSolutionPlanUnitCollection(
		workSolution.Random(),
		workSolution.UnPlannedPlanUnits().SolutionPlanUnits(),
	)

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		default:
			if unplannedPlanUnits.Size() == 0 {
				break Loop
			}

			planUnits := unplannedPlanUnits.RandomDraw(
				d.GroupSize().Value(),
			)

			move := NewNotExecutableMove()

			for _, planUnit := range planUnits {
				planUnitMove := workSolution.BestMove(ctx, planUnit)

				if !planUnitMove.IsExecutable() {
					unplannedPlanUnits.Remove(planUnit)
				} else {
					move = move.TakeBest(planUnitMove)
				}
			}

			if move.IsExecutable() {
				if move.Value() <= 0 {
					_, err := move.Execute(ctx)
					if err != nil {
						return err
					}
				}
				unplannedPlanUnits.Remove(move.PlanUnit())
			}
		}
	}
	return nil
}
