package nextroute

import (
	"context"
	"fmt"

	"github.com/nextmv-io/sdk/common"
)

// NewSolverOperatorOr creates a new solve-or-operator. The probability must be
// between 0 and 1. The number of operators with probability larger than zero
// must be greater than 0.
func NewSolverOperatorOr(
	probability float64,
	operators SolveOperators,
) (SolveOperatorOr, error) {
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
	weights := common.Map(operators, func(operator SolveOperator) float64 {
		return operator.Probability()
	})
	alias, err := common.NewAlias(weights)
	if err != nil {
		return nil, err
	}
	return &solveOperatorOrImpl{
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
		alias:     alias,
	}, nil
}

// SolveOperatorOrImpl is the implementation of the SolveOperatorOr interface.
type solveOperatorOrImpl struct {
	SolveOperator
	alias     common.Alias
	operators SolveOperators
}

func (s *solveOperatorOrImpl) Execute(
	ctx context.Context,
	runTimeInformation SolveInformation,
) error {
	return s.operators[s.alias.Sample(runTimeInformation.Solver().Random())].Execute(ctx, runTimeInformation)
}

func (s *solveOperatorOrImpl) Parameters() SolveParameters {
	return common.MapSlice(
		s.operators,
		func(operator SolveOperator) []SolveParameter {
			return operator.Parameters()
		},
	)
}

func (s *solveOperatorOrImpl) Operators() SolveOperators {
	return s.operators
}

func (s *solveOperatorOrImpl) OnStartSolve(solveInformation SolveInformation) {
	for _, operator := range s.operators {
		if interested, ok := operator.(InterestedInStartSolve); ok {
			interested.OnStartSolve(solveInformation)
		}
	}
}

func (s *solveOperatorOrImpl) OnBetterSolution(solveInformation SolveInformation) {
	for _, operator := range s.operators {
		if interested, ok := operator.(InterestedInBetterSolution); ok {
			interested.OnBetterSolution(solveInformation)
		}
	}
}
