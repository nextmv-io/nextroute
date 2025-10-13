// © 2019-present nextmv.io inc

package nextroute

// NewMinStopsObjective returns a new MinStopsObjective.
func NewMinStopsObjective(minStops, minStopsPenalty VehicleTypeExpression) *MinStopsObjective {
	return &MinStopsObjective{
		minStops:        minStops,
		minStopsPenalty: minStopsPenalty,
	}
}

// MinStopsObjective is an objective that tries to ensure that each vehicle has
// at least a minimum number of stops assigned to it.
type MinStopsObjective struct {
	minStops        VehicleTypeExpression
	minStopsPenalty VehicleTypeExpression
}

// EstimateDeltaValue implements the ModelObjective interface.
func (t *MinStopsObjective) EstimateDeltaValue(move SolutionMoveStops) float64 {
	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	modelVehicle := vehicle.ModelVehicle()
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

// Value implements the ModelObjective interface.
func (t *MinStopsObjective) Value(solution *Solution) float64 {
	penaltySum := 0.0
	for _, vehicle := range solution.vehicles {
		vehicleNumberOfStops := vehicle.NumberOfStops()
		if vehicleNumberOfStops == 0 {
			continue
		}
		modelVehicle := vehicle.ModelVehicle()
		minimum := int(t.minStops.ValueForVehicleType(modelVehicle.vehicleType))
		if vehicleNumberOfStops < minimum {
			penaltySum += t.minStopsPenalty.ValueForVehicleType(modelVehicle.vehicleType) *
				(float64(minimum) - float64(vehicleNumberOfStops)) *
				(float64(minimum) - float64(vehicleNumberOfStops))
		}
	}
	return penaltySum
}

// String returns the string representation of the objective.
func (t *MinStopsObjective) String() string {
	return "min_stops"
}
