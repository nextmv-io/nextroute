// © 2019-present nextmv.io inc

package nextroute

// MaximumStopsConstraint is a constraint that limits the maximum number of
// stops a vehicle type can have. The maximum number of stops is defined by
// the maximum stops expression. The first stop of a vehicle is not counted
// as a stop and the last stop of a vehicle is not counted as a stop.
type MaximumStopsConstraint struct {
	maximumStops              VehicleTypeExpression
	maximumStopsByVehicleType []int
}

// NewMaximumStopsConstraint returns a new MaximumStopsConstraint.
func NewMaximumStopsConstraint(
	maximumStops VehicleTypeExpression,
) (*MaximumStopsConstraint, error) {
	return &MaximumStopsConstraint{
		maximumStops: maximumStops,
	}, nil
}

// Lock locks the constraint to the model. It precomputes the maximum stops
// for each vehicle type.
func (l *MaximumStopsConstraint) Lock(model *Model) error {
	vehicleTypes := model.VehicleTypes()
	l.maximumStopsByVehicleType = make([]int, len(vehicleTypes))
	for _, vehicleType := range vehicleTypes {
		l.maximumStopsByVehicleType[vehicleType.Index()] = int(l.maximumStops.Value(
			vehicleType,
			nil,
			nil,
		))
	}
	return nil
}

// String returns the string representation of the constraint.
func (l *MaximumStopsConstraint) String() string {
	return "maximum_stops"
}

// EstimateIsViolated estimates whether the given move violates the constraint.
func (l *MaximumStopsConstraint) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	moveImpl := move.(*solutionMoveStopsImpl)
	stopPositions := moveImpl.stopPositions
	nrStopsToBeAddedToSolution := len(stopPositions)

	beforeStop := stopPositions[len(stopPositions)-1].Next()
	vehicle := beforeStop.vehicle()

	vehicleType := vehicle.ModelVehicle().VehicleType().Index()
	maximumStops := l.maximumStopsByVehicleType[vehicleType]

	violated := vehicle.NumberOfStops()+nrStopsToBeAddedToSolution > maximumStops
	return violated, skipVehicleIfViolated(violated)
}

// EstimationCost returns the cost of the constraint for estimation purposes.
func (l *MaximumStopsConstraint) EstimationCost() Cost {
	return Constant
}

// MaximumStops returns the maximum stops expression which defines the
// maximum number of stops a vehicle type can have.
func (l *MaximumStopsConstraint) MaximumStops() VehicleTypeExpression {
	return l.maximumStops
}
