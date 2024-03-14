// © 2019-present nextmv.io inc

package factory

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	"github.com/nextmv-io/nextroute/schema"
	"github.com/nextmv-io/sdk/run"
)

// ClusterSolutionOptions configure how the [NewGreedySolution] function builds [nextroute.Solution].
type ClusterSolutionOptions struct {
	Depth int     `json:"depth" usage:"maximum failed tries to add a cluster to a vehicle" default:"10" minimum:"0"`
	Speed float64 `json:"speed" usage:"speed of the vehicle in meters per second" default:"10" minimum:"0"`
}

// FilterAreaOptions configure how the [NewGreedySolution] function builds [nextroute.Solution]. It limits the area
// one vehicle can cover during construction. This limit is only applied during the construction of the solution.
type FilterAreaOptions struct {
	MaximumSide float64 `json:"maximum_side" usage:"maximum side of the square area in meters" default:"100000" minimum:"0"`
}

// GreedySolutionOptions configure how the [NewGreedySolution] function builds [nextroute.Solution].
type GreedySolutionOptions struct {
	ClusterSolutionOptions ClusterSolutionOptions `json:"cluster_solution_options" usage:"options for the cluster solution"`
	FilterAreaOptions      FilterAreaOptions      `json:"filter_area_options" usage:"options for the filter area"`
}

// StopCluster represents a group of stops that can be added to a vehicle.
type StopCluster interface {
	// Stops returns the stops in the stop cluster.
	Stops() []schema.Stop
	// Centroid returns the centroid of the stop cluster.
	Centroid() schema.Location
}

// StopClusterGenerator returns a list of stop clusters for the given input.
type StopClusterGenerator interface {
	// Generate returns a list of stop clusters for the given input.
	// A cluster is a group of stops that can be added to a vehicle. If a stop
	// is added to a cluster all the stops belonging to the same plan units
	// must be added to the same cluster.
	Generate(
		input schema.Input,
		options Options,
		factory ModelFactory,
	) ([]StopCluster, error)
}

// StopClusterSorter returns a sorted list of stop clusters for the given input.
type StopClusterSorter interface {
	// Sort returns a sorted list of stop clusters for the given input.
	Sort(
		input schema.Input,
		clusters []StopCluster,
		factory ModelFactory,
	) ([]StopCluster, error)
}

// StopClusterFilter returns true if the given stop cluster should be filtered
// out.
type StopClusterFilter interface {
	// Filter returns true if the given stop cluster should be filtered out.
	Filter(
		input schema.Input,
		cluster StopCluster,
		factory ModelFactory,
	) (bool, error)
}

// ModelFactory returns a new model for the given input and options.
type ModelFactory interface {
	// NewModel returns a new model for the given input and options.
	NewModel(schema.Input, Options) (nextroute.Model, error)
}

