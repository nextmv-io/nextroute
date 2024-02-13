package nextroute

import (
	"github.com/nextmv-io/sdk/nextroute"
)

// NewTravelDurationObjective returns a new TravelDurationObjective.
func NewTravelDurationObjective() nextroute.TravelDurationObjective {
	return &travelDurationObjectiveImpl{}
}

type travelDurationObjectiveImpl struct{}

func (t *travelDurationObjectiveImpl) ModelExpressions() nextroute.ModelExpressions {
	return nextroute.ModelExpressions{}
}

func (t *travelDurationObjectiveImpl) EstimateDeltaValue(move nextroute.SolutionMoveStops) float64 {
	return move.(*solutionMoveStopsImpl).deltaTravelDurationValue()
}

func (t *travelDurationObjectiveImpl) Value(solution nextroute.Solution) float64 {
	solutionImp := solution.(*solutionImpl)

	score := 0.0
	for _, vehicle := range solutionImp.vehicles {
		score += vehicle.last().CumulativeTravelDurationValue()
	}
	return score
}

func (t *travelDurationObjectiveImpl) String() string {
	return "travel_duration"
}
