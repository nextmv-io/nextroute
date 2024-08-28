// © 2019-present nextmv.io inc

package nextroute

import (
	"math"
	"time"
)

// LatestEnd is a construct that can be added to the model as a constraint or
// as an objective. The latest end of a stop is the latest time a stop can end
// at the location of the stop.
type LatestEnd interface {
	ConstraintReporter
	ModelConstraint
	ModelObjective

	// Latest returns the latest end time expression which defines the latest
	// end of a stop.
	Latest() StopTimeExpression

	// Lateness returns the lateness of a stop. The lateness is the difference
	// between the actual end and its target end time.
	Lateness(stop SolutionStop) float64

	// SetFactor adds a factor with which a deviating stop is multiplied. This
	// is only taken into account if the construct is used as an objective.
	SetFactor(factor float64, stop ModelStop) error

	// Factor returns the multiplication factor for the given stop expression.
	Factor(stop ModelStop) float64
}

// LatestStart is a construct that can be added to the model as a constraint or
// as an objective. The latest start of a stop is the latest time a stop can
// start at the location of the stop.
type LatestStart interface {
	ConstraintReporter
	ModelConstraint
	ModelObjective

	// Latest returns the latest start expression which defines the latest
	// start of a stop.
	Latest() StopTimeExpression

	// Lateness returns the lateness of a stop. The lateness is the difference
	// between the actual start and its target start time.
	Lateness(stop SolutionStop) float64

	// SetFactor adds a factor with which a deviating stop is multiplied. This
	// is only taken into account if the construct is used as an objective.
	SetFactor(factor float64, stop ModelStop) error

	// Factor returns the multiplication factor for the given stop expression.
	Factor(stop ModelStop) float64
}

// LatestArrival is a construct that can be added to the model as a constraint
// or as an objective. The latest arrival of a stop is the latest time a stop
// can arrive at the location of the stop.
type LatestArrival interface {
	ConstraintReporter
	ModelConstraint
	ModelObjective

	// Latest returns the latest arrival expression which defines the latest
	// arrival of a stop.
	Latest() StopTimeExpression

	// Lateness returns the lateness of a stop. The lateness is the difference
	// between the actual arrival and its target arrival time.
	Lateness(stop SolutionStop) float64

	// SetFactor adds a factor with which a deviating stop is multiplied. This
	// is only taken into account if the construct is used as an objective.
	SetFactor(factor float64, stop ModelStop) error

	// Factor returns the multiplication factor for the given stop expression.
	Factor(stop ModelStop) float64
}

// NewLatestEnd returns a new LatestEnd construct.
func NewLatestEnd(
	latestEnd StopTimeExpression,
) (LatestEnd, error) {
	return &latestImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"late_end_penalty",
			ModelExpressions{},
		),
		latest:            latestEnd,
		latenessFactor:    NewStopExpression("lateness_penalty_factor", 1.0),
		temporalReference: OnEnd,
	}, nil
}

// NewLatestStart returns a new LatestStart construct.
func NewLatestStart(
	latestStart StopTimeExpression,
) (LatestStart, error) {
	return &latestImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"late_start_penalty",
			ModelExpressions{},
		),
		latest:            latestStart,
		latenessFactor:    NewStopExpression("lateness_penalty_factor", 1.0),
		temporalReference: OnStart,
	}, nil
}

// NewLatestArrival returns a new LatestArrival construct.
func NewLatestArrival(
	latest StopTimeExpression,
) (LatestArrival, error) {
	return &latestImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"late_arrival_penalty",
			ModelExpressions{},
		),
		latest:            latest,
		latenessFactor:    NewStopExpression("lateness_penalty_factor", 1.0),
		temporalReference: OnArrival,
	}, nil
}

type latestImpl struct {
	latest         StopTimeExpression
	latenessFactor StopExpression
	modelConstraintImpl
	temporalReference TemporalReference
}

func (l *latestImpl) SetFactor(factor float64, stop ModelStop) error {
	if factor >= 0 {
		return l.latenessFactor.SetValue(stop, factor)
	}
	return nil
}

func (l *latestImpl) Factor(stop ModelStop) float64 {
	return l.latenessFactor.Value(nil, nil, stop)
}

func (l *latestImpl) ReportConstraint(stop SolutionStop) map[string]any {
	var t time.Time
	switch l.temporalReference {
	case OnArrival:
		t = stop.Arrival()
	case OnStart:
		t = stop.Start()
	case OnEnd:
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

func (l *latestImpl) Latest() StopTimeExpression {
	return l.latest
}

func (l *latestImpl) EstimationCost() Cost {
	return LinearStop
}

func (l *latestImpl) Lateness(stop SolutionStop) float64 {
	latest := l.latest.Value(nil, nil, stop.ModelStop())
	reference := 0.
	switch l.temporalReference {
	case OnArrival:
		reference = stop.ArrivalValue()
	case OnStart:
		reference = stop.StartValue()
	case OnEnd:
		reference = stop.EndValue()
	}

	return math.Max(0, reference-latest)
}

func (l *latestImpl) Value(s Solution) float64 {
	solution := s.(*solutionImpl)
	value := 0.0
	for _, vehicle := range solution.vehicles {
		solutionStop := vehicle.First().Next()
		lastSolutionStop := vehicle.Last()
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

			solutionStop = solutionStop.Next()
		}
	}

	return value
}

func (l *latestImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	score, hint := l.estimateDeltaScore(
		move.(*solutionMoveStopsImpl),
		true,
	)
	return score != 0.0, hint.(*stopPositionHintImpl)
}

func (l *latestImpl) EstimateDeltaValue(
	move SolutionMoveStops,
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
) (deltaScore float64, stopPositionsHint StopPositionsHint) {
	vehicle := move.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()
	deltaScore = 0.0
	first := true

	arrival, start, end := 0.0, 0.0, 0.0
	previousStop := vehicle.First().ModelStop()
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
		case OnArrival:
			reference = arrival
			currentReference = solutionStop.ArrivalValue()
		case OnStart:
			reference = start
			currentReference = solutionStop.StartValue()
		case OnEnd:
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

func (l *latestImpl) DoesStopHaveViolations(s SolutionStop) bool {
	stop := s
	if !stop.
		vehicle().
		ModelVehicle().
		VehicleType().
		TravelDurationExpression().
		SatisfiesTriangleInequality() {
		latest := l.latest.Value(nil, nil, stop.modelStop())
		switch l.temporalReference {
		case OnArrival:
			return stop.ArrivalValue() > latest
		case OnStart:
			return stop.StartValue() > latest
		case OnEnd:
			return stop.EndValue() > latest
		}
	}

	return false
}

func (l *latestImpl) IsTemporal() bool {
	return true
}
