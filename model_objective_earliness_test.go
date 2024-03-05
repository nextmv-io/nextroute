// © 2019-present nextmv.io inc

package nextroute_test

import (
	"testing"

	"github.com/nextmv-io/nextroute"
)

// Simply test that we can add a new earliness objective to the model.
func TestAddEarlinessObjective(t *testing.T) {
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

	targetTimeExpression := nextroute.NewStopTimeExpression("target_time", model.MaxTime())
	factorExpression := nextroute.NewStopExpression(
		"earliness_penalty_factor",
		1.0,
	)

	earlinessObjective, err := nextroute.NewEarlinessObjective(
		targetTimeExpression,
		factorExpression,
		nextroute.OnArrival,
	)
	if err != nil {
		t.Error(err)
	}
	_, err = model.Objective().NewTerm(1.0, earlinessObjective)
	if err != nil {
		t.Error(err)
	}

	if len(model.Objective().Terms()) != 1 {
		t.Error("model objective should have an objective")
	}
}

// This test simulates a move on a solution and checks that the objective is
// being measured correctly based on a constant expression.
func TestEarlinessObjective_EstimateDeltaValue(_ *testing.T) {
	// TODO: write test here
}
