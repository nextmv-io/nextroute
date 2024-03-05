// © 2019-present nextmv.io inc

package factory

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/nextmv-io/nextroute"
	nmerror "github.com/nextmv-io/nextroute/common/errors"
)

// NewDurationGroupsExpression returns a duration group expression.
func NewDurationGroupsExpression(numberOfStops, numberOfVehicles int) DurationGroupsExpression {
	durationGroupExpression := &durationGroupDurationImpl{
		index:           nextroute.NewModelExpressionIndex(),
		durations:       make([]float64, numberOfStops+2*numberOfVehicles),
		groupDuration:   make([]float64, numberOfStops+2*numberOfVehicles),
		toGroupIndex:    make([]int64, numberOfStops+2*numberOfVehicles),
		stopIndexToStop: make([]nextroute.ModelStop, numberOfStops+2*numberOfVehicles),
		groupCount:      0,
	}
	for i := 0; i < len(durationGroupExpression.toGroupIndex); i++ {
		durationGroupExpression.toGroupIndex[i] = -1
	}
	return durationGroupExpression
}

// DurationGroup is a group of stops with a duration.
type DurationGroup struct {
	Stops    nextroute.ModelStops
	Duration time.Duration
}

// DurationGroupsExpression is an interface implementing semantics of duration
// groups.
type DurationGroupsExpression interface {
	nextroute.DurationExpression

	// SetStopDuration sets the process duration for a stop.
	SetStopDuration(nextroute.ModelStop, time.Duration)
	// SetGroupDuration sets the process duration for a group.
	SetGroupDuration(nextroute.ModelStops, time.Duration) error
	// AddGroup adds a group of stops and their duration to the expression.
	AddGroup(nextroute.ModelStops, time.Duration) error
	// Groups returns the groups of stops.
	Groups() []DurationGroup
	// Durations returns the durations of all stops.
	Durations() map[nextroute.ModelStop]time.Duration
}

type durationGroupDurationImpl struct {
	groupDuration   []float64
	toGroupIndex    []int64
	durations       []float64
	stopIndexToStop []nextroute.ModelStop
	index           int
	groupCount      int64
}

// Groups implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) Groups() []DurationGroup {
	returnGroup := make([]DurationGroup, d.groupCount)
	for i := range returnGroup {
		returnGroup[i].Duration = time.Duration(d.groupDuration[i]) * time.Second
	}
	for stopIndex, group := range d.toGroupIndex {
		if group < 0 {
			continue
		}
		returnGroup[group].Stops = append(returnGroup[group].Stops, d.stopIndexToStop[stopIndex])
	}
	return returnGroup
}

// SetGroupDuration implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) SetGroupDuration(stops nextroute.ModelStops, duration time.Duration) error {
	if len(stops) == 0 {
		return nmerror.NewInputDataError(fmt.Errorf("cannot set duration for empty group"))
	}
	// ensure that all stops are in the same group
	group := d.toGroupIndex[stops[0].Index()]
	if group < 0 {
		return nmerror.NewInputDataError(fmt.Errorf("the stop %s is not in a group", stops[0].ID()))
	}
	for _, stop := range stops[1:] {
		otherGroup := d.toGroupIndex[stop.Index()]
		if otherGroup < 0 {
			return nmerror.NewInputDataError(fmt.Errorf("the stop %s is not in a group", stop.ID()))
		}
		if group != otherGroup {
			return nmerror.NewInputDataError(fmt.Errorf("all stops must be in the same group"))
		}
	}
	d.groupDuration[group] = duration.Seconds()
	return nil
}

// Duration implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) Duration(
	_ nextroute.ModelVehicleType,
	from nextroute.ModelStop,
	to nextroute.ModelStop,
) time.Duration {
	return time.Duration(d.Value(nil, from, to)) * time.Second
}

// HasNegativeValues implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) HasNegativeValues() bool {
	return false
}

// HasPositiveValues implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) HasPositiveValues() bool {
	return true
}

// Index implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) Index() int {
	return d.index
}

// Name implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) Name() string {
	return "duration_group__expression"
}

// SetStopDuration implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) SetStopDuration(
	stop nextroute.ModelStop,
	duration time.Duration,
) {
	d.durations[stop.Index()] = duration.Seconds()
	d.stopIndexToStop[stop.Index()] = stop
}

// Durations returns the durations of all stops.
func (d *durationGroupDurationImpl) Durations() map[nextroute.ModelStop]time.Duration {
	copiedMap := make(map[nextroute.ModelStop]time.Duration, len(d.durations))
	for k, v := range d.durations {
		stop := d.stopIndexToStop[k]
		if stop == nil {
			continue
		}
		copiedMap[stop] = time.Duration(v) * time.Second
	}
	return copiedMap
}

// GroupDurations returns the durations of all groups.
func (d *durationGroupDurationImpl) GroupDurations() map[int]time.Duration {
	copiedMap := make(map[int]time.Duration, len(d.groupDuration))
	for k, v := range d.groupDuration {
		if v >= 0 {
			copiedMap[k] = time.Duration(v) * time.Second
		}
	}
	return copiedMap
}

// AddGroup implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) AddGroup(stops nextroute.ModelStops, duration time.Duration) error {
	groupCount := atomic.AddInt64(&d.groupCount, 1) - 1
	for _, stop := range stops {
		if d.toGroupIndex[stop.Index()] >= 0 {
			return nmerror.NewInputDataError(fmt.Errorf("the stop %s is in two groups or twice in the same group."+
				" a stop can only be assigned once", stop.ID()))
		}
		d.toGroupIndex[stop.Index()] = groupCount
		d.stopIndexToStop[stop.Index()] = stop
	}
	d.groupDuration[groupCount] = duration.Seconds()
	return nil
}

// SetName implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) SetName(string) {
}

// Value implements DurationGroupsExpression.
func (d *durationGroupDurationImpl) Value(
	_ nextroute.ModelVehicleType,
	from nextroute.ModelStop,
	to nextroute.ModelStop,
) float64 {
	toIndex := to.Index()
	toGroup := d.toGroupIndex[toIndex]
	if toGroup == -1 {
		return d.durations[toIndex]
	}
	fromIndex := from.Index()
	fromGroup := d.toGroupIndex[fromIndex]
	if fromGroup == toGroup {
		return d.durations[toIndex]
	}
	return d.durations[toIndex] + d.groupDuration[toGroup]
}
