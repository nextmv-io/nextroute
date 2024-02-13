package nextroute

import (
	"fmt"
	"reflect"
	"strings"

	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/sdk/nextroute"
)

type modelObjectiveImpl struct{}

func newModelObjectiveImpl() modelObjectiveImpl {
	return modelObjectiveImpl{}
}

type modelObjectiveSumImpl struct {
	modelObjectiveImpl
	model *modelImpl
	terms nextroute.ModelObjectiveTerms
}

func (m *modelObjectiveSumImpl) ModelExpressions() nextroute.ModelExpressions {
	return nextroute.ModelExpressions{}
}

func newModelObjectiveSum(m *modelImpl) nextroute.ModelObjectiveSum {
	return &modelObjectiveSumImpl{
		modelObjectiveImpl: newModelObjectiveImpl(),
		terms:              make(nextroute.ModelObjectiveTerms, 0, 1),
		model:              m,
	}
}

func (m *modelObjectiveSumImpl) String() string {
	var sb strings.Builder
	for idx, term := range m.terms {
		if idx > 0 {
			fmt.Fprintf(&sb, " + ")
		}
		fmt.Fprintf(&sb, "%v * %v",
			term.Factor(),
			term.Objective(),
		)
	}
	return sb.String()
}

func (m *modelObjectiveSumImpl) EstimateDeltaValue(move nextroute.SolutionMoveStops) float64 {
	estimateDeltaScore := 0.0
	for _, term := range m.terms {
		estimateDeltaScore += term.Factor() * term.Objective().EstimateDeltaValue(move)
	}
	return estimateDeltaScore
}

func (m *modelObjectiveSumImpl) InternalValue(_ *solutionImpl) float64 {
	panic("use Solution.ObjectiveValue or solution.Score to query objective value")
}

func (m *modelObjectiveSumImpl) Value(_ nextroute.Solution) float64 {
	panic("use Solution.ObjectiveValue or solution.Score to query objective value")
}

func (m *modelObjectiveSumImpl) Terms() nextroute.ModelObjectiveTerms {
	return m.terms
}

func (m *modelObjectiveSumImpl) NewTerm(
	factor float64,
	objective nextroute.ModelObjective,
) (nextroute.ModelObjectiveTerm, error) {
	term := newModelObjectiveTerm(factor, objective)
	for _, existingTerm := range m.terms {
		if &existingTerm == &term {
			return nil, nmerror.NewModelCustomizationError(fmt.Errorf(
				"objective '%v' with address %v already added,"+
					" if objective has not been added: address must be unique",
				term,
				&term,
			))
		}
	}
	if _, ok := objective.(nextroute.ObjectiveDataUpdater); ok {
		return nil, nmerror.NewModelCustomizationError(fmt.Errorf(
			"nextroute.ObjectiveDataUpdater has been deprecated, "+
				"please use nextroute.ObjectiveStopDataUpdater instead, "+
				"rename UpdateObjectiveData to UpdateObjectiveStopData for %s",
			reflect.TypeOf(objective).String(),
		))
	}
	if factor != 0 {
		m.terms = append(m.terms, term)

		if registered, ok := term.Objective().(nextroute.RegisteredModelExpressions); ok {
			for _, expression := range registered.ModelExpressions() {
				m.model.addExpression(expression)
			}
		}
		if _, ok := term.Objective().(nextroute.ObjectiveStopDataUpdater); ok {
			m.model.objectivesWithStopUpdater = append(
				m.model.objectivesWithStopUpdater,
				term.Objective(),
			)
		}
		if _, ok := term.Objective().(nextroute.ObjectiveSolutionDataUpdater); ok {
			m.model.objectivesWithSolutionUpdater = append(
				m.model.objectivesWithSolutionUpdater,
				term.Objective(),
			)
		}
	}
	return term, nil
}
