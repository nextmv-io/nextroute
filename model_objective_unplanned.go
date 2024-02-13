package nextroute

import (
	"github.com/nextmv-io/sdk/nextroute"
)

// NewUnPlannedObjective returns a new UnPlannedObjective.
func NewUnPlannedObjective(
	expression nextroute.StopExpression,
) nextroute.UnPlannedObjective {
	return &unplannedObjectiveImpl{
		expression: expression,
	}
}

type unplannedObjectiveImpl struct {
	expression nextroute.StopExpression
	costs      []float64
}

func (t *unplannedObjectiveImpl) calculateCosts(planUnit nextroute.ModelPlanUnit) float64 {
	switch unit := planUnit.(type) {
	case nextroute.ModelPlanStopsUnit:
		cost := 0.0
		for _, stop := range unit.Stops() {
			cost += t.expression.Value(nil, nil, stop)
		}
		return cost
	case nextroute.ModelPlanUnitsUnit:
		cost := 0.0
		for _, planUnit := range unit.PlanUnits() {
			cost += t.calculateCosts(planUnit)
		}
		if unit.PlanOneOf() {
			// we take the average cost of planing one unit
			return cost / float64(len(unit.PlanUnits()))
		}
		return cost
	default:
		panic("planUnit type is not recognized for the unplanned objective")
	}
}

func (t *unplannedObjectiveImpl) Lock(model nextroute.Model) error {
	units := model.PlanUnits()
	t.costs = make([]float64, len(units))
	for _, planUnit := range units {
		t.costs[planUnit.Index()] = t.calculateCosts(planUnit)
	}
	return nil
}

func (t *unplannedObjectiveImpl) ModelExpressions() nextroute.ModelExpressions {
	return nextroute.ModelExpressions{}
}

func (t *unplannedObjectiveImpl) EstimateDeltaValue(move nextroute.SolutionMoveStops) float64 {
	return -1 * t.costs[move.(*solutionMoveStopsImpl).planUnit.modelPlanStopsUnit.Index()]
}

func (t *unplannedObjectiveImpl) Value(solution nextroute.Solution) float64 {
	unplannedScore := 0.0

	units := solution.UnPlannedPlanUnits().(*solutionPlanUnitCollectionBaseImpl).solutionPlanUnits
	for _, upu := range units {
		switch upu := upu.(type) {
		case *solutionPlanStopsUnitImpl:
			unplannedScore += t.costs[upu.modelPlanStopsUnit.Index()]
		case *solutionPlanUnitsUnitImpl:
			unplannedScore += t.costs[upu.modelPlanUnitsUnit.Index()]
		}
	}
	return unplannedScore
}

func (t *unplannedObjectiveImpl) String() string {
	return "unplanned_penalty"
}
