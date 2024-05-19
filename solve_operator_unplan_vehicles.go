// © 2019-present nextmv.io inc

package nextroute

import (
	"context"

	"github.com/nextmv-io/nextroute/common"
)

// NewSolveOperatorUnPlanVehicles creates a new SolveOperatorUnPlan. .
// The operator un-plans a number of vehicles. The number of vehicles to
// un-plan is determined by the solve-parameter. The solve-parameter can
// change value during the solve run. For each stop in the vehicle un-planned
// the operator will also unplan the stops within distance of this stop (using
// haversine distance). The probability of un-planning a vehicle is determined
// inversely proportional to the number of stops in the vehicle.
func NewSolveOperatorUnPlanVehicles(
	numberOfVehicles SolveParameter,
	distance common.Distance,
	probability float64,
) SolveOperatorUnPlanVehicles {
	return &solveOperatorUnPlanVehiclesImpl{
		SolveOperator: NewSolveOperator(
			probability,
			false,
			SolveParameters{numberOfVehicles},
		),
		distance: distance,
	}
}

// SolveOperatorUnPlanVehicles is a solve-operator which un-plans all the
// stops of a vehicle.
type SolveOperatorUnPlanVehicles interface {
	SolveOperator

	// Distance returns the distance to use for the un-planning.
	Distance() common.Distance

	// NumberOfVehicles returns the number of vehicles to unplan as a solve-parameter.
	// Solve-parameters can change value during the solve run.
	NumberOfVehicles() SolveParameter
}

type solveOperatorUnPlanVehiclesImpl struct {
	SolveOperator
	distance common.Distance
}

func (d *solveOperatorUnPlanVehiclesImpl) NumberOfVehicles() SolveParameter {
	return d.Parameters()[0]
}

func (d *solveOperatorUnPlanVehiclesImpl) Distance() common.Distance {
	return d.distance
}

func (d *solveOperatorUnPlanVehiclesImpl) Execute(
	ctx context.Context,
	runTimeInformation SolveInformation,
) error {
	workSolution := runTimeInformation.
		Solver().
		WorkSolution().(*solutionImpl)

	if workSolution.PlannedPlanUnits().Size() == 0 {
		return nil
	}

	random := runTimeInformation.Solver().Random()

	numberOfVehicles := d.NumberOfVehicles().Value()

	vehicles := common.Filter(
		workSolution.vehicles,
		func(vehicle SolutionVehicle) bool {
			return vehicle.NumberOfStops() > 0
		},
	)

	if len(vehicles) == 0 {
		return nil
	}
	weights := common.Map(vehicles, func(vehicle SolutionVehicle) float64 {
		return 1.0 + float64(workSolution.Model().NumberOfStops()-vehicle.NumberOfStops())
	})
	alias, err := common.NewAlias(weights)
	if err != nil {
		return err
	}
Loop:
	for i := 0; i < numberOfVehicles; i++ {
		select {
		case <-ctx.Done():
			break Loop
		default:
			err = UnplanVehicle(vehicles[alias.Sample(random)], d.distance)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
