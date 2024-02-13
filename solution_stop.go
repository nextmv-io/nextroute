package nextroute

import (
	"fmt"
	"strings"
	"time"

	"github.com/nextmv-io/sdk/nextroute"
)

type solutionStopImpl struct {
	solution *solutionImpl
	index    int
}

func toSolutionStop(solution nextroute.Solution, index int) solutionStopImpl {
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
	constraint nextroute.ModelConstraint,
) any {
	return v.solution.constraintValue(constraint, v.index)
}

func (v solutionStopImpl) ObjectiveData(
	objective nextroute.ModelObjective,
) any {
	return v.solution.objectiveValue(objective, v.index)
}

func (v solutionStopImpl) Value(
	expression nextroute.ModelExpression,
) float64 {
	return v.solution.value(expression, v.index)
}

func (v solutionStopImpl) CumulativeValue(
	expression nextroute.ModelExpression,
) float64 {
	return v.solution.cumulativeValue(expression, v.index)
}

func (v solutionStopImpl) Solution() nextroute.Solution {
	return v.solution
}

func (v solutionStopImpl) PlanStopsUnit() nextroute.SolutionPlanStopsUnit {
	return v.planStopsUnit()
}

func (v solutionStopImpl) planStopsUnit() *solutionPlanStopsUnitImpl {
	return v.solution.stopToPlanUnit[v.index]
}

func (v solutionStopImpl) Index() int {
	return v.index
}

func (v solutionStopImpl) Next() nextroute.SolutionStop {
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

func (v solutionStopImpl) Previous() nextroute.SolutionStop {
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

func (v solutionStopImpl) Vehicle() nextroute.SolutionVehicle {
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

func (v solutionStopImpl) ModelStop() nextroute.ModelStop {
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
