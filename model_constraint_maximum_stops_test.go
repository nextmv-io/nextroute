// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestMaximumStopsConstraint_EstimateIsViolated(t *testing.T) {
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

	maximumStops := nextroute.NewVehicleTypeValueExpression(
		"maximum stops",
		2,
	)

	cnstr, err := nextroute.NewMaximumStopsConstraint(maximumStops)
	if err != nil {
		t.Error(err)
	}

	err = maximumStops.SetValue(model.VehicleTypes()[0], 1)
	if err != nil {
		t.Error(err)
	}
	err = maximumStops.SetValue(model.VehicleTypes()[1], 2)
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
	sequencePlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() > 1
	})

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	solutionSingleStopPlanUnit := solution.SolutionPlanStopsUnit(singleStopPlanUnits[0])
	position, err := nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionSingleStopPlanUnit.SolutionStops()[0],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	moveSingleOnVehicle0, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}
	position, err = nextroute.NewStopPosition(
		solution.Vehicles()[1].First(),
		solutionSingleStopPlanUnit.SolutionStops()[0],
		solution.Vehicles()[1].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSingleOnVehicle1, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle0); violated {
		t.Error("constraint is violated")
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle1); violated {
		t.Error("constraint is violated")
	}

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
		solution.Vehicles()[0].Last(),
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

	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[1].First(),
		solutionSequencePlanUnit.SolutionStops()[0],
		solutionSequencePlanUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionSequencePlanUnit.SolutionStops()[0],
		solutionSequencePlanUnit.SolutionStops()[1],
		solution.Vehicles()[1].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	moveSequenceOnVehicle1, err := nextroute.NewMoveStops(
		solutionSequencePlanUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle0); !violated {
		t.Error("constraint is not violated")
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle1); violated {
		t.Error("constraint is violated")
	}

	planned, err := moveSingleOnVehicle0.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("move is not planned")
	}
	planned, err = moveSequenceOnVehicle1.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("move is not planned")
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
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[1].First(),
		solutionSequencePlanUnit.SolutionStops()[0],
		solutionSequencePlanUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionSequencePlanUnit.SolutionStops()[0],
		solutionSequencePlanUnit.SolutionStops()[1],
		solution.Vehicles()[1].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSequenceOnVehicle1, err = nextroute.NewMoveStops(
		solutionSequencePlanUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle0); !violated {
		t.Error("constraint is not violated")
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle1); !violated {
		t.Error("constraint is not violated")
	}
}

func TestMaximumStopsConstraint(t *testing.T) {
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

	maximumStops := nextroute.NewVehicleTypeValueExpression(
		"maximum stops",
		2,
	)

	cnstr, err := nextroute.NewMaximumStopsConstraint(maximumStops)
	if err != nil {
		t.Error(err)
	}

	if cnstr.MaximumStops().Value(model.VehicleTypes()[0], nil, nil) != 2 {
		t.Errorf(
			"maximum stops is not correct, expected 2 got %v",
			cnstr.MaximumStops().Value(model.VehicleTypes()[0], nil, nil),
		)
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
