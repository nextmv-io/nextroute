// © 2019-present nextmv.io inc

package common_test

import (
	"math"
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

func TestStatistics(t *testing.T) {
	data := make([]float64, 0)

	statistics := common.NewStatistics(data, func(d float64) float64 {
		return d
	})

	testResult(t, data, common.Statistics{
		Count:              0,
		Sum:                0,
		Average:            0,
		Maximum:            -math.MaxFloat64,
		MaximumIndex:       -1,
		Minimum:            math.MaxFloat64,
		MinimumIndex:       -1,
		StandardDeviation:  0,
		Quartile1:          0,
		Quartile2:          0,
		Quartile3:          0,
		InterQuartileRange: 0,
		MidHinge:           0,
	}, statistics)

	data = []float64{1}

	statistics = common.NewStatistics(data, func(d float64) float64 {
		return d
	})

	testResult(t, data, common.Statistics{
		Count:              1,
		Sum:                1,
		Average:            1,
		Maximum:            1,
		MaximumIndex:       0,
		Minimum:            1,
		MinimumIndex:       0,
		StandardDeviation:  0,
		Quartile1:          math.NaN(),
		Quartile2:          1,
		Quartile3:          math.NaN(),
		InterQuartileRange: math.NaN(),
		MidHinge:           math.NaN(),
	}, statistics)

	data = []float64{1, 2}

	statistics = common.NewStatistics(data, func(d float64) float64 {
		return d
	})

	testResult(t, data, common.Statistics{
		Count:              2,
		Sum:                3,
		Average:            1.5,
		Maximum:            2,
		MaximumIndex:       1,
		Minimum:            1,
		MinimumIndex:       0,
		StandardDeviation:  0.7071067811865476,
		Quartile1:          1,
		Quartile2:          1.5,
		Quartile3:          2,
		InterQuartileRange: 1,
		MidHinge:           1.75,
	}, statistics)
}

func testResult(
	t *testing.T,
	data []float64,
	expected, actual common.Statistics,
) {
	if testNotEqual(expected.Count, actual.Count) {
		t.Errorf(
			"Count: expected %v, actual %v, data %v",
			expected.Count,
			actual.Count,
			data,
		)
	}
	if testNotEqual(expected.Sum, actual.Sum) {
		t.Errorf(
			"Sum: expected %v, actual %v, data %v",
			expected.Sum,
			actual.Sum,
			data,
		)
	}
	if testNotEqual(expected.Average, actual.Average) {
		t.Errorf(
			"Average: expected %v, actual %v, data %v",
			expected.Average,
			actual.Average,
			data,
		)
	}
	if testNotEqual(expected.Maximum, actual.Maximum) {
		t.Errorf(
			"Maximum: expected %v, actual %v, data %v",
			expected.Maximum,
			actual.Maximum,
			data,
		)
	}

	if expected.MaximumIndex != actual.MaximumIndex {
		t.Errorf(
			"MaximumIndex: expected %v, actual %v, data %v",
			expected.MaximumIndex,
			actual.MaximumIndex,
			data,
		)
	}
	if testNotEqual(expected.Minimum, actual.Minimum) {
		t.Errorf(
			"Minimum: expected %v, actual %v, data %v",
			expected.Minimum,
			actual.Minimum,
			data,
		)
	}
	if expected.MinimumIndex != actual.MinimumIndex {
		t.Errorf(
			"MinimumIndex: expected %v, actual %v, data %v",
			expected.MinimumIndex,
			actual.MinimumIndex,
			data,
		)
	}
	if testNotEqual(expected.StandardDeviation, actual.StandardDeviation) {
		t.Errorf(
			"StandardDeviation: expected %v, actual %v, data %v",
			expected.StandardDeviation,
			actual.StandardDeviation,
			data,
		)
	}
	if testNotEqual(expected.Quartile1, actual.Quartile1) {
		t.Errorf(
			"Quartile1: expected %v, actual %v, data %v",
			expected.Quartile1,
			actual.Quartile1,
			data,
		)
	}
	if testNotEqual(expected.Quartile2, actual.Quartile2) {
		t.Errorf(
			"Quartile2: expected %v, actual %v, data %v",
			expected.Quartile2,
			actual.Quartile2,
			data,
		)
	}
	if testNotEqual(expected.Quartile3, actual.Quartile3) {
		t.Errorf(
			"Quartile3: expected %v, actual %v, data %v",
			expected.Quartile3,
			actual.Quartile3,
			data,
		)
	}
	if testNotEqual(expected.InterQuartileRange, actual.InterQuartileRange) {
		t.Errorf(
			"InterQuartileRange: expected %v, actual %v, data %v",
			expected.InterQuartileRange,
			actual.InterQuartileRange,
			data,
		)
	}
	if testNotEqual(expected.MidHinge, actual.MidHinge) {
		t.Errorf(
			"MidHinge: expected %v, actual %v, data %v",
			expected.MidHinge,
			actual.MidHinge,
			data,
		)
	}
}

func testEquals(expected, actual float64) bool {
	return (math.IsNaN(expected) && math.IsNaN(actual)) ||
		math.Abs(expected-actual) < 0.0000001
}

func testNotEqual(expected, actual float64) bool {
	return !testEquals(expected, actual)
}
