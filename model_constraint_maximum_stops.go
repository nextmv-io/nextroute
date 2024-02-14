package nextroute

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

	beforeStop := stopPositions[len(stopPositions)-1].next()
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
