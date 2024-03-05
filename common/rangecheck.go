// © 2019-present nextmv.io inc

// Package common provides common functionality for the nextroute plugin.
package common

import (
	"fmt"
	"math"
	"slices"
	"time"

	nmerror "github.com/nextmv-io/nextroute/common/errors"
)

// Interval represents a time interval.
type Interval struct {
	// Min is the minimum time in the interval (inclusive; int seconds since
	// epoch).
	Min float64 `json:"min"`
	// Max is the maximum time in the interval (exclusive; int seconds since
	// epoch).
	Max float64 `json:"max"`
}

// IntervalChecker is used to check whether a given time is in an interval.
type IntervalChecker interface {
	// Check returns whether the time is in an interval. Furthermore, it returns
	// the next time an interval will open (if any).
	Check(t float64) (inInterval bool, earliestNext float64)
}

// secondsToMinutes converts seconds to minutes. This is used to discretize the
// time slots.
func secondsToMinutes(seconds float64) int {
	return int(seconds / 60)
}

// slotInfo contains information about a time slot.
type slotInfo struct {
	// interval is the interval overlapping this time slot (if any; nil
	// otherwise).
	interval *Interval
	// next is the interval that will open next after this time slot (if any; nil
	// otherwise).
	next *Interval
	// inInterval is true if the time slot is covered by a interval.
	inInterval bool
}

type infoTuple struct {
	info slotInfo
	i    int
}

// toSlotInfo converts the given intervals to a slice of slotInfo structs.
func toSlotInfo(intervals []Interval) ([]infoTuple, int) {
	minimum := int(intervals[0].Min / 60)
	maximum := int(intervals[len(intervals)-1].Max/60) + 1
	timeslots := make([]infoTuple, maximum-minimum)

	for i := minimum; i < maximum; i++ {
		second := float64(i * 60)
		var next *Interval
		isInInterval := false
		var inInterval *Interval
		for interval := range intervals {
			if second >= intervals[interval].Min && second < intervals[interval].Max {
				isInInterval = true
				inInterval = &intervals[interval]
				break
			}
			if second < intervals[interval].Min && i-1 >= 0 && second >= intervals[interval-1].Max {
				next = &intervals[interval]
				break
			}
			if second >= intervals[interval].Max && i+1 < len(intervals) && second < intervals[interval+1].Min {
				next = &intervals[interval+1]
				break
			}
		}

		timeslots[i-minimum] = infoTuple{
			info: slotInfo{
				interval:   inInterval,
				next:       next,
				inInterval: isInInterval,
			},
			i: i,
		}
	}

	return timeslots, minimum
}

// inInterval returns true if the given time is in the interval.
func inInterval(w Interval, t float64) bool {
	return t >= w.Min && t < w.Max
}

func processIntervals(intervals [][2]float64) ([]Interval, error) {
	converted := make([]Interval, len(intervals))
	for i, iv := range intervals {
		converted[i] = Interval{
			Min: iv[0],
			Max: iv[1],
		}
	}

	slices.SortFunc(converted, func(i, j Interval) int {
		if i.Min-j.Min < 0 {
			return -1
		}
		if i.Min-j.Min > 0 {
			return 1
		}
		return 0
	})

	for i, w1 := range converted {
		for j, w2 := range converted {
			if i == j {
				continue
			}
			if inInterval(w1, w2.Min) || inInterval(w1, w2.Max-1) {
				return nil, nmerror.NewInputDataError(fmt.Errorf("intervals %s and %s overlap",
					time.Unix(int64(math.Round(w1.Min)), 0).Format(time.RFC3339),
					time.Unix(int64(math.Round(w2.Min)), 0).Format(time.RFC3339)))
			}
		}
	}

	for _, w := range converted {
		if w.Min < 0 || w.Max < 0 {
			return nil, nmerror.NewInputDataError(fmt.Errorf("interval %s has negative time",
				time.Unix(int64(math.Round(w.Min)), 0).Format(time.RFC3339)))
		}
	}

	return converted, nil
}

// >>> Slice lookup implementation

type intervalCheckerSliceLookup struct {
	// timeSlots tracks per time-slot whether it is in a interval.
	timeSlots []slotInfo
	// sliceOffset is the offset of the slice of time slots in terms of unix
	// time seconds.
	sliceOffset int
	// earliestNext is the earliest interval opening.
	earliestNext float64
}

// NewIntervalCheckerSliceLookup returns a new slice based interval checker.
func NewIntervalCheckerSliceLookup(intervals [][2]float64) (IntervalChecker, error) {
	ivs, err := processIntervals(intervals)
	if err != nil {
		return nil, err
	}
	slots, offset := toSlotInfo(ivs)
	infos := make([]slotInfo, len(slots))
	for i, t := range slots {
		infos[i] = t.info
	}
	return &intervalCheckerSliceLookup{
		timeSlots:    infos,
		sliceOffset:  offset,
		earliestNext: ivs[0].Min,
	}, nil
}

func (w *intervalCheckerSliceLookup) Check(tf float64) (bool, float64) {
	t := secondsToMinutes(tf)
	idx := t - w.sliceOffset
	if idx < 0 {
		return false, w.earliestNext
	}
	if idx >= len(w.timeSlots) {
		return false, -1
	}
	slot := w.timeSlots[idx]
	if slot.next != nil {
		return slot.inInterval, slot.next.Min
	}
	return slot.inInterval, -1
}