// NewStartSolution returns a start solution. It uses input, factoryOptions and
// modelFactory to create a model to create a start solution. The start solution
// is created using the given solveOptions and clusterSolutionOptions. The
// solveOptions is used to limit the duration and the number of parallel runs at
// the same time. The clusterSolutionOptions is used to create the clusters to
// create the start solution, see [NewClusterSolution].
func NewStartSolution(
	cont context.Context,
	input schema.Input,
	factoryOptions Options,
	modelFactory ModelFactory,
	solveOptions nextroute.ParallelSolveOptions,
	clusterSolutionOptions ClusterSolutionOptions,
) (nextroute.Solution, error) {
	if cont.Value(run.Start) == nil {
		cont = context.WithValue(cont, run.Start, time.Now())
	}
	ctx, cancelFn := context.WithDeadline(
		cont,
		cont.Value(run.Start).(time.Time).Add(solveOptions.Duration),
	)

	defer cancelFn()

	model, err := modelFactory.NewModel(input, factoryOptions)
	if err != nil {
		return nil, err
	}

	bestSolution, err := nextroute.NewSolution(model)
	if err != nil {
		return nil, err
	}

	if len(input.Vehicles) <= 1 {
		return bestSolution, nil
	}

	maxPlanUnitSide := 0.0

	for _, planUnit := range model.PlanUnits() {
		stops := getStops(planUnit)
		boundingBox := common.NewBoundingBox(
			common.Map(stops, func(stop nextroute.ModelStop) common.Location {
				return stop.Location()
			}),
		)

		side := math.Max(
			boundingBox.Width().Value(common.Meters),
			boundingBox.Height().Value(common.Meters),
		)

		if side > maxPlanUnitSide {
			maxPlanUnitSide = side
		}
	}

	if maxPlanUnitSide == 0.0 {
		return bestSolution, nil
	}

	boundingBox := common.NewBoundingBox(
		common.Map(input.Stops, func(stop schema.Stop) common.Location {
			l, _ := common.NewLocation(stop.Location.Lon, stop.Location.Lat)
			return l
		}),
	)

	side := math.Max(
		boundingBox.Width().Value(common.Meters),
		boundingBox.Height().Value(common.Meters),
	)

	nrExperiments := 2.0

	if (solveOptions.ParallelRuns == -1 || solveOptions.ParallelRuns > 2.0) &&
		runtime.NumCPU() > 2 {
		nrExperiments = math.Min(
			math.Max(float64(solveOptions.ParallelRuns), math.MaxFloat64),
			float64(runtime.NumCPU()),
		)
	}

	minSide := maxPlanUnitSide
	maxSide := side

	lowerBoundSide := minSide
	upperBoundSide := maxSide

	for {
		select {
		case <-ctx.Done():
			return bestSolution, nil
		default:
			if upperBoundSide-lowerBoundSide < maxPlanUnitSide {
				return bestSolution, nil
			}

			delta := (upperBoundSide - lowerBoundSide) / (nrExperiments - 1)

			n := int(nrExperiments)

			experimentSides := make([]float64, n)

			for i := 0; i < n; i++ {
				experimentSides[i] = lowerBoundSide + float64(i)*delta
			}

			var waitGroup sync.WaitGroup
			waitGroup.Add(n)

			type experimentResult struct {
				side     float64
				solution nextroute.Solution
			}

			experimentResults := make(chan experimentResult, n)

			for _, experimentSide := range experimentSides {
				go func(side float64) {
					solution, _ := NewClusterSolution(
						ctx,
						input,
						factoryOptions,
						NewPlanUnitStopClusterGenerator(),
						NewSortStopClustersRandom(),
						NewSortStopClustersOnDistanceFromCentroid(),
						NewStopClusterFilterArea(
							common.NewDistance(
								side,
								common.Meters,
							),
						),
						clusterSolutionOptions,
						modelFactory,
					)
					experimentResults <- experimentResult{
						side:     side,
						solution: solution,
					}
					waitGroup.Done()
				}(experimentSide)
			}

			waitGroup.Wait()
			close(experimentResults)

			bestSize := -1.0
			bestScore := math.MaxFloat64

			for result := range experimentResults {
				if bestSolution == nil ||
					result.solution.Score() < bestSolution.Score() {
					bestSolution = result.solution
				}
				if result.solution.Score() < bestScore {
					bestSize = result.side
					bestScore = result.solution.Score()
				}
			}
			if bestSize == minSide || bestSize == maxSide {
				lowerBoundSide += delta / 3.0
				upperBoundSide -= delta / 3.0
				continue
			}

			lowerBoundSide = bestSize + delta/nrExperiments
			upperBoundSide = bestSize - delta/nrExperiments
		}
	}
}

// NewGreedySolution returns a greedy solution for the given side.
func NewGreedySolution(
	ctx context.Context,
	input schema.Input,
	options Options,
	greedySolutionOptions GreedySolutionOptions,
	modelFactory ModelFactory,
) (nextroute.Solution, error) {
	return NewClusterSolution(
		ctx,
		input,
		options,
		NewPlanUnitStopClusterGenerator(),
		NewSortStopClustersRandom(),
		NewSortStopClustersOnDistanceFromCentroid(),
		NewStopClusterFilterArea(
			common.NewDistance(
				greedySolutionOptions.FilterAreaOptions.MaximumSide,
				common.Meters,
			),
		),
		greedySolutionOptions.ClusterSolutionOptions,
		modelFactory,
	)
}

