// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"

	"github.com/nextmv-io/nextroute/common"
)

// Cluster is both a constraint and an objective that limits/prefers the
// vehicles a plan cluster will be added to. If used as a constraint a plan
// cluster can only be added to a vehicles whose centroid is closer to the plan
// cluster than the centroid of any other vehicle. In case of using it as an
// objective, those vehicles will be preferred.
type Cluster struct {
	includeFirst bool
	includeLast  bool
}

// NewCluster creates a new cluster component. It needs to be added as a
// constraint or as an objective to the model to be taken into account.
// By default, the first and last stop of a vehicle are not included in the
// centroid calculation.
func NewCluster() (*Cluster, error) {
	return &Cluster{
		includeFirst: false,
		includeLast:  false,
	}, nil
}

// IncludeFirst returns whether the first stop of the vehicle is included in the
// centroid calculation. The centroid is used to determine the distance
// between a new stop and the cluster.
func (l *Cluster) IncludeFirst() bool {
	return l.includeFirst
}

// IncludeLast returns whether the last stop of the vehicle is included in
// the centroid calculation. The centroid is used to determine the distance
// between a new stop and the cluster.
func (l *Cluster) IncludeLast() bool {
	return l.includeLast
}

// SetIncludeFirst sets whether the first stop of the vehicle is included in
// the centroid calculation. The centroid is used to determine the distance
// between a new stop and the cluster.
func (l *Cluster) SetIncludeFirst(includeFirst bool) {
	l.includeFirst = includeFirst
}

// SetIncludeLast sets whether the last stop of the vehicle is included in
// the centroid calculation. The centroid is used to determine the distance
// between a new stop and the cluster.
func (l *Cluster) SetIncludeLast(includeLast bool) {
	l.includeLast = includeLast
}

type centroidData struct {
	location common.Location
	// this will be 0.0 for all stops but the last and only if this is used as
	// an objective.
	compactness float64
}

func (c *centroidData) Copy() Copier {
	return &centroidData{
		location:    c.location,
		compactness: c.compactness,
	}
}

func (c *centroidData) String() string {
	return fmt.Sprintf("%v", c.location)
}

// String implements fmt.Stringer.
func (l *Cluster) String() string {
	return "cluster"
}

// EstimationCost returns the cost of the constraint for estimation purposes.
func (l *Cluster) EstimationCost() Cost {
	return LinearVehicle
}

// UpdateObjectiveStopData implements ObjectiveStopDataUpdater.
func (l *Cluster) UpdateObjectiveStopData(
	solutionStop SolutionStop,
) (Copier, error) {
	return l.updateData(solutionStop, true)
}

// UpdateConstraintStopData implements ConstraintSolutionDataUpdater.
func (l *Cluster) UpdateConstraintStopData(
	solutionStop SolutionStop,
) (Copier, error) {
	return l.updateData(solutionStop, false)
}

func (l *Cluster) updateData(
	solutionStop SolutionStop,
	asObjective bool,
) (Copier, error) {
	if solutionStop.IsFirst() {
		location, err := common.NewLocation(0, 0)
		if err != nil {
			return nil, err
		}
		return &centroidData{
			location: location,
		}, nil
	}

	if solutionStop.IsLast() {
		if asObjective {
			centroid := solutionStop.Previous().ObjectiveData(l).(*centroidData)
			stops := l.getSolutionStops(solutionStop.vehicle())
			compact := compactness(stops, centroid.location, SolutionStop{}, false)
			centroid.compactness = compact
			return centroid, nil
		}
		return solutionStop.Previous().ConstraintData(l).(*centroidData), nil
	}
	nrStops := solutionStop.Position()

	var centroid *centroidData
	if asObjective {
		centroid = solutionStop.Previous().ObjectiveData(l).(*centroidData)
	} else {
		centroid = solutionStop.Previous().ConstraintData(l).(*centroidData)
	}

	location, err := common.NewLocation(
		centroid.location.Longitude()+
			(solutionStop.ModelStop().Location().Longitude()-
				centroid.location.Longitude())/float64(nrStops),
		centroid.location.Latitude()+
			(solutionStop.ModelStop().Location().Latitude()-
				centroid.location.Latitude())/float64(nrStops),
	)
	if err != nil {
		return nil, err
	}

	return &centroidData{
		location: location,
	}, nil
}

