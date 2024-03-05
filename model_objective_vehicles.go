// © 2019-present nextmv.io inc

package nextroute

// VehiclesObjective is an objective that uses the number of vehicles as an
// objective. Each vehicle that is not empty is scored by the given expression.
// A vehicle is empty if it has no stops assigned to it (except for the first
// and last visit).
type VehiclesObjective interface {
	ModelObjective
	// ActivationPenalty returns the activation penalty expression.
	ActivationPenalty() VehicleTypeExpression
}

// NewVehiclesObjective returns a new VehiclesObjective.
func NewVehiclesObjective(
	expression VehicleTypeExpression,
) VehiclesObjective {
	return &vehiclesObjectiveImpl{
		expression: expression,
	}
}

type vehiclesObjectiveImpl struct {
	expression VehicleTypeExpression
}

func (t *vehiclesObjectiveImpl) ModelExpressions() ModelExpressions {
	return ModelExpressions{}
}

func (t *vehiclesObjectiveImpl) EstimateDeltaValue(move SolutionMoveStops) float64 {
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

func (t *vehiclesObjectiveImpl) Value(solution Solution) float64 {
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

func (t *vehiclesObjectiveImpl) ActivationPenalty() VehicleTypeExpression {
	return t.expression
}
