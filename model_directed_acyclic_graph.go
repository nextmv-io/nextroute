// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"slices"

	"github.com/nextmv-io/nextroute/common"
)

// Arc is a directed connection between two nodes ([ModelStops]) that specifies
// that the origin stop must be planned before the destination stop on a
// vehicle's route.
type Arc interface {
	// Origin returns the origin node ([ModelStop]) of the arc.
	Origin() ModelStop
	// Destination returns the destination node ([ModelStop]) of the arc.
	Destination() ModelStop
	// IsDirect returns true if the Destination has to be a direct successor of
	// the Origin, otherwise returns false.
	IsDirect() bool
}

// Arcs is a collection of [Arc]s.
type Arcs []Arc

// DirectedAcyclicGraph is a set of nodes (of type [ModelStop]) connected by
// arcs that does not contain cycles. It restricts the sequence in which the
// stops can be planned on the vehicle. An arc (u -> v) indicates that the stop
// u must be planned before the stop v on the vehicle's route.
type DirectedAcyclicGraph interface {
	// Arcs returns all [Arcs] in the graph.
	Arcs() Arcs

	// IndependentDirectedAcyclicGraphs returns all the independent
	// [DirectedAcyclicGraph]s in the graph. An independent
	// [DirectedAcyclicGraph] is a [DirectedAcyclicGraph] that does not share
	// any [ModelStop]s with any other [DirectedAcyclicGraph]s.
	IndependentDirectedAcyclicGraphs() ([]DirectedAcyclicGraph, error)

	// IsAllowed returns true if the sequence of stops is allowed by the DAG,
	// otherwise returns false.
	IsAllowed(stops ModelStops) (bool, error)

	// HasDirectArc returns true if there is a direct arc between the origin and
	// destination stops, otherwise returns false.
	HasDirectArc(origin, destination ModelStop) bool

	// ModelStops returns all [ModelStops] in the graph.
	ModelStops() ModelStops
	// AddArc adds a new [Arc] in the graph if it was not already added. The new
	// [Arc] should not cause a cycle.
	AddArc(origin, destination ModelStop) error
	// AddDirectArc adds a new [Arc] marked direct in the graph if it was not
	// already added. The new [Arc] should not cause a cycle. The destination
	// stop should be the next stop after the origin stop.
	AddDirectArc(origin, destination ModelStop) error
	// OutboundArcs returns all [Arcs] that have the given [ModelStop] as their
	// origin.
	OutboundArcs(stop ModelStop) Arcs
}

// NewDirectedAcyclicGraph connects NewDirectedAcyclicGraph.
func NewDirectedAcyclicGraph() DirectedAcyclicGraph {
	return &directedAcyclicGraphImpl{
		arcs:               []Arc{},
		adjacencyList:      map[int][]int{},
		outboundArcs:       map[int]Arcs{},
		outboundDirectArcs: map[int]Arc{},
		inboundDirectArcs:  map[int]Arc{},
	}
}

// directedAcyclicGraphImpl implements DirectedAcyclicGraph.
type directedAcyclicGraphImpl struct {
	adjacencyList      map[int][]int
	outboundArcs       map[int]Arcs
	outboundDirectArcs map[int]Arc
	inboundDirectArcs  map[int]Arc
	arcs               Arcs
}

func (d *directedAcyclicGraphImpl) addArc(origin, destination ModelStop, isDirect bool) error {
	if isDirect {
		if arc, alreadyDefined := d.outboundDirectArcs[origin.Index()]; alreadyDefined {
			if arc.Destination().Index() != destination.Index() {
				return fmt.Errorf(
					"origin stop already has a direct arc: %v -> %v",
					origin,
					arc.Destination(),
				)
			}
			return nil
		}
		if arc, alreadyDefined := d.inboundDirectArcs[destination.Index()]; alreadyDefined {
			return fmt.Errorf(
				"destination stop already has a direct arc: %v -> %v",
				arc.Origin(),
				destination,
			)
		}
	}

	d.addEdge(origin.Index(), destination.Index())
	if d.isCyclic() {
		return fmt.Errorf(
			"arc would create a cycle and cannot be added to the DAG: %v -> %v",
			origin,
			destination,
		)
	}

	arc := arcImpl{
		origin:      origin,
		destination: destination,
		isDirect:    isDirect,
	}
	d.arcs = append(d.arcs, arc)
	d.outboundArcs[origin.Index()] = append(d.outboundArcs[origin.Index()], arc)
	if isDirect {
		d.outboundDirectArcs[origin.Index()] = arc
		d.inboundDirectArcs[destination.Index()] = arc
	}
	return nil
}

func (d *directedAcyclicGraphImpl) AddArc(origin, destination ModelStop) error {
	if origin == nil {
		return fmt.Errorf("origin stop cannot be nil")
	}
	if destination == nil {
		return fmt.Errorf("destination stop cannot be nil")
	}
	if origin.Model().IsLocked() {
		return fmt.Errorf(lockErrorMessage, "add arc")
	}
	err := d.addArc(origin, destination, false)
	if err != nil {
		return err
	}

	return nil
}

func (d *directedAcyclicGraphImpl) HasDirectArc(origin, destination ModelStop) bool {
	return d.hasDirectArc(origin.Index(), destination.Index())
}

func (d *directedAcyclicGraphImpl) hasDirectArc(originIndex, destinationIndex int) bool {
	if arc, ok := d.outboundDirectArcs[originIndex]; ok {
		return arc.Destination().Index() == destinationIndex
	}
	return false
}

