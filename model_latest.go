package nextroute

import (
	"math"
	"time"

	"github.com/nextmv-io/sdk/nextroute"
)

// NewLatestEnd returns a new LatestEnd construct.
func NewLatestEnd(
	latestEnd nextroute.StopTimeExpression,
) (nextroute.LatestEnd, error) {
	return &latestImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"late_end_penalty",
			nextroute.ModelExpressions{},
		),
		latest:            latestEnd,
		latenessFactor:    NewStopExpression("lateness_penalty_factor", 1.0),
		temporalReference: nextroute.OnEnd,
	}, nil
}

// NewLatestStart returns a new LatestStart construct.
func NewLatestStart(
	latestStart nextroute.StopTimeExpression,
) (nextroute.LatestStart, error) {
	return &latestImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"late_start_penalty",
			nextroute.ModelExpressions{},
		),
		latest:            latestStart,
		latenessFactor:    NewStopExpression("lateness_penalty_factor", 1.0),
		temporalReference: nextroute.OnStart,
	}, nil
}

// NewLatestArrival returns a new LatestArrival construct.
func NewLatestArrival(
	latest nextroute.StopTimeExpression,
) (nextroute.LatestArrival, error) {
	return &latestImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"late_arrival_penalty",
			nextroute.ModelExpressions{},
		),
		latest:            latest,
		latenessFactor:    NewStopExpression("lateness_penalty_factor", 1.0),
		temporalReference: nextroute.OnArrival,
	}, nil
}

type latestImpl struct {
	latest         nextroute.StopTimeExpression
	latenessFactor nextroute.StopExpression
	modelConstraintImpl
	temporalReference nextroute.TemporalReference
}

func (l *latestImpl) SetFactor(factor float64, stop nextroute.ModelStop) {
	if factor >= 0 {
		l.latenessFactor.SetValue(stop, factor)
	}
}

func (l *latestImpl) Factor(stop nextroute.ModelStop) float64 {
	return l.latenessFactor.Value(nil, nil, stop)
}

func (l *latestImpl) ReportConstraint(stop nextroute.SolutionStop) map[string]any {
	var t time.Time
	switch l.temporalReference {
	case nextroute.OnArrival:
		t = stop.Arrival()
	case nextroute.OnStart:
		t = stop.Start()
	case nextroute.OnEnd:
		t = stop.End()
	}

	return map[string]any{
		"latest": l.latest.Value(nil, nil, stop.ModelStop()),
		"start":  t,
	}
}

func (l *latestImpl) String() string {
	return l.name
}

func (l *latestImpl) Latest() nextroute.StopTimeExpression {
	return l.latest
}

func (l *latestImpl) EstimationCost() nextroute.Cost {
	return nextroute.LinearStop
}

func (l *latestImpl) Lateness(stop nextroute.SolutionStop) float64 {
	latest := l.latest.Value(nil, nil, stop.ModelStop())
	reference := 0.
	switch l.temporalReference {
	case nextroute.OnArrival:
		reference = stop.ArrivalValue()
	case nextroute.OnStart:
		reference = stop.StartValue()
	case nextroute.OnEnd:
		reference = stop.EndValue()
	}

	return math.Max(0, reference-latest)
}

func (l *latestImpl) Value(s nextroute.Solution) float64 {
	solution := s.(*solutionImpl)
	value := 0.0
	for _, vehicle := range solution.vehicles {
		solutionStop := vehicle.first().next()
		lastSolutionStop := vehicle.last()
		for {
			latenessFactor := l.latenessFactor.Value(
				nil,
				nil,
				solutionStop.ModelStop(),
			)
			value += l.Lateness(solutionStop) * latenessFactor

			if solutionStop == lastSolutionStop {
				break
			}

			solutionStop = solutionStop.next()
		}
	}

	return value
}

func (l *latestImpl) EstimateIsViolated(
	move nextroute.SolutionMoveStops,
) (isViolated bool, stopPositionsHint nextroute.StopPositionsHint) {
	score, hint := l.estimateDeltaScore(
		move.(*solutionMoveStopsImpl),
		true,
	)
	return score != 0.0, hint.(*stopPositionHintImpl)
}

func (l *latestImpl) EstimateDeltaValue(
	move nextroute.SolutionMoveStops,
) float64 {
	score, _ := l.estimateDeltaScore(
		move.(*solutionMoveStopsImpl),
		false,
	)
	return score
}

func (l *latestImpl) estimateDeltaScore(
	move *solutionMoveStopsImpl,
	asConstraint bool,
) (deltaScore float64, stopPositionsHint nextroute.StopPositionsHint) {
	vehicle := move.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()
	deltaScore = 0.0
	first := true

	arrival, start, end := 0.0, 0.0, 0.0
	previousStop := vehicle.first().ModelStop()
	generator := newSolutionStopGenerator(*move, false, true)
	defer generator.release()

	for solutionStop, ok := generator.next(); ok; solutionStop, ok = generator.next() {
		if first {
			previousStop = solutionStop.ModelStop()
			end = solutionStop.EndValue()
			first = false
			continue
		}

		modelStop := solutionStop.ModelStop()
		_, arrival, start, end = vehicleType.TemporalValues(
			end,
			previousStop,
			modelStop,
		)

		previousStop = modelStop
		reference, currentReference := 0.0, 0.0

		switch l.temporalReference {
		case nextroute.OnArrival:
			reference = arrival
			currentReference = solutionStop.ArrivalValue()
		case nextroute.OnStart:
			reference = start
			currentReference = solutionStop.StartValue()
		case nextroute.OnEnd:
			reference = end
			currentReference = solutionStop.EndValue()
		}

		latest := l.latest.Value(nil, nil, modelStop)

		if reference <= latest {
			continue
		}

		if asConstraint {
			return 1.0, constNoPositionsHint
		}

		factor := l.latenessFactor.Value(nil, nil, modelStop)
		violation := (reference - latest) * factor
		deltaScore += violation

		if !solutionStop.IsPlanned() {
			continue
		}

		currentScore := 0.0

		if currentReference > latest {
			currentScore = (currentReference - latest) * factor
		}

		deltaScore -= currentScore
	}

	return deltaScore, constNoPositionsHint
}

func (l *latestImpl) DoesStopHaveViolations(s nextroute.SolutionStop) bool {
	stop := s.(solutionStopImpl)
	if !stop.
		vehicle().
		ModelVehicle().
		VehicleType().
		TravelDurationExpression().
		SatisfiesTriangleInequality() {
		latest := l.latest.Value(nil, nil, stop.modelStop())
		switch l.temporalReference {
		case nextroute.OnArrival:
			return stop.ArrivalValue() > latest
		case nextroute.OnStart:
			return stop.StartValue() > latest
		case nextroute.OnEnd:
			return stop.EndValue() > latest
		}
	}

	return false
}

func (l *latestImpl) IsTemporal() bool {
	return true
}
