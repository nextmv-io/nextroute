// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestSolutionVehicleImpl_Unplan(t *testing.T) {
	model, err := createModel(singleVehiclePlanSequenceModel())
	if err != nil {
		t.Fatal(err)
	}
	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}
	solutionVehicle := solution.Vehicles()[0]

	solutionPlanUnit := solution.UnPlannedPlanUnits().SolutionPlanUnits()[0]

	move := solutionVehicle.BestMove(context.Background(), solutionPlanUnit)
	if !move.IsExecutable() {
		t.Fatal("move should be executable")
	}
	unplannedCount := len(solution.UnPlannedPlanUnits().SolutionPlanUnits())

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("move should be planned")
	}
	if len(solution.UnPlannedPlanUnits().SolutionPlanUnits()) != unplannedCount-1 {
		t.Fatal("unplanned plan unit should be removed")
	}

	solutionPlanUnit = solution.UnPlannedPlanUnits().SolutionPlanUnits()[0]

	move = solutionVehicle.BestMove(context.Background(), solutionPlanUnit)
	if !move.IsExecutable() {
		t.Fatal("move should be executable")
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("move should be planned")
	}
	if len(solution.UnPlannedPlanUnits().SolutionPlanUnits()) != unplannedCount-2 {
		t.Fatal("unplanned plan unit should be removed")
	}
	unplanned, err := solutionVehicle.Unplan()
	if err != nil {
		t.Fatal(err)
	}
	if !unplanned {
		t.Fatal("solution vehicle should be unplanned")
	}
	if len(solution.UnPlannedPlanUnits().SolutionPlanUnits()) != unplannedCount {
		t.Fatal("unplanned plan unit should be added")
	}
}

func BenchmarkSolutionVehicleImpl_BestMove(b *testing.B) {
	model, err := createModel(singleVehiclePlanSequenceModel())
	if err != nil {
		b.Fatal(err)
	}
	solution, err := nextroute.NewSolution(model)
	if err != nil {
		b.Fatal(err)
	}
	solutionVehicle := solution.Vehicles()[0]

	solutionPlanUnit := solution.UnPlannedPlanUnits().SolutionPlanUnits()[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = solutionVehicle.BestMove(context.Background(), solutionPlanUnit)
	}
}
