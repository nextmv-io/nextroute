// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestDirectPrecedenceConstraint_EstimateIsViolated(t *testing.T) {
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
			nil,
		),
	)
	if err != nil {
		t.Error(err)
	}

	directPrecedence, err := nextroute.NewDirectPrecedencesConstraint()
	if err != nil {
		t.Error(err)
	}

	stop0 := model.Stops()[0]
	stop2 := model.Stops()[2]

	err = directPrecedence.DisallowSuccessors(stop0, []nextroute.ModelStop{stop2})
	if err != nil {
		t.Error(err)
	}

	err = model.AddConstraint(directPrecedence)
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	solutionStop0 := solution.SolutionStop(stop0)
	solutionStop2 := solution.SolutionStop(stop2)
	vehicle0 := solution.Vehicles()[0]

	stopPosition, err := nextroute.NewStopPosition(vehicle0.First(), solutionStop0, vehicle0.Last())
	if err != nil {
		t.Error(err)
	}
	move, err := nextroute.NewMoveStops(solutionStop0.PlanStopsUnit(), []nextroute.StopPosition{stopPosition})
	if err != nil {
		t.Error(err)
	}

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}

	if !planned {
		t.Error("expected move to be planned")
	}

	stopPosition, err = nextroute.NewStopPosition(solutionStop0, solutionStop2, vehicle0.Last())
	if err != nil {
		t.Error(err)
	}
	move, err = nextroute.NewMoveStops(solutionStop2.PlanStopsUnit(), []nextroute.StopPosition{stopPosition})
	if err != nil {
		t.Error(err)
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}

	if planned {
		t.Error("expected move to not be planned")
	}
}
