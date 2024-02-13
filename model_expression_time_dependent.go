package nextroute

import (
	"fmt"
	"math"
	"strings"
	"time"

	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

const maxTimeDependentExpressionInterval = 24 * 7 * time.Hour

// NewTimeDependentDurationExpression returns a new
// TimeDependentDurationExpression.
func NewTimeDependentDurationExpression(
	model nextroute.Model,
	expression nextroute.DurationExpression,
) (nextroute.TimeDependentDurationExpression, error) {
	if model.Epoch().Second() != 0 || model.Epoch().Nanosecond() != 0 {
		return nil,
			nmerror.NewArgumentMismatchError(fmt.Errorf(
				"model epoch %v is not on a minute boundary",
				model.Epoch(),
			))
	}

	if expression.HasNegativeValues() {
		return nil,
			nmerror.NewArgumentMismatchError(fmt.Errorf(
				"expression %v has negative values, time travel is not supported",
				expression.Name(),
			))
	}

	return &timeDependentDurationExpressionImpl{
		model:                       model,
		index:                       NewModelExpressionIndex(),
		defaultExpression:           expression,
		expressions:                 []nextroute.DurationExpression{},
		elements:                    make(map[int64]*expressionElement),
		name:                        "time_dependent_expression",
		satisfiesTriangleInequality: false,
	}, nil
}

// NewTimeIndependentDurationExpression returns a new
// TimeInDependentDurationExpression.
func NewTimeIndependentDurationExpression(
	expression nextroute.DurationExpression,
) nextroute.TimeDependentDurationExpression {
	return &timeIndependentDurationExpressionImpl{
		expression:                  expression,
		name:                        "time_independent_expression",
		satisfiesTriangleInequality: false,
	}
}

type expressionElement struct {
	expression nextroute.DurationExpression
	next       *expressionElement
	previous   *expressionElement
	start      float64
	end        float64
}

type timeDependentDurationExpressionImpl struct {
	model                       nextroute.Model
	defaultExpression           nextroute.DurationExpression
	elements                    map[int64]*expressionElement
	startElement                *expressionElement
	endElement                  *expressionElement
	name                        string
	expressions                 []nextroute.DurationExpression
	earliest                    float64
	latest                      float64
	index                       int
	satisfiesTriangleInequality bool
}

func (t *timeDependentDurationExpressionImpl) SatisfiesTriangleInequality() bool {
	return t.satisfiesTriangleInequality
}

func (t *timeDependentDurationExpressionImpl) SetSatisfiesTriangleInequality(satisfies bool) {
	t.satisfiesTriangleInequality = satisfies
}

func (t *timeDependentDurationExpressionImpl) String() string {
	var sb strings.Builder
	if len(t.elements) == 0 {
		fmt.Fprintf(&sb, "[%v]-[%v] %v\n",
			t.earliest,
			t.latest,
			t.defaultExpression.Name(),
		)
		return sb.String()
	}

	element := t.startElement
	for element.next != nil {
		fmt.Fprintf(&sb, "[%v]-[%v] %v\n",
			t.model.ValueToTime(element.start),
			t.model.ValueToTime(element.end),
			element.expression.Name(),
		)
		element = element.next
	}
	fmt.Fprintf(&sb, "[%v]-[%v] %v\n",
		t.model.ValueToTime(element.start),
		t.model.ValueToTime(element.end),
		element.expression.Name(),
	)
	return sb.String()
}

func (t *timeDependentDurationExpressionImpl) Index() int {
	return t.index
}

func (t *timeDependentDurationExpressionImpl) Name() string {
	return t.name
}

func (t *timeDependentDurationExpressionImpl) Expressions() []nextroute.DurationExpression {
	return common.DefensiveCopy(t.expressions)
}

func (t *timeDependentDurationExpressionImpl) Value(
	vehicleType nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
) float64 {
	if t.IsDependentOnTime() {
		panic("asking for a value on a time dependent expression, require a time to be passed in, use ValueAtTime")
	}
	return t.defaultExpression.Value(vehicleType, from, to)
}

func (t *timeDependentDurationExpressionImpl) Duration(
	vehicleType nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
) time.Duration {
	if t.IsDependentOnTime() {
		panic("asking for a duration on a time dependent expression," +
			" requires a time to be passed in")
	}
	return t.defaultExpression.Duration(vehicleType, from, to)
}

func (t *timeDependentDurationExpressionImpl) HasNegativeValues() bool {
	hasNegativeValues := t.defaultExpression.HasNegativeValues()
	for _, expression := range t.expressions {
		hasNegativeValues = hasNegativeValues || expression.HasNegativeValues()
		if hasNegativeValues {
			break
		}
	}
	return hasNegativeValues
}

func (t *timeDependentDurationExpressionImpl) HasPositiveValues() bool {
	hasPositiveValues := t.defaultExpression.HasPositiveValues()
	for _, expression := range t.expressions {
		hasPositiveValues = hasPositiveValues || expression.HasPositiveValues()
		if hasPositiveValues {
			break
		}
	}
	return hasPositiveValues
}

func (t *timeDependentDurationExpressionImpl) DefaultExpression() nextroute.DurationExpression {
	return t.defaultExpression
}

func (t *timeDependentDurationExpressionImpl) SetName(name string) {
	t.name = name
}

func (t *timeDependentDurationExpressionImpl) updateMap() {
	t.elements = make(map[int64]*expressionElement)

	element := t.startElement
	t.startElement.end = t.startElement.next.start
	for {
		element = element.next
		if element.next == nil {
			break
		}
		t.endElement.start = element.end
		increment := t.model.DurationToValue(time.Minute)
		for v := element.start; v < element.end; v += increment {
			t.elements[int64(v)] = element
		}
	}
}

func (t *timeDependentDurationExpressionImpl) SetExpression(
	start, end time.Time,
	expression nextroute.DurationExpression,
) error {
	if start.Before(t.model.Epoch()) {
		return nmerror.NewArgumentMismatchError(fmt.Errorf(
			"start time %v is before model epoch %v",
			start,
			t.model.Epoch(),
		))
	}
	if start.After(end) {
		return nmerror.NewArgumentMismatchError(fmt.Errorf("start time %v is after end time %v", start, end))
	}
	if start.Second() != 0 || start.Nanosecond() != 0 {
		return nmerror.NewArgumentMismatchError(fmt.Errorf("start time %v is not on a minute boundary", start))
	}
	if end.Second() != 0 || end.Nanosecond() != 0 {
		return nmerror.NewArgumentMismatchError(fmt.Errorf("end time %v is not on a minute boundary", end))
	}
	if expression.HasNegativeValues() {
		return nmerror.NewArgumentMismatchError(fmt.Errorf(
			"expression %v has negative values,"+
				" time travel is not supported",
			expression.Name(),
		))
	}
	t.expressions = append(t.expressions, expression)
	t.expressions = common.UniqueDefined(
		t.expressions,
		func(durationExpression nextroute.DurationExpression) int {
			return durationExpression.Index()
		},
	)
	startMinutesFromEpoch := t.model.TimeToValue(start)
	endMinutesFromEpoch := t.model.TimeToValue(end)

	earliest := t.earliest
	if startMinutesFromEpoch < t.earliest || t.earliest == 0 {
		earliest = startMinutesFromEpoch
	}
	latest := t.latest
	if endMinutesFromEpoch > t.latest || t.latest == 0 {
		latest = endMinutesFromEpoch
	}
	duration := t.model.DurationUnit() * time.Duration(latest-earliest)

	if duration > maxTimeDependentExpressionInterval {
		return nmerror.NewArgumentMismatchError(fmt.Errorf(
			"time dependent expression is too large,"+
				" expressions are defined from %v till %v,"+
				" limited to interval of size %v",
			earliest,
			latest,
			maxTimeDependentExpressionInterval,
		))
	}
	t.earliest = earliest
	t.latest = latest

	newElement := &expressionElement{
		start:      startMinutesFromEpoch,
		end:        endMinutesFromEpoch,
		expression: expression,
		previous:   nil,
		next:       nil,
	}

	if t.startElement == nil {
		t.startElement = &expressionElement{
			start:      0,
			end:        startMinutesFromEpoch,
			expression: t.defaultExpression,
			previous:   nil,
			next:       newElement,
		}
		newElement.previous = t.startElement
		newElement.next = &expressionElement{
			start:      endMinutesFromEpoch,
			end:        t.model.TimeToValue(t.model.MaxTime()),
			expression: t.defaultExpression,
			previous:   newElement,
			next:       nil,
		}
		t.endElement = newElement.next
	} else {
		element := t.startElement
		for element.next != nil && element.start < startMinutesFromEpoch && element.end <= startMinutesFromEpoch {
			element = element.next
		}

		if element.expression != t.defaultExpression && element.start <
			newElement.end {
			return nmerror.NewArgumentMismatchError(fmt.Errorf(
				"new time dependent expression %s [%v, %v] overlaps with existing"+
					" expression %s [%v, %v]",
				expression.Name(),
				t.model.ValueToTime(newElement.start),
				t.model.ValueToTime(newElement.end),
				element.expression.Name(),
				t.model.ValueToTime(element.start),
				t.model.ValueToTime(element.end),
			))
		}

		splitElement1 := &expressionElement{
			start:      element.start,
			end:        startMinutesFromEpoch,
			expression: element.expression,
			previous:   element.previous,
			next:       newElement,
		}

		splitElement2 := &expressionElement{
			start:      endMinutesFromEpoch,
			end:        element.end,
			expression: element.expression,
			previous:   newElement,
			next:       element.next,
		}

		newElement.previous = splitElement1
		newElement.next = splitElement2
		if element.previous != nil {
			element.previous.next = splitElement1
		} else {
			t.startElement = splitElement1
		}
	}

	t.updateMap()

	return nil
}

func (t *timeDependentDurationExpressionImpl) getElementAtValue(
	value float64,
) *expressionElement {
	if len(t.elements) == 0 {
		return nil
	}
	valuesInAMinute := t.model.DurationToValue(time.Minute)
	minute := math.Floor(value/valuesInAMinute) * valuesInAMinute
	if minute < t.startElement.end {
		return t.startElement
	}
	if minute >= t.endElement.start {
		return t.endElement
	}
	if element, ok := t.elements[int64(minute)]; ok {
		return element
	}
	return nil
}

func (t *timeDependentDurationExpressionImpl) ExpressionAtTime(
	atTime time.Time,
) nextroute.DurationExpression {
	return t.ExpressionAtValue(t.model.TimeToValue(atTime))
}

func (t *timeDependentDurationExpressionImpl) ExpressionAtValue(
	value float64,
) nextroute.DurationExpression {
	if len(t.elements) == 0 {
		return t.defaultExpression
	}
	valuesInAMinute := t.model.DurationToValue(time.Minute)
	minute := math.Floor(value/valuesInAMinute) * valuesInAMinute
	if element, ok := t.elements[int64(minute)]; ok {
		return element.expression
	}
	return t.defaultExpression
}

func (t *timeDependentDurationExpressionImpl) ValueAtTime(
	atTime time.Time,
	vehicleType nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
) float64 {
	return t.ValueAtValue(
		t.model.TimeToValue(atTime),
		vehicleType,
		from,
		to,
	)
}

func (t *timeDependentDurationExpressionImpl) ValueAtValue(
	value float64,
	vehicleType nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
) float64 {
	if len(t.elements) == 0 {
		return t.defaultExpression.Value(vehicleType, from, to)
	}

	element := t.getElementAtValue(value)

	duration := element.expression.Value(vehicleType, from, to)

	if duration == 0 {
		return 0
	}

	if duration < 0 {
		panic("duration is negative, time travel not allowed")
	}

	fractionCovered := (element.end - value) / duration

	if fractionCovered >= 1 {
		return duration
	}

	duration = fractionCovered * duration

	for fractionCovered < 1 {
		element = element.next
		requiredDuration := (1 - fractionCovered) *
			element.expression.Value(vehicleType, from, to)

		if requiredDuration == 0 {
			return duration
		}

		fractionCurrentElementCanProvide := (element.end - element.start) /
			requiredDuration

		if fractionCurrentElementCanProvide >= 1 {
			return duration + requiredDuration
		}

		duration += fractionCurrentElementCanProvide * requiredDuration

		fractionCovered += fractionCurrentElementCanProvide *
			(1 - fractionCovered)
	}
	return duration
}

func (t *timeDependentDurationExpressionImpl) IsDependentOnTime() bool {
	return len(t.expressions) != 0
}

type timeIndependentDurationExpressionImpl struct {
	expression                  nextroute.DurationExpression
	name                        string
	satisfiesTriangleInequality bool
}

func (t *timeIndependentDurationExpressionImpl) SatisfiesTriangleInequality() bool {
	return t.satisfiesTriangleInequality
}

func (t *timeIndependentDurationExpressionImpl) SetSatisfiesTriangleInequality(satisfies bool) {
	t.satisfiesTriangleInequality = satisfies
}

func (t *timeIndependentDurationExpressionImpl) Expressions() []nextroute.DurationExpression {
	return []nextroute.DurationExpression{}
}

func (t *timeIndependentDurationExpressionImpl) Duration(
	vehicleType nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
) time.Duration {
	return t.expression.Duration(vehicleType, from, to)
}

func (t *timeIndependentDurationExpressionImpl) Index() int {
	return t.expression.Index()
}

func (t *timeIndependentDurationExpressionImpl) Name() string {
	return t.name
}

func (t *timeIndependentDurationExpressionImpl) Value(
	vehicleType nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
) float64 {
	return t.expression.Value(vehicleType, from, to)
}

func (t *timeIndependentDurationExpressionImpl) HasNegativeValues() bool {
	return t.expression.HasNegativeValues()
}

func (t *timeIndependentDurationExpressionImpl) HasPositiveValues() bool {
	return t.expression.HasPositiveValues()
}

func (t *timeIndependentDurationExpressionImpl) SetName(s string) {
	t.name = s
}

func (t *timeIndependentDurationExpressionImpl) DefaultExpression() nextroute.DurationExpression {
	return t.expression
}

func (t *timeIndependentDurationExpressionImpl) SetExpression(
	_, _ time.Time,
	_ nextroute.DurationExpression,
) error {
	return nmerror.NewModelCustomizationError(
		fmt.Errorf("should not be invoked on time in-dependent expression"),
	)
}

func (t *timeIndependentDurationExpressionImpl) ExpressionAtTime(
	_ time.Time,
) nextroute.DurationExpression {
	return t.expression
}

func (t *timeIndependentDurationExpressionImpl) ExpressionAtValue(
	_ float64,
) nextroute.DurationExpression {
	return t.expression
}

func (t *timeIndependentDurationExpressionImpl) ValueAtTime(
	_ time.Time,
	vehicleType nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
) float64 {
	return t.expression.Value(vehicleType, from, to)
}

func (t *timeIndependentDurationExpressionImpl) ValueAtValue(
	_ float64,
	vehicleType nextroute.ModelVehicleType,
	from, to nextroute.ModelStop,
) float64 {
	return t.expression.Value(vehicleType, from, to)
}

func (t *timeIndependentDurationExpressionImpl) IsDependentOnTime() bool {
	return false
}
