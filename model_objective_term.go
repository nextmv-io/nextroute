// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// ModelObjectiveTerm is a term in a model objective sum.
type ModelObjectiveTerm struct {
	objective ModelObjective
	factor    float64
}

// Factor returns the factor by which the objective is multiplied.
func (m ModelObjectiveTerm) Factor() float64 {
	return m.factor
}

// Objective returns the objective that is multiplied by the factor.
func (m ModelObjectiveTerm) Objective() ModelObjective {
	return m.objective
}

// String returns the string representation of the model objective term.
func (m ModelObjectiveTerm) String() string {
	return fmt.Sprintf("%v * %v", m.factor, m.objective)
}
