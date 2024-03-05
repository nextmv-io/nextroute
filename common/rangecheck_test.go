// © 2019-present nextmv.io inc

package common_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/nextmv-io/nextroute/common"
)

// >>> Auxiliary functions

const testHorizonMinutes = 12 * 60

var testBaseTime = time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC)

// SampleTimes returns a number of random times within the test horizon.
func SampleTimes(samples int) []float64 {
	rand := rand.New(rand.NewSource(0))
	times := make([]float64, samples)
	for i := 0; i < len(times); i++ {
		times[i] = float64(testBaseTime.Add(time.Duration(rand.Intn(testHorizonMinutes*60)) * time.Second).Unix())
	}
	return times
}

// CreateIntervals returns a number of time intervals within the test horizon.
func CreateIntervals(count int) [][2]float64 {
	intervals := make([][2]float64, count)
	intervalLength := testHorizonMinutes / count / 2
	intervalDistance := testHorizonMinutes / count
	for i := 0; i < len(intervals); i++ {
		intervals[i] = [2]float64{
			float64(testBaseTime.Add(time.Duration(i*intervalDistance) * time.Minute).Unix()),
			float64(testBaseTime.Add(time.Duration(i*intervalDistance+intervalLength) * time.Minute).Unix()),
		}
	}
	return intervals
}

// >>> Unit tests

func TestIntervalLookup(t *testing.T) {
	checkers := []struct {
		checker func([][2]float64) (common.IntervalChecker, error)
		name    string
	}{
		{
			name:    "slice-lookup",
			checker: common.NewIntervalCheckerSliceLookup,
		},
	}
	for _, c := range checkers {
		intervals := CreateIntervals(5)
		times := SampleTimes(100000)
		checker, err := c.checker(intervals)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", c.name, err)
		}
		for _, ti := range times {
			expectedInInterval := false
			var expectedInterval [2]float64
			expectedEarliestNext := -1.0

			for i, w := range intervals {
				if w[0] <= ti && ti < w[1] {
					expectedInInterval = true
					expectedInterval = w
					break
				}
				if ti < w[0] && i-1 >= 0 && ti >= intervals[i-1][1] {
					expectedEarliestNext = w[0]
					break
				}
				if ti >= w[1] && i+1 < len(intervals) && ti < intervals[i+1][0] {
					expectedEarliestNext = intervals[i+1][0]
					break
				}
			}

			gotInInterval, gotEarliestNext := checker.Check(ti)

			if gotEarliestNext != expectedEarliestNext {
				t.Errorf("%s: expected earliest next %f, got %f for time %f",
					c.name,
					expectedEarliestNext,
					gotEarliestNext,
					ti)
			}

			if gotInInterval != expectedInInterval {
				formattedInterval := "nil"
				if expectedInInterval {
					formattedInterval = fmt.Sprintf("(%f,%f)", expectedInterval[0], expectedInterval[1])
				}
				t.Errorf("%s: expected %v, got %v for time %f; expected interval: %v",
					c.name,
					expectedInInterval,
					gotInInterval,
					ti,
					formattedInterval)
			}
		}
	}
}

// >>> Benchmarks

var BenchConfigs = []struct {
	intervalCount int
	timeCount     int
}{
	{intervalCount: 1, timeCount: 1000},
	{intervalCount: 5, timeCount: 1000},
	{intervalCount: 10, timeCount: 1000},
	{intervalCount: 50, timeCount: 1000},
	{intervalCount: 100, timeCount: 1000},
}

func benchInt(b *testing.B, c func([][2]float64) (common.IntervalChecker, error), w, t int) {
	intervals := CreateIntervals(w)
	times := SampleTimes(t)
	l, err := c(intervals)
	if err != nil {
		b.Errorf("unexpected error: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, t := range times {
			l.Check(t)
		}
	}
}

// BenchmarkIntervalSliceLookup is a benchmark for the slice based interval
// checker.
func BenchmarkIntervalSliceLookup(b *testing.B) {
	for _, config := range BenchConfigs {
		b.Run(strconv.Itoa(config.intervalCount)+"-"+strconv.Itoa(config.timeCount), func(b *testing.B) {
			benchInt(b, common.NewIntervalCheckerSliceLookup, config.intervalCount, config.timeCount)
		})
	}
}
