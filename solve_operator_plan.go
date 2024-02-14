package nextroute

import (
	"context"
)

// NewSolveOperatorPlan creates a new SolveOperatorPlan.
func NewSolveOperatorPlan(
	groupSize SolveParameter,
) (SolveOperatorPlan, error) {
	return &solveOperatorPlanImpl{
		SolveOperator: NewSolveOperator(
			1.0,
			true,
			SolveParameters{groupSize},
		),
	}, nil
}

type solveOperatorPlanImpl struct {
	SolveOperator
}

func (d *solveOperatorPlanImpl) GroupSize() SolveParameter {
	return d.Parameters()[0]
}

func (d *solveOperatorPlanImpl) Execute(
	ctx context.Context,
	runTimeInformation SolveInformation,
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
