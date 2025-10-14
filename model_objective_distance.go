// © 2019-present nextmv.io inc

package nextroute

// NewDistanceObjective returns a new DistanceObjective.
func NewDistanceObjective() *DistanceObjective {
	return &DistanceObjective{}
}

// DistanceObjective minimizes total traveled distance.
type DistanceObjective struct{}

// distanceObjectiveVehicleData keeps track of the cumulative distance traveled
// by a vehicle.
type distanceObjectiveVehicleData struct {
	cumulativeDistance float64
}

// UpdateObjectiveVehicleData implements the ObjectiveVehicleDataUpdater interface.
func (d *DistanceObjective) UpdateObjectiveVehicleData(s SolutionVehicle) (Copier, error) {
	distance := 0.0
	vehicleType := s.ModelVehicle().vehicleType
	distanceExpr := vehicleType.distance
	var previousStop *ModelStop

	s.iterateStops(func(stop SolutionStop) bool {
		modelStop := stop.ModelStop()
		if previousStop != nil {
			distance += distanceExpr.Value(vehicleType, previousStop, modelStop)
		}
		previousStop = modelStop
		return true
	})
	return distanceObjectiveVehicleData{
		cumulativeDistance: distance,
	}, nil
}

// Copy implements the Copier interface.
func (d distanceObjectiveVehicleData) Copy() Copier {
	return distanceObjectiveVehicleData{
		cumulativeDistance: d.cumulativeDistance,
	}
}

// EstimateDeltaValue implements the ModelObjective interface.
func (d *DistanceObjective) EstimateDeltaValue(move SolutionMoveStops) float64 {
	impl := move.(*solutionMoveStopsImpl)
	vehicle := impl.vehicle()
	vehicleType := vehicle.ModelVehicle().VehicleType()
	distanceExpr := vehicleType.DistanceExpression()

	delta := 0.0
	for _, pos := range impl.stopPositions {
		modelStop := pos.Stop().ModelStop()
		previous := pos.Previous()

		// We always add the distance from the previous stop to this one
		// (independent of whether the previous stop is planned or not).
		delta += distanceExpr.Value(vehicleType, previous.ModelStop(), modelStop)

		// If the previous stop is planned, remove the distance from it to its
		// original successor. The connection has been replaced by going through
		// the new stop instead.
		if previous.IsPlanned() {
			delta -= distanceExpr.Value(vehicleType, previous.ModelStop(), previous.Next().ModelStop())
		}

		// If the next stop is planned we need to add the distance from the new
		// stop to it too to complete the triangle of the new connection.
		successor := pos.Next()
		if successor.IsPlanned() {
			delta += distanceExpr.Value(vehicleType, modelStop, successor.ModelStop())
		}
	}

	return delta
}

// Value implements the ModelObjective interface.
func (d *DistanceObjective) Value(solution *Solution) float64 {
	total := 0.0
	for _, v := range solution.vehicles {
		data := v.ObjectiveData(d).(distanceObjectiveVehicleData)
		total += data.cumulativeDistance
	}
	return total
}

// String returns the string representation of the distance objective.
func (d *DistanceObjective) String() string {
	return "distance"
}
