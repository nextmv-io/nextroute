// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"math"
	"time"

	"github.com/nextmv-io/nextroute/common"
)

// ModelStops is a slice of stops.
type ModelStops []*ModelStop

// ModelStop is a stop to be assigned to a vehicle.
type ModelStop struct {
	windowChecker common.IntervalChecker
	planUnit      ModelPlanStopsUnit
	location      common.Location
	modelDataImpl
	vehicle           *ModelVehicle
	model             *modelImpl
	id                string
	closest           ModelStops
	windows           [][2]float64
	earliestStartTime float64
	index             int
	measureIndex      int
	firstOrLast       bool
	fixed             bool
}

// Model returns the model of the stop.
func (s *ModelStop) Model() Model {
	return s.model
}

func (s *ModelStop) Vehicle() *ModelVehicle {
	return s.vehicle
}

func (s *ModelStop) String() string {
	return fmt.Sprintf("stop{%s[%v]}",
		s.id,
		s.index,
	)
}

// IsFirstOrLast returns true if the stop is the first or last stop of one
// or more vehicles. A stop which is the first or last stop of one or more
// vehicles is not allowed to be part of a plan unit. A stop which is the
// first or last stop of one or more vehicles is by definition fixed.
func (s *ModelStop) IsFirstOrLast() bool {
	return s.firstOrLast
}

// IsFixed returns true if fixed.
func (s *ModelStop) IsFixed() bool {
	return s.fixed
}

