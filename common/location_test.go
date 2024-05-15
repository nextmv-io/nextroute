// © 2019-present nextmv.io inc

package common_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

func BenchmarkLocation(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	lon, lat := r.Float64()*360-180, r.Float64()*180-90
	l, _ := common.NewLocation(lon, lat)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.IsValid()
	}
}

func BenchmarkUnique(b *testing.B) {
	// test for different number of locations
	for _, n := range []int{10, 100, 1_000, 10_000} {
		locations := make(common.Locations, 0, n)
		r := rand.New(rand.NewSource(0))
		for i := 0; i < n; i++ {
			l, _ := common.NewLocation(r.Float64()*360-180, r.Float64()*180-90)
			if !l.IsValid() {
				b.Error("invalid location")
			}
			locations = append(locations, l)
		}
		b.Run(fmt.Sprintf("n=%v", n), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = locations.Unique()
			}
		})
	}
}

func TestLocation(t *testing.T) {
	l, err := common.NewLocation(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !l.IsValid() {
		t.Error("expected valid location")
	}
	invalid := common.NewInvalidLocation()
	if invalid.IsValid() {
		t.Error("expected invalid location")
	}
	l, err = common.NewLocation(180.1, 0)
	if err == nil {
		t.Error("expected error")
	}
	if l.IsValid() {
		t.Error("expected invalid location")
	}
}

func TestUnique(t *testing.T) {
	newLocation := func(lon, lat float64) common.Location {
		l, err := common.NewLocation(lon, lat)
		if err != nil {
			t.Fatal(err)
		}
		return l
	}
	locations := common.Locations{
		newLocation(0, 0),
		newLocation(123.234983434, 80.234983434),
		newLocation(123.234983434, 80.234983434),
		newLocation(0, 0),
		newLocation(0, 0),
		newLocation(0, 0),
		newLocation(0, 0),
	}
	unique := locations.Unique()
	if len(unique) != 2 {
		t.Errorf("expected 2 unique locations, got %v", len(unique))
	}
}

func BenchmarkCentroid(b *testing.B) {
	locations := make(common.Locations, 0, 2_000)
	r := rand.New(rand.NewSource(0))
	for i := 0; i < 2_000; i++ {
		l, _ := common.NewLocation(r.Float64()*360-180, r.Float64()*180-90)
		if !l.IsValid() {
			b.Error("invalid location")
		}
		locations = append(locations, l)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = locations.Centroid()
	}
}
