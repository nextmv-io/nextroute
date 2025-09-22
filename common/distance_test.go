package common_test

import (
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

func BenchmarkDistance(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := common.NewDistance(1, common.Meters)
		_ = d.Value(common.Meters)
	}
}
