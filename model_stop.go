package nextroute

import (
	"fmt"
	"math"
	"time"

	common_internal "github.com/nextmv-io/nextroute/common"
	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

type stopImpl struct {
	windowChecker common_internal.IntervalChecker
	planUnit      nextroute.ModelPlanStopsUnit
	location      common.Location
	modelDataImpl
	vehicle           *modelVehicleImpl
	model             *modelImpl
	id                string
	closest           nextroute.ModelStops
	windows           [][2]float64
	earliestStartTime float64
	index             int
	measureIndex      int
	firstOrLast       bool
	fixed             bool
}

func (s *stopImpl) Model() nextroute.Model {
	return s.model
}

func (s *stopImpl) Vehicle() nextroute.ModelVehicle {
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

func (s *stopImpl) cacheClosestStops() {
	if s.HasPlanStopsUnit() {
		n := 20
		modelStopsDistanceQueries, err := NewModelStopsDistanceQueries(
			common.Filter(s.model.Stops(), func(stop nextroute.ModelStop) bool {
				return stop.Location().IsValid()
			}),
		)
		if err != nil {
			panic(err)
		}
		s.closest, err = modelStopsDistanceQueries.NearestStops(s, n)
		if err != nil {
			panic(err)
		}
	}
}

func (s *stopImpl) closestStops() nextroute.ModelStops {
	if s.closest == nil {
		s.model.mutex.Lock()
		defer s.model.mutex.Unlock()
		if s.closest == nil {
			s.cacheClosestStops()
		}
	}
	return s.closest
}

func (s *stopImpl) ClosestStops() nextroute.ModelStops {
	closest := make(nextroute.ModelStops, len(s.closestStops()))
	copy(closest, s.closestStops())
	return s.closest
}

func (s *stopImpl) HasPlanStopsUnit() bool {
	return s.planUnit != nil
}

func (s *stopImpl) PlanStopsUnit() nextroute.ModelPlanStopsUnit {
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
		panic("model is isLocked, a model is isLocked once a solution" +
			" has been created using this model")
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

	checker, err := common_internal.NewIntervalCheckerSliceLookup(windowsInSeconds)
	if err != nil {
		return err
	}
	s.windowChecker = checker

	return nil
}

func (s *stopImpl) SetEarliestStart(t time.Time) error {
	if s.model.IsLocked() {
		panic("model is isLocked, a model is isLocked once a solution" +
			" has been created using this model")
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
