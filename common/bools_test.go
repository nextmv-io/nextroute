package common_test

import (
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

func TestBools(t *testing.T) {
	b := common.NewBools(100, true)
	for i := 0; i < 100; i++ {
		if !b.Get(i) {
			t.Errorf("Expected true, got false")
		}
	}
	b.Set(50, false)
	if b.Get(50) {
		t.Errorf("Expected false, got true")
	}
	b.Set(50, false)
	b.Set(50, false)
	if b.Get(50) {
		t.Errorf("Expected false, got true")
	}
	b.Set(50, true)
	if !b.Get(50) {
		t.Errorf("Expected true, got false")
	}
	b.Set(50, true)
	b.Set(50, true)
	if !b.Get(50) {
		t.Errorf("Expected true, got false")
	}
	b.Set(0, false)
	if b.Get(0) {
		t.Errorf("Expected false, got true")
	}
	b.Set(100, false)
	if b.Get(100) {
		t.Errorf("Expected false, got true")
	}
}

func BenchmarkBools(b *testing.B) {
	bools := common.NewBools(100, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 100; i++ {
			bools.Set(i, false)
		}
		for i := 0; i < 100; i++ {
			bools.Get(i)
		}
	}
}
