package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

func checkCanBeUsedInPlanUnit(stop nextroute.ModelStop) error {
	if stop == nil {
		return fmt.Errorf("stop cannot be nil")
	}

	if stop.HasPlanStopsUnit() {
		return fmt.Errorf(
			"stop %s [%v] already has a plan planUnit, cannot be used in multiple model plan units",
			stop.ID(),
			stop.Index(),
		)
	}

	if stop.IsFirstOrLast() {
		return fmt.Errorf(
			"stop %s [%v] is first or last, cannot be used in a model plan unit",
			stop.ID(),
			stop.Index(),
		)
	}

	return nil
}

func newPlanSingleStop(
	index int,
	stop nextroute.ModelStop,
) (nextroute.ModelPlanStopsUnit, error) {
	err := checkCanBeUsedInPlanUnit(stop)

	if err != nil {
		return nil, err
	}

	planUnit := &planMultipleStopsImpl{
		modelDataImpl: newModelDataImpl(),
		index:         index,
		stops:         nextroute.ModelStops{stop},
		dag:           NewDirectedAcyclicGraph(),
	}
	stop.(*stopImpl).planUnit = planUnit

	return planUnit, nil
}

func newPlanMultipleStops(
	index int,
	modelStops nextroute.ModelStops,
	sequence nextroute.DirectedAcyclicGraph,
) (nextroute.ModelPlanStopsUnit, error) {
	if len(modelStops) < 2 {
		return nil, fmt.Errorf("multiple stops plan must have at least 2 stops")
	}

	planUnit := &planMultipleStopsImpl{
		modelDataImpl: newModelDataImpl(),
		index:         index,
		stops:         common.DefensiveCopy(modelStops),
		dag:           sequence,
	}
	inStops := make(map[int]bool)
	stops := make([]nextroute.ModelStop, len(modelStops))
	for s, modelStop := range modelStops {
		err := checkCanBeUsedInPlanUnit(modelStop)

		if err != nil {
			return nil, err
		}

		if _, ok := inStops[modelStop.Index()]; ok {
			return nil,
				fmt.Errorf(
					"duplicate stop %s [%v] in model plan unit",
					modelStop.ID(),
					modelStop.Index(),
				)
		}

		inStops[modelStop.Index()] = true
		modelStop.(*stopImpl).planUnit = planUnit
		stops[s] = modelStop
	}

	for _, arc := range sequence.Arcs() {
		if _, ok := inStops[arc.Origin().Index()]; !ok {
			return nil, fmt.Errorf(
				"arc (origin, destination) (%v,%v) has origin not present in model stops",
				arc.Origin(),
				arc.Origin(),
			)
		}
		if _, ok := inStops[arc.Destination().Index()]; !ok {
			return nil, fmt.Errorf(
				"arc (origin, destination) (%v,%v) has destination not present in model stops",
				arc.Origin(),
				arc.Origin(),
			)
		}
	}

	return planUnit, nil
}

// planMultipleStopsImpl implements nextroute.ModelPlanMultipleStops.
type planMultipleStopsImpl struct {
	dag nextroute.DirectedAcyclicGraph
	modelDataImpl
	stops         nextroute.ModelStops
	index         int
	planUnitsUnit nextroute.ModelPlanUnitsUnit
}

func (p *planMultipleStopsImpl) PlanUnitsUnit() (nextroute.ModelPlanUnitsUnit, bool) {
	return p.planUnitsUnit, p.planUnitsUnit != nil
}

func (p *planMultipleStopsImpl) setPlanUnitsUnit(planUnitsUnit nextroute.ModelPlanUnitsUnit) error {
	if p.planUnitsUnit != nil {
		return fmt.Errorf("plan unit %v already has a plan units unit", p)
	}

	p.planUnitsUnit = planUnitsUnit
	return nil
}

func (p *planMultipleStopsImpl) String() string {
	return fmt.Sprintf("plan_multiple_stops{%v}", p.stops)
}

func (p *planMultipleStopsImpl) Centroid() (common.Location, error) {
	locations := common.Map(p.stops, func(stop nextroute.ModelStop) common.Location {
		return stop.Location()
	})
	return common.Locations(locations).Centroid()
}

func (p *planMultipleStopsImpl) Index() int {
	return p.index
}

func (p *planMultipleStopsImpl) IsFixed() bool {
	for _, stop := range p.stops {
		if stop.IsFixed() {
			return true
		}
	}
	return false
}

func (p *planMultipleStopsImpl) NumberOfStops() int {
	return len(p.stops)
}

func (p *planMultipleStopsImpl) Stops() nextroute.ModelStops {
	return common.DefensiveCopy(p.stops)
}

func (p *planMultipleStopsImpl) DirectedAcyclicGraph() nextroute.DirectedAcyclicGraph {
	return p.dag
}
