// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// MaximumDurationConstraint is a constraint that limits the
// duration of a vehicle.
type MaximumDurationConstraint interface {
	ModelConstraint

	// Maximum returns the maximum expression which defines the maximum
	// duration of a vehicle type.
	Maximum() VehicleTypeDurationExpression
}

// NewMaximumDurationConstraint returns a new MaximumDurationConstraint.
func NewMaximumDurationConstraint(
	maximum VehicleTypeDurationExpression,
) (MaximumDurationConstraint, error) {
	return &maximumDurationConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"maximum_duration",
			ModelExpressions{},
		),
		maximum: maximum,
	}, nil
}

type maximumDurationConstraintImpl struct {
	maximum VehicleTypeDurationExpression
	modelConstraintImpl
}

func (l *maximumDurationConstraintImpl) String() string {
	return fmt.Sprintf("MaximumDuration '%v', maxima: %v",
		l.name,
		l.maximum,
	)
}

func (l *maximumDurationConstraintImpl) EstimationCost() Cost {
	return Constant
}

func (l *maximumDurationConstraintImpl) Maximum() VehicleTypeDurationExpression {
	return l.maximum
}

func (l *maximumDurationConstraintImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()

	dependentOnTime := vehicleType.TravelDurationExpression().IsDependentOnTime()

	maximumValue := l.maximum.Value(vehicleType, nil, nil)

	startValue := vehicle.First().StartValue()
	previous, _ := moveImpl.previous()
	endValue := previous.EndValue()

	generator := newSolutionStopGenerator(
		*moveImpl,
		false,
		dependentOnTime,
	)
	defer generator.release()

	previousStop, _ := generator.next()

	for solutionStop, ok := generator.next(); ok; solutionStop, ok = generator.next() {
		_, _, _, endValue = vehicleType.TemporalValues(
			endValue,
			previousStop.ModelStop(),
			solutionStop.ModelStop(),
		)

		if endValue-startValue > maximumValue {
			return true, constNoPositionsHint
		}

		previousStop = solutionStop
	}

	deltaEnd := endValue - previousStop.EndValue() - previousStop.SlackValue()

	if vehicle.DurationValue()+deltaEnd > maximumValue {
		return true, constNoPositionsHint
	}

	return false, constNoPositionsHint
}

func (l *maximumDurationConstraintImpl) DoesVehicleHaveViolations(vehicle SolutionVehicle) bool {
	return vehicle.DurationValue() >
		l.maximum.Value(vehicle.ModelVehicle().VehicleType(), nil, nil)
}

func (l *maximumDurationConstraintImpl) IsTemporal() bool {
	return true
}
