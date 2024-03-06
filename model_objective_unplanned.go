// © 2019-present nextmv.io inc

package nextroute

import "fmt"

// UnPlannedObjective is an objective that uses the un-planned stops as an
// objective. Each unplanned stop is scored by the given expression.
type UnPlannedObjective interface {
	ModelObjective
}

// NewUnPlannedObjective returns a new UnPlannedObjective.
func NewUnPlannedObjective(
	expression StopExpression,
) UnPlannedObjective {
	return &unplannedObjectiveImpl{
		expression: expression,
	}
}

type unplannedObjectiveImpl struct {
	expression StopExpression
	costs      []float64
}

func (t *unplannedObjectiveImpl) calculateCosts(
	planUnit ModelPlanUnit,
) (float64, error) {
	switch unit := planUnit.(type) {
	case ModelPlanStopsUnit:
		cost := 0.0
		for _, stop := range unit.Stops() {
			cost += t.expression.Value(nil, nil, stop)
		}
		return cost, nil
	case ModelPlanUnitsUnit:
		cost := 0.0
		for _, planUnit := range unit.PlanUnits() {
			c, err := t.calculateCosts(planUnit)
			if err != nil {
				return 0, err
			}
			cost += c
		}
		if unit.PlanOneOf() {
			// we take the average cost of planing one unit
			return cost / float64(len(unit.PlanUnits())), nil
		}
		return cost, nil
	default:
		return 0, fmt.Errorf(
			"model plan unit type is not recognized for the unplanned objective",
		)
	}
}

func (t *unplannedObjectiveImpl) Lock(model Model) error {
	units := model.PlanUnits()
	t.costs = make([]float64, len(units))
	for _, planUnit := range units {
		cost, err := t.calculateCosts(planUnit)
		if err != nil {
			return err
		}
		t.costs[planUnit.Index()] = cost
	}
	return nil
}

func (t *unplannedObjectiveImpl) ModelExpressions() ModelExpressions {
	return ModelExpressions{}
}

func (t *unplannedObjectiveImpl) EstimateDeltaValue(move SolutionMoveStops) float64 {
	return -1 * t.costs[move.(*solutionMoveStopsImpl).planUnit.modelPlanStopsUnit.Index()]
}

func (t *unplannedObjectiveImpl) Value(solution Solution) float64 {
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
