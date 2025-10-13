// © 2019-present nextmv.io inc

package nextroute

// VehiclesObjective is an objective that uses the number of vehicles as an
// objective. Each vehicle that is not empty is scored by the given expression.
// A vehicle is empty if it has no stops assigned to it (except for the first
// and last visit).
type VehiclesObjective struct {
	expression VehicleTypeExpression
}

// NewVehiclesObjective returns a new VehiclesObjective.
func NewVehiclesObjective(
	expression VehicleTypeExpression,
) *VehiclesObjective {
	return &VehiclesObjective{
		expression: expression,
	}
}

// ModelExpressions implements the RegisteredModelExpressions interface.
func (t *VehiclesObjective) ModelExpressions() ModelExpressions {
	return ModelExpressions{}
}

// EstimateDeltaValue implements the ModelObjective interface.
func (t *VehiclesObjective) EstimateDeltaValue(move SolutionMoveStops) float64 {
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

// Value implements the ModelObjective interface.
func (t *VehiclesObjective) Value(solution *Solution) float64 {
	vehicleCost := 0.0
	for _, vehicle := range solution.Vehicles() {
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

// String returns the string representation of the objective.
func (t *VehiclesObjective) String() string {
	return "vehicle_activation_penalty"
}

// ActivationPenalty returns the activation penalty expression.
func (t *VehiclesObjective) ActivationPenalty() VehicleTypeExpression {
	return t.expression
}
