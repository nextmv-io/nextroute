// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestNonExecutableMove(t *testing.T) {
	m := nextroute.NewNotExecutableMove()
	if m.IsExecutable() {
		t.Error("move should not be executable")
	}
	if m.IsExecutable() {
		t.Error("move should not be an improvement")
	}
	if m.ValueSeen() != 0 {
		t.Error("value seen should be 0")
	}
	planned, err := m.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if planned {
		t.Error("move should not be planned")
	}
}

func TestMove_TakeBest(t *testing.T) {
	model, err := createModel(singleVehiclePlanSingleStopsModel())
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	notExecutableMove := nextroute.NewNotExecutableMove()

	move1 := solution.BestMove(context.Background(), solution.UnPlannedPlanUnits().SolutionPlanUnits()[0])

	if !move1.IsExecutable() {
		t.Error("move1 is not executable")
	}

	bestMove := notExecutableMove.TakeBest(move1)

	if !bestMove.IsExecutable() {
		t.Error("best move is not executable")
	}
	if bestMove.Value() != move1.Value() {
		t.Error("best move value is not correct")
	}
	if bestMove.ValueSeen() != 1 {
		t.Errorf(
			"best move value seen is not correct, expected 1, got %d",
			bestMove.ValueSeen(),
		)
	}

	bestMove = bestMove.TakeBest(bestMove)

	if bestMove.ValueSeen() != 2 {
		t.Errorf(
			"best move value seen is not correct, expected 2, got %d",
			bestMove.ValueSeen(),
		)
	}

	bestMove = move1.TakeBest(notExecutableMove)

	if !bestMove.IsExecutable() {
		t.Error("best move is not executable")
	}
	if bestMove.Value() != move1.Value() {
		t.Error("best move value is not correct")
	}
	if bestMove.ValueSeen() != 1 {
		t.Errorf(
			"best move value seen is not correct, expected 1, got %d",
			bestMove.ValueSeen(),
		)
	}
}

func TestVehicleBestMoveSinglePlanUnit(t *testing.T) {
	model, err := createModel(singleVehiclePlanSingleStopsModel())
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	solutionVehicle := solution.SolutionVehicle(model.Vehicle(0))

	if solutionVehicle.IsZero() {
		t.Error("solutionVehicle is nil")
	}

	spss1 := solution.SolutionPlanStopsUnit(model.PlanStopsUnits()[0])
	spss2 := solution.SolutionPlanStopsUnit(model.PlanStopsUnits()[1])

	move1 := solutionVehicle.BestMove(context.Background(), spss1)

	if !move1.IsExecutable() {
		t.Error("move1 is not executable")
	}
	moveStops := move1.(nextroute.SolutionMoveStops)
	if moveStops.Vehicle().Index() != solutionVehicle.Index() {
		t.Error("vehicle index is not correct")
	}
	if len(moveStops.StopPositions()) != 1 {
		t.Error("stop positions length is not correct")
	}
	stopPosition := moveStops.StopPositions()[0]

	if stopPosition.Previous().Index() != solutionVehicle.First().Index() {
		t.Error("after index is not correct")
	}

	if stopPosition.Stop().Index() != spss1.SolutionStops()[0].Index() {
		t.Error("stop index is not correct")
	}

	if stopPosition.Next().Index() != solutionVehicle.Last().Index() {
		t.Error("before index is not correct")
	}

	if move1.Value() != 0.0 {
		t.Error("value is not correct, expected 0.0 (no objective)")
	}

	if move1.ValueSeen() != 1 {
		t.Error("value seen is not correct, expected 1")
	}

	planned, err := move1.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}

	if !planned {
		t.Error("move1 is not planned")
	}

	move2 := solutionVehicle.BestMove(context.Background(), spss2)

	if !move2.IsExecutable() {
		t.Error("move2 is not executable")
	}

	move2Stops := move2.(nextroute.SolutionMoveStops)

	if move2Stops.Vehicle().Index() != solutionVehicle.Index() {
		t.Error("vehicle index is not correct")
	}
	if len(move2Stops.StopPositions()) != 1 {
		t.Error("stop positions length is not correct")
	}

	if move2.Value() != 0.0 {
		t.Error("value is not correct, expected 0.0 (no objective)")
	}

	if move2.ValueSeen() != 1 {
		t.Error("value seen is not correct, expected 1")
	}
}

func TestVehicleBestMoveSequencePlanUnit(t *testing.T) {
	model, err := createModel(singleVehiclePlanSequenceModel())
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	solutionVehicle := solution.SolutionVehicle(model.Vehicle(0))

	if solutionVehicle.IsZero() {
		t.Error("solutionVehicle is nil")
	}

	spss1s2 := solution.UnPlannedPlanUnits().SolutionPlanUnits()[0].(nextroute.SolutionPlanStopsUnit)
	spss3s4 := solution.UnPlannedPlanUnits().SolutionPlanUnits()[1].(nextroute.SolutionPlanStopsUnit)

	move1 := solutionVehicle.BestMove(context.Background(), spss1s2)

	if !move1.IsExecutable() {
		t.Error("move1 is not executable")
	}
	move1Stops := move1.(nextroute.SolutionMoveStops)

	if move1Stops.Vehicle().Index() != solutionVehicle.Index() {
		t.Error("vehicle index is not correct")
	}
	if len(move1Stops.StopPositions()) != 2 {
		t.Error("stop positions length is not correct")
	}

	stopPosition := move1Stops.StopPositions()[0]

	if stopPosition.Previous().Index() != solutionVehicle.First().Index() {
		t.Error("after index is not correct")
	}

	if stopPosition.Stop().Index() != spss1s2.SolutionStops()[0].Index() {
		t.Error("stop index is not correct")
	}

	if stopPosition.Next().Index() != spss1s2.SolutionStops()[1].Index() {
		t.Error("before index is not correct")
	}

	if move1.Value() != 0.0 {
		t.Error("value is not correct, expected 0.0 (no objective)")
	}

	if move1.ValueSeen() != 1 {
		// depot -> s1 -> s2 -> depot
		t.Error("value seen is not correct, expected 1")
	}

	planned, err := move1.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}

	if !planned {
		t.Error("move1 is not planned")
	}

	move2 := solutionVehicle.BestMove(context.Background(), spss3s4)

	if !move2.IsExecutable() {
		t.Error("move2 is not executable")
	}

	move2Stops := move2.(nextroute.SolutionMoveStops)

	if move2Stops.Vehicle().Index() != solutionVehicle.Index() {
		t.Error("vehicle index is not correct")
	}
	if len(move2Stops.StopPositions()) != 2 {
		t.Error("stop positions length is not correct")
	}

	if move2.Value() != 0.0 {
		t.Error("value is not correct, expected 0.0 (no objective)")
	}

	if move2.ValueSeen() != 6 {
		// depot -> s3 -> s4 -> s1 -> s2 -> depot
		// depot -> s3 -> s1 -> s4 -> s2 -> depot
		// depot -> s3 -> s1 -> s2 -> s4 -> depot
		// depot -> s1 -> s3 -> s4 -> s2 -> depot
		// depot -> s1 -> s3 -> s2 -> s4 -> depot
		// depot -> s1 -> s2 -> s3 -> s4 -> depot
		t.Errorf("value seen is not correct, got %v, expected 6",
			move2.ValueSeen(),
		)
	}
}
