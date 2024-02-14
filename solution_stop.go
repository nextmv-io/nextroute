package nextroute

import (
	"fmt"
	"strings"
	"time"
)

// A SolutionStop is a stop that is planned to be visited by a vehicle. It is
// part of a SolutionPlanUnit and is based on a ModelStop.
type SolutionStop interface {
	// Arrival returns the arrival time of the stop. If the stop is unplanned,
	// the arrival time has no semantic meaning.
	Arrival() time.Time
	// ArrivalValue returns the arrival time of the stop as a float64. If the
	// stop is unplanned, the arrival time has no semantic meaning.
	ArrivalValue() float64

	// ConstraintData returns the value of the constraint for the stop. The
	// constraint value of a stop is set by the ConstraintStopDataUpdater.
	// UpdateConstrainStopData method of the constraint. If the constraint is
	// not set on the stop, nil is returned. If the stop is unplanned, the
	// constraint value has no semantic meaning.
	ConstraintData(constraint ModelConstraint) any
	// CumulativeTravelDurationValue returns the cumulative travel duration of
	// the stop as a float64. The cumulative travel duration is the sum of the
	// travel durations of all stops that are visited before the stop. If the
	// stop is unplanned, the cumulative travel duration has no semantic
	// meaning. The returned value is the number of Model.DurationUnit units.
	CumulativeTravelDurationValue() float64
	// CumulativeTravelDuration returns the cumulative value of the expression
	// for the stop as a time.Duration. The cumulative travel duration is the
	// sum of the travel durations of all stops that are visited before the
	// stop and the stop itself. If the stop is unplanned, the cumulative
	// travel duration has no semantic meaning.
	CumulativeTravelDuration() time.Duration
	// CumulativeValue returns the cumulative value of the expression for the
	// stop as a float64. The cumulative value is the sum of the values of the
	// expression for all stops that are visited before the stop and the stop
	// itself. If the stop is unplanned, the cumulative value has no semantic
	// meaning.
	CumulativeValue(expression ModelExpression) float64

	// End returns the end time of the stop. If the stop is unplanned, the end
	// time has no semantic meaning.
	End() time.Time
	// EndValue returns the end time of the stop as a float64. If the stop is
	// unplanned, the end time has no semantic meaning. The returned value is
	// the number of Model.DurationUnit units since Model.Epoch.
	EndValue() float64

	// Index returns the index of the stop in the Solution.
	Index() int
	// IsFixed returns true if the stop is fixed. A fixed stop is a stop that
	// that can not transition form being planned to unplanned or vice versa.
	IsFixed() bool
	// IsFirst returns true if the stop is the first stop of a vehicle.
	IsFirst() bool
	// IsLast returns true if the stop is the last stop of a vehicle.
	IsLast() bool
	// IsPlanned returns true if the stop is planned. A planned stop is a stop
	// that is visited by a vehicle. An unplanned stop is a stop that is not
	// visited by a vehicle.
	IsPlanned() bool

	// ModelStop returns the ModelStop that is the basis of the SolutionStop.
	ModelStop() ModelStop
	// ModelStopIndex is the index of the ModelStop in the Model.
	ModelStopIndex() int

	// Next returns the next stop the vehicle will visit after the stop. If
	// the stop is the last stop of a vehicle, the solution stop itself is
	// returned. If the stop is unplanned, the next stop has no semantic
	// meaning and the stop itself is returned.
	Next() SolutionStop
	// NextIndex returns the index of the next solution stop the vehicle will
	// visit after the stop. If the stop is the last stop of a vehicle,
	// the index of the stop itself is returned. If the stop is unplanned,
	// the next stop has no semantic meaning and the index of the stop itself
	// is returned.
	NextIndex() int

	// ObjectiveData returns the value of the objective for the stop. The
	// objective value of a stop is set by the
	// ObjectiveStopDataUpdater.UpdateObjectiveStopData method of the objective.
	// If the objective is not set on the stop, nil is returned. If the stop is
	// unplanned, the objective value has no semantic meaning.
	ObjectiveData(objective ModelObjective) any

	// PlanStopsUnit returns the [SolutionPlanStopsUnit] that the stop is
	// associated with.
	PlanStopsUnit() SolutionPlanStopsUnit
	// Previous returns the previous stop the vehicle visited before the stop.
	// If the stop is the first stop of a vehicle, the solution stop itself is
	// returned. If the stop is unplanned, the previous stop has no semantic
	// meaning and the stop itself is returned.
	Previous() SolutionStop
	// PreviousIndex returns the index of the previous solution stop the
	// vehicle visited before the stop. If the stop is the first stop of a
	// vehicle, the index of the stop itself is returned. If the stop is
	// unplanned, the previous stop has no semantic meaning and the index of
	// the stop itself is returned.
	PreviousIndex() int

	// Slack returns the slack of the stop as a time.Duration. Slack is defined
	// as the duration you can start the invoking stop later without
	// postponing the last stop of the vehicle. If the stop is unplanned,
	// the slack has no semantic meaning. Slack is a consequence of the
	// earliest start of stops, if no earliest start is set, the slack is
	// always zero.
	Slack() time.Duration
	// SlackValue returns the slack of the stop as a float64.
	SlackValue() float64

	// Vehicle returns the SolutionVehicle that visits the stop. If the stop
	// is unplanned, the vehicle has no semantic meaning and a panic will be
	// raised.
	Vehicle() SolutionVehicle
	// VehicleIndex returns the index of the SolutionVehicle that visits the
	// stop. If the stop is unplanned, a panic will be raised.
	VehicleIndex() int

	// Solution returns the Solution that the stop is part of.
	Solution() Solution
	// Start returns the start time of the stop. If the stop is unplanned, the
	// start time has no semantic meaning.
	Start() time.Time
	// StartValue returns the start time of the stop as a float64. If the stop
	// is unplanned, the start time has no semantic meaning. The returned
	// value is the number of Model.DurationUnit units since Model.Epoch.
	StartValue() float64
	// Position returns the position of the stop in the vehicle starting with
	// 0 for the first stop. If the stop is unplanned, a panic will be raised.
	Position() int

	// TravelDuration returns the travel duration of the stop as a
	// time.Duration. If the stop is unplanned, the travel duration has no
	// semantic meaning. The travel duration is the time it takes to get to
	// the invoking stop.
	TravelDuration() time.Duration
	// TravelDurationValue returns the travel duration of the stop as a
	// float64. If the stop is unplanned, the travel duration has no semantic
	// meaning. The travel duration is the time it takes to get to the
	// invoking stop. The returned value is the number of
	// Model.DurationUnit units.
	TravelDurationValue() float64

	// Value returns the value of the expression for the stop as a float64.
	// If the stop is unplanned, the value has no semantic meaning.
	Value(expression ModelExpression) float64
}

