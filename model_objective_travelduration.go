// © 2019-present nextmv.io inc

package nextroute

// TravelDurationObjective is an objective that uses the travel duration as an
// objective.
type TravelDurationObjective interface {
	ModelObjective
}

// NewTravelDurationObjective returns a new TravelDurationObjective.
func NewTravelDurationObjective() TravelDurationObjective {
	return &travelDurationObjectiveImpl{}
}

type travelDurationObjectiveImpl struct{}

func (t *travelDurationObjectiveImpl) ModelExpressions() ModelExpressions {
	return ModelExpressions{}
}

func (t *travelDurationObjectiveImpl) EstimateDeltaValue(move SolutionMoveStops) float64 {
	return move.(*solutionMoveStopsImpl).deltaTravelDurationValue()
}

func (t *travelDurationObjectiveImpl) Value(solution Solution) float64 {
	solutionImp := solution.(*solutionImpl)

	score := 0.0
	for _, vehicle := range solutionImp.vehicles {
		score += vehicle.Last().CumulativeTravelDurationValue()
	}
	return score
}

func (t *travelDurationObjectiveImpl) String() string {
	return "travel_duration"
}
