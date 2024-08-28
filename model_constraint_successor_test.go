// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestNewSuccessorConstraintConstraint_EstimateIsViolated(t *testing.T) {
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
			nil,
		),
	)
	if err != nil {
		t.Error(err)
	}

	successorConstraint, err := nextroute.NewSuccessorConstraint()
	if err != nil {
		t.Error(err)
	}

	stop0 := model.Stops()[0]
	stop2 := model.Stops()[2]

	err = successorConstraint.DisallowSuccessors(stop0, []nextroute.ModelStop{stop2})
	if err != nil {
		t.Error(err)
	}

	err = model.AddConstraint(successorConstraint)
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	solutionStop0 := solution.SolutionStop(stop0)
	solutionStop2 := solution.SolutionStop(stop2)
	vehicle0 := solution.Vehicles()[0]

	stopPosition, err := nextroute.NewStopPosition(vehicle0.First(), solutionStop0, vehicle0.Last())
	if err != nil {
		t.Error(err)
	}
	move, err := nextroute.NewMoveStops(solutionStop0.PlanStopsUnit(), []nextroute.StopPosition{stopPosition})
	if err != nil {
		t.Error(err)
	}

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}

	if !planned {
		t.Error("expected move to be planned")
	}

	stopPosition, err = nextroute.NewStopPosition(solutionStop0, solutionStop2, vehicle0.Last())
	if err != nil {
		t.Error(err)
	}
	move, err = nextroute.NewMoveStops(solutionStop2.PlanStopsUnit(), []nextroute.StopPosition{stopPosition})
	if err != nil {
		t.Error(err)
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}

	if planned {
		t.Error("expected move to not be planned")
	}
}

func TestSuccessorMovesGenerated(t *testing.T) {
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
			nil,
		),
	)
	if err != nil {
		t.Error(err)
	}

	successorConstraint, err := nextroute.NewSuccessorConstraint()
	if err != nil {
		t.Error(err)
	}

	stop0 := model.Stops()[0]
	stop1 := model.Stops()[1]
	stop2 := model.Stops()[2]

	err = successorConstraint.DisallowSuccessors(stop0, []nextroute.ModelStop{stop2})
	if err != nil {
		t.Error(err)
	}

	err = model.AddConstraint(successorConstraint)
	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	vehicle0 := solution.Vehicles()[0]

	solutionStop0 := solution.SolutionStop(stop0)

	position, err := nextroute.NewStopPosition(
		vehicle0.First(),
		solutionStop0,
		vehicle0.Last(),
	)
	if err != nil {
		t.Error(err)
	}
	move, err := nextroute.NewMoveStops(
		solutionStop0.PlanStopsUnit(),
		[]nextroute.StopPosition{
			position,
		},
	)
	if err != nil {
		t.Error(err)
	}

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("expected move to be planned")
	}

	solutionStop2 := solution.SolutionStop(stop2)

	count := 0
	alloc := nextroute.NewPreAllocatedMoveContainer(solutionStop2.PlanStopsUnit())
	nextroute.SolutionMoveStopsGeneratorTest(
		vehicle0,
		solutionStop2.PlanStopsUnit(),
		func(move nextroute.SolutionMoveStops) {
			if move.StopPositions()[0].Next() != solutionStop0 {
				t.Errorf("expected stop 0 to be next")
			}
			count++
		},
		solutionStop2.PlanStopsUnit().SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)
	if count != 1 {
		t.Errorf("expected 1 move, got %d", count)
	}

	solutionStop1 := solution.SolutionStop(stop1)

	count = 0
	alloc = nextroute.NewPreAllocatedMoveContainer(solutionStop1.PlanStopsUnit())
	nextroute.SolutionMoveStopsGeneratorTest(
		vehicle0,
		solutionStop1.PlanStopsUnit(),
		func(_ nextroute.SolutionMoveStops) {
			count++
		},
		solutionStop1.PlanStopsUnit().SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)
	if count != 2 {
		t.Errorf("expected 2 moves, got %d", count)
	}
}