// NewDefaultModelFactory returns a new default model factory.
// The default model factory creates a new model for the given side and
// options.
func NewDefaultModelFactory() ModelFactory {
	return defaultModelFactoryImpl{}
}

type defaultModelFactoryImpl struct{}

func (d defaultModelFactoryImpl) NewModel(
	input schema.Input,
	options Options,
) (nextroute.Model, error) {
	return NewModel(input, options)
}

// NewStopCluster returns a new stop cluster for the given stops.
func NewStopCluster(
	stops []schema.Stop) (StopCluster, error) {
	if len(stops) == 0 {
		return nil, fmt.Errorf("cannot create stop cluster with no stops")
	}
	return stopCluster{
		stops:    slices.Clone(stops),
		centroid: CentroidLocation(stops),
	}, nil
}

// NewPlanUnitStopClusterGenerator returns a list of stop clusters based
// upon unplanned plan units.
func NewPlanUnitStopClusterGenerator() StopClusterGenerator {
	return &planUnitStopClusterGeneratorImpl{}
}

type planUnitStopClusterGeneratorImpl struct {
}

func (s *planUnitStopClusterGeneratorImpl) Generate(
	input schema.Input,
	modelOptions Options,
	modelFactory ModelFactory,
) ([]StopCluster, error) {
	clusters := make([]StopCluster, 0, len(input.Stops))

	model, err := modelFactory.NewModel(input, modelOptions)
	if err != nil {
		return nil, err
	}

	solution, err := nextroute.NewSolution(model)
	if err != nil {
		return nil, err
	}

	for _, solutionPlanUnit := range solution.UnPlannedPlanUnits().SolutionPlanUnits() {
		stops := getInputStops(solutionPlanUnit.ModelPlanUnit())
		cluster, err := NewStopCluster(stops)
		if err != nil {
			return nil, err
		}
		clusters = append(
			clusters,
			cluster,
		)
	}
	return clusters, nil
}

// NewSortStopClustersRandom returns StopClusterSorter which sorts the stop
// clusters randomly.
func NewSortStopClustersRandom() StopClusterSorter {
	return &sortStopClustersRandomImpl{}
}

type sortStopClustersRandomImpl struct {
}

func (s *sortStopClustersRandomImpl) Sort(
	_ schema.Input,
	clusters []StopCluster,
	_ ModelFactory,
) ([]StopCluster, error) {
	rand.Shuffle(len(clusters), func(i, j int) { clusters[i], clusters[j] = clusters[j], clusters[i] })
	return clusters, nil
}

// NewSortStopClustersOnDistanceFromCentroid sorts the stop clusters based upon
// the distance from the centroid of the stop cluster to the centroid of all
// stops.
func NewSortStopClustersOnDistanceFromCentroid() StopClusterSorter {
	return &sortStopClustersOnDistanceFromCentroidImpl{}
}

type sortStopClustersOnDistanceFromCentroidImpl struct {
}

func (s *sortStopClustersOnDistanceFromCentroidImpl) Sort(
	input schema.Input,
	clusters []StopCluster,
	_ ModelFactory,
) ([]StopCluster, error) {
	centroid := CentroidLocation(input.Stops)
	var err error
	sort.Slice(clusters, func(i, j int) bool {
		distanceI, e := HaversineDistance(clusters[i].Centroid(), centroid)
		if e != nil {
			err = e
		}
		distanceJ, e := HaversineDistance(clusters[j].Centroid(), centroid)
		if e != nil {
			err = e
		}
		return distanceI.Value(common.Meters) < distanceJ.Value(common.Meters)
	})
	return clusters, err
}

