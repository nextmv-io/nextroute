package common_test

import (
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

func BenchmarkHaversine(b *testing.B) {
	from, _ := common.NewLocation(4.899, 19.372)
	to, _ := common.NewLocation(4.901, 23.373)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = common.HaversineUnsafe(from, to)
	}
}
