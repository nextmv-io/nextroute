// © 2019-present nextmv.io inc

package nextroute

// MaximumTravelDurationConstraint is a constraint that limits the
// total travel duration of a vehicle.
type MaximumTravelDurationConstraint struct {
	maximum VehicleTypeDurationExpression
}

// NewMaximumTravelDurationConstraint returns a new
// MaximumTravelDurationConstraint.
func NewMaximumTravelDurationConstraint(
	maximum VehicleTypeDurationExpression,
) (*MaximumTravelDurationConstraint, error) {
	return &MaximumTravelDurationConstraint{
		maximum: maximum,
	}, nil
}

// String returns the string representation of the constraint.
func (l *MaximumTravelDurationConstraint) String() string {
	return "maximum_travel_duration"
}

// EstimationCost returns the cost of the constraint for estimation purposes.
func (l *MaximumTravelDurationConstraint) EstimationCost() Cost {
	return Constant
}

// Maximum returns the maximum expression which defines the maximum
// travel duration of a vehicle type.
func (l *MaximumTravelDurationConstraint) Maximum() VehicleTypeDurationExpression {
	return l.maximum
}

// EstimateIsViolated estimates whether the given move violates the constraint.
func (l *MaximumTravelDurationConstraint) EstimateIsViolated(
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

// DoesVehicleHaveViolations returns true if the vehicle has violations of
// the constraint.
func (l *MaximumTravelDurationConstraint) DoesVehicleHaveViolations(vehicle SolutionVehicle) bool {
	return vehicle.Last().CumulativeTravelDurationValue() >
		l.maximum.Value(vehicle.ModelVehicle().VehicleType(), nil, nil)
}

// IsTemporal returns true if the constraint is temporal.
func (l *MaximumTravelDurationConstraint) IsTemporal() bool {
	return true
}
