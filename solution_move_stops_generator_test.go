// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
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
	planUnit0 := solution.UnPlannedPlanUnits().RandomElement().(nextroute.SolutionPlanStopsUnit)
	position, err := nextroute.NewStopPosition(
		v.First(),
		planUnit0.SolutionStops()[0],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move0, err := nextroute.NewMoveStops(
		planUnit0,
		nextroute.StopPositions{position},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit0,
		[]nextroute.SolutionMoveStops{move0},
	)
	position, err = nextroute.NewStopPosition(
		v.First(),
		planUnit0.SolutionStops()[0],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err := nextroute.NewMoveStops(
		planUnit0,
		nextroute.StopPositions{position},
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

	planUnit1 := solution.UnPlannedPlanUnits().RandomElement().(nextroute.SolutionPlanStopsUnit)
	position, err = nextroute.NewStopPosition(
		v.First(),
		planUnit1.SolutionStops()[0],
		planUnit0.SolutionStops()[0],
	)
	if err != nil {
		t.Fatal(err)
	}
	move10, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position},
	)
	if err != nil {
		t.Fatal(err)
	}
	position, err = nextroute.NewStopPosition(
		planUnit0.SolutionStops()[0],
		planUnit1.SolutionStops()[0],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move11, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit1,
		[]nextroute.SolutionMoveStops{move10, move11},
	)
	position, err = nextroute.NewStopPosition(
		planUnit0.SolutionStops()[0],
		planUnit1.SolutionStops()[0],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move1, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position},
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

	planUnit2 := solution.UnPlannedPlanUnits().RandomElement().(nextroute.SolutionPlanStopsUnit)
	position, err = nextroute.NewStopPosition(
		v.First(),
		planUnit2.SolutionStops()[0],
		planUnit0.SolutionStops()[0],
	)
	if err != nil {
		t.Fatal(err)
	}
	move20, err := nextroute.NewMoveStops(
		planUnit2,
		nextroute.StopPositions{position},
	)
	if err != nil {
		t.Fatal(err)
	}
	position, err = nextroute.NewStopPosition(
		planUnit0.SolutionStops()[0],
		planUnit2.SolutionStops()[0],
		planUnit1.SolutionStops()[0],
	)
	if err != nil {
		t.Fatal(err)
	}
	move21, err := nextroute.NewMoveStops(
		planUnit2,
		nextroute.StopPositions{position},
	)
	if err != nil {
		t.Fatal(err)
	}
	position, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit2.SolutionStops()[0],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move22, err := nextroute.NewMoveStops(
		planUnit2,
		nextroute.StopPositions{position},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit2,
		[]nextroute.SolutionMoveStops{move20, move21, move22},
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
	planUnit0 := solution.UnPlannedPlanUnits().RandomElement().(nextroute.SolutionPlanStopsUnit)
	position1, err := nextroute.NewStopPosition(
		v.First(),
		planUnit0.SolutionStops()[0],
		planUnit0.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err := nextroute.NewStopPosition(
		planUnit0.SolutionStops()[0],
		planUnit0.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err := nextroute.NewMoveStops(
		planUnit0,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit0,
		[]nextroute.SolutionMoveStops{
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

	planUnit1 := solution.UnPlannedPlanUnits().RandomElement().(nextroute.SolutionPlanStopsUnit)
	position1, err = nextroute.NewStopPosition(
		v.First(),
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
		v.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m1, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	position1, err = nextroute.NewStopPosition(
		v.First(),
		planUnit1.SolutionStops()[0],
		v.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		v.First().Next(),
		planUnit1.SolutionStops()[1],
		v.First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	position1, err = nextroute.NewStopPosition(
		v.First(),
		planUnit1.SolutionStops()[0],
		v.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		v.Last().Previous(),
		planUnit1.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m3, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	position1, err = nextroute.NewStopPosition(
		v.First().Next(),
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
		v.First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m4, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	position1, err = nextroute.NewStopPosition(
		v.First().Next(),
		planUnit1.SolutionStops()[0],
		v.First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		v.Last().Previous(),
		planUnit1.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m5, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	position1, err = nextroute.NewStopPosition(
		v.Last().Previous(),
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m6, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit1,
		[]nextroute.SolutionMoveStops{m1, m2, m3, m4, m5, m6},
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

	planUnit := solution.UnPlannedPlanUnits().SolutionPlanUnits()[0].(nextroute.SolutionPlanStopsUnit)
	alloc := nextroute.NewPreAllocatedMoveContainer(planUnit)
	for move := range nextroute.SolutionMoveStopsGeneratorChannelTest(
		vehicle, planUnit, quit, planUnit.SolutionStops(), alloc,
	) {
		count++
		_ = move
	}
}

func TestMoveGeneratorMustBeNeighbors1(t *testing.T) {
	model, err := createModel(singleVehiclePlanSequenceModel())
	if err != nil {
		t.Fatal(err)
	}

	for _, planUnit := range model.PlanUnits() {
		planStopsUnit := planUnit.(nextroute.ModelPlanStopsUnit)
		err = planStopsUnit.DirectedAcyclicGraph().AddDirectArc(
			planStopsUnit.Stops()[0],
			planStopsUnit.Stops()[1],
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	v := solution.Vehicles()[0]
	planUnit0 := solution.UnPlannedPlanUnits().RandomElement().(nextroute.SolutionPlanStopsUnit)
	position1, err := nextroute.NewStopPosition(
		v.First(),
		planUnit0.SolutionStops()[0],
		planUnit0.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err := nextroute.NewStopPosition(
		planUnit0.SolutionStops()[0],
		planUnit0.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err := nextroute.NewMoveStops(
		planUnit0,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit0,
		[]nextroute.SolutionMoveStops{
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

	planUnit1 := solution.UnPlannedPlanUnits().RandomElement().(nextroute.SolutionPlanStopsUnit)
	position1, err = nextroute.NewStopPosition(
		v.First(),
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
		v.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m1, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	position1, err = nextroute.NewStopPosition(
		v.Last().Previous(),
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit1,
		[]nextroute.SolutionMoveStops{m1, m2},
	)
}

func TestMoveGeneratorMustBeNeighbors2(t *testing.T) {
	model, err := createModel(singleVehiclePlanSequenceModel())
	if err != nil {
		t.Fatal(err)
	}

	modelPlanUnit0 := model.PlanUnits()[0].(nextroute.ModelPlanStopsUnit)
	modelPlanUnit1 := model.PlanUnits()[1].(nextroute.ModelPlanStopsUnit)
	err = modelPlanUnit1.DirectedAcyclicGraph().AddDirectArc(
		modelPlanUnit1.Stops()[0],
		modelPlanUnit1.Stops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	v := solution.Vehicles()[0]
	planUnit0 := solution.SolutionPlanUnit(modelPlanUnit0).(nextroute.SolutionPlanStopsUnit)
	position1, err := nextroute.NewStopPosition(
		v.First(),
		planUnit0.SolutionStops()[0],
		planUnit0.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err := nextroute.NewStopPosition(
		planUnit0.SolutionStops()[0],
		planUnit0.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err := nextroute.NewMoveStops(
		planUnit0,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit0,
		[]nextroute.SolutionMoveStops{
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

	planUnit1 := solution.SolutionPlanUnit(modelPlanUnit1).(nextroute.SolutionPlanStopsUnit)
	position1, err = nextroute.NewStopPosition(
		v.First(),
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
		v.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m1, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	position1, err = nextroute.NewStopPosition(
		v.First().Next(),
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
		v.First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	position1, err = nextroute.NewStopPosition(
		v.Last().Previous(),
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m3, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit1,
		[]nextroute.SolutionMoveStops{m1, m2, m3},
	)
}

func TestMoveGeneratorMustBeNeighbors3(t *testing.T) {
	model, err := createModel(singleVehiclePlanSequenceModel())
	if err != nil {
		t.Fatal(err)
	}

	modelPlanUnit0 := model.PlanUnits()[0].(nextroute.ModelPlanStopsUnit)
	modelPlanUnit1 := model.PlanUnits()[1].(nextroute.ModelPlanStopsUnit)
	err = modelPlanUnit0.DirectedAcyclicGraph().AddDirectArc(
		modelPlanUnit0.Stops()[0],
		modelPlanUnit0.Stops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	v := solution.Vehicles()[0]
	planUnit0 := solution.SolutionPlanUnit(modelPlanUnit0).(nextroute.SolutionPlanStopsUnit)
	position1, err := nextroute.NewStopPosition(
		v.First(),
		planUnit0.SolutionStops()[0],
		planUnit0.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err := nextroute.NewStopPosition(
		planUnit0.SolutionStops()[0],
		planUnit0.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err := nextroute.NewMoveStops(
		planUnit0,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit0,
		[]nextroute.SolutionMoveStops{
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

	planUnit1 := solution.SolutionPlanUnit(modelPlanUnit1).(nextroute.SolutionPlanStopsUnit)
	position1, err = nextroute.NewStopPosition(
		v.First(),
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
		v.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m1, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	position1, err = nextroute.NewStopPosition(
		v.First(),
		planUnit1.SolutionStops()[0],
		v.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		v.Last().Previous(),
		planUnit1.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	position1, err = nextroute.NewStopPosition(
		v.Last().Previous(),
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		planUnit1.SolutionStops()[0],
		planUnit1.SolutionStops()[1],
		v.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	m3, err := nextroute.NewMoveStops(
		planUnit1,
		nextroute.StopPositions{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	testMoves(
		t,
		v,
		planUnit1,
		[]nextroute.SolutionMoveStops{m1, m2, m3},
	)
}

func testMoves(
	t *testing.T,
	vehicle nextroute.SolutionVehicle,
	planUnit nextroute.SolutionPlanStopsUnit,
	moves []nextroute.SolutionMoveStops,
) {
	count := 0
	quit := make(chan struct{})
	defer close(quit)

	alloc := nextroute.NewPreAllocatedMoveContainer(planUnit)
	nextroute.SolutionMoveStopsGeneratorTest(
		vehicle,
		planUnit,
		func(move nextroute.SolutionMoveStops) {
			if count == len(moves) {
				t.Errorf("more moves than expected")
			}
			if len(move.StopPositions()) != len(moves[count].StopPositions()) {
				t.Errorf("move %d is not correct, expected %v, got %v",
					count+1,
					moves[count],
					move,
				)
			}
			for i, stopPosition := range move.StopPositions() {
				if moves[count].StopPositions()[i].Previous() != stopPosition.Previous() {
					t.Errorf("move %d is not correct, stop position %v, previous stop",
						count+1,
						i,
					)
				}
				if moves[count].StopPositions()[i].Stop() != stopPosition.Stop() {
					t.Errorf("move %d is not correct, stop position %v, stop",
						count+1,
						i,
					)
				}
				if moves[count].StopPositions()[i].Next() != stopPosition.Next() {
					t.Errorf("move %d is not correct, stop position %v, next stop",
						count+1,
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

func TestSolutionMoveStopsGeneratorInterleaved(
	t *testing.T,
) {
	model, planUnits, _ := createModel2(t)
	xPlanUnit := planUnits[0]
	yPlanUnit := planUnits[1]
	iPlanUnit := planUnits[2]
	jPlanUnit := planUnits[3]

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	v := solution.Vehicles()[0]

	xSolutionPlanUnit := solution.SolutionPlanUnit(xPlanUnit)

	move := v.BestMove(context.Background(), xSolutionPlanUnit)
	if move == nil {
		t.Fatal("move should not be nil")
	}
	if !move.IsExecutable() {
		t.Fatal("move should be executable")
	}
	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("move should be planned")
	}

	// i can not be interleaved anywhere in 'warehouse - a - b - c - d - e - warehouse'
	// but can go first or last
	iSolutionPlanUnit := solution.SolutionPlanUnit(iPlanUnit).(nextroute.SolutionPlanStopsUnit)

	moveCount := 0

	alloc := nextroute.NewPreAllocatedMoveContainer(iSolutionPlanUnit)
	nextroute.SolutionMoveStopsGeneratorTest(
		v,
		iSolutionPlanUnit,
		func(move nextroute.SolutionMoveStops) {
			moveCount += 1
			if moveCount > 2 {
				t.Fatal("move count should not exceed 2")
			}
			if moveCount == 1 {
				if move.StopPositions()[0].Previous() != v.First() {
					t.Fatal("previous stop should be the first stop")
				}
			}
			if moveCount == 2 {
				if move.StopPositions()[0].Next() != v.Last() {
					t.Fatal("next stop should be the last stop")
				}
			}
		},
		iSolutionPlanUnit.SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)

	// j can go anywhere in 'warehouse - a - b - c - d - e - warehouse'
	jSolutionPlanUnit := solution.SolutionPlanUnit(jPlanUnit).(nextroute.SolutionPlanStopsUnit)

	moveCount = 0

	alloc = nextroute.NewPreAllocatedMoveContainer(jSolutionPlanUnit)
	nextroute.SolutionMoveStopsGeneratorTest(
		v,
		jSolutionPlanUnit,
		func(move nextroute.SolutionMoveStops) {
			moveCount += 1
			if moveCount > 6 {
				t.Fatal("move count should not exceed 6")
			}
		},
		jSolutionPlanUnit.SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)
	if moveCount != 6 {
		t.Fatal("move count should be 6")
	}

	hPlanUnit := yPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[1]

	hSolutionPlanUnit := solution.SolutionPlanUnit(hPlanUnit).(nextroute.SolutionPlanStopsUnit)

	// h can go only at start and end of 'warehouse - a - b - c - d - e - warehouse'
	// h is part of Y and Y can not be interleaved with X and a,b,c,d and e are
	// part of X so h can only go at start and end
	moveCount = 0

	alloc = nextroute.NewPreAllocatedMoveContainer(hSolutionPlanUnit)
	nextroute.SolutionMoveStopsGeneratorTest(
		v,
		hSolutionPlanUnit,
		func(move nextroute.SolutionMoveStops) {
			moveCount += 1
		},
		hSolutionPlanUnit.SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)
	if moveCount != 2 {
		t.Fatal("move for h count should be 2, it is", moveCount)
	}

	fgPlanUnit := yPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[0]

	fgSolutionPlanUnit := solution.SolutionPlanUnit(fgPlanUnit).(nextroute.SolutionPlanStopsUnit)

	// fg can can not interleave a-..- e in in 'warehouse - a - b - c - d - e - warehouse'
	// and fg can not be interleaved by any a-..-e, so 2 moves

	moveCount = 0

	alloc = nextroute.NewPreAllocatedMoveContainer(fgSolutionPlanUnit)
	nextroute.SolutionMoveStopsGeneratorTest(
		v,
		fgSolutionPlanUnit,
		func(move nextroute.SolutionMoveStops) {
			moveCount += 1
		},
		fgSolutionPlanUnit.SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)
	if moveCount != 2 {
		t.Fatal("move for fg count should be 2, it is", moveCount)
	}

	hFirstPosition, err := nextroute.NewStopPosition(v.First(), hSolutionPlanUnit.SolutionStops()[0], v.First().Next())
	if err != nil {
		t.Fatal(err)
	}

	moveH, err := nextroute.NewMoveStops(
		hSolutionPlanUnit,
		nextroute.StopPositions{
			hFirstPosition,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	planned, err = moveH.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("move should be planned")
	}
	//fmt.Println(common.Map(v.SolutionStops(), func(stop nextroute.SolutionStop) string {
	//	return stop.ModelStop().ID()
	//}))

	moveCount = 0

	nextroute.SolutionMoveStopsGeneratorTest(
		v,
		fgSolutionPlanUnit,
		func(move nextroute.SolutionMoveStops) {
			moveCount += 1
		},
		fgSolutionPlanUnit.SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)
	if moveCount != 3 {
		t.Fatal("move for fg count should be 3, it is", moveCount)
	}

	unplanned, err := hSolutionPlanUnit.UnPlan()
	if err != nil {
		t.Fatal(err)
	}
	if !unplanned {
		t.Fatal("h should be unplanned")
	}

	hLastPosition, err := nextroute.NewStopPosition(v.Last().Previous(), hSolutionPlanUnit.SolutionStops()[0], v.Last())
	if err != nil {
		t.Fatal(err)
	}

	moveH, err = nextroute.NewMoveStops(
		hSolutionPlanUnit,
		nextroute.StopPositions{
			hLastPosition,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	planned, err = moveH.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("move should be planned")
	}

	moveCount = 0

	nextroute.SolutionMoveStopsGeneratorTest(
		v,
		fgSolutionPlanUnit,
		func(move nextroute.SolutionMoveStops) {
			moveCount += 1
		},
		fgSolutionPlanUnit.SolutionStops(),
		alloc,
		func() bool {
			return false
		},
	)
	if moveCount != 4 {
		t.Fatal("move for fg count should be 4, it is", moveCount)
	}
}