// SolutionStops is a slice of SolutionStop.
type SolutionStops []SolutionStop

type solutionStopImpl struct {
	solution *solutionImpl
	index    int
}

func toSolutionStop(solution Solution, index int) solutionStopImpl {
	return solutionStopImpl{
		index:    index,
		solution: solution.(*solutionImpl),
	}
}

func (v solutionStopImpl) String() string {
	var sb strings.Builder
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		fmt.Fprintf(&sb, "%v;%v;%v;%v;%v;%v;%v",
			v.Position(),
			v.ModelStop().Index(),
			v.TravelDuration(),
			v.Arrival().Format(v.solution.model.TimeFormat()),
			v.Start().Format(v.solution.model.TimeFormat()),
			v.End().Sub(v.Start()),
			v.End().Format(v.solution.model.TimeFormat()),
		)
	} else {
		fmt.Fprintf(&sb, "-;%v;0;-;-;-;-",
			v.ModelStop().Index(),
		)
	}
	return sb.String()
}

func (v solutionStopImpl) ConstraintData(
	constraint ModelConstraint,
) any {
	return v.solution.constraintValue(constraint, v.index)
}

func (v solutionStopImpl) ObjectiveData(
	objective ModelObjective,
) any {
	return v.solution.objectiveValue(objective, v.index)
}

func (v solutionStopImpl) Value(
	expression ModelExpression,
) float64 {
	return v.solution.value(expression, v.index)
}

func (v solutionStopImpl) CumulativeValue(
	expression ModelExpression,
) float64 {
	return v.solution.cumulativeValue(expression, v.index)
}

func (v solutionStopImpl) Solution() Solution {
	return v.solution
}

func (v solutionStopImpl) PlanStopsUnit() SolutionPlanStopsUnit {
	return v.planStopsUnit()
}

func (v solutionStopImpl) planStopsUnit() *solutionPlanStopsUnitImpl {
	return v.solution.stopToPlanUnit[v.index]
}

func (v solutionStopImpl) Index() int {
	return v.index
}

func (v solutionStopImpl) Next() SolutionStop {
	return v.solution.stopByIndexCache[v.solution.next[v.index]]
}

func (v solutionStopImpl) next() solutionStopImpl {
	return solutionStopImpl{
		index:    v.solution.next[v.index],
		solution: v.solution,
	}
}

func (v solutionStopImpl) NextIndex() int {
	return v.solution.next[v.index]
}

func (v solutionStopImpl) IsPlanned() bool {
	return v.solution.next[v.index] != v.solution.previous[v.index]
}

func (v solutionStopImpl) Previous() SolutionStop {
	return v.solution.stopByIndexCache[v.solution.previous[v.index]]
}

