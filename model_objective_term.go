// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

func newModelObjectiveTerm(
	factor float64,
	objective ModelObjective,
) ModelObjectiveTerm {
	return modelObjectiveTermImpl{
		factor:    factor,
		objective: objective,
	}
}

type modelObjectiveTermImpl struct {
	objective ModelObjective
	factor    float64
}

func (m modelObjectiveTermImpl) Factor() float64 {
	return m.factor
}

func (m modelObjectiveTermImpl) Objective() ModelObjective {
	return m.objective
}

func (m modelObjectiveTermImpl) String() string {
	return fmt.Sprintf("%v * %v", m.factor, m.objective)
}
