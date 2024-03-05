// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"
)

// SolveOperator is a solve-operator. A solve-operator is a function that
// modifies the current solution. The function is executed with a certain
// probability. The probability is set by the SetProbability method. The
// probability is used by the solver to determine if the solve-operator should
// be executed. The probability is a value between 0 and 1. The manipulation of
// the solution is implemented in the Execute method. The Execute method will be
// invoked by the solver. The Execute method receives a SolveInformation
// instance that contains information about the current solution and the
// solver. The Execute method should modify the current solution. The
// Execute method should not modify the SolveInformation instance. The Execute
// method should not modify the SolveOperator instance.
type SolveOperator interface {
	// CanResultInImprovement returns true if the solve-operator can result in
	// an improvement compared to the best solution. The solver uses this
	// information to determine if the best solution should be replaced by
	// the new solution.
	CanResultInImprovement() bool

	// Execute executes the solve-operator. The Execute method is called by the
	// solver. The Execute method is passed a SolveInformation instance that
	// contains information about the current solution and the solver. The
	// Execute method should modify the current solution.
	Execute(context.Context, SolveInformation) error

	// Probability returns the probability of the solve-operator.
	// The probability is a value between 0 and 1. The solver uses the
	// probability to determine if the solve-operator should be executed in
	// the current iteration. Each iteration the solver will execute a
	// solve-operator with a probability equal to the probability of the
	// solve-operator. The probability is set by the SetProbability method.
	Probability() float64

	// SetProbability sets the probability of the solve-operator. Returns an
	// error if the probability is not in the range [0, 1].
	SetProbability(probability float64) error
	// Parameters returns the solve-parameters of the solve-operator.
	Parameters() SolveParameters
}

// InterestedInBetterSolution is an interface that can be implemented by
// solve-operators that are interested in being notified when a better solution
// is found. The solver will call the OnBetterSolution method when a better
// best-solution is found.
type InterestedInBetterSolution interface {
	OnBetterSolution(SolveInformation)
}

// InterestedInStartSolve is an interface that can be implemented by
// solve-operators that are interested in being notified when the solver starts
// solving. The solver will call the OnStartSolve method when the solver starts
// solving.
type InterestedInStartSolve interface {
	OnStartSolve(SolveInformation)
}

// SolveOperators is a slice of solve-operators.
type SolveOperators []SolveOperator

// NewSolveOperator returns a new solve operator.
func NewSolveOperator(
	probability float64,
	canResultInImprovement bool,
	parameters SolveParameters,
) SolveOperator {
	return &solveOperatorImpl{
		probability:            probability,
		canResultInImprovement: canResultInImprovement,
		parameters:             parameters,
	}
}

type solveOperatorImpl struct {
	parameters             SolveParameters
	probability            float64
	canResultInImprovement bool
}

func (s *solveOperatorImpl) Execute(
	_ context.Context,
	_ SolveInformation,
) error {
	panic("implement me")
}

func (s *solveOperatorImpl) Parameters() SolveParameters {
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
