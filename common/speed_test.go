// © 2019-present nextmv.io inc

package common_test

import (
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

func BenchmarkSpeedValue(b *testing.B) {
	speed := common.NewSpeed(100, common.KilometersPerHour)
	for i := 0; i < b.N; i++ {
		_ = speed.Value(common.MilesPerHour)
	}
}
