// © 2019-present nextmv.io inc

package nextroute

// TravelDurationObjective is an objective that uses the travel duration as an
// objective.
type TravelDurationObjective struct{}

// NewTravelDurationObjective returns a new TravelDurationObjective.
func NewTravelDurationObjective() *TravelDurationObjective {
	return &TravelDurationObjective{}
}

// ModelExpressions implements the RegisteredModelExpressions interface.
func (t *TravelDurationObjective) ModelExpressions() ModelExpressions {
	return ModelExpressions{}
}

// EstimateDeltaValue implements the ModelObjective interface.
func (t *TravelDurationObjective) EstimateDeltaValue(move SolutionMoveStops) float64 {
	return move.(*solutionMoveStopsImpl).deltaTravelDurationValue()
}

// Value implements the ModelObjective interface.
func (t *TravelDurationObjective) Value(solution *Solution) float64 {
	score := 0.0
	for _, vehicle := range solution.vehicles {
		score += vehicle.Last().CumulativeTravelDurationValue()
	}
	return score
}

// String returns the string representation of the objective.
func (t *TravelDurationObjective) String() string {
	return "travel_duration"
}
