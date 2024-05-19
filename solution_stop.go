// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"strings"
	"time"
)

// A SolutionStop is a stop that is planned to be visited by a vehicle. It is
// part of a SolutionPlanUnit and is based on a ModelStop.
type SolutionStop struct {
	solution *solutionImpl
	index    int
}

// SolutionStops is a slice of SolutionStop.
type SolutionStops []SolutionStop

func toSolutionStop(solution Solution, index int) SolutionStop {
	return SolutionStop{
		index:    index,
		solution: solution.(*solutionImpl),
	}
}

func (v SolutionStop) String() string {
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

// ConstraintData returns the value of the constraint for the stop. The
// constraint value of a stop is set by the ConstraintStopDataUpdater.
// UpdateConstrainStopData method of the constraint. If the constraint is
// not set on the stop, nil is returned. If the stop is unplanned, the
// constraint value has no semantic meaning.
func (v SolutionStop) ConstraintData(
	constraint ModelConstraint,
) any {
	return v.solution.constraintValue(constraint, v.index)
}

// ObjectiveData returns the value of the objective for the stop. The
// objective value of a stop is set by the
// ObjectiveStopDataUpdater.UpdateObjectiveStopData method of the objective.
// If the objective is not set on the stop, nil is returned. If the stop is
// unplanned, the objective value has no semantic meaning.
func (v SolutionStop) ObjectiveData(
	objective ModelObjective,
) any {
	return v.solution.objectiveValue(objective, v.index)
}

// Value returns the value of the expression for the stop as a float64.
// If the stop is unplanned, the value has no semantic meaning.
func (v SolutionStop) Value(
	expression ModelExpression,
) float64 {
	return v.solution.value(expression, v.index)
}

// CumulativeValue returns the cumulative value of the expression for the
// stop as a float64. The cumulative value is the sum of the values of the
// expression for all stops that are visited before the stop and the stop
// itself. If the stop is unplanned, the cumulative value has no semantic
// meaning.
func (v SolutionStop) CumulativeValue(
	expression ModelExpression,
) float64 {
	return v.solution.cumulativeValue(expression, v.index)
}

// Solution returns the Solution that the stop is part of.
func (v SolutionStop) Solution() Solution {
	return v.solution
}

// PlanStopsUnit returns the [SolutionPlanStopsUnit] that the stop is
// associated with.
func (v SolutionStop) PlanStopsUnit() SolutionPlanStopsUnit {
	return v.planStopsUnit()
}

func (v SolutionStop) planStopsUnit() *solutionPlanStopsUnitImpl {
	return v.solution.stopToPlanUnit[v.index]
}

// Index returns the index of the stop in the Solution.
func (v SolutionStop) Index() int {
	return v.index
}

// Next returns the next stop the vehicle will visit after the stop. If
// the stop is the last stop of a vehicle, the solution stop itself is
// returned. If the stop is unplanned, the next stop has no semantic
// meaning and the stop itself is returned.
func (v SolutionStop) Next() SolutionStop {
	return SolutionStop{
		index:    v.solution.next[v.index],
		solution: v.solution,
	}
}

// NextIndex returns the index of the next solution stop the vehicle will
// visit after the stop. If the stop is the last stop of a vehicle,
// the index of the stop itself is returned. If the stop is unplanned,
// the next stop has no semantic meaning and the index of the stop itself
// is returned.
func (v SolutionStop) NextIndex() int {
	return v.solution.next[v.index]
}

// IsPlanned returns true if the stop is planned. A planned stop is a stop
// that is visited by a vehicle. An unplanned stop is a stop that is not
// visited by a vehicle.
func (v SolutionStop) IsPlanned() bool {
	return v.solution.next[v.index] != v.solution.previous[v.index]
}

// Previous returns the previous stop the vehicle visited before the stop.
// If the stop is the first stop of a vehicle, the solution stop itself is
// returned. If the stop is unplanned, the previous stop has no semantic
// meaning and the stop itself is returned.
func (v SolutionStop) Previous() SolutionStop {
	return SolutionStop{
		index:    v.solution.previous[v.index],
		solution: v.solution,
	}
}

// PreviousIndex returns the index of the previous solution stop the
// vehicle visited before the stop. If the stop is the first stop of a
// vehicle, the index of the stop itself is returned. If the stop is
// unplanned, the previous stop has no semantic meaning and the index of
// the stop itself is returned.
func (v SolutionStop) PreviousIndex() int {
	return v.solution.previous[v.index]
}

// ArrivalValue returns the arrival time of the stop as a float64. If the
// stop is unplanned, the arrival time has no semantic meaning.
func (v SolutionStop) ArrivalValue() float64 {
	return v.solution.arrival[v.index]
}

// Arrival returns the arrival time of the stop. If the stop is unplanned,
// the arrival time has no semantic meaning.
func (v SolutionStop) Arrival() time.Time {
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		return v.solution.model.Epoch().
			Add(
				time.Duration(v.ArrivalValue()) *
					v.solution.model.DurationUnit())
	}
	return time.Time{}
}