func (d *directedAcyclicGraphImpl) AddDirectArc(origin, destination ModelStop) error {
	if origin == nil {
		return fmt.Errorf("origin stop cannot be nil")
	}
	if destination == nil {
		return fmt.Errorf("destination stop cannot be nil")
	}
	if origin.Model().IsLocked() {
		return fmt.Errorf(lockErrorMessage, "add arc")
	}
	err := d.addArc(origin, destination, true)
	if err != nil {
		return err
	}

	return nil
}

func (d *directedAcyclicGraphImpl) Arcs() Arcs {
	return slices.Clone(d.arcs)
}

func (d *directedAcyclicGraphImpl) updateColors(
	dag DirectedAcyclicGraph,
	parent int,
	colors map[int]bool,
) error {
	colors[parent] = true
	if arcs, ok := d.outboundArcs[parent]; ok {
		for _, arc := range arcs {
			colors[arc.Origin().Index()] = true
			err := dag.AddArc(arc.Origin(), arc.Destination())
			if err != nil {
				return err
			}
			err = d.updateColors(dag, arc.Destination().Index(), colors)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *directedAcyclicGraphImpl) IndependentDirectedAcyclicGraphs() ([]DirectedAcyclicGraph, error) {
	if len(d.adjacencyList) == 0 {
		return []DirectedAcyclicGraph{}, nil
	}
	colors := make(map[int]bool)

	dags := make([]DirectedAcyclicGraph, 0)

	for key := range d.adjacencyList {
		if _, ok := colors[key]; !ok {
			dag := NewDirectedAcyclicGraph()

			err := d.updateColors(dag, key, colors)
			if err != nil {
				return nil, err
			}
			dags = append(dags, dag)
		}
	}

	return dags, nil
}

func (d *directedAcyclicGraphImpl) IsAllowed(stops ModelStops) (bool, error) {
	if len(stops) < 2 {
		return true, nil
	}

	uniqueStops := common.UniqueDefined(stops, func(stop ModelStop) int {
		return stop.Index()
	})

	if len(uniqueStops) != len(stops) {
		return false, fmt.Errorf("stops are not unique")
	}

	c := directedAcyclicGraphImpl{
		adjacencyList:      make(map[int][]int, len(d.adjacencyList)),
		outboundArcs:       make(map[int]Arcs, len(d.outboundArcs)),
		outboundDirectArcs: make(map[int]Arc, len(d.outboundDirectArcs)),
		inboundDirectArcs:  make(map[int]Arc, len(d.inboundDirectArcs)),
		arcs:               make(Arcs, 0, len(d.arcs)),
	}
	for _, arc := range d.arcs {
		err := c.addArc(arc.Origin(), arc.Destination(), arc.IsDirect())
		if err != nil {
			return false, err
		}
	}

LoopStops:
	for idx := 1; idx < len(stops); idx++ {
		origin, destination := stops[idx-1], stops[idx]
		for _, arc := range c.arcs {
			// if arc is a direct arc, then destination should be the next stop
			// after origin
			if arc.IsDirect() {
				if arc.Origin().Index() == origin.Index() &&
					arc.Destination().Index() != destination.Index() {
					return false, nil
				}
				if arc.Destination().Index() == destination.Index() &&
					arc.Origin().Index() != origin.Index() {
					return false, nil
				}
			}
			if arc.Origin().Index() == origin.Index() && arc.Destination().Index() == destination.Index() {
				continue LoopStops
			}
		}
		c.addEdge(origin.Index(), destination.Index())
		if c.isCyclic() {
			return false, nil
		}
	}

	return true, nil
}

func (d *directedAcyclicGraphImpl) ModelStops() ModelStops {
	modelStops := make(ModelStops, 0)
	modelStopAdded := make(map[int]struct{})
	for _, arc := range d.arcs {
		if _, ok := modelStopAdded[arc.Origin().Index()]; !ok {
			modelStops = append(modelStops, arc.Origin())
			modelStopAdded[arc.Origin().Index()] = struct{}{}
		}
		if _, ok := modelStopAdded[arc.Destination().Index()]; !ok {
			modelStops = append(modelStops, arc.Destination())
			modelStopAdded[arc.Destination().Index()] = struct{}{}
		}
	}
	return modelStops
}

func (d *directedAcyclicGraphImpl) OutboundArcs(stop ModelStop) Arcs {
	return slices.Clone(d.outboundArcs[stop.Index()])
}

func (d *directedAcyclicGraphImpl) addEdge(u int, v int) {
	d.adjacencyList[u] = append(d.adjacencyList[u], v)
}

func (d *directedAcyclicGraphImpl) isCyclic() bool {
	visited := make(map[int]bool)
	stack := make(map[int]bool)

	returnValue := false
	common.RangeMap(d.adjacencyList, func(vertex int, _ []int) bool {
		if d.isCyclicUtil(vertex, visited, stack) {
			returnValue = true
			return true
		}
		return false
	})

	return returnValue
}

func (d *directedAcyclicGraphImpl) isCyclicUtil(vertex int, visited map[int]bool, stack map[int]bool) bool {
	visited[vertex] = true
	stack[vertex] = true

	for _, adjVertex := range d.adjacencyList[vertex] {
		if !visited[adjVertex] && d.isCyclicUtil(adjVertex, visited, stack) {
			return true
		} else if stack[adjVertex] {
			return true
		}
	}

	stack[vertex] = false

	return false
}

// arcImpl implements Arc.
type arcImpl struct {
	origin      ModelStop
	destination ModelStop
	isDirect    bool
}

func (a arcImpl) Origin() ModelStop {
	return a.origin
}

func (a arcImpl) Destination() ModelStop {
	return a.destination
}

func (a arcImpl) IsDirect() bool {
	return a.isDirect
}
