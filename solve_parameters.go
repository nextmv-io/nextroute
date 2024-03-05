// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// SolveParameter is an interface for a parameter that can change
// during the solving. The parameter can be used to control the
// behavior of the solver and it's operators.
type SolveParameter interface {
	// Update updates the parameter based on the given solve information.
	// Update is invoked after each iteration of the solver.
	Update(SolveInformation)

	// Value returns the current value of the parameter.
	Value() int
}

// SolveParameters is a slice of solve parameters.
type SolveParameters []SolveParameter

// NewConstSolveParameter creates a new constant solve parameter.
func NewConstSolveParameter(value int) SolveParameter {
	return &constParameterImpl{value: value}
}

// NewSolveParameter creates a new solve parameter.
//   - startValue is the initial value of the parameter.
//   - deltaAfterIterations is the number of iterations without an improvement
//     before the value is changed.
//   - delta is the initial change in value after deltaAfterIterations.
//   - minValue is the minimum value of the parameter.
//   - maxValue is the maximum value of the parameter.
//   - snapBackAfterImprovement is a flag that indicates if the value should
//     snap back to the start value after an improvement.
//   - zigzag is a flag that indicates if the value should zigzag between
//     the min and max value. If the value is at the min value and delta is
//     negative, the delta is changed to positive. If the value is at the
//     max value and delta is positive, the delta is changed to negative.
func NewSolveParameter(
	startValue int,
	deltaAfterIterations int,
	delta int,
	minValue int,
	maxValue int,
	snapBackAfterImprovement bool,
	zigzag bool,
) (SolveParameter, error) {
	if deltaAfterIterations < 0 {
		return nil,
			fmt.Errorf(
				"NewSolveParameter, deltaAfterIterations %v must be greater than 0",
				deltaAfterIterations,
			)
	}
	if startValue < minValue {
		return nil,
			fmt.Errorf(
				"NewSolveParameter, startValue %v must be greater than or equal minValue %v",
				startValue,
				minValue,
			)
	}
	if startValue > maxValue {
		return nil,
			fmt.Errorf(
				"NewSolveParameter, startValue %v must be smaller than or equal to maxValue %v",
				startValue,
				maxValue,
			)
	}

	if startValue == maxValue && delta < 0 {
		delta = -delta
	}
	if startValue == minValue && delta > 0 {
		delta = -delta
	}
	return &intParameterImpl{
		startValue:               startValue,
		startDelta:               delta,
		deltaAfterIterations:     deltaAfterIterations,
		delta:                    delta,
		minValue:                 minValue,
		maxValue:                 maxValue,
		value:                    startValue,
		snapBackAfterImprovement: snapBackAfterImprovement,
		zigzag:                   zigzag,
	}, nil
}

type intParameterImpl struct {
	startValue               int
	startDelta               int
	deltaAfterIterations     int
	delta                    int
	maxValue                 int
	minValue                 int
	value                    int
	snapBackAfterImprovement bool
	zigzag                   bool
	iterations               int
}

func (i *intParameterImpl) Value() int {
	return i.value
}

func (i *intParameterImpl) Update(solveInformation SolveInformation) {
	if solveInformation.DeltaScore() < 0.0 {
		i.iterations = 0
		if i.snapBackAfterImprovement && i.value != i.startValue {
			i.delta = i.startDelta
			i.value = i.startValue
		}
		return
	}
	i.iterations++
	if i.iterations > i.deltaAfterIterations {
		if i.value == i.maxValue || i.value == i.minValue {
			i.delta = -i.delta
		}

		i.iterations = 0
		i.value += i.delta
		if i.value > i.maxValue {
			i.value = i.maxValue
		}
		if i.value < i.minValue {
			i.value = i.minValue
		}
	}
}

type constParameterImpl struct {
	value int
}

func (c *constParameterImpl) Update(_ SolveInformation) {
}

func (c *constParameterImpl) Value() int {
	return c.value
}
