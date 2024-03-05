// © 2019-present nextmv.io inc

package nextroute_test

import (
	"slices"
	"testing"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/sdk/common"
)

func TestAttributesConstraint_EstimateIsViolated(t *testing.T) {
	model, err := createModel(
		input(
			vehicleTypes("truck", "car", "bike"),
			[]Vehicle{
				vehicles(
					"truck",
					depot(),
					1,
				)[0],
				vehicles(
					"car",
					depot(),
					1,
				)[0],
				vehicles(
					"bike",
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

	cnstr, err := nextroute.NewAttributesConstraint()
	if err != nil {
		t.Error(err)
	}
	err = model.AddConstraint(cnstr)
	if err != nil {
		t.Error(err)
	}

	attribute0 := "attribute-0"
	attribute1 := "attribute-1"
	attribute2 := "attribute-2"

	singleStopPlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() == 1
	})

	cnstr.SetStopAttributes(singleStopPlanUnits[0].Stops()[0], []string{attribute0})
	cnstr.SetStopAttributes(singleStopPlanUnits[1].Stops()[0], []string{attribute1})
	cnstr.SetStopAttributes(singleStopPlanUnits[2].Stops()[0], []string{})

	sequencePlanUnits := common.Filter(model.PlanStopsUnits(), func(planUnit nextroute.ModelPlanStopsUnit) bool {
		return planUnit.NumberOfStops() > 1
	})

	cnstr.SetStopAttributes(sequencePlanUnits[0].Stops()[0], []string{attribute0, attribute1})
	cnstr.SetStopAttributes(sequencePlanUnits[0].Stops()[1], []string{attribute1, attribute2})

	cnstr.SetStopAttributes(sequencePlanUnits[1].Stops()[0], []string{attribute1})
	cnstr.SetStopAttributes(sequencePlanUnits[1].Stops()[1], []string{attribute1, attribute2})

	cnstr.SetVehicleTypeAttributes(model.VehicleTypes()[0], []string{attribute0, attribute1})
	cnstr.SetVehicleTypeAttributes(model.VehicleTypes()[1], []string{attribute1, attribute2})
	cnstr.SetVehicleTypeAttributes(model.VehicleTypes()[2], []string{})

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Error(err)
	}

	truckIndex := slices.IndexFunc(solution.Vehicles(), func(solutionVehicle nextroute.SolutionVehicle) bool {
		return solutionVehicle.ModelVehicle().VehicleType().ID() == "truck"
	})
	if truckIndex == -1 {
		t.Error("truck not found")
	}
	truck := solution.Vehicles()[truckIndex]

	carIndex := slices.IndexFunc(solution.Vehicles(), func(solutionVehicle nextroute.SolutionVehicle) bool {
		return solutionVehicle.ModelVehicle().VehicleType().ID() == "car"
	})
	if carIndex == -1 {
		t.Error("car not found")
	}
	car := solution.Vehicles()[carIndex]

	bikeIndex := slices.IndexFunc(solution.Vehicles(), func(solutionVehicle nextroute.SolutionVehicle) bool {
		return solutionVehicle.ModelVehicle().VehicleType().ID() == "bike"
	})
	if bikeIndex == -1 {
		t.Error("bike not found")
	}
	bike := solution.Vehicles()[bikeIndex]

	{
		moveSingle0OnTruck, err := nextroute.NewMoveStops(
			solution.SolutionPlanStopsUnit(singleStopPlanUnits[0]),
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					truck.First(),
					solution.SolutionStop(singleStopPlanUnits[0].Stops()[0]),
					truck.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSingle0OnTruck); violated {
			t.Errorf("moveSingle0OnTruck should not be violated, share attribute-0")
		}

		moveSingle0OnCar, err := nextroute.NewMoveStops(
			solution.SolutionPlanStopsUnit(singleStopPlanUnits[0]),
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					car.First(),
					solution.SolutionStop(singleStopPlanUnits[0].Stops()[0]),
					car.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSingle0OnCar); !violated {
			t.Errorf("moveSingle0OnCar should be violated, car no attribute-0")
		}

		moveSingle0OnBike, err := nextroute.NewMoveStops(
			solution.SolutionPlanStopsUnit(singleStopPlanUnits[0]),
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					bike.First(),
					solution.SolutionStop(singleStopPlanUnits[0].Stops()[0]),
					bike.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSingle0OnBike); !violated {
			t.Errorf("moveSingle0OnBike should not be violated, bike no attributes")
		}
	}
	{
		moveSingle1OnTruck, err := nextroute.NewMoveStops(
			solution.SolutionPlanStopsUnit(singleStopPlanUnits[1]),
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					truck.First(),
					solution.SolutionStop(singleStopPlanUnits[1].Stops()[0]),
					truck.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSingle1OnTruck); violated {
			t.Errorf("moveSingle1OnTruck should not be violated, share attribute-1")
		}

		moveSingle1OnCar, err := nextroute.NewMoveStops(
			solution.SolutionPlanStopsUnit(singleStopPlanUnits[1]),
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					car.First(),
					solution.SolutionStop(singleStopPlanUnits[1].Stops()[0]),
					car.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSingle1OnCar); violated {
			t.Errorf("moveSingle1OnCar should not be violated, share attribute-1")
		}

		moveSingle1OnBike, err := nextroute.NewMoveStops(
			solution.SolutionPlanStopsUnit(singleStopPlanUnits[1]),
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					bike.First(),
					solution.SolutionStop(singleStopPlanUnits[1].Stops()[0]),
					bike.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSingle1OnBike); !violated {
			t.Errorf("moveSingle1OnBike should be violated, bike no attributes")
		}
	}
	{
		moveSingle2OnTruck, err := nextroute.NewMoveStops(
			solution.SolutionPlanStopsUnit(singleStopPlanUnits[2]),
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					truck.First(),
					solution.SolutionStop(singleStopPlanUnits[2].Stops()[0]),
					truck.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSingle2OnTruck); violated {
			t.Errorf("moveSingle2OnTruck should not be violated, stop no attributes")
		}

		moveSingle2OnCar, err := nextroute.NewMoveStops(
			solution.SolutionPlanStopsUnit(singleStopPlanUnits[2]),
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					car.First(),
					solution.SolutionStop(singleStopPlanUnits[2].Stops()[0]),
					car.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSingle2OnCar); violated {
			t.Errorf("moveSingle1OnCar should not be violated, stop no attributes")
		}

		moveSingle2OnBike, err := nextroute.NewMoveStops(
			solution.SolutionPlanStopsUnit(singleStopPlanUnits[2]),
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					bike.First(),
					solution.SolutionStop(singleStopPlanUnits[2].Stops()[0]),
					bike.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSingle2OnBike); violated {
			t.Errorf("moveSingle1OnBike should not be violated, stop no attributes")
		}
	}

	{
		sequencePlanUnit := solution.SolutionPlanStopsUnit(sequencePlanUnits[0])

		moveSequence0OnTruck, err := nextroute.NewMoveStops(
			sequencePlanUnit,
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					truck.First(),
					sequencePlanUnit.SolutionStops()[0],
					sequencePlanUnit.SolutionStops()[1],
				),
				nextroute.NewStopPosition(
					sequencePlanUnit.SolutionStops()[0],
					sequencePlanUnit.SolutionStops()[1],
					truck.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSequence0OnTruck); violated {
			t.Errorf("moveSequence0OnTruck should not be violated, stop shares attributes with truck")
		}

		moveSequence0OnCar, err := nextroute.NewMoveStops(
			sequencePlanUnit,
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					car.First(),
					sequencePlanUnit.SolutionStops()[0],
					sequencePlanUnit.SolutionStops()[1],
				),
				nextroute.NewStopPosition(
					sequencePlanUnit.SolutionStops()[0],
					sequencePlanUnit.SolutionStops()[1],
					car.Last(),
				),
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSequence0OnCar); violated {
			t.Errorf("moveSequence0OnCar should not be violated, stops share attribute-1 with car")
		}

		moveSequence0OnBike, err := nextroute.NewMoveStops(
			sequencePlanUnit,
			[]nextroute.StopPosition{
				nextroute.NewStopPosition(
					bike.First(),
					sequencePlanUnit.SolutionStops()[0],
					sequencePlanUnit.SolutionStops()[1],
				),
				nextroute.NewStopPosition(
					sequencePlanUnit.SolutionStops()[0],
					sequencePlanUnit.SolutionStops()[1],
					bike.Last(),
				),
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		if violated, _ := cnstr.EstimateIsViolated(moveSequence0OnBike); !violated {
			t.Errorf("moveSequence0OnBike should not be violated, bike has not attributes")
		}
	}
}

func TestAttributesConstraint(t *testing.T) {
	model, err := createModel(
		input(
			vehicleTypes("truck"),
			vehicles(
				"truck",
				depot(),
				2,
			),
			planSingleStops(),
			nil,
		),
	)
	if err != nil {
		t.Error(err)
	}

	constraint, err := nextroute.NewAttributesConstraint()
	if err != nil {
		t.Error(err)
	}

	for _, vt := range model.VehicleTypes() {
		attributes := constraint.VehicleTypeAttributes(vt)

		if len(attributes) != 0 {
			t.Errorf(
				"number of attributes is not correct, expected 0 got %v",
				len(attributes),
			)
		}
	}

	for _, stop := range model.Stops() {
		attributes := constraint.StopAttributes(stop)

		if len(attributes) != 0 {
			t.Errorf(
				"number of attributes is not correct, expected 0 got %v",
				len(attributes),
			)
		}
	}

	vehicleTypeAttributes := []string{"attribute-1", "attribute-2"}

	for _, vt := range model.VehicleTypes() {
		constraint.SetVehicleTypeAttributes(vt, vehicleTypeAttributes)
		attributes := constraint.VehicleTypeAttributes(vt)
		if len(attributes) != 2 {
			t.Errorf(
				"number of attributes is not correct, expected 2 got %v",
				len(attributes),
			)
		}
		if !slices.ContainsFunc(attributes, func(s string) bool {
			return s == vehicleTypeAttributes[0]
		}) {
			t.Errorf(
				"attribute is not correct, expected %v to be in attributes %v",
				vehicleTypeAttributes[0],
				attributes,
			)
		}
		if !slices.ContainsFunc(attributes, func(s string) bool {
			return s == vehicleTypeAttributes[1]
		}) {
			t.Errorf(
				"attribute is not correct, expected %v to be in attributes %v",
				vehicleTypeAttributes[1],
				attributes,
			)
		}
	}

	constraint.SetVehicleTypeAttributes(model.VehicleTypes()[0], []string{})
	constraint.SetStopAttributes(model.Stops()[0], []string{})

	stopAttributes := []string{"attribute-2", "attribute-3"}

	for _, stop := range model.Stops() {
		constraint.SetStopAttributes(stop, stopAttributes)

		attributes := constraint.StopAttributes(stop)

		if len(attributes) != 2 {
			t.Errorf(
				"number of attributes is not correct, expected 2 got %v",
				len(attributes),
			)
		}
		if !slices.ContainsFunc(attributes, func(s string) bool {
			return s == stopAttributes[0]
		}) {
			t.Errorf(
				"attribute is not correct, expected %v to be in attributes %v",
				stopAttributes[0],
				attributes,
			)
		}
		if !slices.ContainsFunc(attributes, func(s string) bool {
			return s == stopAttributes[1]
		}) {
			t.Errorf(
				"attribute is not correct, expected %v to be in attributes %v",
				stopAttributes[1],
				attributes,
			)
		}
	}

	constraint.SetStopAttributes(
		model.Stops()[0],
		[]string{"A", "B", "C", "A", "B", "C"},
	)

	if len(constraint.StopAttributes(model.Stops()[0])) != 3 {
		t.Errorf(
			"number of attributes is not correct, expected 3 got %v",
			len(constraint.StopAttributes(model.Stops()[0])),
		)
	}

	constraint.SetVehicleTypeAttributes(
		model.VehicleTypes()[0],
		[]string{"A", "B", "C", "A", "B", "C"},
	)

	if len(constraint.VehicleTypeAttributes(model.VehicleTypes()[0])) != 3 {
		t.Errorf(
			"number of attributes is not correct, expected 3 got %v",
			len(constraint.VehicleTypeAttributes(model.VehicleTypes()[0])),
		)
	}

	err = model.AddConstraint(constraint)

	if err != nil {
		t.Error(err)
	}

	if len(model.Constraints()) != 1 {
		t.Errorf(
			"number of constraints is not correct, expected 1 got %v",
			len(model.Constraints()),
		)
	}

	if model.Constraints()[0] != constraint {
		t.Errorf(
			"constraint is not correct, expected %v got %v",
			constraint,
			model.Constraints()[0],
		)
	}
}
