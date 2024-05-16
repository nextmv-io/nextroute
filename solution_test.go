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
