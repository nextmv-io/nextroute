package nextroute

import (
	"github.com/nextmv-io/sdk/nextroute"
)

// NewVehiclesObjective returns a new VehiclesObjective.
func NewVehiclesObjective(
	expression nextroute.VehicleTypeExpression,
) nextroute.VehiclesObjective {
	return &vehiclesObjectiveImpl{
		expression: expression,
	}
}

type vehiclesObjectiveImpl struct {
	expression nextroute.VehicleTypeExpression
}

func (t *vehiclesObjectiveImpl) ModelExpressions() nextroute.ModelExpressions {
	return nextroute.ModelExpressions{}
}

func (t *vehiclesObjectiveImpl) EstimateDeltaValue(move nextroute.SolutionMoveStops) float64 {
	vehicle := move.(*solutionMoveStopsImpl).vehicle()

	if vehicle.NumberOfStops() == 0 {
		return t.expression.Value(
			vehicle.ModelVehicle().VehicleType(),
			nil,
			nil,
		)
	}

	return 0.0
}

func (t *vehiclesObjectiveImpl) Value(solution nextroute.Solution) float64 {
	vehicleCost := 0.0
	for _, vehicle := range solution.(*solutionImpl).vehiclesMutable() {
		if vehicle.NumberOfStops() > 0 {
			vehicleCost += t.expression.Value(
				vehicle.ModelVehicle().VehicleType(),
				nil,
				nil,
			)
		}
	}
	return vehicleCost
}

func (t *vehiclesObjectiveImpl) String() string {
	return "vehicle_activation_penalty"
}

func (t *vehiclesObjectiveImpl) ActivationPenalty() nextroute.VehicleTypeExpression {
	return t.expression
}
