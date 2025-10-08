// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
)

// NewSolveOperatorUnPlanLocation creates a new SolveOperatorUnPlanLocation.
func NewSolveOperatorUnPlanLocation(
	numberOfUnits SolveParameter,
) (*SolveOperatorUnPlanLocation, error) {
	return &SolveOperatorUnPlanLocation{
		solveOperator: solveOperator{
			probability:            1.0,
			canResultInImprovement: false,
			parameters:             SolveParameters{numberOfUnits},
		},
	}, nil
}

// SolveOperatorUnPlanLocation is a solve-operator which un-plans planned plan-units.
// It is used to remove planned plan-units. from the solution.
// In each iteration of the solve run, the number of plan-units. to un-plan
// is determined by the number of units. The number of units is a
// solve-parameter which can be configured by the user. In each iteration, the
// number of units is sampled from a uniform distribution. The number of units
// is always an integer between 1 and the number of units.
type SolveOperatorUnPlanLocation struct {
	solveOperator
}

func (d *SolveOperatorUnPlanLocation) unplanLocation(
	planUnit SolutionPlanUnit,
) (int, error) {
	count := 0
	unPlanUnits := make(SolutionPlanUnits, 1, 64)
	unPlanUnits[0] = planUnit
	plannedPlanStopsUnits := planUnit.PlannedPlanStopsUnits()
	for _, plannedPlanStopsUnit := range plannedPlanStopsUnits {
		for _, solutionStop := range plannedPlanStopsUnit.(*solutionPlanStopsUnitImpl).solutionStops {
			location := solutionStop.ModelStop().Location()
			stop := solutionStop.Next()
			for location.Equals(stop.ModelStop().Location()) && !stop.IsLast() {
				unPlanUnits = append(unPlanUnits, stop.PlanStopsUnit())
				stop = stop.Next()
			}
			stop = solutionStop.Previous()
			for location.Equals(stop.ModelStop().Location()) && !stop.IsFirst() {
				unPlanUnits = append(unPlanUnits, stop.PlanStopsUnit())
				stop = stop.Previous()
			}
		}
	}
	for _, unPlanUnit := range unPlanUnits {
		unplanned, err := unPlanUnit.UnPlan()
		if err != nil {
			return count, err
		}
		if unplanned {
			count++
		}
	}

	return count, nil
}

// NumberOfUnits returns the number of units to unplan as a solve-parameter.
func (d *SolveOperatorUnPlanLocation) NumberOfUnits() SolveParameter {
	return d.Parameters()[0]
}

// Execute implements the SolveOperator interface.
func (d *SolveOperatorUnPlanLocation) Execute(
	ctx context.Context,
	runTimeInformation SolveInformation,
) error {
	workSolution := runTimeInformation.
		Solver().
		WorkSolution()

	numberOfUnits := d.NumberOfUnits().Value()

	if workSolution.PlannedPlanUnits().Size() == 0 {
		return nil
	}

Loop:
	for i := 0; i < numberOfUnits &&
		workSolution.PlannedPlanUnits().Size() > 0; i++ {
		select {
		case <-ctx.Done():
			break Loop
		default:
			plannedPlanUnit := workSolution.PlannedPlanUnits().RandomElement()
			units, err := d.unplanLocation(plannedPlanUnit)
			if err != nil {
				return err
			}

			i += units
		}
	}
	return nil
}
