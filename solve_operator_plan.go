// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
)

// SolveOperatorPlan is a solve-operator that tries to plan all unplanned
// plan-units in each iteration. The group-size is a solve-parameter which
// can be configured by the user. Group-size determines how many random
// plan-units are selected for which the best move is determined. The best
// move is then executed. In one iteration of the solve run, the operator will
// continue to select a random group-size number of unplanned plan-units
// and execute the best move until all unplanned plan-units are planned or
// no more moves can be executed. In an unconstrained model all plan-units
// will be planned after one iteration of this operator.
type SolveOperatorPlan interface {
	SolveOperator

	// GroupSize returns the group size of the solve operator.
	GroupSize() SolveParameter
}

// NewSolveOperatorPlan creates a new solve operator for nextroute that
// plans units.
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
