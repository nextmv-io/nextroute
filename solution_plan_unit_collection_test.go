// © 2019-present nextmv.io inc

package nextroute_test

import (
	"slices"
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestSolutionPlanUnitCollection(t *testing.T) {
	model, err := createModel(singleVehiclePlanSingleStopsModel())
	if err != nil {
		t.Fatal(err)
	}
	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	sourcePlanUnits := slices.Clone(solution.UnPlannedPlanUnits().SolutionPlanUnits())

	unplannedPlanUnitCollection := nextroute.NewSolutionPlanUnitCollection(
		solution.Random(),
		sourcePlanUnits,
	)

	if unplannedPlanUnitCollection.Size() != 3 {
		t.Error("unplannedPlanUnitCollection.Size() should be 3")
	}

	if len(unplannedPlanUnitCollection.SolutionPlanUnits()) != 3 {
		t.Error("len(unplannedPlanUnitCollection.SolutionPlanUnits()) should be 3")
	}

	sourcePlanUnits[2] = nil

	if unplannedPlanUnitCollection.Size() != 3 {
		t.Error("unplannedPlanUnitCollection.Size() should be 3")
	}

	for _, planUnit := range unplannedPlanUnitCollection.SolutionPlanUnits() {
		if planUnit == nil {
			t.Error("planUnit should not be nil")
		}
	}

	if len(unplannedPlanUnitCollection.SolutionPlanUnits()) != 3 {
		t.Error("len(unplannedPlanUnitCollection.SolutionPlanUnits()) should be 3")
	}

	elements := unplannedPlanUnitCollection.RandomDraw(2)

	if len(elements) != 2 {
		t.Error("len(elements) should be 2")
	}

	unplannedPlanUnitCollection.Remove(elements[0])

	elements = unplannedPlanUnitCollection.RandomDraw(2)

	if len(elements) != 2 {
		t.Error("len(elements) should be 2")
	}

	unplannedPlanUnitCollection.Remove(elements[1])

	elements = unplannedPlanUnitCollection.RandomDraw(2)

	if len(elements) != 1 {
		t.Error("len(elements) should be 1")
	}
}
