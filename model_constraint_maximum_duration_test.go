// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestMaximumDurationConstraint_EstimateIsViolated(t *testing.T) {
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

	maximum := nextroute.NewVehicleTypeDurationExpression(
		"maximum duration",
		3*time.Minute,
	)

	cnstr, err := nextroute.NewMaximumTravelDurationConstraint(maximum)
	if err != nil {
		t.Error(err)
	}

	err = model.AddConstraint(cnstr)
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}
	singleStopPlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() == 1
	})

	sequencePlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() > 1
	})

	_ = sequencePlanUnits

	solutionSingleStopPlanUnit0 := solution.SolutionPlanStopsUnit(singleStopPlanUnits[0])
	position, err := nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionSingleStopPlanUnit0.SolutionStops()[0],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	moveSingleOnVehicle, err := nextroute.NewMoveStops(
		solutionSingleStopPlanUnit0,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle); violated {
		t.Error("move should not be violated")
	}

	planned, err := moveSingleOnVehicle.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("move should be planned")
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
	moveSingleOnVehicle, err = nextroute.NewMoveStops(
		solutionSingleStopPlanUnit1,
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	if violated, _ := cnstr.EstimateIsViolated(moveSingleOnVehicle); !violated {
		t.Error("move should be violated")
	}

	// TODO add sequence test
}

func TestMaximumDurationConstraint(t *testing.T) {
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

	maximum := nextroute.NewVehicleTypeDurationExpression(
		"maximum duration",
		123*time.Second,
	)

	cnstr, err := nextroute.NewMaximumDurationConstraint(
		maximum,
	)
	if err != nil {
		t.Error(err)
	}

	for _, vt := range model.VehicleTypes() {
		if cnstr.Maximum().Value(vt, nil, nil) != 123 {
			t.Errorf(
				"maximum  is not correct, expected 123 got %v",
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