// NewAndStopClusterFilter returns a StopClusterFilter that filters out stop clusters that are filtered out by all
// the given filters.
func NewAndStopClusterFilter(
	filter StopClusterFilter,
	filters ...StopClusterFilter,
) StopClusterFilter {
	// combine filters
	allFilters := make([]StopClusterFilter, 0, len(filters)+1)
	allFilters = append(allFilters, filter)
	return &nAryStopClusterFilterImpl{
		filters:     allFilters,
		conjunction: true,
	}
}

// NewOrStopClusterFilter returns a StopClusterFilter that filters out stop clusters that are filtered out by any of
// the given filters.
func NewOrStopClusterFilter(
	filter StopClusterFilter,
	filters ...StopClusterFilter,
) StopClusterFilter {
	// combine filters
	allFilters := make([]StopClusterFilter, 0, len(filters)+1)
	allFilters = append(allFilters, filter)
	return &nAryStopClusterFilterImpl{
		filters:     allFilters,
		conjunction: false,
	}
}

type nAryStopClusterFilterImpl struct {
	filters     []StopClusterFilter
	conjunction bool
}

func (n *nAryStopClusterFilterImpl) Filter(
	input schema.Input,
	cluster StopCluster,
	modelFactory ModelFactory,
) (bool, error) {
	for _, filter := range n.filters {
		filter, err := filter.Filter(input, cluster, modelFactory)
		if err != nil {
			return true, err
		}
		if filter && !n.conjunction {
			return filter, nil
		}
		if !filter && n.conjunction {
			return filter, nil
		}
	}
	return false, nil
}

// NewStopClusterFilterArea returns a StopClusterFilter that filters out stop
// clusters resulting in covering a square area defined by the parameter.
// The area is approximated using haversine distances.
func NewStopClusterFilterArea(
	side common.Distance,
) StopClusterFilter {
	return &filterSortStopClusterAreaImpl{
		side: side.Value(common.Meters),
	}
}

type filterSortStopClusterAreaImpl struct {
	side float64
}

func (f *filterSortStopClusterAreaImpl) Filter(
	input schema.Input,
	cluster StopCluster,
	_ ModelFactory,
) (bool, error) {
	if f.side <= 0 {
		return true, nil
	}

	combinedStops := make([]schema.Stop, len(input.Stops)+len(cluster.Stops()))
	copy(combinedStops, input.Stops)
	copy(combinedStops[len(input.Stops):], cluster.Stops())

	if f.side < math.MaxFloat64 {
		boundingBox := common.NewBoundingBox(common.Map(combinedStops, func(stop schema.Stop) common.Location {
			loc, _ := common.NewLocation(stop.Location.Lon, stop.Location.Lat)
			return loc
		}))

		width := boundingBox.Width()
		if width.Value(common.Meters) > f.side {
			return true, nil
		}
		height := boundingBox.Height()
		if height.Value(common.Meters) > f.side {
			return true, nil
		}
	}

	return false, nil
}

