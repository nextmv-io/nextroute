// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestModel_NewPlanAllPlanUnits(t *testing.T) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}
	s1, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s1.SetID("s1")

	s2, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s2.SetID("s2")

	s3, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s3.SetID("s3")

	s1Unit, err := model.NewPlanSingleStop(s1)
	if err != nil {
		t.Fatal(err)
	}
	s2Unit, err := model.NewPlanSingleStop(s2)
	if err != nil {
		t.Fatal(err)
	}
	s3Unit, err := model.NewPlanSingleStop(s3)
	if err != nil {
		t.Fatal(err)
	}
	_, err = model.NewPlanAllPlanUnits(true)
	if err == nil {
		t.Fatal(err)
	}
	_, err = model.NewPlanAllPlanUnits(false)
	if err == nil {
		t.Fatal(err)
	}
	cpu1, err := model.NewPlanAllPlanUnits(true, s1Unit)
	if err != nil {
		t.Fatal(err)
	}
	if cpu1 == nil {
		t.Fatal("cpu1 is nil")
	}
	if cpu1.SameVehicle() != true {
		t.Fatal("cpu1.SameVehicle() != true")
	}
	if cpu1.PlanAll() != true {
		t.Fatal("cpu1.PlanAll() != true")
	}
	if cpu1.PlanOneOf() != false {
		t.Fatal("cpu1.PlanOneOf() != false")
	}
	if len(cpu1.PlanUnits()) != 1 {
		t.Fatal("len(cpu1.PlanUnits()) != 1")
	}
	if cpu1.PlanUnits()[0] != s1Unit {
		t.Fatal("cpu1.PlanUnits()[0] != s1Unit")
	}
	if planUnitsUnit, ok := s1Unit.PlanUnitsUnit(); !ok || planUnitsUnit != cpu1 {
		t.Fatal("s1Unit.PlanUnitsUnit() != cpu1")
	}
	_, err = model.NewPlanAllPlanUnits(true, s1Unit)
	if err == nil {
		t.Fatal(err)
	}
	_, err = model.NewPlanAllPlanUnits(true, s2Unit, s2Unit)
	if err == nil {
		t.Fatal(err)
	}
	cpu2, err := model.NewPlanAllPlanUnits(false, s2Unit, s3Unit)
	if err != nil {
		t.Fatal(err)
	}
	if cpu2.SameVehicle() != false {
		t.Fatal("cpu2.SameVehicle() != false")
	}
	if len(cpu2.PlanUnits()) != 2 {
		t.Fatal("len(cpu2.PlanUnits()) != 2")
	}
	if cpu2.PlanUnits()[0] == cpu2.PlanUnits()[1] {
		t.Fatal("cpu2.PlanUnits()[0] == cpu2.PlanUnits()[1]")
	}
	if cpu2.PlanUnits()[0] != s2Unit && cpu2.PlanUnits()[0] != s3Unit {
		t.Fatal("cpu2.PlanUnits()[0] != s2Unit && cpu2.PlanUnits()[0] != s3Unit")
	}
	if cpu2.PlanUnits()[1] != s2Unit && cpu2.PlanUnits()[1] != s3Unit {
		t.Fatal("cpu2.PlanUnits()[1] != s2Unit && cpu2.PlanUnits()[1] != s3Unit")
	}
}

