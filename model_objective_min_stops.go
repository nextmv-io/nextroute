// © 2019-present nextmv.io inc

package nextroute

// NewMinStopsObjective returns a new MinStopsObjective.
func NewMinStopsObjective(minStops, minStopsPenalty VehicleTypeExpression) ModelObjective {
	return &minStopsObjectiveImpl{
		minStops:        minStops,
		minStopsPenalty: minStopsPenalty,
	}
}

type minStopsObjectiveImpl struct {
	minStops        VehicleTypeExpression
	minStopsPenalty VehicleTypeExpression
}

func (t *minStopsObjectiveImpl) EstimateDeltaValue(move SolutionMoveStops) float64 {
	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	modelVehicle := vehicle.ModelVehicle().(*modelVehicleImpl)
	minimum := int(t.minStops.ValueForVehicleType(modelVehicle.vehicleType))

	vehicleStops := vehicle.NumberOfStops()
	if vehicleStops >= minimum {
		return 0
	}

	moveStops := len(moveImpl.stopPositions)

	if vehicle.IsEmpty() {
		if moveStops >= minimum {
			return 0
		}
		return t.minStopsPenalty.ValueForVehicleType(modelVehicle.vehicleType) *
			(float64(minimum) - float64(moveStops)) *
			(float64(minimum) - float64(moveStops))
	}

	oldDelta := minimum - vehicleStops
	newDelta := minimum - vehicleStops - moveStops

	if newDelta >= 0 {
		return t.minStopsPenalty.ValueForVehicleType(modelVehicle.vehicleType) *
			(float64(newDelta)*float64(newDelta) - float64(oldDelta)*float64(oldDelta))
	}

	return t.minStopsPenalty.ValueForVehicleType(modelVehicle.vehicleType) *
		-float64(oldDelta) * float64(oldDelta)
}

func (t *minStopsObjectiveImpl) Value(solution Solution) float64 {
	solutionImpl := solution.(*solutionImpl)
	penaltySum := 0.0
	for _, vehicle := range solutionImpl.vehicles {
		vehicleNumberOfStops := vehicle.NumberOfStops()
		if vehicleNumberOfStops == 0 {
			continue
		}
		modelVehicle := vehicle.ModelVehicle().(*modelVehicleImpl)
		minimum := int(t.minStops.ValueForVehicleType(modelVehicle.vehicleType))
		if vehicleNumberOfStops < minimum {
			penaltySum += t.minStopsPenalty.ValueForVehicleType(modelVehicle.vehicleType) *
				(float64(minimum) - float64(vehicleNumberOfStops)) *
				(float64(minimum) - float64(vehicleNumberOfStops))
		}
	}
	return penaltySum
}

func (t *minStopsObjectiveImpl) String() string {
	return "min_stops"
}
