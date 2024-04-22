// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestClusterConstraint_EstimateIsViolated(t *testing.T) {
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

	cnstr, err := nextroute.NewCluster()
	if err != nil {
		t.Error(err)
	}

	err = model.AddConstraint(cnstr)
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

	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle0); violated {
		t.Error("constraint is violated")
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

	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle1); violated {
		t.Error("constraint is violated")
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

	// This move should violate the constraint because the third point is closer
	// to the second point than the first point.
	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle2); !violated {
		t.Error("constraint should be violated")
	}
	position, err = nextroute.NewStopPosition(
		solution.Vehicles()[1].First(),
		solutionSingleStopPlanUnit2.SolutionStops()[0],
		solution.Vehicles()[1].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSingleOnVehicle3, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit2,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle3); violated {
		t.Error("constraint is violated")
	}

	sequencePlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() > 1
	})
	solutionSequencePlanUnit := solution.SolutionPlanStopsUnit(sequencePlanUnits[0])
	position1, err := nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionSequencePlanUnit.SolutionStops()[0],
		solutionSequencePlanUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err := nextroute.NewStopPosition(
		solutionSequencePlanUnit.SolutionStops()[0],
		solutionSequencePlanUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSequenceOnVehicle0, err := nextroute.NewMoveStops(
		solutionSequencePlanUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle0); !violated {
		t.Error("constraint is not violated")
	}

	b, err = moveSequenceOnVehicle0.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if b {
		t.Error("move resulted in planned planunit although it results in infeasible solution")
	}

	solutionSequencePlanUnit = solution.SolutionPlanStopsUnit(sequencePlanUnits[1])
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionSequencePlanUnit.SolutionStops()[0],
		solutionSequencePlanUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionSequencePlanUnit.SolutionStops()[0],
		solutionSequencePlanUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSequenceOnVehicle0, err = nextroute.NewMoveStops(
		solutionSequencePlanUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle0); !violated {
		t.Error("constraint is not violated")
	}

	b, err = moveSequenceOnVehicle0.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if b {
		t.Error("move is executed and planned")
	}
}

func TestClusterConstraint(t *testing.T) {
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

	cnstr, err := nextroute.NewCluster()
	if err != nil {
		t.Error(err)
	}

	err = model.AddConstraint(cnstr)
	if err != nil {
		t.Error(err)
	}

	if len(model.Constraints()) != 1 {
		t.Errorf(
			"number of constraints is not correct, expected 1 got %v",
			len(model.Constraints()),
		)
	}

	if model.Constraints()[0] != cnstr {
		t.Errorf(
			"constraint is not correct, expected %v got %v",
			cnstr,
			model.Constraints()[0],
		)
	}
}
