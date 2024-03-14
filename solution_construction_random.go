// © 2019-present nextmv.io inc

package nextroute

import (
	"context"

	"github.com/nextmv-io/nextroute/common"
)

// NewRandomSolution returns a random solution for the given model.
func NewRandomSolution(ctx context.Context, model Model) (Solution, error) {
	solution, err := NewSolution(model)
	if err != nil {
		return nil, err
	}
	return RandomSolutionConstruction(ctx, solution)
}

// RandomSolutionConstruction returns a random solution by populating the
// empty input with a random plan unit. The remaining plan units
// are added to the solution in a random order at the best possible position.
func RandomSolutionConstruction(ctx context.Context, s Solution) (Solution, error) {
	solution := s.Copy()

	emptyVehicles := common.Filter(
		solution.Vehicles(),
		func(v SolutionVehicle) bool {
			return v.IsEmpty()
		},
	)

LoopVehicles:
	for _, vehicle := range emptyVehicles {
		unplannedPlanUnits := NewSolutionPlanUnitCollection(
			solution.Random(),
			solution.UnPlannedPlanUnits().SolutionPlanUnits(),
		)

	UnplannedUnitsLoop:
		for unplannedPlanUnits.Size() > 0 {
			select {
			case <-ctx.Done():
				break LoopVehicles
			default:
				unplannedPlanUnit := unplannedPlanUnits.RandomElement()

				m := vehicle.BestMove(ctx, unplannedPlanUnit)

				if m.IsImprovement() {
					result, err := m.Execute(ctx)
					if err != nil {
						return nil, err
					}
					if result {
						break UnplannedUnitsLoop
					}
				}

				unplannedPlanUnits.Remove(unplannedPlanUnit)
			}
		}
	}

	unplannedPlanUnits := NewSolutionPlanUnitCollection(
		solution.Random(),
		solution.UnPlannedPlanUnits().SolutionPlanUnits(),
	)

LoopUnplannedPlanUnits:
	for unplannedPlanUnits.Size() > 0 {
		select {
		case <-ctx.Done():
			break LoopUnplannedPlanUnits
		default:
			unplannedPlanUnit := unplannedPlanUnits.RandomElement()

			m := solution.BestMove(ctx, unplannedPlanUnit)

			if m.IsImprovement() {
				_, err := m.Execute(ctx)
				if err != nil {
					return nil, err
				}
			}

			unplannedPlanUnits.Remove(unplannedPlanUnit)
		}
	}
	return solution, nil
}
