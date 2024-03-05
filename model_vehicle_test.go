// © 2019-present nextmv.io inc

package nextroute_test

import (
	"fmt"
	"testing"

	"github.com/nextmv-io/nextroute"
)

func TestModelVehicleImpl_AddStop(t *testing.T) {
	model, err := createModel(
		input(
			vehicleTypes("truck"),
			vehicles(
				"truck",
				depot(),
				2,
			),
			planSingleStops(),
			planPairSequences(),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	v1 := model.Vehicles()[0]

	count := 0
	for idx, planUnit := range model.PlanStopsUnits() {
		for _, stop := range planUnit.Stops() {
			if err := v1.AddStop(
				stop,
				idx < 3,
			); err != nil {
				t.Fatal(err)
			}
			count++
			if len(v1.Stops()) != count {
				t.Fatalf("expected %v stops, got %v", count, len(v1.Stops()))
			}
		}
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}

	sv1 := solution.SolutionVehicle(v1)

	for _, stop := range sv1.SolutionStops() {
		fmt.Println(stop)
	}
}