func (s *ModelStop) cacheClosestStops() error {
	if s.HasPlanStopsUnit() {
		n := 20
		modelStopsDistanceQueries, err := NewModelStopsDistanceQueries(
			common.Filter(s.model.Stops(), func(stop *ModelStop) bool {
				return stop.Location().IsValid()
			}),
		)
		if err != nil {
			return err
		}
		s.closest, err = modelStopsDistanceQueries.NearestStops(s, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ModelStop) closestStops() (ModelStops, error) {
	if s.closest == nil {
		s.model.mutex.Lock()
		defer s.model.mutex.Unlock()
		if s.closest == nil {
			err := s.cacheClosestStops()
			if err != nil {
				return nil, err
			}
		}
	}
	return s.closest, nil
}

// ClosestStops returns a slice containing the closest stops to the
// invoking stop. The slice is sorted by increasing distance to the
// location. The slice first stop is the stop itself. The distance used
// is the common.Haversine distance between the stops. All the stops
// in the model are used in the slice. Slice with similar distance are
// sorted by their index (increasing).
func (s *ModelStop) ClosestStops() (ModelStops, error) {
	closestStops, err := s.closestStops()
	if err != nil {
		return nil, err
	}
	closest := make(ModelStops, len(closestStops))
	copy(closest, closestStops)
	return s.closest, nil
}

// HasPlanStopsUnit returns true if the stop belongs to a plan unit. For example,
// start and end stops of a vehicle do not belong to a plan unit.
func (s *ModelStop) HasPlanStopsUnit() bool {
	return s.planUnit != nil
}

// PlanStopsUnit returns the [ModelPlanStopsUnit] associated with the stop.
// A stop is associated with at most one plan unit. Can be nil if the stop
// is not part of a stops plan unit.
func (s *ModelStop) PlanStopsUnit() ModelPlanStopsUnit {
	return s.planUnit
}

// SetID sets the identifier of the stop. This identifier is not used by
// nextroute, and therefore it does not have to be unique for nextroute
// internally. However, to make this ID useful for debugging and reporting
// it should be made unique.
func (s *ModelStop) SetID(id string) {
	s.id = id
}

// Index returns the unique index of the stop.
func (s *ModelStop) Index() int {
	return s.index
}

// MeasureIndex returns the measure index of the invoking stop . This index
// is not necessarily unique.
// This index is used by the model expression constructed by the factory
// NewMeasureByIndexExpression to calculate the value of the measure
// expression. By default, the measure index is the same as the index of
// the stop.
func (s *ModelStop) MeasureIndex() int {
	return s.measureIndex
}

// SetMeasureIndex sets the reference index of the stop, by default the
// measure index is the same as the index of the stop.
// This index is used by the model expression constructed by the factory
// NewMeasureByIndexExpression to calculate the value of the measure
// expression.
func (s *ModelStop) SetMeasureIndex(index int) {
	s.measureIndex = index
}

// ID returns the identifier of the stop.
func (s *ModelStop) ID() string {
	return s.id
}

// Location returns the location of the stop.
func (s *ModelStop) Location() common.Location {
	return s.location
}

// Windows returns the time windows of the stop.
func (s *ModelStop) Windows() [][2]time.Time {
	windows := make([][2]time.Time, len(s.windows))
	for i, window := range s.windows {
		windows[i] = [2]time.Time{
			s.model.Epoch().Add(time.Duration(window[0]) * s.model.DurationUnit()),
			s.model.Epoch().Add(time.Duration(window[1]) * s.model.DurationUnit()),
		}
	}
	return windows
}

// SetWindows sets the time windows of the stop.
func (s *ModelStop) SetWindows(windows [][2]time.Time) error {
	if s.model.IsLocked() {
		return fmt.Errorf("can not set window of stop %s once the model is locked",
			s.ID(),
		)
	}

	if len(windows) == 0 {
		return nil
	}

	for i, window := range windows {
		startTime := window[0]
		endTime := window[1]
		if startTime.After(endTime) {
			return fmt.Errorf("window %d is invalid, start time %s is after end time %s", i,
				startTime.Format(time.RFC3339),
				endTime.Format(time.RFC3339),
			)
		}
		if i > 0 && startTime.Before(windows[i-1][1]) {
			return fmt.Errorf("windows %d and %d are overlapping, start time %s is before end time %s", i-1, i,
				windows[i-1][1].Format(time.RFC3339),
				startTime.Format(time.RFC3339),
			)
		}
		if startTime.Second() != 0 || startTime.Nanosecond() != 0 {
			return fmt.Errorf("window %d is invalid, start time %v is not on a minute boundary", i, startTime)
		}
		if endTime.Second() != 0 || endTime.Nanosecond() != 0 {
			return fmt.Errorf("window %d is invalid, end time %v is not on a minute boundary", i, endTime)
		}
	}

	windowsInSeconds := make([][2]float64, len(windows))
	for i, window := range windows {
		windowsInSeconds[i] = [2]float64{
			window[0].Sub(s.model.Epoch()).Seconds(),
			window[1].Sub(s.model.Epoch()).Seconds(),
		}
	}
	s.windows = windowsInSeconds

	checker, err := common.NewIntervalCheckerSliceLookup(windowsInSeconds)
	if err != nil {
		return err
	}
	s.windowChecker = checker

	return nil
}

// SetEarliestStart sets the earliest start time of the stop.
func (s *ModelStop) SetEarliestStart(t time.Time) error {
	if s.model.IsLocked() {
		return fmt.Errorf("can not set earliest start of stop %s once the model is locked",
			s.ID(),
		)
	}

	s.earliestStartTime = t.Sub(s.model.Epoch()).Seconds()
	return nil
}

// EarliestStart returns the earliest start time of the stop.
func (s *ModelStop) EarliestStart() (t time.Time) {
	return s.model.Epoch().Add(time.Duration(s.earliestStartTime) * time.Second)
}

func (s *ModelStop) validate() error {
	if s.earliestStartTime != 0.0 && s.windows != nil {
		return fmt.Errorf(
			"stop `%v` has both earliest start and windows set",
			s.ID(),
		)
	}
	return nil
}

// ToEarliestStartValue returns the earliest start time if the vehicle
// arrives at the stop at the given arrival time in seconds since
// [Model.Epoch].
func (s *ModelStop) ToEarliestStartValue(arrivalTime float64) float64 {
	if s.windowChecker != nil {
		inWindow, windowOpening := s.windowChecker.Check(arrivalTime)
		if inWindow {
			return arrivalTime
		} else if windowOpening > 0 {
			return windowOpening
		}
		// arrivalTime is after the last window closes, so it is not clear what
		// the earliest start time should be. to return arrivalTime
		return arrivalTime
	}
	if s.earliestStartTime == 0.0 {
		return arrivalTime
	}
	return math.Max(arrivalTime, s.earliestStartTime)
}

func (s *ModelStop) canIncurWaitingTime() bool {
	return s.windowChecker != nil || s.earliestStartTime != 0.0
}
