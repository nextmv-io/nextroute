// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"

	"github.com/nextmv-io/nextroute/common"
)

// SolveOperatorAnd is a solve-operator which executes a set of solve-operators
// in each iteration.
type SolveOperatorAnd interface {
	SolveOperator

	// Operators returns the solve-operators that will be executed in each
	// iteration.
	Operators() SolveOperators
}

// NewSolverOperatorAnd creates a new solve-and-operator.
func NewSolverOperatorAnd(
	probability float64,
	operators SolveOperators,
) (SolveOperatorAnd, error) {
	if probability < 0 || probability > 1 {
		return nil,
			fmt.Errorf(
				"the probability must be between 0 and 1",
			)
	}
	operators = common.Filter(operators, func(operator SolveOperator) bool {
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
				func(operator SolveOperator) bool {
					return operator.CanResultInImprovement()
				},
			),
			common.MapSlice(
				operators,
				func(operator SolveOperator) []SolveParameter {
					return operator.Parameters()
				},
			),
		),
		operators: operators,
	}, nil
}

// solveOperatorAndImpl is the implementation of the SolveOperatorAnd interface.
type solveOperatorAndImpl struct {
	operators SolveOperators
	SolveOperator
}

func (s *solveOperatorAndImpl) Execute(
	ctx context.Context,
	runTimeInformation SolveInformation,
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

func (s *solveOperatorAndImpl) Parameters() SolveParameters {
	return common.MapSlice(
		s.operators,
		func(operator SolveOperator) []SolveParameter {
			return operator.Parameters()
		},
	)
}

func (s *solveOperatorAndImpl) Operators() SolveOperators {
	return s.operators
}

func (s *solveOperatorAndImpl) OnStartSolve(solveInformation SolveInformation) {
	for _, operator := range s.operators {
		if interested, ok := operator.(InterestedInStartSolve); ok {
			interested.OnStartSolve(solveInformation)
		}
	}
}

func (s *solveOperatorAndImpl) OnBetterSolution(solveInformation SolveInformation) {
	for _, operator := range s.operators {
		if interested, ok := operator.(InterestedInBetterSolution); ok {
			interested.OnBetterSolution(solveInformation)
		}
	}
}