// Slack returns the slack of the stop as a time.Duration. Slack is defined
// as the duration you can start the invoking stop later without
// postponing the last stop of the vehicle. If the stop is unplanned,
// the slack has no semantic meaning. Slack is a consequence of the
// earliest start of stops, if no earliest start is set, the slack is
// always zero.
func (v SolutionStop) Slack() time.Duration {
	return time.Duration(v.SlackValue()) *
		v.solution.model.DurationUnit()
}

// SlackValue returns the slack of the stop as a float64.
func (v SolutionStop) SlackValue() float64 {
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		return v.solution.slack[v.index]
	}
	return 0.0
}

// StartValue returns the start time of the stop as a float64. If the stop
// is unplanned, the start time has no semantic meaning. The returned
// value is the number of Model.DurationUnit units since Model.Epoch.
func (v SolutionStop) StartValue() float64 {
	return v.solution.start[v.index]
}

// Start returns the start time of the stop. If the stop is unplanned, the
// start time has no semantic meaning.
func (v SolutionStop) Start() time.Time {
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		return v.solution.model.Epoch().
			Add(time.Duration(v.StartValue()) *
				v.solution.model.DurationUnit())
	}
	return time.Time{}
}

// EndValue returns the end time of the stop as a float64. If the stop is
// unplanned, the end time has no semantic meaning. The returned value is
// the number of Model.DurationUnit units since Model.Epoch.
func (v SolutionStop) EndValue() float64 {
	return v.solution.end[v.index]
}

// End returns the end time of the stop. If the stop is unplanned, the end
// time has no semantic meaning.
func (v SolutionStop) End() time.Time {
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		return v.solution.model.Epoch().
			Add(time.Duration(v.EndValue()) *
				v.solution.model.DurationUnit())
	}
	return time.Time{}
}

// DurationValue returns the duration of the stop as a float64. If the stop
// is unplanned, the duration has no semantic meaning.
func (v SolutionStop) DurationValue() float64 {
	return v.EndValue() - v.StartValue()
}

// TravelDurationValue returns the travel duration of the stop as a
// float64. If the stop is unplanned, the travel duration has no semantic
// meaning. The travel duration is the time it takes to get to the
// invoking stop. The returned value is the number of
// Model.DurationUnit units.
func (v SolutionStop) TravelDurationValue() float64 {
	return v.CumulativeTravelDurationValue() - v.Previous().CumulativeTravelDurationValue()
}

// TravelDuration returns the travel duration of the stop as a
// time.Duration. If the stop is unplanned, the travel duration has no
// semantic meaning. The travel duration is the time it takes to get to
// the invoking stop.
func (v SolutionStop) TravelDuration() time.Duration {
	return time.Duration(v.TravelDurationValue()) *
		v.solution.model.DurationUnit()
}

// CumulativeTravelDurationValue returns the cumulative travel duration of
// the stop as a float64. The cumulative travel duration is the sum of the
// travel durations of all stops that are visited before the stop. If the
// stop is unplanned, the cumulative travel duration has no semantic
// meaning. The returned value is the number of Model.DurationUnit units.
func (v SolutionStop) CumulativeTravelDurationValue() float64 {
	if v.solution.next[v.index] != v.solution.previous[v.index] {
		return v.solution.cumulativeTravelDuration[v.index]
	}
	return 0
}

