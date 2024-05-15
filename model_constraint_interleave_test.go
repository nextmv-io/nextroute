// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
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
}

func TestInterleaveConstraint0(t *testing.T) {
	model, planUnits, modelStops := createModel1(t, false)
	xPlanUnit := planUnits[0]
	aPlanUnit := xPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[0]

	yPlanUnit := planUnits[1]
	cPlanUnit := yPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[0]
	dPlanUnit := yPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[1]

	a := modelStops[0]
	c := modelStops[2]
	d := modelStops[3]

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	solutionVehicle := solution.Vehicles()[0]

	aSolutionStop := solution.SolutionStop(a)
	cSolutionStop := solution.SolutionStop(c)
	dSolutionStop := solution.SolutionStop(d)

	aSolutionPlanUnit := solution.SolutionPlanUnit(aPlanUnit).(nextroute.SolutionPlanStopsUnit)
	cSolutionPlanUnit := solution.SolutionPlanUnit(cPlanUnit).(nextroute.SolutionPlanStopsUnit)
	dSolutionPlanUnit := solution.SolutionPlanUnit(dPlanUnit).(nextroute.SolutionPlanStopsUnit)

	// F - c - L
	stopPositionc, err := nextroute.NewStopPosition(
		solutionVehicle.First(),
		cSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err := nextroute.NewMoveStops(
		cSolutionPlanUnit,
		nextroute.StopPositions{stopPositionc},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}

	// F - c - d - L
	stopPositiond, err := nextroute.NewStopPosition(
		cSolutionStop,
		dSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		dSolutionPlanUnit,
		nextroute.StopPositions{stopPositiond},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}

	// F - c - a - d - L
	stopPositiona, err := nextroute.NewStopPosition(
		cSolutionStop,
		aSolutionStop,
		dSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		aSolutionPlanUnit,
		nextroute.StopPositions{stopPositiona},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}
}

func TestInterleaveConstraint1(t *testing.T) {
	model, planUnits, modelStops := createModel1(t, true)

	xPlanUnit := planUnits[0]
	aPlanUnit := xPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[0]
	bPlanUnit := xPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[1]

	yPlanUnit := planUnits[1]
	cPlanUnit := yPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[0]
	dPlanUnit := yPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[1]

	zPlanUnit := planUnits[2]
	ePlanUnit := zPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[0]
	fPlanUnit := zPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[1]

	a := modelStops[0]
	b := modelStops[1]
	c := modelStops[2]
	d := modelStops[3]
	e := modelStops[4]
	f := modelStops[5]

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	solutionVehicle := solution.Vehicles()[0]

	aSolutionStop := solution.SolutionStop(a)
	bSolutionStop := solution.SolutionStop(b)
	cSolutionStop := solution.SolutionStop(c)
	dSolutionStop := solution.SolutionStop(d)
	eSolutionStop := solution.SolutionStop(e)
	fSolutionStop := solution.SolutionStop(f)

	aSolutionPlanUnit := solution.SolutionPlanUnit(aPlanUnit).(nextroute.SolutionPlanStopsUnit)
	bSolutionPlanUnit := solution.SolutionPlanUnit(bPlanUnit).(nextroute.SolutionPlanStopsUnit)

	cSolutionPlanUnit := solution.SolutionPlanUnit(cPlanUnit).(nextroute.SolutionPlanStopsUnit)
	dSolutionPlanUnit := solution.SolutionPlanUnit(dPlanUnit).(nextroute.SolutionPlanStopsUnit)

	eSolutionPlanUnit := solution.SolutionPlanUnit(ePlanUnit).(nextroute.SolutionPlanStopsUnit)
	fSolutionPlanUnit := solution.SolutionPlanUnit(fPlanUnit).(nextroute.SolutionPlanStopsUnit)

	// F - a - L
	stopPositiona, err := nextroute.NewStopPosition(
		solutionVehicle.First(),
		aSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err := nextroute.NewMoveStops(
		aSolutionPlanUnit,
		nextroute.StopPositions{stopPositiona},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}

	// F - a - b - L
	stopPositionb, err := nextroute.NewStopPosition(
		aSolutionStop,
		bSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		bSolutionPlanUnit,
		nextroute.StopPositions{stopPositionb},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}

	// F - a - b - c - L
	stopPositionc, err := nextroute.NewStopPosition(
		bSolutionStop,
		cSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		cSolutionPlanUnit,
		nextroute.StopPositions{stopPositionc},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}

	// F - a - b - c - d - L
	stopPositiond, err := nextroute.NewStopPosition(
		cSolutionStop,
		dSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		dSolutionPlanUnit,
		nextroute.StopPositions{stopPositiond},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}

	// F - e - a - b - c - d - L allowed
	stopPositione, err := nextroute.NewStopPosition(
		solutionVehicle.First(),
		eSolutionStop,
		aSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		eSolutionPlanUnit,
		nextroute.StopPositions{stopPositione},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	// F - a - e - b - c - d - L not allowed
	stopPositione, err = nextroute.NewStopPosition(
		aSolutionStop,
		eSolutionStop,
		bSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		eSolutionPlanUnit,
		nextroute.StopPositions{stopPositione},
	)
	if err != nil {
		t.Fatal(err)
	}

	if move.IsExecutable() {
		t.Fatal("expected move not to be executable")
	}

	// F - a - b - e  - c - d - L allowed
	stopPositione, err = nextroute.NewStopPosition(
		bSolutionStop,
		eSolutionStop,
		cSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		eSolutionPlanUnit,
		nextroute.StopPositions{stopPositione},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	// F - a - b - c - e - d - L not allowed
	stopPositione, err = nextroute.NewStopPosition(
		cSolutionStop,
		eSolutionStop,
		dSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		eSolutionPlanUnit,
		nextroute.StopPositions{stopPositione},
	)
	if err != nil {
		t.Fatal(err)
	}

	if move.IsExecutable() {
		t.Fatal("expected move to be not executable")
	}

	// F - a - b  - c - d - e - L allowed
	stopPositione, err = nextroute.NewStopPosition(
		dSolutionStop,
		eSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		eSolutionPlanUnit,
		nextroute.StopPositions{stopPositione},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	_ = fSolutionStop
	_ = fSolutionPlanUnit
}

func TestInterleaveConstraint2(t *testing.T) {
	model, planUnits, modelStops := createModel2(t)
	xPlanUnit := planUnits[0]
	yPlanUnit := planUnits[1]
	iPlanUnit := planUnits[2]

	a := modelStops[0]
	b := modelStops[1]
	c := modelStops[2]
	d := modelStops[3]
	e := modelStops[4]
	f := modelStops[5]
	g := modelStops[6]
	h := modelStops[7]
	i := modelStops[8]

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	solutionVehicle := solution.Vehicles()[0]

	iSolutionStop := solution.SolutionStop(i)
	iSolutionPlanUnit := solution.SolutionPlanUnit(iPlanUnit)

	stopPosition, err := nextroute.NewStopPosition(
		solutionVehicle.First(),
		iSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err := nextroute.NewMoveStops(
		iSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPosition},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}

	// we now have: F - i - L

	eSolutionStop := solution.SolutionStop(e)

	ePlanUnit := xPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[2]
	if len(ePlanUnit.(nextroute.ModelPlanStopsUnit).Stops()) != 1 {
		t.Fatal("expected plan unit to have 1 stop (e)")
	}
	eSolutionPlanUnit := solution.SolutionPlanUnit(ePlanUnit)

	stopPosition, err = nextroute.NewStopPosition(
		solutionVehicle.First(),
		eSolutionStop,
		solutionVehicle.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		eSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPosition},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}

	// we now have: F - e - i - L
	aSolutionStop := solution.SolutionStop(a)
	bSolutionStop := solution.SolutionStop(b)

	abPlanUnit := xPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[0]
	if len(abPlanUnit.(nextroute.ModelPlanStopsUnit).Stops()) != 2 {
		t.Fatal("expected plan unit to have 2 stops (a, b)")
	}
	abSolutionPlanUnit := solution.SolutionPlanUnit(abPlanUnit)

	// Check F - a - b - e - i - L is accepted
	stopPositiona, err := nextroute.NewStopPosition(
		solutionVehicle.First(),
		aSolutionStop,
		bSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositionb, err := nextroute.NewStopPosition(
		aSolutionStop,
		bSolutionStop,
		solutionVehicle.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		abSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositiona, stopPositionb},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}
	unplanned, err := abSolutionPlanUnit.UnPlan()
	if err != nil {
		t.Fatal(err)
	}
	if !unplanned {
		t.Fatal("expected plan unit to be unplanned")
	}
	// Check F - a - e - b  - i - L is accepted
	stopPositiona, err = nextroute.NewStopPosition(
		solutionVehicle.First(),
		aSolutionStop,
		eSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositionb, err = nextroute.NewStopPosition(
		eSolutionStop,
		bSolutionStop,
		iSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		abSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositiona, stopPositionb},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}
	unplanned, err = abSolutionPlanUnit.UnPlan()
	if err != nil {
		t.Fatal(err)
	}
	if !unplanned {
		t.Fatal("expected plan unit to be unplanned")
	}
	// Check F - a - e - i - b - L is rejected
	stopPositiona, err = nextroute.NewStopPosition(
		solutionVehicle.First(),
		aSolutionStop,
		eSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositionb, err = nextroute.NewStopPosition(
		iSolutionStop,
		bSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		abSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositiona, stopPositionb},
	)
	if err != nil {
		t.Fatal(err)
	}
	if move.IsExecutable() {
		t.Fatal(
			"expected move to be not executable, i can not be interleaved with a-b" +
				", a before i and b after i is not allowed",
		)
	}
	// Check F - e - i - a - b - L is rejected
	stopPositiona, err = nextroute.NewStopPosition(
		solutionVehicle.Last().Previous(),
		aSolutionStop,
		bSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositionb, err = nextroute.NewStopPosition(
		aSolutionStop,
		bSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		abSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositiona, stopPositionb},
	)
	if err != nil {
		t.Fatal(err)
	}
	if move.IsExecutable() {
		t.Fatal(
			"expected move to be not executable, i can not be interleaved with a-b" +
				", a-b not allowed after i as e is already before i",
		)
	}
	// Check F - e - a - b - i - L is accepted
	stopPositiona, err = nextroute.NewStopPosition(
		eSolutionStop,
		aSolutionStop,
		bSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositionb, err = nextroute.NewStopPosition(
		aSolutionStop,
		bSolutionStop,
		iSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		abSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositiona, stopPositionb},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}

	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}
	// We now have F - e - a - b - i - L
	cSolutionStop := solution.SolutionStop(c)
	dSolutionStop := solution.SolutionStop(d)

	cdPlanUnit := xPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[1]
	if len(cdPlanUnit.(nextroute.ModelPlanStopsUnit).Stops()) != 2 {
		t.Fatal("expected plan unit to have 2 stops (c, d)")
	}
	cdSolutionPlanUnit := solution.SolutionPlanUnit(cdPlanUnit)

	// Check F - c - e - a - b - i - d - L is rejected
	stopPositionc, err := nextroute.NewStopPosition(
		solutionVehicle.First(),
		cSolutionStop,
		eSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositiond, err := nextroute.NewStopPosition(
		iSolutionStop,
		dSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		cdSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionc, stopPositiond},
	)
	if err != nil {
		t.Fatal(err)
	}
	if move.IsExecutable() {
		t.Fatal("expected move not to be executable, d not allowed after i as b is already before i")
	}
	// Check F - e - a - b - i - c - d - L is rejected
	stopPositionc, err = nextroute.NewStopPosition(
		iSolutionStop,
		cSolutionStop,
		dSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositiond, err = nextroute.NewStopPosition(
		cSolutionStop,
		dSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		cdSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionc, stopPositiond},
	)
	if err != nil {
		t.Fatal(err)
	}
	if move.IsExecutable() {
		t.Fatal("expected move not to be executable, c - d not allowed after i as b is already before i")
	}
	// Check F - e - a - b - c - d - i - L is accepted
	stopPositionc, err = nextroute.NewStopPosition(
		bSolutionStop,
		cSolutionStop,
		dSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositiond, err = nextroute.NewStopPosition(
		cSolutionStop,
		dSolutionStop,
		iSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		cdSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionc, stopPositiond},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !move.IsExecutable() {
		t.Fatal("expected move to be executable, c - d is before i")
	}
	planned, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !planned {
		t.Fatal("expected move to be planned")
	}

	// We now have F - e - a - b - c - d - i - L

	fSolutionStop := solution.SolutionStop(f)
	gSolutionStop := solution.SolutionStop(g)

	fgPlanUnit := yPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[0]
	if len(fgPlanUnit.(nextroute.ModelPlanStopsUnit).Stops()) != 2 {
		t.Fatal("expected plan unit to have 2 stops (f, g)")
	}
	fgSolutionPlanUnit := solution.SolutionPlanUnit(fgPlanUnit)

	// Check F - f - e - a - b - c - d - i - g - L is rejected
	stopPositionf, err := nextroute.NewStopPosition(
		solutionVehicle.First(),
		fSolutionStop,
		eSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositiong, err := nextroute.NewStopPosition(
		iSolutionStop,
		gSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		fgSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionf, stopPositiong},
	)
	if err != nil {
		t.Fatal(err)
	}
	if move.IsExecutable() {
		t.Fatal("expected move not to be executable, x not allowed to interleave f - g")
	}
	// Check F - e - a - b - c - d - f - i - g - L is accepted
	stopPositionf, err = nextroute.NewStopPosition(
		dSolutionStop,
		fSolutionStop,
		iSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositiong, err = nextroute.NewStopPosition(
		iSolutionStop,
		gSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		fgSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionf, stopPositiong},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !move.IsExecutable() {
		t.Fatal("expected move to be executable, i can interleave f - g")
	}
	// Check F - f - g - e - a - b - c - d - i - L is accepted
	stopPositionf, err = nextroute.NewStopPosition(
		solutionVehicle.First(),
		fSolutionStop,
		gSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	stopPositiong, err = nextroute.NewStopPosition(
		fSolutionStop,
		gSolutionStop,
		solutionVehicle.First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		fgSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionf, stopPositiong},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !move.IsExecutable() {
		t.Fatal("expected move to be executable, i can interleave f - g")
	}

	hSolutionStop := solution.SolutionStop(h)
	hPlanUnit := yPlanUnit.(nextroute.ModelPlanUnitsUnit).PlanUnits()[1]
	if len(hPlanUnit.(nextroute.ModelPlanStopsUnit).Stops()) != 1 {
		t.Fatal("expected plan unit to have 1 stop (h)")
	}
	hSolutionPlanUnit := solution.SolutionPlanUnit(hPlanUnit)

	// Check F - h - e - a - b - c - d - i - L is accepted
	stopPositionh, err := nextroute.NewStopPosition(
		solutionVehicle.First(),
		hSolutionStop,
		eSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		hSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionh},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !move.IsExecutable() {
		t.Fatal("expected move to be executable, h is allowed first")
	}
	// Check F - e - h - a - b - c - d - i - L is rejected
	stopPositionh, err = nextroute.NewStopPosition(
		eSolutionStop,
		hSolutionStop,
		aSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		hSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionh},
	)
	if err != nil {
		t.Fatal(err)
	}
	if move.IsExecutable() {
		t.Fatal("expected move not to be executable, h cannot be interleaved with e and a-b")
	}
	// Check F - e - a - h - b - c - d - i - L is rejected
	stopPositionh, err = nextroute.NewStopPosition(
		aSolutionStop,
		hSolutionStop,
		bSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		hSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionh},
	)
	if err != nil {
		t.Fatal(err)
	}
	if move.IsExecutable() {
		t.Fatal("expected move not to be executable, h cannot be interleaved with a-b")
	}

	// Check F - e - a - b - c - d - h - i - L is accepted
	stopPositionh, err = nextroute.NewStopPosition(
		dSolutionStop,
		hSolutionStop,
		iSolutionStop,
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		hSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionh},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}
	// Check F - e - a - b - c - d - i - h - L is accepted
	stopPositionh, err = nextroute.NewStopPosition(
		iSolutionStop,
		hSolutionStop,
		solutionVehicle.Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		hSolutionPlanUnit.(nextroute.SolutionPlanStopsUnit),
		nextroute.StopPositions{stopPositionh},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !move.IsExecutable() {
		t.Fatal("expected move to be executable")
	}
}

// createModel2 creates a model with 10 stops, 4 plan units, 2 vehicles.
// Creates 10 stops A, B, C, D, E, F, G, H, I, J
// Creates 4 plan units:
// 1. Plan units unit x consisting out of 3 plan stops units:
//   - Plan stops unit x1 consisting out of stops A, B, A has to go before B
//   - Plan stops unit x2 consisting out of stops C, D, C has to go before D
//   - Plan stops unit x3 consisting out of stop E
//
// 2. Plan units unit y consisting out of 2 plan stops units:
//   - Plan stops unit y1 consisting out of stops F, G, F has to go before G
//   - Plan stops unit y2 consisting out of stop H
//
// 3. Plan stops unit i consisting out of stop I
// 4. Plan stops unit j consisting out of stop J
//
// Creates an interleave constraint that:
// - Disallows interleaving of plan unit y and i with plan unit x
// - Disallows interleaving of plan unit x and j with plan unit y
// Adds the interleave constraint to the model
// Returns the model, plan units [x,y, i, j], and stops [A, B, C, D, E, F, G, H, I, J].
func createModel2(t *testing.T) (
	nextroute.Model,
	nextroute.ModelPlanUnits,
	nextroute.ModelStops,
) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}
	center, err := common.NewLocation(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	a, err := model.NewStop(center)
	a.SetID("a")
	if err != nil {
		t.Fatal(err)
	}
	b, err := model.NewStop(center)
	b.SetID("b")
	if err != nil {
		t.Fatal(err)
	}
	c, err := model.NewStop(center)
	c.SetID("c")
	if err != nil {
		t.Fatal(err)
	}
	d, err := model.NewStop(center)
	d.SetID("d")
	if err != nil {
		t.Fatal(err)
	}
	e, err := model.NewStop(center)
	e.SetID("e")
	if err != nil {
		t.Fatal(err)
	}
	x1DAG := nextroute.NewDirectedAcyclicGraph()
	err = x1DAG.AddArc(a, b)
	if err != nil {
		t.Fatal(err)
	}
	x1, err := model.NewPlanMultipleStops([]nextroute.ModelStop{a, b}, x1DAG)
	if err != nil {
		t.Fatal(err)
	}
	x2DAG := nextroute.NewDirectedAcyclicGraph()
	err = x2DAG.AddArc(c, d)
	if err != nil {
		t.Fatal(err)
	}
	x2, err := model.NewPlanMultipleStops([]nextroute.ModelStop{c, d}, x2DAG)
	if err != nil {
		t.Fatal(err)
	}
	x3, err := model.NewPlanSingleStop(e)
	if err != nil {
		t.Fatal(err)
	}
	xPlanUnit, err := model.NewPlanAllPlanUnits(true, x1, x2, x3)
	if err != nil {
		t.Fatal(err)
	}

	f, err := model.NewStop(center)
	f.SetID("f")
	if err != nil {
		t.Fatal(err)
	}
	g, err := model.NewStop(center)
	g.SetID("g")
	if err != nil {
		t.Fatal(err)
	}
	h, err := model.NewStop(center)
	h.SetID("h")
	if err != nil {
		t.Fatal(err)
	}
	y1DAG := nextroute.NewDirectedAcyclicGraph()
	err = y1DAG.AddArc(f, g)
	if err != nil {
		t.Fatal(err)
	}
	y1, err := model.NewPlanMultipleStops([]nextroute.ModelStop{f, g}, y1DAG)
	if err != nil {
		t.Fatal(err)
	}
	y2, err := model.NewPlanSingleStop(h)
	if err != nil {
		t.Fatal(err)
	}
	yPlanUnit, err := model.NewPlanAllPlanUnits(true, y1, y2)
	if err != nil {
		t.Fatal(err)
	}

	i, err := model.NewStop(center)
	i.SetID("i")
	if err != nil {
		t.Fatal(err)
	}
	j, err := model.NewStop(center)
	j.SetID("j")
	if err != nil {
		t.Fatal(err)
	}
	iPlanUnit, err := model.NewPlanSingleStop(i)
	if err != nil {
		t.Fatal(err)
	}
	jPlanUnit, err := model.NewPlanSingleStop(j)
	if err != nil {
		t.Fatal(err)
	}

	interleaveConstraint, err := nextroute.NewInterleaveConstraint()
	if err != nil {
		t.Fatal(err)
	}
	err = interleaveConstraint.DisallowInterleaving(xPlanUnit, []nextroute.ModelPlanUnit{yPlanUnit, iPlanUnit})
	if err != nil {
		t.Fatal(err)
	}
	err = interleaveConstraint.DisallowInterleaving(yPlanUnit, []nextroute.ModelPlanUnit{xPlanUnit, jPlanUnit})
	if err != nil {
		t.Fatal(err)
	}

	err = model.AddConstraint(interleaveConstraint)
	if err != nil {
		t.Fatal(err)
	}

	vt, err := model.NewVehicleType(
		nextroute.NewTimeIndependentDurationExpression(
			nextroute.NewTravelDurationExpression(
				nextroute.NewHaversineExpression(),
				common.NewSpeed(
					10.0,
					common.MetersPerSecond,
				),
			),
		),
		nextroute.NewDurationExpression(
			"travelDuration",
			nextroute.NewStopDurationExpression("serviceDuration", 0.0),
			common.Second,
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	warehouse, err := model.NewStop(center)
	if err != nil {
		t.Fatal(err)
	}
	warehouse.SetID("warehouse")

	v, err := model.NewVehicle(vt, model.Epoch(), warehouse, warehouse)
	v.SetID("v1")
	if err != nil {
		t.Fatal(err)
	}
	v, err = model.NewVehicle(vt, model.Epoch(), warehouse, warehouse)
	v.SetID("v2")
	if err != nil {
		t.Fatal(err)
	}

	return model,
		nextroute.ModelPlanUnits{xPlanUnit, yPlanUnit, iPlanUnit, jPlanUnit},
		nextroute.ModelStops{a, b, c, d, e, f, g, h, i, j}
}

func createModel1(t *testing.T, postAll bool) (
	nextroute.Model,
	nextroute.ModelPlanUnits,
	nextroute.ModelStops,
) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}
	center, err := common.NewLocation(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	a, err := model.NewStop(center)
	a.SetID("a")
	if err != nil {
		t.Fatal(err)
	}
	b, err := model.NewStop(center)
	b.SetID("b")
	if err != nil {
		t.Fatal(err)
	}
	c, err := model.NewStop(center)
	c.SetID("c")
	if err != nil {
		t.Fatal(err)
	}
	d, err := model.NewStop(center)
	d.SetID("d")
	if err != nil {
		t.Fatal(err)
	}
	e, err := model.NewStop(center)
	e.SetID("e")
	if err != nil {
		t.Fatal(err)
	}
	f, err := model.NewStop(center)
	f.SetID("f")
	if err != nil {
		t.Fatal(err)
	}

	aPlanUnit, err := model.NewPlanSingleStop(a)
	if err != nil {
		t.Fatal(err)
	}
	bPlanUnit, err := model.NewPlanSingleStop(b)
	if err != nil {
		t.Fatal(err)
	}
	xPlanUnit, err := model.NewPlanAllPlanUnits(true, aPlanUnit, bPlanUnit)
	if err != nil {
		t.Fatal(err)
	}
	cPlanUnit, err := model.NewPlanSingleStop(c)
	if err != nil {
		t.Fatal(err)
	}
	dPlanUnit, err := model.NewPlanSingleStop(d)
	if err != nil {
		t.Fatal(err)
	}
	yPlanUnit, err := model.NewPlanAllPlanUnits(true, cPlanUnit, dPlanUnit)
	if err != nil {
		t.Fatal(err)
	}
	ePlanUnit, err := model.NewPlanSingleStop(e)
	if err != nil {
		t.Fatal(err)
	}
	fPlanUnit, err := model.NewPlanSingleStop(f)
	if err != nil {
		t.Fatal(err)
	}
	zPlanUnit, err := model.NewPlanAllPlanUnits(true, ePlanUnit, fPlanUnit)
	if err != nil {
		t.Fatal(err)
	}
	interleaveConstraint, err := nextroute.NewInterleaveConstraint()
	if err != nil {
		t.Fatal(err)
	}
	err = interleaveConstraint.DisallowInterleaving(xPlanUnit, []nextroute.ModelPlanUnit{yPlanUnit, zPlanUnit})
	if err != nil {
		t.Fatal(err)
	}

	if postAll {
		err = interleaveConstraint.DisallowInterleaving(yPlanUnit, []nextroute.ModelPlanUnit{xPlanUnit, zPlanUnit})
		if err != nil {
			t.Fatal(err)
		}
		err = interleaveConstraint.DisallowInterleaving(zPlanUnit, []nextroute.ModelPlanUnit{xPlanUnit, yPlanUnit})
		if err != nil {
			t.Fatal(err)
		}
	}
	err = model.AddConstraint(interleaveConstraint)
	if err != nil {
		t.Fatal(err)
	}

	vt, err := model.NewVehicleType(
		nextroute.NewTimeIndependentDurationExpression(
			nextroute.NewTravelDurationExpression(
				nextroute.NewHaversineExpression(),
				common.NewSpeed(
					10.0,
					common.MetersPerSecond,
				),
			),
		),
		nextroute.NewDurationExpression(
			"travelDuration",
			nextroute.NewStopDurationExpression("serviceDuration", 0.0),
			common.Second,
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	warehouse, err := model.NewStop(center)
	if err != nil {
		t.Fatal(err)
	}
	warehouse.SetID("warehouse")

	v, err := model.NewVehicle(vt, model.Epoch(), warehouse, warehouse)
	v.SetID("v1")
	if err != nil {
		t.Fatal(err)
	}

	return model,
		nextroute.ModelPlanUnits{xPlanUnit, yPlanUnit, zPlanUnit},
		nextroute.ModelStops{a, b, c, d, e, f}
}
