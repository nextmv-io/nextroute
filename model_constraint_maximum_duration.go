// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// NewMaximumDurationConstraint returns a new MaximumDurationConstraint.
func NewMaximumDurationConstraint(
	maximum VehicleTypeDurationExpression,
) (*MaximumDurationConstraint, error) {
	return &MaximumDurationConstraint{
		maximum: maximum,
	}, nil
}

// MaximumDurationConstraint is a constraint that limits the
// duration of a vehicle.
type MaximumDurationConstraint struct {
	maximum VehicleTypeDurationExpression
}

// String returns a string representation of the MaximumDurationConstraint.
func (l *MaximumDurationConstraint) String() string {
	return fmt.Sprintf("MaximumDuration maximum_duration', maxima: %v",
		l.maximum,
	)
}

// EstimationCost returns the cost of the constraint for estimation purposes.
func (l *MaximumDurationConstraint) EstimationCost() Cost {
	return Constant
}

// Maximum returns the maximum expression which defines the maximum
// duration of a vehicle type.
func (l *MaximumDurationConstraint) Maximum() VehicleTypeDurationExpression {
	return l.maximum
}

// EstimateIsViolated estimates whether the given move violates the constraint.
func (l *MaximumDurationConstraint) EstimateIsViolated(
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
			return true, NoPositionsHint()
		}

		previousStop = solutionStop
	}

	deltaEnd := endValue - previousStop.EndValue() - previousStop.SlackValue()

	if vehicle.DurationValue()+deltaEnd > maximumValue {
		return true, NoPositionsHint()
	}

	return false, NoPositionsHint()
}

// DoesVehicleHaveViolations returns whether the given vehicle has violations
// of the constraint.
func (l *MaximumDurationConstraint) DoesVehicleHaveViolations(vehicle SolutionVehicle) bool {
	return vehicle.DurationValue() >
		l.maximum.Value(vehicle.ModelVehicle().VehicleType(), nil, nil)
}

// IsTemporal returns whether the constraint is temporal.
func (l *MaximumDurationConstraint) IsTemporal() bool {
	return true
}
