// © 2019-present nextmv.io inc

package nextroute_test

import (
	"context"
	"testing"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestSweepOneDepot(t *testing.T) {
	input := singleVehiclePlanSingleStopsModel()
	input.Vehicles = append(input.Vehicles, vehicles("truck", depot(), 1)...)
	model, err := createModel(input)
	if err != nil {
		t.Fatal(err)
	}

	solution, err := nextroute.NewSweepSolution(context.Background(), model)
	if err != nil {
		t.Fatal(err)
	}

	if len(solution.Vehicles()) != 2 {
		t.Errorf("expected 2 vehicles, got %v", len(solution.Vehicles()))
	}
}

func TestSweepTwoDepots(t *testing.T) {
	input := singleVehiclePlanSingleStopsModel()

	location := Location{
		Lat: 0,
		Lon: 0,
	}
	input.Vehicles = append(input.Vehicles, vehicles("truck", location, 1)...)
	model, err := createModel(input)
	if err != nil {
		t.Fatal(err)
	}

	_, err = nextroute.NewSweepSolution(context.Background(), model)
	if err != nil {
		if err.Error() != "sweep construction, not implemented for multiple start-end locations of input" {
			t.Fatal(err)
		}
	}
}

func TestSweepStartAndEndDifferent(t *testing.T) {
	input := singleVehiclePlanSingleStopsModel()

	location := Location{
		Lat:     common.NewInvalidLocation().Latitude(),
		Lon:     common.NewInvalidLocation().Longitude(),
		IsValid: false,
	}
	input.Vehicles[0].StartLocation = location
	input.Vehicles[0].EndLocation = depot()
	model, err := createModel(input)
	if err != nil {
		t.Fatal(err)
	}

	_, err = nextroute.NewSweepSolution(context.Background(), model)
	if err != nil {
		t.Fatal(err)
	}
}
