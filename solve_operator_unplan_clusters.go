// © 2019-present nextmv.io inc

package nextroute

import (
	"context"

	"github.com/nextmv-io/nextroute/common"
)

// SolveOperatorUnPlanUnits is a solve-operator which un-plans all the
// stops of a vehicle.
type SolveOperatorUnPlanUnits struct {
	solveOperator
	distance common.Distance
}

// NewSolveOperatorUnPlanUnits creates a new SolveOperatorUnPlanUnits.
// The operator un-plans a number of units. The number of units to
// un-plan is determined by the solve-parameter. The solve-parameter can
// change value during the solve run. For each stop in the unit un-planned
// the operator will also unplan the stops within distance of this stop (using
// haversine distance).
func NewSolveOperatorUnPlanUnits(
	numberOfUnits SolveParameter,
	distance common.Distance,
	probability float64,
) *SolveOperatorUnPlanUnits {
	return &SolveOperatorUnPlanUnits{
		solveOperator: solveOperator{
			probability:            probability,
			canResultInImprovement: false,
			parameters:             SolveParameters{numberOfUnits},
		},
		distance: distance,
	}
}

// Distance returns the distance to use for the un-planning.
func (d *SolveOperatorUnPlanUnits) Distance() common.Distance {
	return d.distance
}

// NumberOfUnits returns the number of units to unplan as a solve-parameter.
// Solve-parameters can change value during the solve run.
func (d *SolveOperatorUnPlanUnits) NumberOfUnits() SolveParameter {
	return d.Parameters()[0]
}

// Execute implements the SolveOperator interface.
func (d *SolveOperatorUnPlanUnits) Execute(
	ctx context.Context,
	runTimeInformation SolveInformation,
) error {
	numberOfUnits := d.NumberOfUnits().Value()

	if numberOfUnits == 0 {
		return nil
	}

	workSolution := runTimeInformation.
		Solver().
		WorkSolution()
Loop:
	for i := 0; i < numberOfUnits &&
		workSolution.PlannedPlanUnits().Size() > 0; i++ {
		select {
		case <-ctx.Done():
			break Loop
		default:
			plannedPlanUnit := workSolution.PlannedPlanUnits().RandomElement()

			err := UnplanIsland(plannedPlanUnit, d.Distance())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
