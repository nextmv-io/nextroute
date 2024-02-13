package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/nextroute"
)

func newModelObjectiveTerm(
	factor float64,
	objective nextroute.ModelObjective,
) nextroute.ModelObjectiveTerm {
	return modelObjectiveTermImpl{
		factor:    factor,
		objective: objective,
	}
}

type modelObjectiveTermImpl struct {
	objective nextroute.ModelObjective
	factor    float64
}

func (m modelObjectiveTermImpl) Factor() float64 {
	return m.factor
}

func (m modelObjectiveTermImpl) Objective() nextroute.ModelObjective {
	return m.objective
}

func (m modelObjectiveTermImpl) String() string {
	return fmt.Sprintf("%v * %v", m.factor, m.objective)
}
