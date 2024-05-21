// © 2019-present nextmv.io inc

package nextroute

import (
	"github.com/nextmv-io/nextroute/common"
)

// UnplanIsland un-plans planUnit and all stops (with their plan-units)
// that are located at the same location in the same vehicle and are within
// the given distance of the stops in planUnit not necessary on the same
// vehicle. The function returns a list of un-planned plan-units.
func UnplanIsland(
	planUnit SolutionPlanUnit,
	distance common.Distance,
) error {
	unplanUnits := SolutionPlanUnits{planUnit}
	for _, solutionStop := range planUnit.(*solutionPlanStopsUnitImpl).solutionStops {
		location := solutionStop.ModelStop().Location()
		stop := solutionStop.Next()
		for location.Equals(stop.ModelStop().Location()) && !stop.IsLast() {
			unplanUnits = append(unplanUnits, solutionStop.PlanStopsUnit())
			stop = stop.Next()
		}
		stop = solutionStop.Previous()
		for location.Equals(stop.ModelStop().Location()) && !stop.IsFirst() {
			unplanUnits = append(unplanUnits, solutionStop.PlanStopsUnit())
			stop = stop.Previous()
		}
		if distance.Value(common.Meters) > 0 {
			closestStops, err := solutionStop.modelStop().closestStops()
			if err != nil {
				return err
			}
			for _, closeModelStop := range closestStops {
				d := haversineDistance(
					solutionStop.ModelStop().Location(),
					closeModelStop.Location()).Value(common.Meters)
				if d <= distance.Value(common.Meters) {
					unplanUnits = append(unplanUnits, solutionStop.PlanStopsUnit())
					break
				}
			}
		}
	}

	for _, unplanUnit := range unplanUnits {
		_, err := unplanUnit.UnPlan()
		if err != nil {
			return err
		}
	}

	return nil
}

// UnplanVehicle un-plans all stops of a vehicle that are not fixed. The
// un-planning is done by calling UnplanIsland for each stop. The distance
// parameter is passed to UnplanIsland. The function returns a list of
// un-planned plan-units.
func UnplanVehicle(
	vehicle SolutionVehicle,
	distance common.Distance,
) error {
	// TODO: this can be optimized a lot, now we calculate everything for
	// each stop, but we can do it once after detaching the stops from the
	// vehicle.
	stops := vehicle.SolutionStops()

	for _, stop := range stops {
		if stop.IsPlanned() && !stop.IsFixed() {
			err := UnplanIsland(stop.PlanStopsUnit(), distance)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
