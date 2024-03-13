// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestNoMixConstraint(t *testing.T) {
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
			[]PlanSequence{
				{
					Stops: []Stop{
						{
							Name: "s1",
							Location: Location{
								Lon: -74.04866,
								Lat: 4.69018,
							},
						},
						{
							Name: "s2",
							Location: Location{
								Lon: -74.044215,
								Lat: 4.693907,
							},
						},
					},
				},
				{
					Stops: []Stop{
						{
							Name: "s3",
							Location: Location{
								Lon: -74.04866,
								Lat: 4.693907,
							},
						},
						{
							Name: "s4",
							Location: Location{
								Lon: -74.044215,
								Lat: 4.69018,
							},
						},
					},
				},
				{
					Stops: []Stop{
						{
							Name: "s5",
							Location: Location{
								Lon: -74.04866,
								Lat: 4.693907,
							},
						},
						{
							Name: "s6",
							Location: Location{
								Lon: -74.044215,
								Lat: 4.69018,
							},
						},
					},
				},
			},
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	sequencePlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() > 1
	})

	deltas := make(map[nextroute.ModelStop]nextroute.MixItem)

	deltas[sequencePlanUnits[0].Stops()[0]] = nextroute.MixItem{
		Name:     "A",
		Quantity: 1,
	}
	deltas[sequencePlanUnits[0].Stops()[1]] = nextroute.MixItem{
		Name:     "A",
		Quantity: -1,
	}

	deltas[sequencePlanUnits[1].Stops()[0]] = nextroute.MixItem{
		Name:     "B",
		Quantity: 1,
	}
	deltas[sequencePlanUnits[1].Stops()[1]] = nextroute.MixItem{
		Name:     "B",
		Quantity: -1,
	}

	deltas[sequencePlanUnits[2].Stops()[0]] = nextroute.MixItem{
		Name:     "A",
		Quantity: 1,
	}
	deltas[sequencePlanUnits[2].Stops()[1]] = nextroute.MixItem{
		Name:     "A",
		Quantity: -1,
	}

	cnstr, err := nextroute.NewNoMixConstraint(deltas)
	if err != nil {
		t.Fatal(err)
	}
	err = model.AddConstraint(cnstr)
	if err != nil {
		t.Fatal(err)
	}
	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}
	solutionPlanStopsUnit := solution.SolutionPlanStopsUnit(sequencePlanUnits[0])

	position1, err := nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err := nextroute.NewStopPosition(
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err := nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}

	isViolated, _ := cnstr.EstimateIsViolated(move)
	if isViolated {
		t.Fatal("constraint is violated, should be possible [+A][-A]")
	}
	executed, err := move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !executed {
		t.Fatal("move should be executed")
	}

	solutionPlanStopsUnit = solution.SolutionPlanStopsUnit(sequencePlanUnits[1])
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if isViolated {
		t.Fatal("constraint is violated, it should fit [+B][-B]+A-A")
	}

	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if !isViolated {
		t.Fatal("constraint is not violated, it should not fit [+B]+A[-B]-A")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if !isViolated {
		t.Fatal("constraint is not violated, it should not fit +A[+B][-B]-A")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next().Next(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if !isViolated {
		t.Fatal("constraint is not violated, it should not fit +A[+B]-A[-B]")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if isViolated {
		t.Fatal("constraint is violated, it should fit +A-A[+B][-B]")
	}
	executed, err = move.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !executed {
		t.Fatal("move should be executed")
	}

	solutionPlanStopsUnit = solution.SolutionPlanStopsUnit(sequencePlanUnits[2])

	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if isViolated {
		t.Fatal("constraint is violated, it should fit [+A][-A]+A-A+B-BB")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if isViolated {
		t.Fatal("constraint is violated, it should fit  [+A]+A[-A]-A+B-B")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next().Next(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if isViolated {
		t.Fatal("constraint is violated, it should not fit  [+A]+A-A[-A]+B-B")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous().Previous(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last().Previous(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if !isViolated {
		t.Fatal("constraint is violated, it should not fit  [+A]+A-A+B[-A]-B")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if !isViolated {
		t.Fatal("constraint is violated, it should not fit  [+A]+A-A+B-B[-A]")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if isViolated {
		t.Fatal("constraint is violated, it should fit +A[+A][-A]-A+B-BB")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next().Next(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if isViolated {
		t.Fatal("constraint is violated, it should fit  +A[+A]-A[-A]+B-B")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next().Next().Next(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next().Next().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if !isViolated {
		t.Fatal("constraint is not violated, it should not fit  +A[+A]-A+B[-A]-B")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if !isViolated {
		t.Fatal("constraint is violated, it should not fit  +A[+A]-A+B-B[-A]")
	}

	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}

	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if isViolated {
		t.Fatal("constraint is violated, it should fit +A-A[+A][-A]+B-B")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next().Next().Next(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].First().Next().Next().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if !isViolated {
		t.Fatal("constraint is not violated, it should not fit +A-A[+A]+B[-A]-B")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if !isViolated {
		t.Fatal("constraint is not violated, it should not fit +A-A[+A]+B-B[-A]")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next().Next().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next().Next().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous().Previous(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last().Previous(),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err == nil {
		t.Fatal("planned stop must be after the previous planned stop, stop s3 is not")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].First().Next().Next().Next(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solution.Vehicles()[0].First().Next().Next().Next().Next(),
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous(),
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if !isViolated {
		t.Fatal("constraint is not violated, it should not fit +A-A+B[+A]-B[-A]")
	}
	position1, err = nextroute.NewStopPosition(
		solution.Vehicles()[0].Last().Previous(),
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
	)
	if err != nil {
		t.Fatal(err)
	}
	position2, err = nextroute.NewStopPosition(
		solutionPlanStopsUnit.SolutionStops()[0],
		solutionPlanStopsUnit.SolutionStops()[1],
		solution.Vehicles()[0].Last(),
	)
	if err != nil {
		t.Fatal(err)
	}
	move, err = nextroute.NewMoveStops(
		solutionPlanStopsUnit,
		[]nextroute.StopPosition{position1, position2},
	)
	if err != nil {
		t.Fatal(err)
	}
	isViolated, _ = cnstr.EstimateIsViolated(move)
	if isViolated {
		t.Fatal("constraint is violated, it should fit +A-A+B-B[+A][-A]")
	}
}

func TestNoMixConstraint_ArgumentMismatch(t *testing.T) {
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
			[]PlanSequence{
				{
					Stops: []Stop{
						{
							Name: "s1",
							Location: Location{
								Lon: -74.04866,
								Lat: 4.69018,
							},
						},
						{
							Name: "s2",
							Location: Location{
								Lon: -74.044215,
								Lat: 4.693907,
							},
						},
					},
				},
				{
					Stops: []Stop{
						{
							Name: "s3",
							Location: Location{
								Lon: -74.04866,
								Lat: 4.693907,
							},
						},
						{
							Name: "s4",
							Location: Location{
								Lon: -74.044215,
								Lat: 4.69018,
							},
						},
					},
				},
				{
					Stops: []Stop{
						{
							Name: "s5",
							Location: Location{
								Lon: -74.04866,
								Lat: 4.693907,
							},
						},
						{
							Name: "s6",
							Location: Location{
								Lon: -74.044215,
								Lat: 4.69018,
							},
						},
					},
				},
			},
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	sequencePlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() > 1
	})

	{
		cnstr, err := nextroute.NewNoMixConstraint(
			map[nextroute.ModelStop]nextroute.MixItem{},
		)
		if err != nil {
			t.Fatal(err)
		}
		err = model.AddConstraint(cnstr)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		deltas := make(map[nextroute.ModelStop]nextroute.MixItem)

		deltas[sequencePlanUnits[0].Stops()[0]] = nextroute.MixItem{
			Name:     "A",
			Quantity: 1,
		}
		deltas[sequencePlanUnits[0].Stops()[1]] = nextroute.MixItem{
			Name:     "A",
			Quantity: -2,
		}

		cnstr, err := nextroute.NewNoMixConstraint(deltas)
		if err != nil {
			t.Fatal(err)
		}
		err = model.AddConstraint(cnstr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = nextroute.NewSolution(model)
		if err == nil {
			t.Fatal("should not be possible to create constraint with" +
				" sum insert and remove not equal across plan-units")
		}
	}
	{
		stop := model.Stops()[0]

		deltas := make(map[nextroute.ModelStop]nextroute.MixItem)

		deltas[stop] = nextroute.MixItem{
			Name:     "A",
			Quantity: 1,
		}

		cnstr, err := nextroute.NewNoMixConstraint(deltas)
		if err != nil {
			t.Fatal(err)
		}
		err = model.AddConstraint(cnstr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = nextroute.NewSolution(model)
		if err == nil {
			t.Fatal("should not be possible to create constraint with missing pickup")
		}
	}
	{
		deltas := make(map[nextroute.ModelStop]nextroute.MixItem)

		deltas[sequencePlanUnits[0].Stops()[0]] = nextroute.MixItem{
			Name:     "A",
			Quantity: 1,
		}
		deltas[sequencePlanUnits[0].Stops()[1]] = nextroute.MixItem{
			Name:     "B",
			Quantity: -1,
		}

		cnstr, err := nextroute.NewNoMixConstraint(deltas)
		if err != nil {
			t.Fatal(err)
		}
		err = model.AddConstraint(cnstr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = nextroute.NewSolution(model)
		if err == nil {
			t.Fatal("should not be possible to create constraint with different types per plan-unit")
		}
	}
}