func TestMultipleDisallowedSuccessors(t *testing.T) {
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
			nil,
		),
	)
	if err != nil {
		t.Error(err)
	}

	successorConstraint, err := nextroute.NewSuccessorConstraint()
	if err != nil {
		t.Error(err)
	}

	stop0 := model.Stops()[0]
	stop1 := model.Stops()[1]
	stop2 := model.Stops()[2]

	err = successorConstraint.DisallowSuccessors(stop0, []nextroute.ModelStop{stop1, stop2})

	if err != nil {
		t.Error(err)
	}

	err = model.AddConstraint(successorConstraint)

	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	vehicle0 := solution.Vehicles()[0]

	solutionStop0 := solution.SolutionStop(stop0)

	position, err := nextroute.NewStopPosition(
		vehicle0.First(),
		solutionStop0,
		vehicle0.Last(),
	)
	if err != nil {
		t.Error(err)
	}
	move, err := nextroute.NewMoveStops(
		solutionStop0.PlanStopsUnit(),
		[]nextroute.StopPosition{
			position,
		},
	)
	if err != nil {
		t.Error(err)
	}

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("expected move to be planned")
	}
	solutionStop1 := solution.SolutionStop(stop1)

	count := 0
	alloc := nextroute.NewPreAllocatedMoveContainer(solutionStop1.PlanStopsUnit())
	nextroute.SolutionMoveStopsGeneratorTest(
		vehicle0,
		solutionStop1.PlanStopsUnit(),
		func(move nextroute.SolutionMoveStops) {
			if move.StopPositions()[0].Next() != solutionStop0 {
				t.Errorf("expected stop 0 to be next")
			}
			count++
		},
		solutionStop1.PlanStopsUnit().SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)
	if count != 1 {
		t.Errorf("expected 1 move, got %d", count)
	}

	position, err = nextroute.NewStopPosition(
		vehicle0.First(),
		solutionStop1,
		vehicle0.First().Next(),
	)
	if err != nil {
		t.Error(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionStop1.PlanStopsUnit(),
		[]nextroute.StopPosition{
			position,
		},
	)
	if err != nil {
		t.Error(err)
	}

	_, err = move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}

	solutionStop2 := solution.SolutionStop(stop2)

	count = 0
	alloc = nextroute.NewPreAllocatedMoveContainer(solutionStop2.PlanStopsUnit())
	nextroute.SolutionMoveStopsGeneratorTest(
		vehicle0,
		solutionStop2.PlanStopsUnit(),
		func(move nextroute.SolutionMoveStops) {
			if move.StopPositions()[0].Previous() == solutionStop0 {
				t.Errorf("stop 2 is disallowed successor of stop 0")
			}
			count++
		},
		solutionStop2.PlanStopsUnit().SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)
	if count != 2 {
		t.Errorf("expected 1 move, got %d", count)
	}
}

func TestUnplanSuccessorConstrained(t *testing.T) {
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
			nil,
		),
	)
	if err != nil {
		t.Error(err)
	}

	successorConstraint, err := nextroute.NewSuccessorConstraint()
	if err != nil {
		t.Error(err)
	}

	stop0 := model.Stops()[0]
	stop1 := model.Stops()[1]
	stop2 := model.Stops()[2]

	err = successorConstraint.DisallowSuccessors(stop0, []nextroute.ModelStop{stop1})

	if err != nil {
		t.Error(err)
	}

	err = model.AddConstraint(successorConstraint)

	if err != nil {
		t.Error(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	vehicle0 := solution.Vehicles()[0]

	solutionStop0 := solution.SolutionStop(stop0)

	position, err := nextroute.NewStopPosition(
		vehicle0.First(),
		solutionStop0,
		vehicle0.Last(),
	)
	if err != nil {
		t.Error(err)
	}
	move, err := nextroute.NewMoveStops(
		solutionStop0.PlanStopsUnit(),
		[]nextroute.StopPosition{
			position,
		},
	)
	if err != nil {
		t.Error(err)
	}

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("expected move to be planned")
	}
	solutionStop2 := solution.SolutionStop(stop2)

	position, err = nextroute.NewStopPosition(
		vehicle0.Last().Previous(),
		solutionStop2,
		vehicle0.Last(),
	)
	if err != nil {
		t.Error(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionStop2.PlanStopsUnit(),
		[]nextroute.StopPosition{
			position,
		},
	)
	if err != nil {
		t.Error(err)
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("expected move to be planned")
	}
	solutionStop1 := solution.SolutionStop(stop1)

	position, err = nextroute.NewStopPosition(
		vehicle0.Last().Previous(),
		solutionStop1,
		vehicle0.Last(),
	)
	if err != nil {
		t.Error(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionStop1.PlanStopsUnit(),
		[]nextroute.StopPosition{
			position,
		},
	)
	if err != nil {
		t.Error(err)
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !planned {
		t.Error("expected move to be planned")
	}

	unplanned, err := solutionStop2.PlanStopsUnit().UnPlan()
	if err != nil {
		t.Error(err)
	}
	if unplanned {
		t.Error("expected stop 2 to not be unplanned, it would violate the constraint")
	}
}