func (v solutionStopImpl) previous() solutionStopImpl {
	return solutionStopImpl{
		index:    v.solution.previous[v.index],
		solution: v.solution,
	}
}

func (v solutionStopImpl) PreviousIndex() int {
	return v.solution.previous[v.index]
}

func (v solutionStopImpl) ArrivalValue() float64 {
	return v.solution.arrival[v.index]
}

func (v solutionStopImpl) Arrival() time.Time {
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		return v.solution.model.Epoch().
			Add(
				time.Duration(v.ArrivalValue()) *
					v.solution.model.DurationUnit())
	}
	return time.Time{}
}

func (v solutionStopImpl) Slack() time.Duration {
	return time.Duration(v.SlackValue()) *
		v.solution.model.DurationUnit()
}

func (v solutionStopImpl) SlackValue() float64 {
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		return v.solution.slack[v.index]
	}
	return 0.0
}

func (v solutionStopImpl) StartValue() float64 {
	return v.solution.start[v.index]
}

func (v solutionStopImpl) Start() time.Time {
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		return v.solution.model.Epoch().
			Add(time.Duration(v.StartValue()) *
				v.solution.model.DurationUnit())
	}
	return time.Time{}
}

func (v solutionStopImpl) EndValue() float64 {
	return v.solution.end[v.index]
}

func (v solutionStopImpl) End() time.Time {
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		return v.solution.model.Epoch().
			Add(time.Duration(v.EndValue()) *
				v.solution.model.DurationUnit())
	}
	return time.Time{}
}

func (v solutionStopImpl) DurationValue() float64 {
	return v.EndValue() - v.StartValue()
}

func (v solutionStopImpl) TravelDurationValue() float64 {
	return v.CumulativeTravelDurationValue() - v.previous().CumulativeTravelDurationValue()
}

func (v solutionStopImpl) TravelDuration() time.Duration {
	return time.Duration(v.TravelDurationValue()) *
		v.solution.model.DurationUnit()
}

func (v solutionStopImpl) CumulativeTravelDurationValue() float64 {
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		return v.solution.cumulativeTravelDuration[v.index]
	}
	return 0
}

func (v solutionStopImpl) CumulativeTravelDuration() time.Duration {
	return time.Duration(v.CumulativeTravelDurationValue()) *
		v.solution.model.DurationUnit()
}

func (v solutionStopImpl) Vehicle() SolutionVehicle {
	if v.solution.next[v.index] == v.solution.previous[v.index] {
		panic("cannot get route of unplanned visit")
	}
	return v.solution.solutionVehicles[v.solution.inVehicle[v.index]]
}

func (v solutionStopImpl) vehicle() solutionVehicleImpl {
	return solutionVehicleImpl{
		index:    v.solution.inVehicle[v.index],
		solution: v.solution,
	}
}

func (v solutionStopImpl) VehicleIndex() int {
	if v.solution.next[v.index] == v.solution.previous[v.index] {
		panic("cannot get route index of unplanned visit")
	}
	return v.solution.inVehicle[v.index]
}

func (v solutionStopImpl) Position() int {
	if v.solution.next[v.index] == v.solution.previous[v.index] {
		panic("cannot get stop position of unplanned stop")
	}
	return v.solution.stopPosition[v.index]
}

func (v solutionStopImpl) IsFixed() bool {
	return v.ModelStop().IsFixed()
}

func (v solutionStopImpl) IsLast() bool {
	return v.solution.next[v.index] == v.index &&
		v.solution.previous[v.index] != v.index
}

func (v solutionStopImpl) IsFirst() bool {
	return v.solution.previous[v.index] == v.index &&
		v.solution.next[v.index] != v.index
}

func (v solutionStopImpl) ModelStop() ModelStop {
	return v.solution.model.(*modelImpl).stops[v.solution.stop[v.index]]
}

func (v solutionStopImpl) modelStop() *stopImpl {
	return v.solution.model.(*modelImpl).stops[v.solution.stop[v.index]].(*stopImpl)
}

func (v solutionStopImpl) ModelStopIndex() int {
	return v.solution.stop[v.index]
}

func (v solutionStopImpl) detach() {
	previousIndex := v.solution.previous[v.index]
	nextIndex := v.solution.next[v.index]

	v.solution.next[previousIndex] = nextIndex
	v.solution.previous[nextIndex] = previousIndex
	v.solution.next[v.index] = v.index
	v.solution.previous[v.index] = v.index
	v.solution.inVehicle[v.index] = -1
}

func (v solutionStopImpl) attach(after int) int {
	v.solution.previous[v.index] = after
	v.solution.next[v.index] = v.solution.next[after]

	v.solution.previous[v.solution.next[after]] = v.index
	v.solution.next[after] = v.index

	v.solution.inVehicle[v.index] = v.solution.inVehicle[after]

	return after
}
