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

// NewSolveOperator returns a new solve operator that cannot be executed.
func NewSolveOperator(
	probability float64,
	canResultInImprovement bool,
	parameters SolveParameters,
) SolveOperator {
	return solveOperator{
		probability:            probability,
		canResultInImprovement: canResultInImprovement,
		parameters:             parameters,
	}
}

type solveOperator struct {
	parameters             SolveParameters
	probability            float64
	canResultInImprovement bool
}

// Execute implements the SolveOperator interface.
func (s solveOperator) Execute(
	_ context.Context,
	_ SolveInformation,
) error {
	panic("The generic SolveOperator cannot be executed. " +
		"Please use a specific implementation.")
}

// Parameters implements the SolveOperator interface.
func (s solveOperator) Parameters() SolveParameters {
	return s.parameters
}

// Probability implements the SolveOperator interface.
func (s solveOperator) Probability() float64 {
	return s.probability
}

// CanResultInImprovement implements the SolveOperator interface.
func (s solveOperator) CanResultInImprovement() bool {
	return s.canResultInImprovement
}

// SetProbability sets the probability of the solve-operator.
func (s *solveOperator) SetProbability(
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
