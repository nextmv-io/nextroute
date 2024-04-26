// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestNewInterleaveConstraint(t *testing.T) {
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
			nil,
			planPairSequences(),
		),
	)
	if err != nil {
		t.Error(err)
	}

	interleaveConstraint, err := nextroute.NewInterleaveConstraint()
	if err != nil {
		t.Error(err)
	}

	if interleaveConstraint == nil {
		t.Error("interleave constraint should not be nil")
	}

	err = interleaveConstraint.DisallowInterleaving(
		model.PlanUnits()[0],
		[]nextroute.ModelPlanUnit{
			model.PlanUnits()[1],
		},
	)
	if err != nil {
		t.Error(err)
	}

	err = interleaveConstraint.DisallowInterleaving(
		model.PlanUnits()[0],
		[]nextroute.ModelPlanUnit{},
	)
	if err != nil {
		t.Error(err)
	}
	err = interleaveConstraint.DisallowInterleaving(
		nil,
		[]nextroute.ModelPlanUnit{
			model.PlanUnits()[1],
		},
	)
	if err == nil {
		t.Error("expected error, target cannot be nil")
	}
	err = interleaveConstraint.DisallowInterleaving(
		model.PlanUnits()[0],
		nil,
	)
	if err == nil {
		t.Error("expected error, sources cannot be nil")
	}
	err = interleaveConstraint.DisallowInterleaving(
		model.PlanUnits()[0],
		[]nextroute.ModelPlanUnit{
			model.PlanUnits()[0],
			nil,
		},
	)
	if err == nil {
		t.Error("expected error, target cannot be nil")
	}
	err = interleaveConstraint.DisallowInterleaving(
		model.PlanUnits()[0],
		[]nextroute.ModelPlanUnit{
			model.PlanUnits()[0],
		},
	)
	if err == nil {
		t.Error("expected error, target is also in a target")
	}

	err = interleaveConstraint.DisallowInterleaving(
		model.PlanUnits()[0],
		[]nextroute.ModelPlanUnit{
			model.PlanUnits()[1],
			model.PlanUnits()[1],
		},
	)
	if err == nil {
		t.Error("expected error, sources has duplicates")
	}

	err = model.AddConstraint(interleaveConstraint)
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	if solution == nil {
		t.Error("solution should not be nil")
	}

	solution.Vehicles()[0].BestMove(
		context.Background(),
		solution.SolutionPlanUnit(model.PlanUnits()[0]),
	)
}
