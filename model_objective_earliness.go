// © 2019-present nextmv.io inc

package nextroute

import (
	"math"
)

// TemporalReference is a representation of OnArrival, OnEnd or OnStart as an
// enum.
type TemporalReference int

const (
	// OnStart refers to the Start at a stop.
	OnStart TemporalReference = iota
	// OnEnd refers to the End at a stop.
	OnEnd
	// OnArrival refers to the Arrival at a stop.
	OnArrival = 2
)

// EarlinessObjective is a construct that can be added to the model as an
// objective. It uses to the difference of Arrival, Start or End to the target
// time to penalize.
type EarlinessObjective interface {
	ModelObjective

	// TargetTime returns the target time expression which defines target time
	// that is compared to either arrival, start or end at the stop - depending
	// on the given TemporalReference.
	TargetTime() StopTimeExpression

	// Earliness returns the earliness of a stop. The earliness is the
	// difference between target time and the actual arrival, start or stop of a
	// stop. Depending on the TemporalReference.
	Earliness(stop SolutionStop) float64

	// TemporalReference represents the arrival, start or stop.
	TemporalReference() TemporalReference
}

// NewEarlinessObjective returns a new EarliestObjective construct.
func NewEarlinessObjective(
	targetTime StopTimeExpression,
	earlinessFactor StopExpression,
	temporalReference TemporalReference,
) (EarlinessObjective, error) {
	return &earlinessObjectiveImpl{
			index:             NewModelExpressionIndex(),
			targetTime:        targetTime,
			earlinessFactor:   earlinessFactor,
			temporalReference: temporalReference,
		},
		nil
}

type earlinessObjectiveImpl struct {
	targetTime        StopTimeExpression
	earlinessFactor   StopExpression
	index             int
	temporalReference TemporalReference
}

func (l *earlinessObjectiveImpl) TemporalReference() TemporalReference {
	return l.temporalReference
}

func (l *earlinessObjectiveImpl) ModelExpressions() ModelExpressions {
	return ModelExpressions{}
}

func (l *earlinessObjectiveImpl) Index() int {
	return l.index
}

func (l *earlinessObjectiveImpl) TargetTime() StopTimeExpression {
	return l.targetTime
}

func (l *earlinessObjectiveImpl) Earliness(stop SolutionStop) float64 {
	return l.earliness(stop)
}

func (l *earlinessObjectiveImpl) earliness(stop SolutionStop) float64 {
	targetTime := l.targetTime.Value(nil, nil, stop.modelStop())
	compare := 0.
	switch l.temporalReference {
	case OnStart:
		compare = stop.StartValue()
	case OnEnd:
		compare = stop.EndValue()
	case OnArrival:
		compare = stop.ArrivalValue()
	}

	return math.Max(0, targetTime-compare)
}

func (l *earlinessObjectiveImpl) Value(solution Solution) float64 {
	value := 0.0
	for _, vehicle := range solution.Vehicles() {
		for s := vehicle.First().Next(); !s.IsLast(); s = s.Next() {
			earlinessFactor := l.earlinessFactor.Value(
				nil,
				nil,
				s.ModelStop(),
			)
			value += l.Earliness(s) * earlinessFactor
		}
	}

	return value
}

func (l *earlinessObjectiveImpl) EstimateDeltaValue(
	move SolutionMoveStops,
) float64 {
	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()
	deltaScore := 0.0

	// Init data
	first := true
	arrival, start, end := 0.0, 0.0, 0.0
	previousStop := vehicle.First()

	// Get sequence starting with the first stop prior to the first stop to be
	// inserted.
	generator := newSolutionStopGenerator(*moveImpl, false, true)
	defer generator.release()

	for solutionStop, ok := generator.next(); ok; solutionStop, ok = generator.next() {
		if first {
			previousStop = solutionStop
			end = solutionStop.EndValue()
			first = false
			continue
		}

		// Get arrival, start and end values for current stop when starting at
		// previous stop's end.
		_, arrival, start, end = vehicleType.TemporalValues(
			end,
			previousStop.ModelStop(),
			solutionStop.ModelStop(),
		)

		// depending on the case we calculate the earliness regarding arrival,
		// start or end time of a stop.
		actualTime := 0.0
		currentReference := 0.0
		switch l.temporalReference {
		case OnArrival:
			currentReference = solutionStop.ArrivalValue()
			actualTime = arrival
		case OnStart:
			currentReference = solutionStop.StartValue()
			actualTime = start
		case OnEnd:
			currentReference = solutionStop.EndValue()
			actualTime = end
		}

		// This is a performance tweak. We can stop calculating the if no
		// further stops will be inserted after the last of the move's stops
		// _and_ this stop has still the same end time as it had without
		// inserting any stops.
		// Given the sequence:
		// 1 -> 2 -> 3 -> A -> 4 -> 5 -> 6
		// We know that if stop 4 has the same end time as without adding A to
		// the sequence, then there won't be any different delta then there was
		// before. So we stop there.
		if solutionStop.IsPlanned() {
			next, _ := moveImpl.next()
			if solutionStop.Position() >= next.Position() &&
				solutionStop.EndValue() == end {
				break
			}
		}

		previousStop = solutionStop
		targetTime := l.targetTime.Value(nil, nil, solutionStop.modelStop())

		earlinessFactor := l.earlinessFactor.Value(
			nil,
			nil,
			solutionStop.ModelStop(),
		)

		// Calculate the cost for adding this stop here in the sequence.
		violation := (targetTime - actualTime) * earlinessFactor
		deltaScore += violation

		// If the stop is new in the sequence (by the move) then we do not need
		// to correct the delta. Otherwise we need to correct it.
		if !solutionStop.IsPlanned() {
			continue
		}

		// Correct the delta by removing the difference from it's currently
		// planned reference time to the target time.
		currentScore := 0.0
		if currentReference < targetTime {
			currentScore = (targetTime - currentReference) * earlinessFactor
		}
		deltaScore -= currentScore
	}

	return deltaScore
}

func (l *earlinessObjectiveImpl) String() string {
	switch l.temporalReference {
	case OnStart:
		return "early_start_penalty"
	case OnEnd:
		return "early_end_penalty"
	case OnArrival:
		return "early_arrival_penalty"
	}
	return "early_undefined_reference"
}
