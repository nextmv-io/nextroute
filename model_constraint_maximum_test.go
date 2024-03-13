// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestMaximumConstraint_EstimateIsViolated1(t *testing.T) {
	model, err := createModel(
		input(
			vehicleTypes("truck"),
			[]Vehicle{
				vehicles(
					"truck",
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
	delta := nextroute.NewFromToExpression(
		"delta level",
		0,
	)

	maximum := nextroute.NewVehicleTypeValueExpression(
		"maximum level",
		0,
	)

	err = maximum.SetValue(model.VehicleTypes()[0], 2)
	if err != nil {
		t.Error(err)
	}

	cnstr, err := nextroute.NewMaximum(
		delta,
		maximum,
	)
	if err != nil {
		t.Error(err)
	}
	cnstr.(nextroute.Identifier).SetID("maximum_constraint")

	err = model.AddConstraint(cnstr)

	if err != nil {
		t.Error(err)
	}

	singleStopPlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() == 1
	})

	err = delta.SetValue(model.Vehicles()[0].First(), singleStopPlanUnits[0].Stops()[0], 1.0)
	if err != nil {
		t.Error(err)
	}
	err = delta.SetValue(singleStopPlanUnits[0].Stops()[0], model.Vehicles()[0].Last(), 1.0)
	if err != nil {
		t.Error(err)
	}
	err = delta.SetValue(model.Vehicles()[0].First(), singleStopPlanUnits[1].Stops()[0], 1.0)
	if err != nil {
		t.Error(err)
	}
	err = delta.SetValue(singleStopPlanUnits[1].Stops()[0], singleStopPlanUnits[0].Stops()[0], 1.0)
	if err != nil {
		t.Error(err)
	}
	err = delta.SetValue(singleStopPlanUnits[1].Stops()[0], model.Vehicles()[0].Last(), 1.0)
	if err != nil {
		t.Error(err)
	}

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
	// F(0) -+1- S1(1) -+1- L(2)
	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle0); violated {
		t.Fatal("constraint is violated")
	}
	planned, err := moveSingleOnVehicle0.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}

	if !planned {
		t.Error("move is not planned")
	}
	solutionSingleStopPlanUnit1 := solution.SolutionPlanStopsUnit(singleStopPlanUnits[1])
	position, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionSingleStopPlanUnit1.SolutionStops()[0],
		solution.Vehicles()[0].First().Next(),
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
	// F(0) -+1- S1(1) -+1- S2 -+1- L(3)
	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle1); !violated {
		t.Fatal("constraint is not violated")
	}
}