// NewClusterSolution returns a solution for the given side using the given
// options.
//
// The solution is constructed by first creating a solution for each vehicle
// and then adding stop groups to the vehicles in a greedy fashion.
//
//   - Raises an error if the side has initial stops on any of the vehicles.
//   - Uses haversine distance independent of the side's distance/duration
//     matrix. Uses the correct
//     distance matrix in the solution returned.
//   - Uses the speed of the vehicle if defined, otherwise the speed defined in
//     the options.
//   - Ignores stop duration groups in construction but not in the solution
//     returned.
//
// # The initial solution is created as following:
//
//	Creates the clusters using the stopClusterGenerator
//
//	In random order of the vehicles in the side:
//
//	 - Add a first cluster to the empty vehicle defined by the
//	   initialStopClusterSorter
//	 - If the vehicle is not solved, the cluster is removed and the next cluster
//	   will be added
//	 - If no clusters can be added, the vehicle will not be used
//	 - If a cluster has been added we continue adding clusters to the vehicle in
//	   the order defined by additionalStopClusterSorter until no more clusters
//	   can be added
//
//	We repeat until no more vehicles or no more clusters to add to the solution.
func NewClusterSolution(
	ctx context.Context,
	input schema.Input,
	options Options,
	stopClusterGenerator StopClusterGenerator,
	initialStopClusterSorter StopClusterSorter,
	additionalStopClusterSorter StopClusterSorter,
	stopClusterFilter StopClusterFilter,
	stopClusterOptions ClusterSolutionOptions,
	modelFactory ModelFactory,
) (nextroute.Solution, error) {
	if modelFactory == nil {
		modelFactory = NewDefaultModelFactory()
	}
	byt, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	copyInput := schema.Input{}

	err = json.Unmarshal(byt, &copyInput)
	if err != nil {
		return nil, err
	}

	// we ignore duration groups for now
	copyInput.DurationGroups = nil

	// we ignore duration matrix for now and need haversine distance
	// anyway as we look for the centroid of the clusters
	copyInput.DurationMatrix = nil

	// we ignore distance matrix for now and need haversine distance
	// anyway as we use the centroid of the clusters
	copyInput.DistanceMatrix = nil

	speed := stopClusterOptions.Speed
	for idx, vehicle := range copyInput.Vehicles {
		if vehicle.InitialStops != nil {
			return nil, fmt.Errorf(
				"initial stops on vehicle %v are not (yet) supported",
				vehicle.ID,
			)
		}
		if vehicle.Speed == nil {
			copyInput.Vehicles[idx].Speed = &speed
		}
	}

	copyInput = applyDefaults(copyInput)

	stops := make(map[string]schema.Stop, len(copyInput.Stops))
	for _, stop := range copyInput.Stops {
		stops[stop.ID] = stop
	}

	clusters, err := stopClusterGenerator.Generate(copyInput, options, modelFactory)

	if err != nil {
		return nil, err
	}

	// We randomize the order of the vehicles we will populate. We could be
	// smarter here, although it is not clear if that would be beneficial.
	rand.Shuffle(len(copyInput.Vehicles), func(i, j int) {
		copyInput.Vehicles[i], copyInput.Vehicles[j] =
			copyInput.Vehicles[j], copyInput.Vehicles[i]
	})
VehicleLoop:
	for vehicleIdx := 0; vehicleIdx < len(copyInput.Vehicles) && len(clusters) > 0; vehicleIdx++ {
		select {
		case <-ctx.Done():
			break VehicleLoop
		default:
			newInput, err := createInputForVehicle(copyInput, vehicleIdx)
			if err != nil {
				return nil, err
			}

			newInput, clusters, err = populateVehicle(
				ctx,
				newInput,
				clusters,
				options,
				initialStopClusterSorter,
				additionalStopClusterSorter,
				stopClusterFilter,
				stopClusterOptions,
				modelFactory,
			)
			if err != nil {
				return nil, err
			}

			copyInput.Vehicles[vehicleIdx] = newInput.Vehicles[0]
		}
	}

	solution, err := newSolution(
		ctx,
		copyInput,
		options,
		nextroute.ParallelSolveOptions{
			Iterations:     1,
			StartSolutions: 0,
			ParallelRuns:   1,
			Duration:       0,
		},
		modelFactory,
	)

	if err != nil {
		return nil, err
	}

	return solution, nil
}

// getStops returns the stops of the given plan unit.
func getStops(planUnit nextroute.ModelPlanUnit) []nextroute.ModelStop {
	return getStopsImpl(planUnit, []nextroute.ModelStop{})
}

// getInputStops returns the stops of the given plan unit.
func getInputStops(planUnit nextroute.ModelPlanUnit) []schema.Stop {
	return common.Map(
		getStops(planUnit),
		func(stop nextroute.ModelStop) schema.Stop {
			return stop.Data().(schema.Stop)
		},
	)
}

