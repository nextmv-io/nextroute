// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestSolutionMoveStops(t *testing.T) {
	model, err := createModel(singleVehiclePlanSequenceModel())
	if err != nil {
		t.Fatal(err)
	}

	s1 := model.Stops()[0]
	s2 := model.Stops()[1]
	s3 := model.Stops()[2]
	s4 := model.Stops()[3]

	mixItems := map[nextroute.ModelStop]nextroute.MixItem{
		s1: {
			Name:     "avocados",
			Quantity: 1,
		},
		s2: {
			Name:     "avocados",
			Quantity: -1,
		},
		s3: {
			Name:     "grapes",
			Quantity: 1,
		},
		s4: {
			Name:     "grapes",
			Quantity: -1,
		},
	}

	noMixConstraint, err := nextroute.NewNoMixConstraint(mixItems)
	if err != nil {
		t.Fatal(err)
	}

	err = model.AddConstraint(noMixConstraint)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	// Let's add a plan unit to the vehicle
	v := solution.Vehicles()[0]
	planUnit0 := solution.UnPlannedPlanUnits().RandomElement().(nextroute.SolutionPlanStopsUnit)
	position1, err := nextroute.NewStopPosition(
		v.First(),
		planUnit0.SolutionStops()[0],
		planUnit0.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err := nextroute.NewStopPosition(
		planUnit0.SolutionStops()[0],
		planUnit0.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move0, err := nextroute.NewMoveStops(
		planUnit0,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	planned, err := move0.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}

	// Let's add the other fruit in the vehicle while there is still the other
	// fruit in the vehicle (which is not allowed)
	planUnit1 := solution.UnPlannedPlanUnits().RandomElement().(nextroute.SolutionPlanStopsUnit)
	position3, err := nextroute.NewStopPosition(
		v.First(),
		planUnit1.SolutionStops()[0],
		v.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position4, err := nextroute.NewStopPosition(
		v.First().Next(),
		planUnit1.SolutionStops()[1],
		v.First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move1, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position3, position4},
	)
	if err != nil {
		t.Fatal(err)
	}
	planned, err = move1.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if planned {
		t.Fatal("expected move to not be planned")
	}
}
