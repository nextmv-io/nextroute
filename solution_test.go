// © 2019-present nextmv.io inc

package nextroute_test

import (
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
)

func BenchmarkAllocationsSolution(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		model, err := createModel(singleVehiclePlanSequenceModel())
		if err != nil {
			b.Error(err)
		}

		maximum := nextroute.NewVehicleTypeDurationExpression(
			"maximum duration",
			3*time.Minute,
		)
		expression := nextroute.NewStopExpression("test", 2.0)

		cnstr, err := nextroute.NewMaximum(expression, maximum)
		if err != nil {
			b.Error(err)
		}

		err = model.AddConstraint(cnstr)
		if err != nil {
			b.Error(err)
		}
		b.StartTimer()
		_, err = nextroute.NewSolution(model)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestLimitAllocations tests the number of allocations in the solution creation.
// We want to ensure that the number of allocations is limited and does not grow
// accidentally.
func TestLimitAllocations(t *testing.T) {
	model, err := createModel(singleVehiclePlanSequenceModel())
	if err != nil {
		t.Error(err)
	}

	maximum := nextroute.NewVehicleTypeDurationExpression(
		"maximum duration",
		3*time.Minute,
	)
	expression := nextroute.NewStopExpression("test", 2.0)

	cnstr, err := nextroute.NewMaximum(expression, maximum)
	if err != nil {
		t.Error(err)
	}

	err = model.AddConstraint(cnstr)
	if err != nil {
		t.Error(err)
	}
	allocs := testing.AllocsPerRun(2, func() {
		_, err = nextroute.NewSolution(model)
		if err != nil {
			t.Fatal(err)
		}
	})
	if allocs > 66 {
		t.Errorf("expected 66 allocations, got %v", allocs)
	}
}
