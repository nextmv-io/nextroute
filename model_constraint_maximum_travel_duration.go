// © 2019-present nextmv.io inc

package nextroute

// MaximumTravelDurationConstraint is a constraint that limits the
// total travel duration of a vehicle.
type MaximumTravelDurationConstraint interface {
	ModelConstraint

	// Maximum returns the maximum expression which defines the maximum
	// travel duration of a vehicle type.
	Maximum() VehicleTypeDurationExpression
}

// NewMaximumTravelDurationConstraint returns a new
// MaximumTravelDurationConstraint.
func NewMaximumTravelDurationConstraint(
	maximum VehicleTypeDurationExpression,
) (MaximumTravelDurationConstraint, error) {
	return &maximumTravelDurationConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"maximum_travel_duration",
			ModelExpressions{},
		),
		maximum: maximum,
	}, nil
}

type maximumTravelDurationConstraintImpl struct {
	maximum VehicleTypeDurationExpression
	modelConstraintImpl
}

func (l *maximumTravelDurationConstraintImpl) String() string {
	return l.name
}

func (l *maximumTravelDurationConstraintImpl) EstimationCost() Cost {
	return Constant
}

func (l *maximumTravelDurationConstraintImpl) Maximum() VehicleTypeDurationExpression {
	return l.maximum
}

func (l *maximumTravelDurationConstraintImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()
	isDependentOnTime := vehicleType.TravelDurationExpression().IsDependentOnTime()

	previous, _ := moveImpl.previous()
	cumulativeDurationAtStart := previous.CumulativeTravelDurationValue()
	maximum := l.maximum.Value(vehicleType, nil, nil)

	value := 0.0

	generator := newSolutionStopGenerator(
		*moveImpl,
		false,
		isDependentOnTime,
	)
	defer generator.release()
	previousStop, _ := generator.next()
	departure := previousStop.EndValue()

	for solutionStop, ok := generator.next(); ok; solutionStop, ok = generator.next() {
		travelDuration, _, _, end := vehicleType.TemporalValues(
			departure,
			previousStop.ModelStop(),
			solutionStop.ModelStop(),
		)

		value += travelDuration

		if value+cumulativeDurationAtStart > maximum {
			return true, constNoPositionsHint
		}

		previousStop = solutionStop
		departure = end
	}

	next, _ := moveImpl.next()
	delta := value - next.CumulativeTravelDurationValue()

	if vehicle.Last().CumulativeTravelDurationValue()+delta > maximum {
		return true, constNoPositionsHint
	}

	return false, constNoPositionsHint
}

func (l *maximumTravelDurationConstraintImpl) DoesVehicleHaveViolations(vehicle SolutionVehicle) bool {
	return vehicle.Last().CumulativeTravelDurationValue() >
		l.maximum.Value(vehicle.ModelVehicle().VehicleType(), nil, nil)
}

func (l *maximumTravelDurationConstraintImpl) IsTemporal() bool {
	return true
}
