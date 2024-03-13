// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

const latestEndName = "latest_end"

func TestLatestEndConstraintSingleStop(t *testing.T) {
	model, err := createModel(singleVehiclePlanSingleStopsModel())
	if err != nil {
		t.Error(err)
	}

	defaultLatestEnd := model.Epoch().Add(3 * time.Minute)

	latestEndTimeExpression := nextroute.NewStopTimeExpression(latestEndName, defaultLatestEnd)

	latestEnd, err := nextroute.NewLatestEnd(latestEndTimeExpression)
	if err != nil {
		t.Error(err)
	}

	if latestEnd.Latest().Index() != latestEndTimeExpression.Index() {
		t.Error("latest end time defaultExpression index is not correct")
	}

	for _, stop := range model.Stops() {
		if latestEnd.Latest().Time(stop) != defaultLatestEnd {
			t.Error("latest end time is not correct")
		}
	}

	err = model.AddConstraint(latestEnd)
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	move := solution.BestMove(context.Background(), solution.UnPlannedPlanUnits().SolutionPlanUnits()[0])

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}

	if !planned {
		t.Error("move should be planned")
	}

	move = solution.BestMove(context.Background(), solution.UnPlannedPlanUnits().SolutionPlanUnits()[1])

	if move.IsExecutable() {
		t.Error("move should not be executable, it should not fit on the vehicle")
	}
}

func TestLatestEndConstraintLastOnVehicle(t *testing.T) {
	model, err := createModel(singleVehiclePlanSingleStopsModel())
	if err != nil {
		t.Error(err)
	}

	name := "latest_end"
	defaultLatestEnd := model.Epoch().Add(2 * time.Minute)

	latestEndTimeExpression := nextroute.NewStopTimeExpression(name, defaultLatestEnd)

	latestEnd, err := nextroute.NewLatestEnd(latestEndTimeExpression)
	if err != nil {
		t.Error(err)
	}

	if latestEnd.Latest().Index() != latestEndTimeExpression.Index() {
		t.Error("latest end time defaultExpression index is not correct")
	}

	for _, stop := range model.Stops() {
		if latestEnd.Latest().Time(stop) != defaultLatestEnd {
			t.Error("latest end time is not correct")
		}
	}

	err = model.AddConstraint(latestEnd)
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	move := solution.BestMove(context.Background(), solution.UnPlannedPlanUnits().SolutionPlanUnits()[0])

	if move.IsExecutable() {
		t.Error("move should not be executable, it should not fit on the vehicle")
	}
}

func TestLatestEndObjectiveSingleStop(t *testing.T) {
	model, err := createModel(singleVehiclePlanSingleStopsModel())
	if err != nil {
		t.Error(err)
	}

	defaultLatestEnd := model.Epoch().Add(3 * time.Minute)

	latestEndTimeExpression := nextroute.NewStopTimeExpression(latestEndName, defaultLatestEnd)

	latestEnd, err := nextroute.NewLatestEnd(latestEndTimeExpression)
	if err != nil {
		t.Error(err)
	}

	if latestEnd.Latest().Index() != latestEndTimeExpression.Index() {
		t.Error("latest end time defaultExpression index is not correct")
	}

	for _, stop := range model.Stops() {
		if latestEnd.Latest().Time(stop) != defaultLatestEnd {
			t.Error("latest end time is not correct")
		}
	}

	_, err = model.Objective().NewTerm(1.0, latestEnd)
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	move := solution.BestMove(context.Background(), solution.UnPlannedPlanUnits().SolutionPlanUnits()[0])

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("move should be planned")
	}

	move = solution.BestMove(context.Background(), solution.UnPlannedPlanUnits().SolutionPlanUnits()[1])

	if !move.IsExecutable() {
		t.Error("move should be executable, it should fit on the vehicle with a penalty")
	}

	planned, err = move.Execute(context.Background())

	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("move should be planned")
	}

	if !common.WithinTolerance(solution.ObjectiveValue(model.Objective()), 34.539, 0.01) {
		t.Error("objective value is not correct, expected 34.539, got ", solution.ObjectiveValue(model.Objective()))
	}
}

func TestLatestEndConstraintSequence(t *testing.T) {
	model, err := createModel(
		input(
			vehicleTypes("truck"),
			vehicles(
				"truck",
				depot(),
				1,
			),
			nil,
			planPairSequences(),
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	defaultLatestEnd := model.Epoch().Add(4 * time.Minute)

	latestEndTimeExpression := nextroute.NewStopTimeExpression(latestEndName, defaultLatestEnd)

	latestEnd, err := nextroute.NewLatestEnd(latestEndTimeExpression)
	if err != nil {
		t.Fatal(err)
	}

	if latestEnd.Latest().Index() != latestEndTimeExpression.Index() {
		t.Fatal("latest end time defaultExpression index is not correct")
	}

	for _, stop := range model.Stops() {
		if latestEnd.Latest().Time(stop) != defaultLatestEnd {
			t.Fatal("latest end time is not correct")
		}
	}

	err = model.AddConstraint(latestEnd)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	move := solution.BestMove(context.Background(), solution.UnPlannedPlanUnits().SolutionPlanUnits()[0])

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("move should be planned")
	}

	move = solution.BestMove(context.Background(), solution.UnPlannedPlanUnits().SolutionPlanUnits()[0])

	planned, err = move.Execute(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if planned {
		t.Fatal("move should not be planned")
	}

	if move.IsExecutable() {
		t.Fatal("move should not be executable, unit should not fit on the vehicle")
	}
}

func TestLatestEndObjectiveSequence(t *testing.T) {
	model, err := createModel(
		input(
			vehicleTypes("truck"),
			vehicles(
				"truck",
				depot(),
				1,
			),
			nil,
			planPairSequences(),
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	defaultLatestEnd := model.Epoch().Add(4 * time.Minute)

	latestEndTimeExpression := nextroute.NewStopTimeExpression(latestEndName, defaultLatestEnd)

	latestEnd, err := nextroute.NewLatestEnd(latestEndTimeExpression)
	if err != nil {
		t.Fatal(err)
	}

	if latestEnd.Latest().Index() != latestEndTimeExpression.Index() {
		t.Fatal("latest end time defaultExpression index is not correct")
	}

	for _, stop := range model.Stops() {
		if latestEnd.Latest().Time(stop) != defaultLatestEnd {
			t.Fatal("latest end time is not correct")
		}
	}

	_, err = model.Objective().NewTerm(1.0, latestEnd)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	move := solution.BestMove(context.Background(), solution.UnPlannedPlanUnits().SolutionPlanUnits()[0])

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("move should be planned")
	}

	move = solution.BestMove(context.Background(), solution.UnPlannedPlanUnits().SolutionPlanUnits()[0])

	if !move.IsExecutable() {
		t.Fatal("move should be executable")
	}

	planned, err = move.Execute(context.Background())

	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("move should be planned")
	}

	if !common.WithinTolerance(solution.ObjectiveValue(model.Objective()), 0.867, 0.01) {
		t.Error("objective value is not correct, expected 0.867, got ", solution.ObjectiveValue(model.Objective()))
	}
}
