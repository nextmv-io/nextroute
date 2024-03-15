// © 2019-present nextmv.io inc

package common_test

import (
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

func BenchmarkFilter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		numberOfValues := 100
		values := make([]int, numberOfValues)
		for i := 0; i < numberOfValues; i++ {
			values[i] = i
		}
		_ = common.Filter(values, func(vehicle int) bool {
			return vehicle%2 == 0
		})
	}
}

func TestDefensiveCopy(t *testing.T) {
	numberOfValues := 100
	values := make([]int, numberOfValues)
	for i := 0; i < numberOfValues; i++ {
		values[i] = i
	}
	copiedValues := common.DefensiveCopy(values)
	for i := 0; i < numberOfValues; i++ {
		if values[i] != copiedValues[i] {
			t.Errorf("Expected %v, got %v", values[i], copiedValues[i])
		}
	}
}
