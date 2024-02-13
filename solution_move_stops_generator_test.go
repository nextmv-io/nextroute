package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
	sdkNextRoute "github.com/nextmv-io/sdk/nextroute"
)

func TestMoveGeneratorSingleStops(t *testing.T) {
	model, err := createModel(singleVehiclePlanSingleStopsModel())
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	v := solution.Vehicles()[0]
	planUnit0 := solution.UnPlannedPlanUnits().RandomElement().(sdkNextRoute.SolutionPlanStopsUnit)

	move0, err := nextroute.NewMoveStops(
		planUnit0,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.First(),
				planUnit0.SolutionStops()[0],
				v.Last(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit0,
		[]sdkNextRoute.SolutionMoveStops{move0},
	)

	move, err := nextroute.NewMoveStops(
		planUnit0,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.First(),
				planUnit0.SolutionStops()[0],
				v.Last(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !planned {
		t.Error("move should be planned")
	}

	planUnit1 := solution.UnPlannedPlanUnits().RandomElement().(sdkNextRoute.SolutionPlanStopsUnit)
	move10, err := nextroute.NewMoveStops(
		planUnit1,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.First(),
				planUnit1.SolutionStops()[0],
				planUnit0.SolutionStops()[0],
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	move11, err := nextroute.NewMoveStops(
		planUnit1,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				planUnit0.SolutionStops()[0],
				planUnit1.SolutionStops()[0],
				v.Last(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit1,
		[]sdkNextRoute.SolutionMoveStops{move10, move11},
	)

	move1, err := nextroute.NewMoveStops(
		planUnit1,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				planUnit0.SolutionStops()[0],
				planUnit1.SolutionStops()[0],
				v.Last(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	planned, err = move1.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Error("move should be planned")
	}

	planUnit2 := solution.UnPlannedPlanUnits().RandomElement().(sdkNextRoute.SolutionPlanStopsUnit)
	move20, err := nextroute.NewMoveStops(
		planUnit2,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.First(),
				planUnit2.SolutionStops()[0],
				planUnit0.SolutionStops()[0],
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	move21, err := nextroute.NewMoveStops(
		planUnit2,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				planUnit0.SolutionStops()[0],
				planUnit2.SolutionStops()[0],
				planUnit1.SolutionStops()[0],
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	move22, err := nextroute.NewMoveStops(
		planUnit2,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				planUnit1.SolutionStops()[0],
				planUnit2.SolutionStops()[0],
				v.Last(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit2,
		[]sdkNextRoute.SolutionMoveStops{move20, move21, move22},
	)
}

func TestMoveGeneratorSequenceStops(t *testing.T) {
	model, err := createModel(singleVehiclePlanSequenceModel())
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	v := solution.Vehicles()[0]
	planUnit0 := solution.UnPlannedPlanUnits().RandomElement().(sdkNextRoute.SolutionPlanStopsUnit)

	move, err := nextroute.NewMoveStops(
		planUnit0,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.First(),
				planUnit0.SolutionStops()[0],
				planUnit0.SolutionStops()[1],
			),
			nextroute.NewStopPosition(
				planUnit0.SolutionStops()[0],
				planUnit0.SolutionStops()[1],
				v.Last(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit0,
		[]sdkNextRoute.SolutionMoveStops{
			move,
		},
	)

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Error("move should be planned")
	}

	planUnit1 := solution.UnPlannedPlanUnits().RandomElement().(sdkNextRoute.SolutionPlanStopsUnit)

	m1, err := nextroute.NewMoveStops(
		planUnit1,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.First(),
				planUnit1.SolutionStops()[0],
				planUnit1.SolutionStops()[1],
			),
			nextroute.NewStopPosition(
				planUnit1.SolutionStops()[0],
				planUnit1.SolutionStops()[1],
				v.First().Next(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := nextroute.NewMoveStops(
		planUnit1,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.First(),
				planUnit1.SolutionStops()[0],
				v.First().Next(),
			),
			nextroute.NewStopPosition(
				v.First().Next(),
				planUnit1.SolutionStops()[1],
				v.First().Next().Next(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	m3, err := nextroute.NewMoveStops(
		planUnit1,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.First(),
				planUnit1.SolutionStops()[0],
				v.First().Next(),
			),
			nextroute.NewStopPosition(
				v.Last().Previous(),
				planUnit1.SolutionStops()[1],
				v.Last(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	m4, err := nextroute.NewMoveStops(
		planUnit1,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.First().Next(),
				planUnit1.SolutionStops()[0],
				planUnit1.SolutionStops()[1],
			),
			nextroute.NewStopPosition(
				planUnit1.SolutionStops()[0],
				planUnit1.SolutionStops()[1],
				v.First().Next().Next(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	m5, err := nextroute.NewMoveStops(
		planUnit1,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.First().Next(),
				planUnit1.SolutionStops()[0],
				v.First().Next().Next(),
			),
			nextroute.NewStopPosition(
				v.Last().Previous(),
				planUnit1.SolutionStops()[1],
				v.Last(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	m6, err := nextroute.NewMoveStops(
		planUnit1,
		sdkNextRoute.StopPositions{
			nextroute.NewStopPosition(
				v.Last().Previous(),
				planUnit1.SolutionStops()[0],
				planUnit1.SolutionStops()[1],
			),
			nextroute.NewStopPosition(
				planUnit1.SolutionStops()[0],
				planUnit1.SolutionStops()[1],
				v.Last(),
			),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit1,
		[]sdkNextRoute.SolutionMoveStops{m1, m2, m3, m4, m5, m6},
	)
}

func TestMoveGeneratorMultipleStops(t *testing.T) {
	model, err := createModel(input(
		vehicleTypes("truck"),
		vehicles("truck", depot(), 1),
		nil,
		planTripleSequence(),
	))
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	vehicle := solution.Vehicles()[0]

	count := 0
	quit := make(chan struct{})
	defer close(quit)

	planUnit := solution.UnPlannedPlanUnits().SolutionPlanUnits()[0].(sdkNextRoute.SolutionPlanStopsUnit)
	alloc := nextroute.NewPreAllocatedMoveContainer(planUnit)
	for move := range nextroute.SolutionMoveStopsGeneratorChannelTest(
		vehicle, planUnit, quit, planUnit.SolutionStops(), alloc,
	) {
		count++
		_ = move
	}
}

func testMoves(
	t *testing.T,
	vehicle sdkNextRoute.SolutionVehicle,
	planUnit sdkNextRoute.SolutionPlanStopsUnit,
	moves []sdkNextRoute.SolutionMoveStops,
) {
	count := 0
	quit := make(chan struct{})
	defer close(quit)

	alloc := nextroute.NewPreAllocatedMoveContainer(planUnit)
	nextroute.SolutionMoveStopsGeneratorTest(
		vehicle,
		planUnit,
		func(move sdkNextRoute.SolutionMoveStops) {
			if count == len(moves) {
				t.Errorf("more moves than expected")
			}
			if len(move.StopPositions()) != len(moves[count].StopPositions()) {
				t.Errorf("move %d is not correct, expected %v, got %v",
					count,
					moves[count],
					move,
				)
			}
			for i, stopPosition := range move.StopPositions() {
				if moves[count].StopPositions()[i].Previous() != stopPosition.Previous() {
					t.Errorf("move %d is not correct, stop position %v, previous stop",
						count,
						i,
					)
				}
				if moves[count].StopPositions()[i].Stop() != stopPosition.Stop() {
					t.Errorf("move %d is not correct, stop position %v, stop",
						count,
						i,
					)
				}
				if moves[count].StopPositions()[i].Next() != stopPosition.Next() {
					t.Errorf("move %d is not correct, stop position %v, next stop",
						count,
						i,
					)
				}
			}
			count++
		},
		planUnit.SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)
	if count != len(moves) {
		t.Errorf("less moves than expected, expected %d, got %d",
			len(moves),
			count,
		)
	}
}