func getStopsImpl(planUnit nextroute.ModelPlanUnit, stops []nextroute.ModelStop) []nextroute.ModelStop {
	if planStopsUnit, ok := planUnit.(nextroute.ModelPlanStopsUnit); ok {
		for _, stop := range planStopsUnit.Stops() {
			stops = append(stops, stop)
		}
	}
	if planUnitsUnit, ok := planUnit.(nextroute.ModelPlanUnitsUnit); ok {
		if planUnitsUnit.PlanAll() {
			for _, childPlanUnit := range planUnitsUnit.PlanUnits() {
				stops = getStopsImpl(childPlanUnit, stops)
			}
		}
		if planUnitsUnit.PlanOneOf() {
			// heuristic: take the first plan unit
			stops = getStopsImpl(planUnitsUnit.PlanUnits()[0], stops)
		}
	}
	return stops
}

// newSolution returns a solution for the given side using the given options.
func newSolution(
	ctx context.Context,
	input schema.Input,
	options Options,
	parallelSolveOptions nextroute.ParallelSolveOptions,
	modelFactory ModelFactory,
) (nextroute.Solution, error) {
	model, err := modelFactory.NewModel(input, options)
	if err != nil {
		return nil, err
	}

	parallelSolver, err := nextroute.NewParallelSolver(model)
	if err != nil {
		return nil, err
	}

	solutions, err := parallelSolver.Solve(ctx, parallelSolveOptions)
	if err != nil {
		return nil, err
	}

	return solutions.Last()
}

// HaversineDistance returns the distance between two locations using the
// haversine formula.
func HaversineDistance(from, to schema.Location) (common.Distance, error) {
	fromLocation, err := common.NewLocation(from.Lon, from.Lat)
	if err != nil {
		return common.Distance{}, err
	}

	toLocation, err := common.NewLocation(to.Lon, to.Lat)
	if err != nil {
		return common.Distance{}, err
	}
	distance, err := common.Haversine(fromLocation, toLocation)
	if err != nil {
		return common.Distance{}, err
	}

	return distance, nil
}

// CentroidLocation returns the centroid of the given stops.
func CentroidLocation(stops []schema.Stop) schema.Location {
	lat := 0.0
	lng := 0.0
	for _, stop := range stops {
		lat += stop.Location.Lat
		lng += stop.Location.Lon
	}
	return schema.Location{
		Lat: lat / float64(len(stops)),
		Lon: lng / float64(len(stops)),
	}
}

// stopCluster represents a group of stops that can be added to a vehicle.
type stopCluster struct {
	stops    []schema.Stop
	centroid schema.Location
}

func (s stopCluster) Stops() []schema.Stop {
	return s.stops
}

func (s stopCluster) Centroid() schema.Location {
	return s.centroid
}

