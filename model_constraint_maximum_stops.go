// © 2019-present nextmv.io inc

package nextroute

// MaximumStopsConstraint is a constraint that limits the maximum number of
// stops a vehicle type can have. The maximum number of stops is defined by
// the maximum stops expression. The first stop of a vehicle is not counted
// as a stop and the last stop of a vehicle is not counted as a stop.
type MaximumStopsConstraint interface {
	ModelConstraint

	// MaximumStops returns the maximum stops expression which defines the
	// maximum number of stops a vehicle type can have.
	MaximumStops() VehicleTypeExpression
}

// NewMaximumStopsConstraint returns a new MaximumStopsConstraint.
func NewMaximumStopsConstraint(
	maximumStops VehicleTypeExpression,
) (MaximumStopsConstraint, error) {
	return &maximumStopsConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"maximum_stops",
			ModelExpressions{},
		),
		maximumStops: maximumStops,
	}, nil
}

type maximumStopsConstraintImpl struct {
	maximumStops              VehicleTypeExpression
	maximumStopsByVehicleType []float64
	modelConstraintImpl
}

func (l *maximumStopsConstraintImpl) Lock(model Model) error {
	vehicleTypes := model.VehicleTypes()
	l.maximumStopsByVehicleType = make([]float64, len(vehicleTypes))
	for _, vehicleType := range vehicleTypes {
		l.maximumStopsByVehicleType[vehicleType.Index()] = l.maximumStops.Value(
			vehicleType,
			nil,
			nil,
		)
	}
	return nil
}

func (l *maximumStopsConstraintImpl) String() string {
	return l.name
}

func (l *maximumStopsConstraintImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	moveImpl := move.(*solutionMoveStopsImpl)
	stopPositions := moveImpl.stopPositions
	nrStopsToBeAddedToSolution := len(stopPositions)

	beforeStop := stopPositions[len(stopPositions)-1].Next()
	vehicle := beforeStop.vehicle()

	vehicleType := vehicle.ModelVehicle().VehicleType().Index()
	maximumStops := l.maximumStopsByVehicleType[vehicleType]

	if float64(vehicle.NumberOfStops()+nrStopsToBeAddedToSolution) >
		maximumStops {
		return true, constSkipVehiclePositionsHint
	}

	return false, constNoPositionsHint
}

func (l *maximumStopsConstraintImpl) EstimationCost() Cost {
	return Constant
}

func (l *maximumStopsConstraintImpl) MaximumStops() VehicleTypeExpression {
	return l.maximumStops
}
