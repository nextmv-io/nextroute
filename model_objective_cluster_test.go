// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"math"
	"testing"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestClusterObjective_EstimateDeltaValue(t *testing.T) {
	model, err := createModel(
		input(
			vehicleTypes("truck", "car"),
			[]Vehicle{
				vehicles(
					"truck",
					depot(),
					1,
				)[0],
				vehicles(
					"car",
					depot(),
					1,
				)[0],
			},
			planSingleStops(),
			planPairSequences(),
		),
	)
	if err != nil {
		t.Error(err)
	}

	obj, err := nextroute.NewCluster()
	if err != nil {
		t.Error(err)
	}

	_, err = model.Objective().NewTerm(1.0, obj)
	if err != nil {
		t.Error(err)
	}

	singleStopPlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() == 1
	})

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	solutionSingleStopPlanUnit0 := solution.SolutionPlanStopsUnit(singleStopPlanUnits[0])
	position, err := nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionSingleStopPlanUnit0.SolutionStops()[0],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSingleOnVehicle0, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit0,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	if delta := obj.EstimateDeltaValue(moveSingleOnVehicle0); delta != 0 {
		t.Error("delta estimation is incorrect")
	}

	b, err := moveSingleOnVehicle0.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !b {
		t.Error("move could not be executed")
	}

	solutionSingleStopPlanUnit1 := solution.SolutionPlanStopsUnit(singleStopPlanUnits[1])
	position, err = nextroute.NewStopPosition(
		solution.Vehicles()[1].First(),
		solutionSingleStopPlanUnit1.SolutionStops()[0],
		solution.Vehicles()[1].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSingleOnVehicle1, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit1,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	if delta := obj.EstimateDeltaValue(moveSingleOnVehicle1); delta != 0 {
		t.Error("delta estimation is incorrect")
	}

	b, err = moveSingleOnVehicle1.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !b {
		t.Error("move could not be executed")
	}

	solutionSingleStopPlanUnit2 := solution.SolutionPlanStopsUnit(singleStopPlanUnits[2])
	position, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionSingleStopPlanUnit2.SolutionStops()[0],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSingleOnVehicle2, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit2,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedValue := 669935.264596651
	if delta := obj.EstimateDeltaValue(moveSingleOnVehicle2); math.Round(delta) != math.Round(expectedValue) {
		t.Error("delta estimation is incorrect")
	}
}

func TestClusterObjective(t *testing.T) {
	model, err := createModel(
		input(
			vehicleTypes("truck"),
			vehicles(
				"truck",
				depot(),
				2,
			),
			planSingleStops(),
			nil,
		),
	)
	if err != nil {
		t.Error(err)
	}

	obj, err := nextroute.NewCluster()
	if err != nil {
		t.Error(err)
	}

	_, err = model.Objective().NewTerm(1.0, obj)
	if err != nil {
		t.Error(err)
	}

	if len(model.Objective().Terms()) != 1 {
		t.Errorf(
			"number of objectives is not correct, expected 1 got %v",
			len(model.Objective().Terms()),
		)
	}

	if model.Objective().Terms()[0].Objective() != obj {
		t.Errorf(
			"objective is not correct, expected %v got %v",
			obj,
			model.Objective().Terms()[0],
		)
	}
}
