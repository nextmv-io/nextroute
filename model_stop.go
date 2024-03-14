// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"math"
	"time"

	"github.com/nextmv-io/nextroute/common"
)

// ModelStop is a stop to be assigned to a vehicle.
type ModelStop interface {
	ModelData

	// ClosestStops returns a slice containing the closest stops to the
	// invoking stop. The slice is sorted by increasing distance to the
	// location. The slice first stop is the stop itself. The distance used
	// is the common.Haversine distance between the stops. All the stops
	// in the model are used in the slice. Slice with similar distance are
	// sorted by their index (increasing).
	ClosestStops() (ModelStops, error)

	// HasPlanStopsUnit returns true if the stop belongs to a plan unit. For example,
	// start and end stops of a vehicle do not belong to a plan unit.
	HasPlanStopsUnit() bool

	// ID returns the identifier of the stop.
	ID() string

	// Index returns the unique index of the stop.
	Index() int

	// IsFirstOrLast returns true if the stop is the first or last stop of one
	// or more vehicles. A stop which is the first or last stop of one or more
	// vehicles is not allowed to be part of a plan unit. A stop which is the
	// first or last stop of one or more vehicles is by definition fixed.
	IsFirstOrLast() bool

	// IsFixed returns true if fixed.
	IsFixed() bool

	// Location returns the location of the stop.
	Location() common.Location

	// Model returns the model of the stop.
	Model() Model

	// EarliestStart returns the earliest start time of the stop.
	EarliestStart() (t time.Time)

	// Windows returns the time windows of the stop.
	Windows() [][2]time.Time

	// PlanStopsUnit returns the [ModelPlanStopsUnit] associated with the stop.
	// A stop is associated with at most one plan unit. Can be nil if the stop
	// is not part of a stops plan unit.
	PlanStopsUnit() ModelPlanStopsUnit

	// MeasureIndex returns the measure index of the invoking stop . This index
	// is not necessarily unique.
	// This index is used by the model expression constructed by the factory
	// NewMeasureByIndexExpression to calculate the value of the measure
	// expression. By default, the measure index is the same as the index of
	// the stop.
	MeasureIndex() int

	// SetEarliestStart sets the earliest start time of the stop.
	SetEarliestStart(t time.Time) error

	// SetMeasureIndex sets the reference index of the stop, by default the
	// measure index is the same as the index of the stop.
	// This index is used by the model expression constructed by the factory
	// NewMeasureByIndexExpression to calculate the value of the measure
	// expression.
	SetMeasureIndex(int)

	// SetWindows sets the time windows of the stop.
	SetWindows(windows [][2]time.Time) error

	// ToEarliestStartValue returns the earliest start time if the vehicle
	// arrives at the stop at the given arrival time in seconds since
	// [Model.Epoch].
	ToEarliestStartValue(arrival float64) float64

	// SetID sets the identifier of the stop. This identifier is not used by
	// nextroute, and therefore it does not have to be unique for nextroute
	// internally. However, to make this ID useful for debugging and reporting
	// it should be made unique.
	SetID(string)
}

// ModelStops is a slice of stops.
type ModelStops []ModelStop

type stopImpl struct {
	windowChecker common.IntervalChecker
	planUnit      ModelPlanStopsUnit
	location      common.Location
	modelDataImpl
	vehicle           *modelVehicleImpl
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

func (s *stopImpl) Model() Model {
	return s.model
}

func (s *stopImpl) Vehicle() ModelVehicle {
	return s.vehicle
}

func (s *stopImpl) String() string {
	return fmt.Sprintf("stop{%s[%v]}",
		s.id,
		s.index,
	)
}

func (s *stopImpl) IsFirstOrLast() bool {
	return s.firstOrLast
}

func (s *stopImpl) IsFixed() bool {
	return s.fixed
}

func (s *stopImpl) cacheClosestStops() error {
	if s.HasPlanStopsUnit() {
		n := 20
		modelStopsDistanceQueries, err := NewModelStopsDistanceQueries(
			common.Filter(s.model.Stops(), func(stop ModelStop) bool {
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

func (s *stopImpl) closestStops() (ModelStops, error) {
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

func (s *stopImpl) ClosestStops() (ModelStops, error) {
	closestStops, err := s.closestStops()
	if err != nil {
		return nil, err
	}
	closest := make(ModelStops, len(closestStops))
	copy(closest, closestStops)
	return s.closest, nil
}

func (s *stopImpl) HasPlanStopsUnit() bool {
	return s.planUnit != nil
}

func (s *stopImpl) PlanStopsUnit() ModelPlanStopsUnit {
	return s.planUnit
}

func (s *stopImpl) SetID(id string) {
	s.id = id
}

func (s *stopImpl) Index() int {
	return s.index
}

func (s *stopImpl) MeasureIndex() int {
	return s.measureIndex
}

func (s *stopImpl) SetMeasureIndex(index int) {
	s.measureIndex = index
}

func (s *stopImpl) ID() string {
	return s.id
}

func (s *stopImpl) Location() common.Location {
	return s.location
}

func (s *stopImpl) Windows() [][2]time.Time {
	windows := make([][2]time.Time, len(s.windows))
	for i, window := range s.windows {
		windows[i] = [2]time.Time{
			s.model.Epoch().Add(time.Duration(window[0]) * s.model.DurationUnit()),
			s.model.Epoch().Add(time.Duration(window[1]) * s.model.DurationUnit()),
		}
	}
	return windows
}

func (s *stopImpl) SetWindows(windows [][2]time.Time) error {
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

func (s *stopImpl) SetEarliestStart(t time.Time) error {
	if s.model.IsLocked() {
		return fmt.Errorf("can not set earliest start of stop %s once the model is locked",
			s.ID(),
		)
	}

	s.earliestStartTime = t.Sub(s.model.Epoch()).Seconds()
	return nil
}

func (s *stopImpl) EarliestStart() (t time.Time) {
	return s.model.Epoch().Add(time.Duration(s.earliestStartTime) * time.Second)
}

func (s *stopImpl) validate() error {
	if s.earliestStartTime != 0.0 && s.windows != nil {
		return fmt.Errorf(
			"stop `%v` has both earliest start and windows set",
			s.ID(),
		)
	}
	return nil
}

// ToEarliestStartValue determines the earliest time to start servicing the stop,
// given the current (given) time.
func (s *stopImpl) ToEarliestStartValue(arrivalTime float64) float64 {
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

func (s *stopImpl) canIncurWaitingTime() bool {
	return s.windowChecker != nil || s.earliestStartTime != 0.0
}
