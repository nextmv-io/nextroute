package nextroute

import (
	"context"
	"fmt"
	"math"

	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
	"golang.org/x/exp/slices"
)

// NewSweepSolution returns a solution for the given model using the sweep
// heuristic.
func NewSweepSolution(ctx context.Context, model nextroute.Model) (nextroute.Solution, error) {
	solution, err := NewSolution(model)
	if err != nil {
		return nil, err
	}
	return SweepSolutionConstruction(ctx, solution)
}

// SweepSolutionConstruction returns a solution by planning the plan units
// in order of a radar sweep around the depot. Will raise an error if there is
// more than one depot location either at the start or end of a vehicle.
// The sweep starts at a random angle and continues clockwise.
func SweepSolutionConstruction(ctx context.Context, s nextroute.Solution) (nextroute.Solution, error) {
	solution := s.Copy()

	emptyVehicles := common.Filter(
		solution.Vehicles(),
		func(vehicle nextroute.SolutionVehicle) bool {
			return vehicle.IsEmpty()
		},
	)

	locations := make(common.Locations, 0, len(emptyVehicles)*2)

	for _, vehicle := range emptyVehicles {
		if vehicle.First().ModelStop().Location().IsValid() {
			locations = append(
				locations,
				vehicle.First().ModelStop().Location(),
			)
		}
		if vehicle.Last().ModelStop().Location().IsValid() {
			locations = append(
				locations,
				vehicle.Last().ModelStop().Location(),
			)
		}
	}

	locations = locations.Unique()

	if len(locations) != 1 {
		return nil, fmt.Errorf(
			"sweep construction, not implemented for multiple" +
				" start-end locations of input",
		)
	}

	depot := locations[0]

	unplannedPlanUnits := solution.UnPlannedPlanUnits().SolutionPlanUnits()

	slices.SortStableFunc(unplannedPlanUnits, func(
		leftSolutionPlanUnit, rightSolutionPlanUnit nextroute.SolutionPlanUnit) int {
		if leftModelPlanStopsUnit, iOK :=
			leftSolutionPlanUnit.ModelPlanUnit().(nextroute.ModelPlanStopsUnit); iOK {
			if rightModelPlanStopsUnit, jOK :=
				rightSolutionPlanUnit.ModelPlanUnit().(nextroute.ModelPlanStopsUnit); jOK {
				leftCentroid, _ := leftModelPlanStopsUnit.Centroid()
				rightCentroid, _ := rightModelPlanStopsUnit.Centroid()
				if clockWise(depot, leftCentroid)-clockWise(depot, rightCentroid) < 0 {
					return -1
				}
			}
		}
		return 1
	})

	startIndex := solution.Random().Intn(len(unplannedPlanUnits))

LoopUnplannedPlanUnits:
	for idx := startIndex; idx < startIndex+len(unplannedPlanUnits); idx++ {
		unplannedUnit := unplannedPlanUnits[idx%len(unplannedPlanUnits)]
		select {
		case <-ctx.Done():
			break LoopUnplannedPlanUnits
		default:
			m := solution.BestMove(ctx, unplannedUnit)
			if m.IsImprovement() {
				_, err := m.Execute(ctx)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return solution, nil
}

func clockWise(center, location common.Location) float64 {
	return math.Atan2(
		location.Latitude()-center.Latitude(),
		location.Longitude()-center.Longitude(),
	)
}