func TestMaximumConstraint_EstimateIsViolated2(t *testing.T) {
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

	delta := nextroute.NewStopExpression(
		"delta level",
		0,
	)

	maximum := nextroute.NewVehicleTypeValueExpression(
		"maximum level",
		0,
	)

	err = maximum.SetValue(model.VehicleTypes()[0], 1)
	if err != nil {
		t.Error(err)
	}
	err = maximum.SetValue(model.VehicleTypes()[1], 2)
	if err != nil {
		t.Error(err)
	}

	cnstr, err := nextroute.NewMaximum(
		delta,
		maximum,
	)
	if err != nil {
		t.Error(err)
	}
	cnstr.(nextroute.Identifier).SetID("maximum_constraint")

	err = model.AddConstraint(cnstr)

	if err != nil {
		t.Error(err)
	}

	singleStopPlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() == 1
	})

	err = delta.SetValue(singleStopPlanUnits[0].Stops()[0], 1)
	if err != nil {
		t.Error(err)
	}
	err = delta.SetValue(singleStopPlanUnits[1].Stops()[0], -1)
	if err != nil {
		t.Error(err)
	}
	err = delta.SetValue(singleStopPlanUnits[2].Stops()[0], 2)
	if err != nil {
		t.Error(err)
	}

	sequencePlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() > 1
	})

	err = delta.SetValue(sequencePlanUnits[0].Stops()[0], 1)
	if err != nil {
		t.Error(err)
	}
	err = delta.SetValue(sequencePlanUnits[0].Stops()[1], -1)
	if err != nil {
		t.Error(err)
	}

	err = delta.SetValue(sequencePlanUnits[1].Stops()[0], 1)
	if err != nil {
		t.Error(err)
	}
	err = delta.SetValue(sequencePlanUnits[1].Stops()[1], -1)
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	solutionSingleStopPlanUnit2 := solution.SolutionPlanStopsUnit(singleStopPlanUnits[2])
	position, err := nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionSingleStopPlanUnit2.SolutionStops()[0],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	moveSingleOnVehicle0, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit2,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	position, err = nextroute.NewStopPosition(
		solution.Vehicles()[1].First(),
		solutionSingleStopPlanUnit2.SolutionStops()[0],
		solution.Vehicles()[1].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSingleOnVehicle1, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit2,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle0); !violated {
		t.Fatal("constraint is not violated, vehicle 0 does not fit 2")
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle1); violated {
		t.Fatal("constraint is not violated, vehicle 1 does  fit 2")
	}

	// vehicle 1 with maximum level 2 has after this a stop which consumes 2
	planned, err := moveSingleOnVehicle1.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("move is not planned")
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
		solution.Vehicles()[1].First().Next(),
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

	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle0); violated {
		t.Error("constraint is violated")
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle1); violated {
		t.Error("constraint is violated")
	}

	// vehicle 0 with maximum level 1 has after this a stop which consumes 1
	// and produces 1 (delta is zero at end of sequence)
	planned, err = moveSequenceOnVehicle0.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("move is not planned")
	}

	solutionSingleStopPlanUnit0 := solution.SolutionPlanStopsUnit(singleStopPlanUnits[0])
	position, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionSingleStopPlanUnit0.SolutionStops()[0],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}

	moveSingleOnVehicle0AtStart, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit0,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	// vehicle 0 with maximum level 1 can not be raised to 1 at start of vehicle
	// as that would violate the constraint for the already planned sequence on
	// vehicle 0
	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle0AtStart); !violated {
		t.Error("constraint is not violated")
	}
	position, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous(),
		solutionSingleStopPlanUnit0.SolutionStops()[0],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSingleOnVehicle0AtEnd, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit0,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	// vehicle 0 with maximum level 1 can be raised to 1 at end of vehicle
	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle0AtEnd); violated {
		t.Error("constraint is violated")
	}
	position, err = nextroute.NewStopPosition(
		solution.Vehicles()[1].First(),
		solutionSingleStopPlanUnit0.SolutionStops()[0],
		solution.Vehicles()[1].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSingleOnVehicle1AtStart, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit0,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	// vehicle 1 with maximum level 2 can not be raised to 1 at start of vehicle
	// as that would violate the constraint for the already planned stop on
	// vehicle 1
	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle1AtStart); !violated {
		t.Error("constraint is not violated")
	}
	position, err = nextroute.NewStopPosition(
		solution.Vehicles()[1].Last().Previous(),
		solutionSingleStopPlanUnit0.SolutionStops()[0],
		solution.Vehicles()[1].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	moveSingleOnVehicle1AtEnd, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit0,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	// vehicle 1 with maximum level 2 can not be raised to 3 at end of vehicle
	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle1AtEnd); !violated {
		t.Error("constraint is not violated")
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
	moveSequenceOnVehicle0AtStart, err := nextroute.NewMoveStops(
		solutionSequencePlanUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous(),
		solutionSequencePlanUnit.SolutionStops()[0],
		solutionSequencePlanUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionSequencePlanUnit.SolutionStops()[0],
		solutionSequencePlanUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	moveSequenceOnVehicle0AtEnd, err := nextroute.NewMoveStops(
		solutionSequencePlanUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionSequencePlanUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous(),
		solutionSequencePlanUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	moveSequenceOnVehicle0Wrapped, err := nextroute.NewMoveStops(
		solutionSequencePlanUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Adding a net-zero sequence before the already planned sequence on vehicle 0
	// should not violate the constraint
	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle0AtStart); violated {
		t.Error("constraint is violated")
	}
	// Adding a net-zero sequence after the already planned sequence on vehicle 0
	// should not violate the constraint
	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle0AtEnd); violated {
		t.Error("constraint is violated")
	}
	// Adding a net-zero sequence around the already planned sequence on vehicle 0
	// should violate the constraint
	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle0Wrapped); !violated {
		t.Error("constraint is not violated")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[1].Last().Previous(),
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
	moveSequenceOnVehicle1AtEnd, err := nextroute.NewMoveStops(
		solutionSequencePlanUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Adding a net-zero sequence around the already planned stop on vehicle 1
	// should violate the constraint (level already at 2)
	if violated, _ := cnstr.EstimateIsViolated(moveSequenceOnVehicle1AtEnd); !violated {
		t.Error("constraint is not violated")
	}
}

func TestMaximumConstraint(t *testing.T) {
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

	delta := nextroute.NewStopExpression(
		"delta level",
		0,
	)

	maximum := nextroute.NewVehicleTypeValueExpression(
		"maximum level",
		1,
	)

	cnstr, err := nextroute.NewMaximum(
		delta,
		maximum,
	)
	if err != nil {
		t.Error(err)
	}

	for _, vt := range model.VehicleTypes() {
		if cnstr.Maximum().Value(vt, nil, nil) != 1 {
			t.Errorf(
				"maximum  is not correct, expected 1 got %v",
				cnstr.Maximum().Value(vt, nil, nil),
			)
		}
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
