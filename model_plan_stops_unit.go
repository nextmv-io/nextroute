// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"slices"

	"github.com/nextmv-io/nextroute/common"
)

// ModelPlanStopsUnit is a set of stops. It is a set of stops
// that are required to be planned together on the same vehicle. For example,
// a unit can be a pickup and a delivery stop that are required to be planned
// together on the same vehicle.
type ModelPlanStopsUnit interface {
	ModelPlanUnit

	// Centroid returns the centroid of the unit. The centroid is the
	// average location of all stops in the unit.
	Centroid() (common.Location, error)

	// DirectedAcyclicGraph returns the [DirectedAcyclicGraph] of the plan
	// unit.
	DirectedAcyclicGraph() DirectedAcyclicGraph

	// NumberOfStops returns the number of stops in the invoking unit.
	NumberOfStops() int

	// Stops returns the stops in the invoking unit.
	Stops() ModelStops
}

// ModelPlanStopsUnits is a slice of model plan stops units .
type ModelPlanStopsUnits []ModelPlanStopsUnit

func checkCanBeUsedInPlanUnit(stop ModelStop) error {
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
	stop ModelStop,
) (ModelPlanStopsUnit, error) {
	err := checkCanBeUsedInPlanUnit(stop)

	if err != nil {
		return nil, err
	}

	planUnit := &planMultipleStopsImpl{
		modelDataImpl: newModelDataImpl(),
		index:         index,
		stops:         ModelStops{stop},
		dag:           NewDirectedAcyclicGraph(),
	}
	stop.(*stopImpl).planUnit = planUnit

	return planUnit, nil
}

func newPlanMultipleStops(
	index int,
	modelStops ModelStops,
	sequence DirectedAcyclicGraph,
) (ModelPlanStopsUnit, error) {
	if len(modelStops) < 2 {
		return nil, fmt.Errorf("multiple stops plan must have at least 2 stops")
	}

	planUnit := &planMultipleStopsImpl{
		modelDataImpl: newModelDataImpl(),
		index:         index,
		stops:         slices.Clone(modelStops),
		dag:           sequence,
	}
	inStops := make(map[int]bool)
	stops := make([]ModelStop, len(modelStops))
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

// planMultipleStopsImpl implements ModelPlanMultipleStops.
type planMultipleStopsImpl struct {
	dag DirectedAcyclicGraph
	modelDataImpl
	stops         ModelStops
	index         int
	planUnitsUnit ModelPlanUnitsUnit
}

func (p *planMultipleStopsImpl) PlanUnitsUnit() (ModelPlanUnitsUnit, bool) {
	return p.planUnitsUnit, p.planUnitsUnit != nil
}

func (p *planMultipleStopsImpl) setPlanUnitsUnit(planUnitsUnit ModelPlanUnitsUnit) error {
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
	locations := common.Map(p.stops, func(stop ModelStop) common.Location {
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

func (p *planMultipleStopsImpl) Stops() ModelStops {
	return slices.Clone(p.stops)
}

func (p *planMultipleStopsImpl) DirectedAcyclicGraph() DirectedAcyclicGraph {
	return p.dag
}
