package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/nextroute"
)

// NewMaximumDurationConstraint returns a new MaximumDurationConstraint.
func NewMaximumDurationConstraint(
	maximum nextroute.VehicleTypeDurationExpression,
) (nextroute.MaximumDurationConstraint, error) {
	return &maximumDurationConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"maximum_duration",
			nextroute.ModelExpressions{},
		),
		maximum: maximum,
	}, nil
}

type maximumDurationConstraintImpl struct {
	maximum nextroute.VehicleTypeDurationExpression
	modelConstraintImpl
}

func (l *maximumDurationConstraintImpl) String() string {
	return fmt.Sprintf("MaximumDuration '%v', maxima: %v",
		l.name,
		l.maximum,
	)
}

func (l *maximumDurationConstraintImpl) EstimationCost() nextroute.Cost {
	return nextroute.Constant
}

func (l *maximumDurationConstraintImpl) Maximum() nextroute.VehicleTypeDurationExpression {
	return l.maximum
}

func (l *maximumDurationConstraintImpl) EstimateIsViolated(
	move nextroute.SolutionMoveStops,
) (isViolated bool, stopPositionsHint nextroute.StopPositionsHint) {
	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()

	dependentOnTime := vehicleType.TravelDurationExpression().IsDependentOnTime()

	maximumValue := l.maximum.Value(vehicleType, nil, nil)

	startValue := vehicle.first().StartValue()
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

func (l *maximumDurationConstraintImpl) DoesVehicleHaveViolations(vehicle nextroute.SolutionVehicle) bool {
	return vehicle.DurationValue() >
		l.maximum.Value(vehicle.ModelVehicle().VehicleType(), nil, nil)
}

func (l *maximumDurationConstraintImpl) IsTemporal() bool {
	return true
}