func compactness(
	stops []SolutionStop,
	centroid common.Location,
	newStop SolutionStop,
	newStopIsSet bool,
) float64 {
	if newStopIsSet {
		numberOfStops := len(stops)
		location := newStop.ModelStop().location
		newLat := (centroid.Latitude()*float64(numberOfStops) +
			location.Latitude()) / float64(numberOfStops+1)
		newLong := (centroid.Longitude()*float64(numberOfStops) +
			location.Longitude()) / float64(numberOfStops+1)
		newLocation, err := common.NewLocation(newLong, newLat)
		if err != nil {
			panic(err)
		}
		centroid = newLocation
		stops = append(stops, newStop)
	}
	compactness := 0.0
	for _, stop := range stops {
		dist := haversineDistance(centroid, stop.ModelStop().Location())
		compactness += dist.Value(common.Meters) * dist.Value(common.Meters)
	}
	return compactness
}

// EstimateDeltaValue estimates the change in the objective value when applying
// the given move.
func (l *Cluster) EstimateDeltaValue(
	move SolutionMoveStops,
) float64 {
	score, _ := l.estimateDeltaScore(
		move,
		false,
	)
	return score
}

// EstimateIsViolated estimates whether the given move would violate the
// constraint.
func (l *Cluster) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	score, hint := l.estimateDeltaScore(
		move,
		true,
	)
	return score != 0.0, hint
}

func (l *Cluster) estimateDeltaScore(
	move SolutionMoveStops,
	asConstraint bool,
) (deltaScore float64, stopPositionsHint StopPositionsHint) {
	solution := move.Solution()
	moveImpl := move.(*solutionMoveStopsImpl)
	stopPositions := moveImpl.stopPositions
	deltaScore = 0.0

	for _, stopPosition := range stopPositions {
		vehicle := moveImpl.vehicle()
		if vehicle.IsEmpty() {
			return deltaScore, NoPositionsHint()
		}

		candidate := stopPosition.Stop()

		var c *centroidData
		if asConstraint {
			c = vehicle.Last().ConstraintData(l).(*centroidData)
		} else {
			c = vehicle.Last().ObjectiveData(l).(*centroidData)
		}

		centroid := c.location

		if asConstraint {
			distanceToCentroid := haversineDistance(
				centroid,
				candidate.ModelStop().Location(),
			).Value(common.Meters)

			for _, otherVehicle := range solution.vehicles {
				if otherVehicle.IsEmpty() ||
					otherVehicle.Index() == vehicle.Index() {
					continue
				}
				centroidOtherVehicle := otherVehicle.
					Last().
					ConstraintData(l).(*centroidData).location

				if haversineDistance(
					centroidOtherVehicle,
					candidate.ModelStop().Location(),
				).Value(common.Meters) < distanceToCentroid {
					return 1.0, SkipVehiclePositionsHint()
				}
			}
		} else {
			stops := l.getSolutionStops(vehicle)
			deltaScore += compactness(stops, centroid, candidate, true)
			// we want to compute the difference in compactness, so we need to
			// subtract the compactness of the current route
			deltaScore -= c.compactness
		}
	}
	return deltaScore, NoPositionsHint()
}

func (l *Cluster) getSolutionStops(vehicle SolutionVehicle) []SolutionStop {
	stops := make([]SolutionStop, 0, vehicle.NumberOfStops())
	for _, stop := range vehicle.SolutionStops() {
		if stop.IsFirst() && !l.includeFirst {
			continue
		}
		if stop.IsLast() && !l.includeLast {
			continue
		}
		stops = append(stops, stop)
	}
	return stops
}

// Value returns the value of the objective for the given solution.
func (l *Cluster) Value(solutionStop *Solution) float64 {
	sum := 0.0
	for _, vehicle := range solutionStop.Vehicles() {
		if vehicle.IsEmpty() {
			continue
		}
		sum += vehicle.Last().ObjectiveData(l).(*centroidData).compactness
	}
	return sum
}
