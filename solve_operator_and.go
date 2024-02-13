package nextroute

import (
	"context"
	"fmt"

	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

// NewSolverOperatorAnd creates a new solve-and-operator.
func NewSolverOperatorAnd(
	probability float64,
	operators nextroute.SolveOperators,
) (nextroute.SolveOperatorAnd, error) {
	if probability < 0 || probability > 1 {
		return nil,
			fmt.Errorf(
				"the probability must be between 0 and 1",
			)
	}
	operators = common.Filter(operators, func(operator nextroute.SolveOperator) bool {
		return operator.Probability() > 0
	})
	if len(operators) == 0 {
		return nil,
			fmt.Errorf(
				"the number of operators with probability larger than" +
					" zero must be greater than 0",
			)
	}
	return &solveOperatorAndImpl{
		SolveOperator: NewSolveOperator(
			probability,
			common.Has(operators,
				true,
				func(operator nextroute.SolveOperator) bool {
					return operator.CanResultInImprovement()
				},
			),
			common.MapSlice(
				operators,
				func(operator nextroute.SolveOperator) []nextroute.SolveParameter {
					return operator.Parameters()
				},
			),
		),
		operators: operators,
	}, nil
}

// solveOperatorAndImpl is the implementation of the SolveOperatorAnd interface.
type solveOperatorAndImpl struct {
	operators nextroute.SolveOperators
	nextroute.SolveOperator
}

func (s *solveOperatorAndImpl) Execute(
	ctx context.Context,
	runTimeInformation nextroute.SolveInformation,
) error {
	random := runTimeInformation.Solver().Random()
Loop:
	for _, operator := range s.operators {
		select {
		case <-ctx.Done():
			break Loop
		default:
			if random.Float64() < s.Probability() {
				err := operator.Execute(ctx, runTimeInformation)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *solveOperatorAndImpl) Parameters() nextroute.SolveParameters {
	return common.MapSlice(
		s.operators,
		func(operator nextroute.SolveOperator) []nextroute.SolveParameter {
			return operator.Parameters()
		},
	)
}

func (s *solveOperatorAndImpl) Operators() nextroute.SolveOperators {
	return s.operators
}

func (s *solveOperatorAndImpl) OnStartSolve(solveInformation nextroute.SolveInformation) {
	for _, operator := range s.operators {
		if interested, ok := operator.(nextroute.InterestedInStartSolve); ok {
			interested.OnStartSolve(solveInformation)
		}
	}
}

func (s *solveOperatorAndImpl) OnBetterSolution(solveInformation nextroute.SolveInformation) {
	for _, operator := range s.operators {
		if interested, ok := operator.(nextroute.InterestedInBetterSolution); ok {
			interested.OnBetterSolution(solveInformation)
		}
	}
}
