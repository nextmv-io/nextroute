package nextroute

import (
	"math"

	sdkNextRoute "github.com/nextmv-io/sdk/nextroute"
)

// NewEarlinessObjective returns a new EarliestObjective construct.
func NewEarlinessObjective(
	targetTime sdkNextRoute.StopTimeExpression,
	earlinessFactor sdkNextRoute.StopExpression,
	temporalReference sdkNextRoute.TemporalReference,
) (sdkNextRoute.EarlinessObjective, error) {
	return &earlinessObjectiveImpl{
			index:             NewModelExpressionIndex(),
			targetTime:        targetTime,
			earlinessFactor:   earlinessFactor,
			temporalReference: temporalReference,
		},
		nil
}

type earlinessObjectiveImpl struct {
	targetTime        sdkNextRoute.StopTimeExpression
	earlinessFactor   sdkNextRoute.StopExpression
	index             int
	temporalReference sdkNextRoute.TemporalReference
}

func (l *earlinessObjectiveImpl) TemporalReference() sdkNextRoute.TemporalReference {
	return l.temporalReference
}

func (l *earlinessObjectiveImpl) ModelExpressions() sdkNextRoute.ModelExpressions {
	return sdkNextRoute.ModelExpressions{}
}

func (l *earlinessObjectiveImpl) Index() int {
	return l.index
}

func (l *earlinessObjectiveImpl) TargetTime() sdkNextRoute.StopTimeExpression {
	return l.targetTime
}

func (l *earlinessObjectiveImpl) Earliness(stop sdkNextRoute.SolutionStop) float64 {
	return l.earliness(stop.(solutionStopImpl))
}

func (l *earlinessObjectiveImpl) earliness(stop solutionStopImpl) float64 {
	targetTime := l.targetTime.Value(nil, nil, stop.modelStop())
	compare := 0.
	switch l.temporalReference {
	case sdkNextRoute.OnStart:
		compare = stop.StartValue()
	case sdkNextRoute.OnEnd:
		compare = stop.EndValue()
	case sdkNextRoute.OnArrival:
		compare = stop.ArrivalValue()
	}

	return math.Max(0, targetTime-compare)
}

func (l *earlinessObjectiveImpl) InternalValue(solution *solutionImpl) float64 {
	value := 0.0
	for _, vehicle := range solution.vehicles {
		for s := vehicle.first().next(); !s.IsLast(); s = s.next() {
			earlinessFactor := l.earlinessFactor.Value(
				nil,
				nil,
				s.ModelStop(),
			)
			value += l.earliness(s) * earlinessFactor
		}
	}

	return value
}

func (l *earlinessObjectiveImpl) Value(solution sdkNextRoute.Solution) float64 {
	return l.InternalValue(solution.(*solutionImpl))
}

func (l *earlinessObjectiveImpl) EstimateDeltaValue(
	move sdkNextRoute.SolutionMoveStops,
) float64 {
	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()
	deltaScore := 0.0

	// Init data
	first := true
	arrival, start, end := 0.0, 0.0, 0.0
	previousStop := vehicle.first()

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
		case sdkNextRoute.OnArrival:
			currentReference = solutionStop.ArrivalValue()
			actualTime = arrival
		case sdkNextRoute.OnStart:
			currentReference = solutionStop.StartValue()
			actualTime = start
		case sdkNextRoute.OnEnd:
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
	case sdkNextRoute.OnStart:
		return "early_start_penalty"
	case sdkNextRoute.OnEnd:
		return "early_end_penalty"
	case sdkNextRoute.OnArrival:
		return "early_arrival_penalty"
	}
	return "early_undefined_reference"
}
