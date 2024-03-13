// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestMaximumWaitVehicleConstraint_EstimateIsViolated(t *testing.T) {
	// Define a start time and some earliest service times for the stops. The
	// first time will be too long to wait for, the second & third will be
	// possible but exhaust the accumulated wait max so that the last time
	// cannot be done anymore.
	startTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	serviceTimes := []time.Time{
		startTime.Add(1 * time.Hour),
		startTime.Add(8 * time.Minute),
		startTime.Add(20 * time.Minute),
		startTime.Add(30 * time.Minute),
	}

	// Define some stops with zero travel time and the vehicle.
	vehicle := Vehicle{
		Name:          "truck",
		StartLocation: Location{Lon: 0, Lat: 0},
		StartTime:     &startTime,
		Type:          "truck",
	}
	stops := []PlanSingleStop{
		{Stop: Stop{Name: "s1", Location: Location{Lon: 0, Lat: 0}}},
		{Stop: Stop{Name: "s2", Location: Location{Lon: 0, Lat: 0}}},
		{Stop: Stop{Name: "s3", Location: Location{Lon: 0, Lat: 0}}},
		{Stop: Stop{Name: "s4", Location: Location{Lon: 0, Lat: 0}}},
	}

	// Create the model and constraints
	model, err := createModel(
		input(
			vehicleTypes("truck"),
			[]Vehicle{vehicle},
			stops,
			[]PlanSequence{},
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	for i, stop := range model.Stops()[:4] {
		err = stop.SetEarliestStart(serviceTimes[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	maxVehicle := nextroute.NewVehicleTypeDurationExpression("maximum vehicle wait", 25*time.Minute)
	cnstr, err := nextroute.NewMaximumWaitVehicleConstraint(maxVehicle)
	if err != nil {
		t.Error(err)
	}
	err = model.AddConstraint(cnstr)
	if err != nil {
		t.Error(err)
	}

	// Create the solution + moves and check them.
	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}
	modelPlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() == 1
	})
	solutionPlanUnits := []nextroute.SolutionPlanStopsUnit{}
	for _, planUnit := range modelPlanUnits {
		solutionPlanUnits = append(solutionPlanUnits, solution.SolutionPlanStopsUnit(planUnit))
	}
	// Try to assign all stops and check success for expectation.
	success := []bool{true, false, false, true}
	for i, solutionPlanUnit := range solutionPlanUnits {
		position, err := nextroute.NewStopPosition(
			solution.Vehicles()[0].Last().Previous(),
			solutionPlanUnit.SolutionStops()[0],
			solution.Vehicles()[0].Last(),
		)
		if err != nil {
			t.Fatal(err)
		}
		move, err := nextroute.NewMoveStops(
			solutionPlanUnit,
			[]nextroute.StopPosition{position},
		)
		if err != nil {
			t.Fatal(err)
		}

		violated, _ := cnstr.EstimateIsViolated(move)
		if violated != success[i] {
			t.Errorf("move %v should be %v", i, success[i])
		}
		if !violated {
			_, err := move.Execute(context.Background())
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestMaximumWaitVehicleConstraint(t *testing.T) {
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

	maxVehicle := nextroute.NewVehicleTypeDurationExpression(
		"maximum vehicle duration",
		234*time.Second,
	)
	cnstr, err := nextroute.NewMaximumWaitVehicleConstraint(maxVehicle)
	if err != nil {
		t.Error(err)
	}

	for _, vt := range model.VehicleTypes() {
		if cnstr.Maximum().Value(vt, nil, nil) != 234 {
			t.Errorf(
				"maximum is not correct, expected 123 got %v",
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
