// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"

	"github.com/nextmv-io/nextroute/common"
)

// SolveOperatorAnd is a solve-operator which executes a set of solve-operators
// in each iteration.
type SolveOperatorAnd struct {
	operators SolveOperators
	solveOperator
}

// NewSolverOperatorAnd creates a new solve-and-operator.
func NewSolverOperatorAnd(
	probability float64,
	operators SolveOperators,
) (*SolveOperatorAnd, error) {
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
	return &SolveOperatorAnd{
		solveOperator: solveOperator{
			probability: probability,
			canResultInImprovement: common.Has(operators,
				true,
				func(operator SolveOperator) bool {
					return operator.CanResultInImprovement()
				},
			),
			parameters: common.MapSlice(
				operators,
				func(operator SolveOperator) []SolveParameter {
					return operator.Parameters()
				},
			),
		},
		operators: operators,
	}, nil
}

// Execute implements the SolveOperator interface.
func (s *SolveOperatorAnd) Execute(
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

// Parameters implements the SolveOperator interface.
func (s *SolveOperatorAnd) Parameters() SolveParameters {
	return common.MapSlice(
		s.operators,
		func(operator SolveOperator) []SolveParameter {
			return operator.Parameters()
		},
	)
}

// Operators implements the SolveOperator interface.
func (s *SolveOperatorAnd) Operators() SolveOperators {
	return s.operators
}

// OnStartSolve implements the InterestedInStartSolve interface.
func (s *SolveOperatorAnd) OnStartSolve(solveInformation SolveInformation) {
	for _, operator := range s.operators {
		if interested, ok := operator.(InterestedInStartSolve); ok {
			interested.OnStartSolve(solveInformation)
		}
	}
}

// OnBetterSolution implements the InterestedInBetterSolution interface.
func (s *SolveOperatorAnd) OnBetterSolution(solveInformation SolveInformation) {
	for _, operator := range s.operators {
		if interested, ok := operator.(InterestedInBetterSolution); ok {
			interested.OnBetterSolution(solveInformation)
		}
	}
}
