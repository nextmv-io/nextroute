package nextroute

import (
	"github.com/nextmv-io/sdk/nextroute"
)

type modelConstraintImpl struct {
	name        string
	expressions nextroute.ModelExpressions
}

func (m *modelConstraintImpl) EstimateIsViolated(
	_ nextroute.SolutionMoveStops,
	_ nextroute.Solution,
) (isViolated bool, stopPositionsHint nextroute.StopPositionsHint) {
	panic("implement me in derived class")
}

func newModelConstraintImpl(
	name string,
	expressions nextroute.ModelExpressions,
) modelConstraintImpl {
	return modelConstraintImpl{
		expressions: expressions,
		name:        name,
	}
}

func (m *modelConstraintImpl) ModelExpressions() nextroute.ModelExpressions {
	return m.expressions
}
