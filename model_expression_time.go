// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"time"
)

// TimeExpression is a ModelExpression that returns a time.
type TimeExpression interface {
	ModelExpression
	// Time returns the time for the given parameters.
	Time(ModelVehicleType, ModelStop, ModelStop) time.Time
}

// StopTimeExpression is a ModelExpression that returns a time per stop and
// allows to set the time per stop.
type StopTimeExpression interface {
	ModelExpression
	// Time returns the time for the given stop.
	Time(stop ModelStop) time.Time
	// SetTime sets the time for the given stop.
	SetTime(ModelStop, time.Time)
}

// NewTimeExpression returns a new TimeExpression.
func NewTimeExpression(
	expression ModelExpression,
	epoch time.Time,
) TimeExpression {
	name := expression.Name() + " since " + epoch.String()
	return &timeExpressionImpl{
		index:      NewModelExpressionIndex(),
		expression: expression,
		epoch:      epoch,
		name:       name,
	}
}

// NewStopTimeExpression returns a new StopTimeExpression.
func NewStopTimeExpression(
	name string,
	defaultTime time.Time,
) StopTimeExpression {
	return &stopTimeExpressionImpl{
		index:        NewModelExpressionIndex(),
		name:         name,
		defaultTime:  defaultTime,
		defaultValue: -1,
	}
}

type timeExpressionImpl struct {
	epoch      time.Time
	expression ModelExpression
	name       string
	index      int
}

func (t *timeExpressionImpl) HasNegativeValues() bool {
	return t.expression.HasNegativeValues()
}

func (t *timeExpressionImpl) HasPositiveValues() bool {
	return t.expression.HasPositiveValues()
}

func (t *timeExpressionImpl) String() string {
	return fmt.Sprintf("Time expression[%v] '%s'",
		t.index,
		t.Name(),
	)
}

func (t *timeExpressionImpl) Index() int {
	return t.index
}

func (t *timeExpressionImpl) Name() string {
	return t.name
}

func (t *timeExpressionImpl) SetName(n string) {
	t.name = n
}

func (t *timeExpressionImpl) Value(
	vehicleType ModelVehicleType,
	from, to ModelStop,
) float64 {
	return t.expression.Value(vehicleType, from, to)
}

func (t *timeExpressionImpl) Time(
	vehicleType ModelVehicleType,
	from, to ModelStop,
) time.Time {
	value := t.expression.Value(vehicleType, from, to)
	return t.epoch.Add(
		time.Duration(value) * vehicleType.Model().DurationUnit(),
	)
}

type stopTimeExpressionImpl struct {
	defaultTime  time.Time
	values       []float64
	hasValue     []bool
	name         string
	index        int
	defaultValue float64
}

func (s *stopTimeExpressionImpl) HasNegativeValues() bool {
	return false
}

func (s *stopTimeExpressionImpl) HasPositiveValues() bool {
	return true
}

func (s *stopTimeExpressionImpl) Index() int {
	return s.index
}

func (s *stopTimeExpressionImpl) Name() string {
	return s.name
}

func (s *stopTimeExpressionImpl) SetName(n string) {
	s.name = n
}

func (s *stopTimeExpressionImpl) Value(
	_ ModelVehicleType,
	_,
	to ModelStop,
) float64 {
	idx := to.Index()
	if idx >= 0 && idx < len(s.hasValue) && s.hasValue[idx] {
		return s.values[idx]
	}
	return s.defaultTimeValue(to.Model())
}

func (s *stopTimeExpressionImpl) defaultTimeValue(model Model) float64 {
	if s.defaultValue < 0 {
		if s.defaultTime.Before(model.Epoch()) {
			panic(
				fmt.Sprintf(
					"Default time %v for expression %s is before model epoch %v",
					s.defaultTime,
					s.Name(),
					model.Epoch(),
				),
			)
		}
		s.defaultValue = s.defaultTime.Sub(model.Epoch()).Seconds()
	}
	return s.defaultValue
}

func (s *stopTimeExpressionImpl) Time(stop ModelStop) time.Time {
	idx := stop.Index()
	if idx >= 0 && idx < len(s.hasValue) && s.hasValue[idx] {
		value := s.values[idx]
		return stop.Model().Epoch().Add(time.Duration(value) * stop.Model().DurationUnit())
	}
	return s.defaultTime
}

func (s *stopTimeExpressionImpl) SetTime(stop ModelStop, t time.Time) {
	if stop.Model().IsLocked() {
		panic(
			fmt.Sprintf(
				"Cannot set time for %v in expression %s in locked model",
				stop,
				s.Name(),
			),
		)
	}
	if t.Before(stop.Model().Epoch()) {
		panic(
			fmt.Sprintf(
				"Time %v before epoch %v, setting time for %v in expression %s",
				t,
				stop.Model().Epoch(),
				stop,
				s.Name(),
			),
		)
	}

	idx := stop.Index()
	if idx >= len(s.values) {
		// we have to grow the slice
		// if the slice is empty we grow it by the number of stops
		// if it's not empty we grow it to 1 + idx
		newLen := idx + 1
		if len(s.values) == 0 {
			newLen = max(stop.Model().NumberOfStops(), newLen)
		}
		newValues := make([]float64, newLen)
		newHasValue := make([]bool, newLen)
		copy(newValues, s.values)
		copy(newHasValue, s.hasValue)
		s.values = newValues
		s.hasValue = newHasValue
	}
	s.values[idx] = t.Sub(stop.Model().Epoch()).Seconds()
	s.hasValue[idx] = true
}
