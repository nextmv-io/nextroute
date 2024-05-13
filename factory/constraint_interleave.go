// © 2019-present nextmv.io inc

package factory

import (
	"fmt"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addInterleaveConstraint adds a constraint which limits stop groups to be
// interleaved.
func addInterleaveConstraint(
	input schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	if input.StopGroups == nil || options.Constraints.Disable.Groups {
		return model, nil
	}

	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}

	planUnits := make(nextroute.ModelPlanUnits, 0, len(*input.StopGroups))
	for idx, stops := range *input.StopGroups {
		if len(stops) == 0 {
			return nil, fmt.Errorf("stop group %v is empty", idx)
		}
		stop := stops[0]
		modelStop, err := model.Stop(data.stopIDToIndex[stop])
		if err != nil {
			return nil, err
		}
		if !modelStop.HasPlanStopsUnit() {
			return nil, fmt.Errorf("stop %s does not have a plan-stops unit", stop)
		}
		planStopsUnit := modelStop.PlanStopsUnit()
		if planUnitsUnit, hasPlanUnitsUnit := planStopsUnit.PlanUnitsUnit(); hasPlanUnitsUnit {
			if numberOfStops(planUnitsUnit) != len(stops) {
				return nil,
					fmt.Errorf(
						"stop group %v starting with stop %v is not a plan unit with %v stops,"+
							" but has %v stops instead, probably because there are inter stop group precedence"+
							" relationships which are not allowed with interleave constraint",
						idx,
						stop,
						len(stops),
						numberOfStops(planUnitsUnit),
					)
			}
			planUnits = append(planUnits, planUnitsUnit)
		} else {
			if numberOfStops(planStopsUnit) != len(stops) {
				return nil,
					fmt.Errorf(
						"stop group %v starting with stop %v is not a plan unit with %v stops,"+
							" but has %v stops instead, probably because there are inter stop group precedence"+
							" relationships which are not allowed with interleave constraint",
						idx,
						stop,
						len(stops),
						numberOfStops(planStopsUnit),
					)
			}
			planUnits = append(planUnits, planStopsUnit)
		}
	}

	if len(planUnits) <= 1 {
		return model, nil
	}

	interleaveConstraint, err := nextroute.NewInterleaveConstraint()
	if err != nil {
		return model, err
	}

	for i, p1 := range planUnits {
		inputPlanUnits := make(nextroute.ModelPlanUnits, 0, len(planUnits)-1)
		for j, p2 := range planUnits {
			if i != j {
				inputPlanUnits = append(inputPlanUnits, p2)
			}
		}
		err = interleaveConstraint.DisallowInterleaving(p1, inputPlanUnits)
		if err != nil {
			return model, err
		}
	}

	err = model.AddConstraint(interleaveConstraint)
	if err != nil {
		return model, err
	}

	return model, nil
}

func numberOfStops(planUnit nextroute.ModelPlanUnit) int {
	if planStopsUnit, isPlanStopsUnit := planUnit.(nextroute.ModelPlanStopsUnit); isPlanStopsUnit {
		return len(planStopsUnit.Stops())
	}
	count := 0
	if planUnitsUnit, isPlanUnitsUnit := planUnit.(nextroute.ModelPlanUnitsUnit); isPlanUnitsUnit {
		for _, unit := range planUnitsUnit.PlanUnits() {
			count += numberOfStops(unit)
		}
	}
	if planUnitsUnit, isPlanUnitsUnit := planUnit.PlanUnitsUnit(); isPlanUnitsUnit {
		for _, unit := range planUnitsUnit.PlanUnits() {
			count += numberOfStops(unit)
		}
	}
	return count
}
