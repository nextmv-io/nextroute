// © 2019-present nextmv.io inc

package nextroute_test

import (
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestTravelDurationObjective_EstimateDeltaValue(_ *testing.T) {
	// TODO implement
}

func TestTravelDurationObjective(t *testing.T) {
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

	travelDurationObjective := nextroute.NewTravelDurationObjective()

	if len(model.Objective().Terms()) != 0 {
		t.Error("model objective should be empty")
	}

	_, err = model.Objective().NewTerm(1.0, travelDurationObjective)
	if err != nil {
		t.Error(err)
	}

	if len(model.Objective().Terms()) != 1 {
		t.Error("model objective should have an objective")
	}
}