// populateVehicle adds as many clusters as possible to the given vehicle
// in the given side.
func populateVehicle(
	ctx context.Context,
	input schema.Input,
	clusters []StopCluster,
	options Options,
	initialStopClusterSorter StopClusterSorter,
	addStopClusterSorter StopClusterSorter,
	stopClusterFilter StopClusterFilter,
	stopClusterOptions ClusterSolutionOptions,
	modelFactory ModelFactory,
) (schema.Input, []StopCluster, error) {
	if len(input.Vehicles) != 1 {
		return input,
			clusters,
			fmt.Errorf(
				"expected 1 vehicle in side schema, got %d",
				len(input.Vehicles),
			)
	}

	if len(clusters) == 0 {
		return input, clusters, nil
	}

	var err error
	if len(input.Stops) == 0 {
		clusters, err = initialStopClusterSorter.Sort(
			input,
			clusters,
			modelFactory,
		)

		if err != nil {
			return input, clusters, err
		}

		initialStops := make([]schema.InitialStop, 0)
		input.Vehicles[0].InitialStops = &initialStops

	ClusterLoop:
		for idx, cluster := range clusters {
			select {
			case <-ctx.Done():
				return input, clusters, err
			default:
				if filter, err := stopClusterFilter.Filter(
					input,
					cluster,
					modelFactory,
				); err != nil || filter {
					if err != nil {
						return input, clusters, err
					}
					continue
				}

				input.Stops = append(input.Stops, cluster.Stops()...)

				solution, err := newSolution(
					ctx,
					input,
					options,
					nextroute.ParallelSolveOptions{
						Iterations:     1,
						StartSolutions: 0,
						ParallelRuns:   1,
						Duration:       30 * time.Second,
					},
					modelFactory,
				)

				if err != nil {
					return input, clusters, err
				}

				if solution.UnPlannedPlanUnits().Size() == 0 {
					clusters[idx] = clusters[len(clusters)-1]
					clusters = clusters[:len(clusters)-1]
					for _, solutionStop := range solution.Vehicles()[0].SolutionStops() {
						if solutionStop.IsFirst() || solutionStop.IsLast() {
							continue
						}
						*input.Vehicles[0].InitialStops = append(
							*input.Vehicles[0].InitialStops,
							schema.InitialStop{
								ID: solutionStop.ModelStop().ID(),
							},
						)
					}
					break ClusterLoop
				}
				input.Stops = input.Stops[:0]
			}
		}
	}

	clusters, err = addStopClusterSorter.Sort(input, clusters, modelFactory)

	if err != nil {
		return input, clusters, err
	}

	successIdx := 0
	for idx := 0; idx < len(clusters); idx++ {
		if idx > successIdx+stopClusterOptions.Depth {
			return input, clusters, nil
		}

		select {
		case <-ctx.Done():
			return input, clusters, err
		default:
			cluster := clusters[idx]

			if filter, err := stopClusterFilter.Filter(
				input,
				cluster,
				modelFactory,
			); err != nil || filter {
				if err != nil {
					return input, clusters, err
				}
				continue
			}

			initialStops := slices.Clone(*input.Vehicles[0].InitialStops)

			newInputVehicle := input.Vehicles[0]
			newInputVehicle.InitialStops = &initialStops

			newInput := schema.Input{
				Defaults: input.Defaults,
				Stops:    slices.Clone(input.Stops),
				Vehicles: []schema.Vehicle{newInputVehicle},
			}

			newInput.Stops = append(newInput.Stops, cluster.Stops()...)

			solution, err := newSolution(
				ctx,
				newInput,
				options,
				nextroute.ParallelSolveOptions{
					Iterations:     1,
					StartSolutions: 0,
					ParallelRuns:   1,
					Duration:       30 * time.Second,
				},
				modelFactory,
			)

			if err != nil {
				return input, clusters, err
			}

			if solution.UnPlannedPlanUnits().Size() > 0 {
				continue
			}

			initialStops = initialStops[:0]

			for _, solutionStop := range solution.Vehicles()[0].SolutionStops() {
				if solutionStop.IsFirst() || solutionStop.IsLast() {
					continue
				}
				initialStops = append(initialStops, schema.InitialStop{
					ID: solutionStop.ModelStop().ID(),
				})
			}

			clusters[idx] = clusters[len(clusters)-1]
			clusters = clusters[:len(clusters)-1]

			input = newInput
			idx--
			successIdx = idx
		}
	}

	return input, clusters, nil
}

// createInputForVehicle returns a new side schema with only the vehicle at the
// given index.
func createInputForVehicle(
	input schema.Input,
	vehicle int,
) (schema.Input, error) {
	if len(input.Vehicles) <= vehicle || vehicle < 0 {
		return input,
			fmt.Errorf(
				"vehicle index %v out of bounds [%v, %v]",
				vehicle,
				0,
				len(input.Vehicles)-1,
			)
	}
	return schema.Input{
		Defaults:       input.Defaults,
		Stops:          []schema.Stop{},
		Vehicles:       []schema.Vehicle{input.Vehicles[vehicle]},
		AlternateStops: input.AlternateStops,
	}, nil
}
