// © 2019-present nextmv.io inc

package common

import (
	"fmt"
	"math"
	"sort"
)

// Statistics describes statistics.
type Statistics struct {
	Count              float64
	Sum                float64
	Average            float64
	Maximum            float64
	MaximumIndex       int
	Minimum            float64
	MinimumIndex       int
	StandardDeviation  float64
	Quartile1          float64
	Quartile2          float64
	Quartile3          float64
	InterQuartileRange float64
	MidHinge           float64
}

// Report returns a string containing a report of the statistics.
func (s Statistics) Report() string {
	return fmt.Sprintf(
		"Count                       : %d\n"+
			"Sum                         : %f\n"+
			"Average                     : %f\n"+
			"Maximum                     : %f\n"+
			"Minimum                     : %f\n"+
			"Standard Deviation          : %f\n"+
			"Quartile 1                  : %f\n"+
			"Quartile 2                  : %f\n"+
			"Quartile 3                  : %f\n"+
			"Inter Quartile Range        : %f\n"+
			"IQR fence                   : [%f, %f]\n"+
			"Mid Hinge                   : %f\n",
		int(s.Count),
		s.Sum,
		s.Average,
		s.Maximum,
		s.Minimum,
		s.StandardDeviation,
		s.Quartile1,
		s.Quartile2,
		s.Quartile3,
		s.InterQuartileRange,
		s.Quartile1-1.5*s.InterQuartileRange,
		s.Quartile3+1.5*s.InterQuartileRange,
		s.MidHinge,
	)
}

// returns the median assuming data is sorted.
func median(data []float64) (float64, error) {
	length := len(data)

	if length == 0 {
		return math.NaN(), fmt.Errorf("no data")
	} else if length%2 == 0 {
		return (data[length/2-1] + data[length/2]) / 2.0, nil
	}

	return data[length/2], nil
}

// NewStatistics creates a new statistics object.
func NewStatistics[T any](v []T, f func(T) float64) Statistics {
	statistics := Statistics{
		Maximum:      -math.MaxFloat64,
		MaximumIndex: -1,
		Minimum:      math.MaxFloat64,
		MinimumIndex: -1,
	}
	length := len(v)

	if length == 0 {
		return statistics
	}

	values := make([]float64, len(v))

	for idx, t := range v {
		v := f(t)
		values[idx] = v
		statistics.Count++
		statistics.Sum += v
		if statistics.Maximum < v {
			statistics.Maximum = v
			statistics.MaximumIndex = idx
		}
		if statistics.Minimum > v {
			statistics.Minimum = v
			statistics.MinimumIndex = idx
		}
	}

	sort.Float64s(values)

	var cutOffPlace1 int
	var cutOffPlace2 int

	if length%2 == 0 {
		cutOffPlace1 = length / 2
		cutOffPlace2 = length / 2
	} else {
		cutOffPlace1 = (length - 1) / 2
		cutOffPlace2 = cutOffPlace1 + 1
	}

	statistics.Quartile1, _ = median(values[:cutOffPlace1])
	statistics.Quartile2, _ = median(values)
	statistics.Quartile3, _ = median(values[cutOffPlace2:])
	statistics.InterQuartileRange = statistics.Quartile3 - statistics.Quartile1
	statistics.MidHinge = statistics.Quartile1 + statistics.Quartile2/2.0

	statistics.Average = statistics.Sum / statistics.Count

	squaredDifferenceSum := 0.0

	for idx := range v {
		difference := values[idx] - statistics.Average
		squaredDifferenceSum += difference * difference
	}

	statistics.StandardDeviation = math.Sqrt(squaredDifferenceSum)

	return statistics
}
