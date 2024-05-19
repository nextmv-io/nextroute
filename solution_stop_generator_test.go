// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestSolutionStopGeneratorSingleStop(t *testing.T) {
	model, err := createModel(singleVehiclePlanSingleStopsModel())
	if err != nil {
		t.Fatal(err)
	}
	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	unplannedPlanUnits := solution.UnPlannedPlanUnits().SolutionPlanUnits()

	if solution.UnPlannedPlanUnits().Size() != 3 {
		t.Fatalf("number of unplanned plan Units is not correct,"+
			" expected 3 got %v",
			solution.UnPlannedPlanUnits().Size(),
		)
	}
	position, err := nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		unplannedPlanUnits[0].(nextroute.SolutionPlanStopsUnit).SolutionStops()[0],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err := nextroute.NewMoveStops(
		unplannedPlanUnits[0].(nextroute.SolutionPlanStopsUnit),
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("empty vehicle, startAtFirst=true, endAtLast=true",
		func(t *testing.T) {
			testMove(
				t,
				move,
				true,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().Last(),
				},
			)
		},
	)

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !planned {
		t.Fatal("move should be planned")
	}

	t.Run("next single stop, startAtFirst=true, endAtLast=true",
		func(t *testing.T) {
			position, err = nextroute.NewStopPosition(
				move.Vehicle().First(),
				unplannedPlanUnits[1].(nextroute.SolutionPlanStopsUnit).SolutionStops()[0],
				move.Vehicle().First().Next(),
			)
			if err != nil {
				t.Fatal(err)
			}
			move, err = nextroute.NewMoveStops(
				unplannedPlanUnits[1].(nextroute.SolutionPlanStopsUnit),
				[]nextroute.StopPosition{position},
			)
			if err != nil {
				t.Fatal(err)
			}
			testMove(
				t,
				move,
				true,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().Last().Previous(),
					move.Vehicle().Last(),
				},
			)

			planned, err = move.Execute(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			if !planned {
				t.Fatal("move should be planned")
			}
		},
	)

	t.Run("next single stop, 2 stops on vehicle, first position",
		func(t *testing.T) {
			position, err = nextroute.NewStopPosition(
				move.Vehicle().First(),
				unplannedPlanUnits[2].(nextroute.SolutionPlanStopsUnit).SolutionStops()[0],
				move.Vehicle().First().Next(),
			)
			if err != nil {
				t.Fatal(err)
			}
			move, err = nextroute.NewMoveStops(
				unplannedPlanUnits[2].(nextroute.SolutionPlanStopsUnit),
				[]nextroute.StopPosition{position},
			)
			if err != nil {
				t.Fatal(err)
			}

			testMove(
				t,
				move,
				false,
				false,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().First().Next(),
				},
			)
			testMove(
				t,
				move,
				false,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().First().Next(),
					move.Vehicle().First().Next().Next(),
					move.Vehicle().Last(),
				},
			)
			testMove(
				t,
				move,
				true,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().First().Next(),
					move.Vehicle().First().Next().Next(),
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("next single stop, 2 stops on vehicle, second position",
		func(t *testing.T) {
			position, err = nextroute.NewStopPosition(
				move.Vehicle().First().Next(),
				unplannedPlanUnits[2].(nextroute.SolutionPlanStopsUnit).SolutionStops()[0],
				move.Vehicle().First().Next().Next(),
			)
			if err != nil {
				t.Fatal(err)
			}
			move, err = nextroute.NewMoveStops(
				unplannedPlanUnits[2].(nextroute.SolutionPlanStopsUnit),
				[]nextroute.StopPosition{position},
			)
			if err != nil {
				t.Fatal(err)
			}
			testMove(
				t,
				move,
				false,
				false,
				nextroute.SolutionStops{
					move.Vehicle().First().Next(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().First().Next().Next(),
				},
			)
			testMove(
				t,
				move,
				false,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First().Next(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().First().Next().Next(),
					move.Vehicle().Last(),
				},
			)
			testMove(
				t,
				move,
				true,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.Vehicle().First().Next(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().First().Next().Next(),
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("next single stop, 2 stops on vehicle, last position",
		func(t *testing.T) {
			position, err = nextroute.NewStopPosition(
				move.Vehicle().Last().Previous(),
				unplannedPlanUnits[2].(nextroute.SolutionPlanStopsUnit).SolutionStops()[0],
				move.Vehicle().Last(),
			)
			if err != nil {
				t.Fatal(err)
			}
			move, err = nextroute.NewMoveStops(
				unplannedPlanUnits[2].(nextroute.SolutionPlanStopsUnit),
				[]nextroute.StopPosition{position},
			)
			if err != nil {
				t.Fatal(err)
			}
			testMove(
				t,
				move,
				false,
				false,
				nextroute.SolutionStops{
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().Last(),
				},
			)
			testMove(
				t,
				move,
				false,
				true,
				nextroute.SolutionStops{
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().Last(),
				},
			)
			testMove(
				t,
				move,
				true,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.Vehicle().First().Next(),
					move.Vehicle().First().Next().Next(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().Last(),
				},
			)
		},
	)
}

var global uint64

func BenchmarkSolutionStopGenerator(b *testing.B) {
	model, err := createModel(singleVehiclePlanSingleStopsModel())
	if err != nil {
		b.Fatal(err)
	}
	solution, err := nextroute.NewSolution(model)
	if err != nil {
		b.Fatal(err)
	}
	unplannedPlanUnits := solution.UnPlannedPlanUnits().SolutionPlanUnits()
	position, err := nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		unplannedPlanUnits[0].(nextroute.SolutionPlanStopsUnit).SolutionStops()[0],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		b.Fatal(err)
	}
	move, err := nextroute.NewMoveStops(
		unplannedPlanUnits[0].(nextroute.SolutionPlanStopsUnit),
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		b.Fatal(err)
	}
	planned, err := move.Execute(context.Background())
	if err != nil {
		b.Fatal(err)
	}

	if !planned {
		b.Fatal("move should be planned")
	}
	position, err = nextroute.NewStopPosition(
		move.Vehicle().First(),
		unplannedPlanUnits[1].(nextroute.SolutionPlanStopsUnit).SolutionStops()[0],
		move.Vehicle().First().Next(),
	)
	if err != nil {
		b.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		unplannedPlanUnits[1].(nextroute.SolutionPlanStopsUnit),
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		b.Fatal(err)
	}
	planned, err = move.Execute(context.Background())
	if err != nil {
		b.Fatal(err)
	}

	if !planned {
		b.Fatal("move should be planned")
	}
	position, err = nextroute.NewStopPosition(
		move.Vehicle().First(),
		unplannedPlanUnits[2].(nextroute.SolutionPlanStopsUnit).SolutionStops()[0],
		move.Vehicle().First().Next(),
	)
	if err != nil {
		b.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		unplannedPlanUnits[2].(nextroute.SolutionPlanStopsUnit),
		[]nextroute.StopPosition{position},
	)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	sum := 0
	for i := 0; i < b.N; i++ {
		generator := nextroute.NewSolutionStopGenerator(move, true, true)
		for solutionStop := generator.Next(); !solutionStop.IsZero(); solutionStop = generator.Next() {
			sum++
		}
	}
	global = uint64(sum)
}

func TestSolutionStopGeneratorSequence(t *testing.T) {
	model, err := createModel(singleVehiclePlanSequenceModel())
	if err != nil {
		t.Fatal(err)
	}

	if len(model.Stops()) != 5 {
		t.Fatalf("expected 5 stops, got %v", len(model.Stops()))
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	unplannedPlanUnits := solution.UnPlannedPlanUnits().SolutionPlanUnits()

	if len(unplannedPlanUnits) != 2 {
		t.Fatalf("expected 2 unplanned plan Units, got %v", len(unplannedPlanUnits))
	}

	vehicle := solution.Vehicles()[0]

	upu := unplannedPlanUnits[0].(nextroute.SolutionPlanStopsUnit)
	position1, err := nextroute.NewStopPosition(
		vehicle.First(),
		upu.SolutionStops()[0],
		upu.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err := nextroute.NewStopPosition(
		upu.SolutionStops()[0],
		upu.SolutionStops()[1],
		vehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err := nextroute.NewMoveStops(
		upu,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("empty vehicle sequence, startAtFirst=false, endAtLast=false",
		func(t *testing.T) {
			testMove(
				t,
				move,
				false,
				false,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("empty vehicle sequence, startAtFirst=true, endAtLast=false",
		func(t *testing.T) {
			testMove(
				t,
				move,
				true,
				false,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("empty vehicle sequence, startAtFirst=true, endAtLast=true",
		func(t *testing.T) {
			testMove(
				t,
				move,
				true,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("move is not planned")
	}
	upu = unplannedPlanUnits[1].(nextroute.SolutionPlanStopsUnit)
	position1, err = nextroute.NewStopPosition(
		move.Vehicle().First(),
		upu.SolutionStops()[0],
		upu.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		upu.SolutionStops()[0],
		upu.SolutionStops()[1],
		move.Vehicle().First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		upu,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add at start vehicle sequence, startAtFirst=true, endAtLast=true",
		func(t *testing.T) {
			testMove(
				t,
				move,
				true,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().First().Next(),
					move.Vehicle().Last().Previous(),
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("add at start vehicle sequence, startAtFirst=false, endAtLast=true",
		func(t *testing.T) {
			testMove(
				t,
				move,
				false,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().First().Next(),
					move.Vehicle().Last().Previous(),
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("add at start vehicle sequence, startAtFirst=false, endAtLast=false",
		func(t *testing.T) {
			testMove(
				t,
				move,
				false,
				false,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().First().Next(),
				},
			)
		},
	)
	upu = unplannedPlanUnits[1].(nextroute.SolutionPlanStopsUnit)
	position1, err = nextroute.NewStopPosition(
		vehicle.First(),
		upu.SolutionStops()[0],
		vehicle.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		vehicle.Last().Previous(),
		upu.SolutionStops()[1],
		vehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		upu,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add at start and split vehicle sequence, startAtFirst=true, endAtLast=true",
		func(t *testing.T) {
			testMove(
				t,
				move,
				true,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().First().Next(),
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("add at start and split vehicle sequence, startAtFirst=false, endAtLast=true",
		func(t *testing.T) {
			testMove(
				t,
				move,
				false,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().First().Next(),
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("add at start and split vehicle sequence, startAtFirst=false, endAtLast=false",
		func(t *testing.T) {
			testMove(
				t,
				move,
				false,
				false,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().First().Next(),
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)
	upu = unplannedPlanUnits[1].(nextroute.SolutionPlanStopsUnit)
	position1, err = nextroute.NewStopPosition(
		move.Vehicle().First().Next(),
		upu.SolutionStops()[0],
		move.Vehicle().Last().Previous(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		vehicle.Last().Previous(),
		upu.SolutionStops()[1],
		vehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		upu,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add middle and split vehicle sequence, startAtFirst=true, endAtLast=true",
		func(t *testing.T) {
			testMove(
				t,
				move,
				true,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.Vehicle().First().Next(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("add middle and split vehicle sequence, startAtFirst=false, endAtLast=true",
		func(t *testing.T) {
			testMove(
				t,
				move,
				false,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First().Next(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("add middle and split vehicle sequence, startAtFirst=false, endAtLast=false",
		func(t *testing.T) {
			testMove(
				t,
				move,
				false,
				false,
				nextroute.SolutionStops{
					move.Vehicle().First().Next(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)

	upu = unplannedPlanUnits[1].(nextroute.SolutionPlanStopsUnit)
	position1, err = nextroute.NewStopPosition(
		move.Vehicle().Last().Previous(),
		upu.SolutionStops()[0],
		move.PlanStopsUnit().SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		upu.SolutionStops()[0],
		upu.SolutionStops()[1],
		vehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		upu,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add to end vehicle sequence, startAtFirst=true, endAtLast=true",
		func(t *testing.T) {
			testMove(
				t,
				move,
				true,
				true,
				nextroute.SolutionStops{
					move.Vehicle().First(),
					move.Vehicle().First().Next(),
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("add to end vehicle sequence, startAtFirst=false, endAtLast=true",
		func(t *testing.T) {
			testMove(
				t,
				move,
				false,
				true,
				nextroute.SolutionStops{
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)

	t.Run("add to end vehicle sequence, startAtFirst=false, endAtLast=false",
		func(t *testing.T) {
			testMove(
				t,
				move,
				false,
				false,
				nextroute.SolutionStops{
					move.Vehicle().Last().Previous(),
					move.PlanStopsUnit().SolutionStops()[0],
					move.PlanStopsUnit().SolutionStops()[1],
					move.Vehicle().Last(),
				},
			)
		},
	)
}

func testMove(
	t *testing.T,
	move nextroute.SolutionMoveStops,
	startAtFirst bool,
	endAtLast bool,
	expected nextroute.SolutionStops,
) {
	count := 0

	generator := nextroute.NewSolutionStopGenerator(move, startAtFirst, endAtLast)

	for stop := generator.Next(); !stop.IsZero(); stop = generator.Next() {
		if count == len(expected) {
			t.Fatalf("too many stops, did not expect %v", stop)
		}
		if stop != expected[count] {
			t.Fatalf("stop is not correct at position %v, expected %v got %v",
				count,
				expected[count],
				stop.Index(),
			)
		}
		count++
	}

	if count != len(expected) {
		t.Fatalf("not enough stops, got %v, expected %v",
			count,
			len(expected),
		)
	}
}