func TestModel_NewPlanOneOfPlanUnits(t *testing.T) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}
	s1, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s1.SetID("s1")

	s2, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s2.SetID("s2")

	s3, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s3.SetID("s3")

	s1Unit, err := model.NewPlanSingleStop(s1)
	if err != nil {
		t.Fatal(err)
	}
	s2Unit, err := model.NewPlanSingleStop(s2)
	if err != nil {
		t.Fatal(err)
	}
	s3Unit, err := model.NewPlanSingleStop(s3)
	if err != nil {
		t.Fatal(err)
	}
	_, err = model.NewPlanOneOfPlanUnits()
	if err == nil {
		t.Fatal(err)
	}

	dpu1, err := model.NewPlanOneOfPlanUnits(s1Unit)
	if err != nil {
		t.Fatal(err)
	}
	if dpu1 == nil {
		t.Fatal("dpu1 is nil")
	}
	if dpu1.PlanAll() != false {
		t.Fatal("dpu1.PlanAll() != false")
	}
	if dpu1.PlanOneOf() != true {
		t.Fatal("dpu1.PlanOneOf() != true")
	}
	if len(dpu1.PlanUnits()) != 1 {
		t.Fatal("len(dpu1.PlanUnits()) != 1")
	}
	if dpu1.PlanUnits()[0] != s1Unit {
		t.Fatal("dpu1.PlanUnits()[0] != s1Unit")
	}
	if planUnitsUnit, ok := s1Unit.PlanUnitsUnit(); !ok || planUnitsUnit != dpu1 {
		t.Fatal("s1Unit.PlanUnitsUnit() != dpu1")
	}
	_, err = model.NewPlanOneOfPlanUnits(s1Unit)
	if err == nil {
		t.Fatal(err)
	}
	_, err = model.NewPlanOneOfPlanUnits(s2Unit, s2Unit)
	if err == nil {
		t.Fatal(err)
	}
	dpu2, err := model.NewPlanOneOfPlanUnits(s2Unit, s3Unit)
	if err != nil {
		t.Fatal(err)
	}
	if len(dpu2.PlanUnits()) != 2 {
		t.Fatal("len(dpu2.PlanUnits()) != 2")
	}
	if dpu2.PlanUnits()[0] == dpu2.PlanUnits()[1] {
		t.Fatal("dpu2.PlanUnits()[0] == dpu2.PlanUnits()[1]")
	}
	if dpu2.PlanUnits()[0] != s2Unit && dpu2.PlanUnits()[0] != s3Unit {
		t.Fatal("dpu2.PlanUnits()[0] != s2Unit && dpu2.PlanUnits()[0] != s3Unit")
	}
	if dpu2.PlanUnits()[1] != s2Unit && dpu2.PlanUnits()[1] != s3Unit {
		t.Fatal("dpu2.PlanUnits()[1] != s2Unit && dpu2.PlanUnits()[1] != s3Unit")
	}

	dpu1ordpu2, err := model.NewPlanOneOfPlanUnits(dpu1, dpu2)
	if err != nil {
		t.Fatal(err)
	}

	warehouse, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	warehouse.SetID("warehouse")

	vehicleType, err := model.NewVehicleType(
		nextroute.NewTimeIndependentDurationExpression(
			nextroute.NewDurationExpression(
				"travelDuration",
				nextroute.NewHaversineExpression(),
				common.Second,
			),
		),
		nextroute.NewConstantDurationExpression("processing_time", 1*time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = model.NewVehicle(vehicleType, model.Epoch(), warehouse, warehouse)
	if err != nil {
		t.Fatal(err)
	}

	_, err = model.NewVehicle(vehicleType, model.Epoch(), warehouse, warehouse)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)

	if err != nil {
		t.Fatal(err)
	}

	move := solution.BestMove(context.Background(), solution.SolutionPlanUnit(dpu1ordpu2))

	if !move.IsExecutable() {
		t.Fatal("move.IsExecutable() != true")
	}

	_, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, stop := range solution.Vehicles()[0].SolutionStops() {
		fmt.Println(stop.ModelStop().ID())
	}
}

func TestPlanUnitsUnit(t *testing.T) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}
	warehouse, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	warehouse.SetID("warehouse")

	s1, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s1.SetID("s1")

	s2, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s2.SetID("s2")

	s3, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}
	s3.SetID("s3")

	s1Unit, err := model.NewPlanSingleStop(s1)
	if err != nil {
		t.Fatal(err)
	}
	s2Unit, err := model.NewPlanSingleStop(s2)
	if err != nil {
		t.Fatal(err)
	}
	s3Unit, err := model.NewPlanSingleStop(s3)
	if err != nil {
		t.Fatal(err)
	}
	s1ors2Unit, err := model.NewPlanAllPlanUnits(true, s1Unit, s2Unit)
	if err != nil {
		t.Fatal(err)
	}

	vehicleType, err := model.NewVehicleType(
		nextroute.NewTimeIndependentDurationExpression(
			nextroute.NewDurationExpression(
				"travelDuration",
				nextroute.NewHaversineExpression(),
				common.Second,
			),
		),
		nextroute.NewConstantDurationExpression("processing_time", 1*time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = model.NewVehicle(vehicleType, model.Epoch(), warehouse, warehouse)
	if err != nil {
		t.Fatal(err)
	}

	_, err = model.NewVehicle(vehicleType, model.Epoch(), warehouse, warehouse)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	if solution == nil {
		t.Fatal("solution should not be nil")
	}

	if solution.UnPlannedPlanUnits().Size() != 2 {
		t.Fatal("solution should have 2 un-planned plan units, it has", solution.UnPlannedPlanUnits().Size())
	}

	move := solution.BestMove(
		context.Background(),
		solution.UnPlannedPlanUnits().SolutionPlanUnit(s1ors2Unit),
	)

	if move == nil {
		t.Fatal("move should not be nil")
	}
	if !move.IsExecutable() {
		t.Fatal("move should be executable")
	}

	success, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !success {
		t.Fatal("move should be successful")
	}
	move = solution.BestMove(
		context.Background(),
		solution.UnPlannedPlanUnits().SolutionPlanUnit(s3Unit),
	)

	if move == nil {
		t.Fatal("move should not be nil")
	}
	if !move.IsExecutable() {
		t.Fatal("move should be executable")
	}

	success, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !success {
		t.Fatal("move should be successful")
	}

	/*
		for _, v := range solution.Vehicles() {
			fmt.Println("-")
			for _, stop := range v.SolutionStops() {
				fmt.Println(stop.ModelStop().ID())
			}
		}
	*/
}
