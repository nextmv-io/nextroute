package nextroute

import (
	"context"
	"fmt"

	"github.com/nextmv-io/sdk/nextroute"
)

// NewSolveOperator returns a new solve operator.
func NewSolveOperator(
	probability float64,
	canResultInImprovement bool,
	parameters nextroute.SolveParameters,
) nextroute.SolveOperator {
	return &solveOperatorImpl{
		probability:            probability,
		canResultInImprovement: canResultInImprovement,
		parameters:             parameters,
	}
}

type solveOperatorImpl struct {
	parameters             nextroute.SolveParameters
	probability            float64
	canResultInImprovement bool
}

func (s *solveOperatorImpl) Execute(
	_ context.Context,
	_ nextroute.SolveInformation,
) error {
	panic("implement me")
}

func (s *solveOperatorImpl) Parameters() nextroute.SolveParameters {
	return s.parameters
}

func (s *solveOperatorImpl) Probability() float64 {
	return s.probability
}

func (s *solveOperatorImpl) SetProbability(
	probability float64,
) error {
	if probability < 0 || probability > 1 {
		return fmt.Errorf(
			"the probability must be between 0 and 1",
		)
	}
	s.probability = probability
	return nil
}

func (s *solveOperatorImpl) CanResultInImprovement() bool {
	return s.canResultInImprovement
}
