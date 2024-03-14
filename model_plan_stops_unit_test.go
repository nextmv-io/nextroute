// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestPlanMultipleStops(t *testing.T) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}

	warehouse, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	warehouse.SetID("warehouse")

	s1, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s1.SetID("s1")

	s2, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s2.SetID("s2")

	s3, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s3.SetID("s3")

	s4, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s4.SetID("s4")

	dag1 := nextroute.NewDirectedAcyclicGraph()

	if err := dag1.AddArc(s1, s2); err != nil {
		t.Fatal(err)
	}
	if err := dag1.AddArc(s1, s3); err != nil {
		t.Fatal(err)
	}
	if err := dag1.AddArc(s3, s4); err != nil {
		t.Fatal(err)
	}

	ms1, err := model.NewPlanMultipleStops(nextroute.ModelStops{
		s1,
		s2,
		s3,
		s4,
	}, dag1)
	if err != nil {
		t.Fatal(err)
	}

	if ms1 == nil {
		t.Fatal("ms1 should not be nil")
	}
	if len(ms1.Stops()) != 4 {
		t.Fatal("ms1 should have 4 stops")
	}
	if slices.IndexFunc(ms1.Stops(), func(stop nextroute.ModelStop) bool {
		return stop == s1
	}) == -1 {
		t.Fatal("ms1 should have s1")
	}
	if slices.IndexFunc(ms1.Stops(), func(stop nextroute.ModelStop) bool {
		return stop == s2
	}) == -1 {
		t.Fatal("ms1 should have s2")
	}
	if slices.IndexFunc(ms1.Stops(), func(stop nextroute.ModelStop) bool {
		return stop == s3
	}) == -1 {
		t.Fatal("ms1 should have s3")
	}
	if slices.IndexFunc(ms1.Stops(), func(stop nextroute.ModelStop) bool {
		return stop == s4
	}) == -1 {
		t.Fatal("ms1 should have s4")
	}

	vehicleType, err := model.NewVehicleType(
		nextroute.NewTimeIndependentDurationExpression(
			nextroute.NewDurationExpression(
				"travelDuration",
				nextroute.NewHaversineExpression(),
				common.Second,
			),
		),
		nextroute.NewConstantDurationExpression("processing_time", 1*time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}

	vehicle, err := model.NewVehicle(vehicleType, model.Epoch(), warehouse, warehouse)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	if solution == nil {
		t.Fatal("solution should not be nil")
	}

	if solution.UnPlannedPlanUnits().Size() != 1 {
		t.Fatal("solution should have 1 un-planned plan units")
	}

	solutionVehicle := solution.SolutionVehicle(vehicle)

	_ = solutionVehicle

	move := solution.BestMove(
		context.Background(),
		solution.UnPlannedPlanUnits().SolutionPlanUnit(ms1),
	)

	if move == nil {
		t.Fatal("move should not be nil")
	}
	if !move.IsExecutable() {
		t.Fatal("move should be executable")
	}

	success, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !success {
		t.Fatal("move should be successful")
	}
}
