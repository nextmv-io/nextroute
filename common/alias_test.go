// © 2019-present nextmv.io inc

package common_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

const (
	samples      = 1_000_000
	errorEpsilon = 0.001
)

func TestAlias(t *testing.T) {
	testAlias(t, []float64{2, 2}, 22)
	testAlias(t, []float64{1, 2, 3}, 123)
	testAlias(t, []float64{6, 2, 1, 4, 2}, 62142)
	testAlias(t, []float64{1000, 1, 3, 10}, 10001310)
}

func testAlias(t *testing.T, weights []float64, seed int64) {
	sum := 0.0
	for i := 0; i < len(weights); i++ {
		sum += weights[i]
	}
	alias, err := common.NewAlias(weights)
	if err != nil {
		t.Fatal(err)
	}

	random := rand.New(rand.NewSource(seed))
	counts := make([]int64, len(weights))

	for i := 0; i < samples; i++ {
		counts[alias.Sample(random)]++
	}

	for i := 0; i < len(weights); i++ {
		count := float64(counts[i]) / samples
		if math.Abs(count-weights[i]/sum) > errorEpsilon {
			t.Errorf(
				"Counts did not match, got %v, expected %v, seed %v",
				count,
				weights[i]/sum,
				seed,
			)
		}
	}
}
