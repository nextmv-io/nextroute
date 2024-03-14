// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"

	"github.com/nextmv-io/nextroute/common"
	"gonum.org/v1/gonum/spatial/kdtree"
)

// ModelStopsDistanceQueries is an interface to query distances between stops.
type ModelStopsDistanceQueries interface {
	// ModelStops returns the original set of stops that the distance queries
	// were created from.
	ModelStops() ModelStops
	// NearestStops returns the n nearest stops to the given stop, the stop must
	// be present in the original set of stops.
	NearestStops(stop ModelStop, n int) (ModelStops, error)
	// WithinDistanceStops returns the stops within the given distance of the
	// given stop, the stop must be present in the original set of stops.
	WithinDistanceStops(
		stop ModelStop,
		distance common.Distance,
	) (ModelStops, error)
}

// NewModelStopsDistanceQueries returns a new ModelStopsDistanceQueries.
// All distances in this interface are calculated using the [common.Haversine]
// formula.
func NewModelStopsDistanceQueries(
	stops ModelStops,
) (ModelStopsDistanceQueries, error) {
	wrappers := make(modelStopWrappers, len(stops))
	present := make(map[ModelStop]struct{})
	for i, stop := range stops {
		if !stop.Location().IsValid() {
			return nil,
				fmt.Errorf("stop %v has invalid location", stop.ID())
		}
		present[stop] = struct{}{}
		wrappers[i] = modelStopWrapper{stop: stop}
	}
	return &modelStopsDistanceQueryImpl{
		stops:   wrappers,
		present: present,
		tree:    kdtree.New(wrappers, false),
	}, nil
}

type modelStopsDistanceQueryImpl struct {
	stops   modelStopWrappers
	present map[ModelStop]struct{}
	tree    *kdtree.Tree
}

func (m modelStopsDistanceQueryImpl) WithinDistanceStops(
	stop ModelStop,
	distance common.Distance,
) (ModelStops, error) {
	if _, ok := m.present[stop]; !ok {
		return nil,
			fmt.Errorf(
				"stop %v not in present in original set of stops",
				stop.ID(),
			)
	}
	if distance.Value(common.Kilometers) < 0.0 {
		return ModelStops{}, nil
	}
	km := distance.Value(common.Kilometers)
	keep := kdtree.NewDistKeeper(km * km)
	m.tree.NearestSet(keep, modelStopWrapper{stop: stop})
	stops := make(ModelStops, 0)
	for _, c := range keep.Heap {
		s := c.Comparable.(modelStopWrapper).stop
		if s.Index() == stop.Index() {
			continue
		}
		stops = append(stops, s)
	}
	return stops, nil
}

func (m modelStopsDistanceQueryImpl) ModelStops() ModelStops {
	stops := make(ModelStops, len(m.stops))
	for i, wrapper := range m.stops {
		stops[i] = wrapper.stop
	}
	return stops
}

func (m modelStopsDistanceQueryImpl) NearestStops(
	stop ModelStop,
	n int,
) (ModelStops, error) {
	if _, ok := m.present[stop]; !ok {
		return nil,
			fmt.Errorf(
				"stop %v not in present in original set of stops",
				stop.ID(),
			)
	}
	if n <= 0 {
		return ModelStops{}, nil
	}
	keep := kdtree.NewNKeeper(n + 1)
	m.tree.NearestSet(keep, modelStopWrapper{stop: stop})
	stops := make(ModelStops, 0, n)
	for _, c := range keep.Heap {
		s := c.Comparable.(modelStopWrapper).stop
		if s.Index() == stop.Index() {
			continue
		}
		stops = append(stops, c.Comparable.(modelStopWrapper).stop)
	}
	return stops, nil
}

type modelStopWrapper struct {
	stop ModelStop
}

func (p modelStopWrapper) Compare(
	c kdtree.Comparable,
	d kdtree.Dim,
) float64 {
	q := c.(modelStopWrapper)
	switch d {
	case 0:
		return p.stop.Location().Longitude() - q.stop.Location().Longitude()
	case 1:
		return p.stop.Location().Latitude() - q.stop.Location().Latitude()
	default:
		panic("illegal dimension")
	}
}

func (p modelStopWrapper) Dims() int {
	return 2
}

func (p modelStopWrapper) Distance(c kdtree.Comparable) float64 {
	q := c.(modelStopWrapper)
	if !p.stop.Location().IsValid() || !q.stop.Location().IsValid() {
		return 0.0
	}
	d, err := common.Haversine(p.stop.Location(), q.stop.Location())
	if err != nil {
		panic(err)
	}
	return d.Value(common.Kilometers) * d.Value(common.Kilometers)
}

type modelStopWrappers []modelStopWrapper

func (p modelStopWrappers) Index(i int) kdtree.Comparable {
	return p[i]
}
func (p modelStopWrappers) Len() int {
	return len(p)
}
func (p modelStopWrappers) Pivot(d kdtree.Dim) int {
	return plane{modelStopWrappers: p, Dim: d}.Pivot()
}
func (p modelStopWrappers) Slice(start, end int) kdtree.Interface {
	return p[start:end]
}

type plane struct {
	kdtree.Dim
	modelStopWrappers
}

func (p plane) Less(i, j int) bool {
	switch p.Dim {
	case 0:
		return p.modelStopWrappers[i].stop.Location().Longitude() <
			p.modelStopWrappers[j].stop.Location().Longitude()
	case 1:
		return p.modelStopWrappers[i].stop.Location().Latitude() <
			p.modelStopWrappers[j].stop.Location().Latitude()
	default:
		panic("illegal dimension")
	}
}
func (p plane) Pivot() int {
	return kdtree.Partition(p, kdtree.MedianOfMedians(p))
}

func (p plane) Slice(start, end int) kdtree.SortSlicer {
	p.modelStopWrappers = p.modelStopWrappers[start:end]
	return p
}
func (p plane) Swap(i, j int) {
	p.modelStopWrappers[i], p.modelStopWrappers[j] =
		p.modelStopWrappers[j], p.modelStopWrappers[i]
}
