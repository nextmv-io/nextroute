// © 2019-present nextmv.io inc

package nextroute_test

import (
	"testing"

	"github.com/nextmv-io/nextroute"
)

// Simply test that we can add a new expression objective to the model.
func TestAddExpressionObjective(t *testing.T) {
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

	if len(model.Objective().Terms()) != 0 {
		t.Error("model objective should be empty")
	}

	expressions := len(model.Expressions())

	e := nextroute.NewConstantExpression("test", 1.0)
	objective := nextroute.NewExpressionObjective(e)
	_, err = model.Objective().NewTerm(1.0, objective)
	if err != nil {
		t.Error(err)
	}

	if len(model.Objective().Terms()) != 1 {
		t.Error("model objective should have an objective")
	}

	if registered, ok := objective.(nextroute.RegisteredModelExpressions); ok {
		if len(registered.ModelExpressions()) != 1 {
			t.Error("objective should have an expression")
		}
	}

	if len(model.Expressions())-expressions != 1 {
		t.Error("expressions should increase by 1")
	}
}

// This test simulates a move on a solution and checks that the objective is
// being measured correctly based on a constant expression.
func TestExpressionObjective_EstimateDeltaValue(t *testing.T) {
	model, err := createModel(
		input(
			vehicleTypes("truck", "car", "bike"),
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

	// Use a constant expression for simplicity, so the insertion cost of a
	// plan unit is always the number of stops times this constant.
	expressionValue := 666.999
	e := nextroute.NewConstantExpression("test", expressionValue)
	objective := nextroute.NewExpressionObjective(e)
	_, err = model.Objective().NewTerm(1.0, objective)
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	planUnits := model.PlanStopsUnits()
	position, err := nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solution.SolutionPlanStopsUnit(planUnits[0]).SolutionStops()[0],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m1, err := nextroute.NewMoveStops(
		solution.SolutionPlanStopsUnit(planUnits[0]),
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}
	position1, err := nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solution.SolutionPlanStopsUnit(planUnits[3]).SolutionStops()[0],
		solution.SolutionPlanStopsUnit(planUnits[3]).SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err := nextroute.NewStopPosition(
		solution.SolutionPlanStopsUnit(planUnits[3]).SolutionStops()[0],
		solution.SolutionPlanStopsUnit(planUnits[3]).SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := nextroute.NewMoveStops(
		solution.SolutionPlanStopsUnit(planUnits[3]),
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		move nextroute.SolutionMoveStops
		want float64
	}{
		{
			name: "single stop added increments value by 1",
			move: m1,
			want: 1 * expressionValue,
		},
		{
			name: "sequence of 2 stops added increments value by 2",
			move: m2,
			want: 2 * expressionValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := objective.EstimateDeltaValue(tt.move)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
