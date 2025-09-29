// © 2019-present nextmv.io inc

package nextroute

import (
	"math"
	"time"
)

// LatestEnd is a construct that can be added to the model as a constraint or
// as an objective. The latest end of a stop is the latest time a stop can end
// at the location of the stop.
type LatestEnd struct {
	*latest
}

// LatestStart is a construct that can be added to the model as a constraint or
// as an objective. The latest start of a stop is the latest time a stop can
// start at the location of the stop.
type LatestStart struct {
	*latest
}

// LatestArrival is a construct that can be added to the model as a constraint
// or as an objective. The latest arrival of a stop is the latest time a stop
// can arrive at the location of the stop.
type LatestArrival struct {
	*latest
}

type latest struct {
	latest            StopTimeExpression
	latenessFactor    StopExpression
	name              string
	temporalReference TemporalReference
}

// NewLatestEnd returns a new LatestEnd construct.
func NewLatestEnd(
	latestEnd StopTimeExpression,
) (*LatestEnd, error) {
	return &LatestEnd{
		latest: &latest{
			latest:            latestEnd,
			latenessFactor:    NewStopExpression("lateness_penalty_factor", 1.0),
			temporalReference: OnEnd,
			name:              "late_end_penalty",
		},
	}, nil
}

// NewLatestStart returns a new LatestStart construct.
func NewLatestStart(
	latestStart StopTimeExpression,
) (*LatestStart, error) {
	return &LatestStart{
		latest: &latest{
			latest:            latestStart,
			latenessFactor:    NewStopExpression("lateness_penalty_factor", 1.0),
			temporalReference: OnStart,
			name:              "late_start_penalty",
		},
	}, nil
}

// NewLatestArrival returns a new LatestArrival construct.
func NewLatestArrival(
	l StopTimeExpression,
) (*LatestArrival, error) {
	return &LatestArrival{
		latest: &latest{
			latest:            l,
			latenessFactor:    NewStopExpression("lateness_penalty_factor", 1.0),
			temporalReference: OnArrival,
			name:              "late_arrival_penalty",
		},
	}, nil
}

// SetFactor adds a factor with which a deviating stop is multiplied. This
// is only taken into account if the construct is used as an objective.
func (l *latest) SetFactor(factor float64, stop *ModelStop) error {
	if factor >= 0 {
		return l.latenessFactor.SetValue(stop, factor)
	}
	return nil
}

// Factor returns the multiplication factor for the given stop expression.
func (l *latest) Factor(stop *ModelStop) float64 {
	return l.latenessFactor.Value(nil, nil, stop)
}

// ReportConstraint returns a report of the constraint for the given stop.
func (l *latest) ReportConstraint(stop SolutionStop) map[string]any {
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

// String returns the name of the constraint or objective.
func (l *latest) String() string {
	return l.name
}

// Latest returns the latest arrival expression which defines the latest
// arrival of a stop.
func (l *LatestArrival) Latest() StopTimeExpression {
	return l.latest.latest
}

// Lateness returns the lateness of a stop. The lateness is the difference
// between the actual arrival and its target arrival time.
func (l *LatestArrival) Lateness(stop SolutionStop) float64 {
	return l.lateness(stop)
}

// Latest returns the latest start expression which defines the latest
// start of a stop.
func (l *LatestStart) Latest() StopTimeExpression {
	return l.latest.latest
}

// Lateness returns the lateness of a stop. The lateness is the difference
// between the actual start and its target start time.
func (l *LatestStart) Lateness(stop SolutionStop) float64 {
	return l.lateness(stop)
}

// Latest returns the latest end time expression which defines the latest
// end of a stop.
func (l *LatestEnd) Latest() StopTimeExpression {
	return l.latest.latest
}

// Lateness returns the lateness of a stop. The lateness is the difference
// between the actual end and its target end time.
func (l *LatestEnd) Lateness(stop SolutionStop) float64 {
	return l.lateness(stop)
}

// EstimationCost returns the cost of estimating whether a move violates the
// constraint.
func (l *latest) EstimationCost() Cost {
	return LinearStop
}

func (l *latest) lateness(stop SolutionStop) float64 {
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

// Value returns objective value for a given solution.
func (l *latest) Value(s *Solution) float64 {
	value := 0.0
	for _, vehicle := range s.vehicles {
		solutionStop := vehicle.First().Next()
		lastSolutionStop := vehicle.Last()
		for {
			latenessFactor := l.latenessFactor.Value(
				nil,
				nil,
				solutionStop.ModelStop(),
			)
			value += l.lateness(solutionStop) * latenessFactor

			if solutionStop == lastSolutionStop {
				break
			}

			solutionStop = solutionStop.Next()
		}
	}

	return value
}

// EstimateIsViolated estimates whether the given move violates the constraint.
func (l *latest) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	score, hint := l.estimateDeltaScore(
		move.(*solutionMoveStopsImpl),
		true,
	)
	return score != 0.0, hint
}

// EstimateDeltaValue estimates the change in value for the given move.
func (l *latest) EstimateDeltaValue(
	move SolutionMoveStops,
) float64 {
	score, _ := l.estimateDeltaScore(
		move.(*solutionMoveStopsImpl),
		false,
	)
	return score
}

func (l *latest) estimateDeltaScore(
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
			return 1.0, NoPositionsHint()
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

	return deltaScore, NoPositionsHint()
}

func (l *latest) DoesStopHaveViolations(s SolutionStop) bool {
	stop := s
	if !stop.
		vehicle().
		ModelVehicle().
		VehicleType().
		TravelDurationExpression().
		SatisfiesTriangleInequality() {
		latest := l.latest.Value(nil, nil, stop.ModelStop())
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

func (l *latest) IsTemporal() bool {
	return true
}
