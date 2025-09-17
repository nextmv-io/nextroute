// © 2019-present nextmv.io inc

package nextroute

// DistanceObjective minimizes total traveled distance.
type DistanceObjective interface {
	ModelObjective
}

// NewDistanceObjective returns a new DistanceObjective.
func NewDistanceObjective() DistanceObjective {
	return &distanceObjectiveImpl{}
}

type distanceObjectiveImpl struct{}

// distanceObjectiveVehicleData keeps track of the cumulative distance traveled
// by a vehicle.
type distanceObjectiveVehicleData struct {
	cumulativeDistance float64
}

func (d *distanceObjectiveImpl) UpdateObjectiveVehicleData(s SolutionVehicle) (Copier, error) {
	distance := 0.0
	vehicleType := s.ModelVehicle().(*modelVehicleImpl).vehicleType
	distanceExpr := vehicleType.(*vehicleTypeImpl).distance
	var previousStop ModelStop
	for _, stop := range s.SolutionStops() {
		modelStop := stop.ModelStop()
		if previousStop != nil {
			distance += distanceExpr.Value(vehicleType, previousStop, modelStop)
		}
		previousStop = modelStop
	}
	return &distanceObjectiveVehicleData{
		cumulativeDistance: distance,
	}, nil
}

func (d *distanceObjectiveVehicleData) Copy() Copier {
	return &distanceObjectiveVehicleData{
		cumulativeDistance: d.cumulativeDistance,
	}
}

func (d *distanceObjectiveImpl) EstimateDeltaValue(move SolutionMoveStops) float64 {
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

func (d *distanceObjectiveImpl) Value(solution Solution) float64 {
	s := solution.(*solutionImpl)
	total := 0.0
	for _, v := range s.vehicles {
		data := v.ObjectiveData(d).(*distanceObjectiveVehicleData)
		total += data.cumulativeDistance
	}
	return total
}

func (d *distanceObjectiveImpl) String() string {
	return "distance"
}