// CumulativeTravelDuration returns the cumulative value of the expression
// for the stop as a time.Duration. The cumulative travel duration is the
// sum of the travel durations of all stops that are visited before the
// stop and the stop itself. If the stop is unplanned, the cumulative
// travel duration has no semantic meaning.
func (v SolutionStop) CumulativeTravelDuration() time.Duration {
	return time.Duration(v.CumulativeTravelDurationValue()) *
		v.solution.model.DurationUnit()
}

// Vehicle returns the SolutionVehicle that visits the stop. If the stop
// is unplanned, the vehicle has no semantic meaning and a panic will be
// raised.
func (v SolutionStop) Vehicle() SolutionVehicle {
	if v.solution.next[v.index] == v.solution.previous[v.index] {
		panic("cannot get route of unplanned visit")
	}
	return v.vehicle()
}

func (v SolutionStop) vehicle() SolutionVehicle {
	return SolutionVehicle{
		index:    v.solution.inVehicle[v.index],
		solution: v.solution,
	}
}

// VehicleIndex returns the index of the SolutionVehicle that visits the
// stop. If the stop is unplanned, a panic will be raised.
func (v SolutionStop) VehicleIndex() int {
	if v.solution.next[v.index] == v.solution.previous[v.index] {
		panic("cannot get route index of unplanned visit")
	}
	return v.solution.inVehicle[v.index]
}

// Position returns the position of the stop in the vehicle starting with
// 0 for the first stop. If the stop is unplanned, a panic will be raised.
func (v SolutionStop) Position() int {
	if v.solution.next[v.index] == v.solution.previous[v.index] {
		panic("cannot get stop position of unplanned stop")
	}
	return v.solution.stopPosition[v.index]
}

// IsFixed returns true if the stop is fixed. A fixed stop is a stop that
// that can not transition form being planned to unplanned or vice versa.
func (v SolutionStop) IsFixed() bool {
	return v.ModelStop().IsFixed()
}

// IsLast returns true if the stop is the last stop of a vehicle.
func (v SolutionStop) IsLast() bool {
	return v.solution.next[v.index] == v.index &&
		v.solution.previous[v.index] != v.index
}

// IsFirst returns true if the stop is the first stop of a vehicle.
func (v SolutionStop) IsFirst() bool {
	return v.solution.previous[v.index] == v.index &&
		v.solution.next[v.index] != v.index
}

// IsZero returns true if the stop is the zero value of SolutionStop.
func (v SolutionStop) IsZero() bool {
	return v.solution == nil && v.index == 0
}

// ModelStop returns the ModelStop that is the basis of the SolutionStop.
func (v SolutionStop) ModelStop() ModelStop {
	return v.solution.model.(*modelImpl).stops[v.solution.stop[v.index]]
}

func (v SolutionStop) modelStop() *stopImpl {
	return v.solution.model.(*modelImpl).stops[v.solution.stop[v.index]].(*stopImpl)
}

// ModelStopIndex is the index of the ModelStop in the Model.
func (v SolutionStop) ModelStopIndex() int {
	return v.solution.stop[v.index]
}

func (v SolutionStop) detach() {
	previousIndex := v.solution.previous[v.index]
	nextIndex := v.solution.next[v.index]

	v.solution.next[previousIndex] = nextIndex
	v.solution.previous[nextIndex] = previousIndex
	v.solution.next[v.index] = v.index
	v.solution.previous[v.index] = v.index
	v.solution.inVehicle[v.index] = -1
}

func (v SolutionStop) attach(after int) int {
	v.solution.previous[v.index] = after
	v.solution.next[v.index] = v.solution.next[after]

	v.solution.previous[v.solution.next[after]] = v.index
	v.solution.next[after] = v.index

	v.solution.inVehicle[v.index] = v.solution.inVehicle[after]

	return after
}
