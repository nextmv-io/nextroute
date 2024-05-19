// © 2019-present nextmv.io inc

package nextroute

import (
	"context"

	"github.com/nextmv-io/nextroute/common"
)

// SolveOperatorUnPlan is a solve operator that un-plans units.
type SolveOperatorUnPlan interface {
	SolveOperator

	// NumberOfUnits returns the number of units of the solve operator.
	NumberOfUnits() SolveParameter
}

// NewSolveOperatorUnPlan creates a new SolveOperatorUnPlan.
// SolveOperatorUnPlan is a solve-operator which un-plans planned plan-units.
// It is used to remove planned plan-units. from the solution.
// In each iteration of the solve run, the number of plan-units. to un-plan
// is determined by the number of units. The number of units is a
// solve-parameter which can be configured by the user. In each iteration, the
// number of units is sampled from a uniform distribution. The number of units
// is always an integer between 1 and the number of units.
func NewSolveOperatorUnPlan(
	numberOfUnits SolveParameter,
) (SolveOperatorUnPlan, error) {
	return &solveOperatorUnPlanImpl{
		SolveOperator: NewSolveOperator(
			1.0,
			false,
			SolveParameters{numberOfUnits},
		),
	}, nil
}

type solveOperatorUnPlanImpl struct {
	SolveOperator
}

func (d *solveOperatorUnPlanImpl) NumberOfUnits() SolveParameter {
	return d.Parameters()[0]
}

func (d *solveOperatorUnPlanImpl) Execute(
	ctx context.Context,
	runTimeInformation SolveInformation,
) error {
	workSolution := runTimeInformation.
		Solver().
		WorkSolution()

	random := runTimeInformation.Solver().Random()

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
			chance := random.Float64()
			switch {
			case chance < 0.01:
				n, err := d.unplanSomeStopsOfOneVehicle(workSolution, 0.2)
				if err != nil {
					return err
				}
				i += n
			case chance < 0.0132:
				n, err := d.unplanOneIsland(workSolution, numberOfUnits-i)
				if err != nil {
					return err
				}
				i += n
			default:
				plannedPlanUnit := workSolution.PlannedPlanUnits().RandomElement()
				n, err := d.unplanLocation(plannedPlanUnit)
				if err != nil {
					return err
				}

				i += n
			}
		}
	}

	return nil
}

func (d *solveOperatorUnPlanImpl) unplanOneStopIsland(
	solutionStop SolutionStop,
	numberOfStops int,
) (int, error) {
	if !solutionStop.IsPlanned() {
		return 0, nil
	}

	unPlanned, err := solutionStop.
		PlanStopsUnit().
		UnPlan()
	if err != nil {
		return 0, err
	}

	if unPlanned {
		if numberOfStops == 1 {
			return 1, nil
		}
	}

	units, err := d.unplanClosestStops(solutionStop, numberOfStops-1)
	if err != nil {
		return 0, err
	}

	return 1 + len(units), nil
}

func (d *solveOperatorUnPlanImpl) unplanClosestStops(
	solutionStop SolutionStop,
	numberOfStops int,
) (SolutionPlanUnits, error) {
	planUnits := make(SolutionPlanUnits, 0)

	solution := solutionStop.Solution()

	stops, err := solutionStop.ModelStop().(*stopImpl).closestStops()
	if err != nil {
		return planUnits, err
	}
	unplannedCount := 0
	for _, stop := range stops {
		if stop.Index() == solutionStop.ModelStop().Index() ||
			!stop.HasPlanStopsUnit() ||
			!solution.SolutionPlanStopsUnit(stop.PlanStopsUnit()).IsPlanned() ||
			solution.Random().Float64() > 0.5 {
			continue
		}
		unPlanned, err := solution.
			SolutionPlanStopsUnit(stop.PlanStopsUnit()).
			UnPlan()
		if err != nil {
			return planUnits, err
		}
		if unPlanned {
			planUnits = append(planUnits, solutionStop.PlanStopsUnit())
			unplannedCount++
		}

		if unplannedCount > 3 && unplannedCount > numberOfStops {
			break
		}
	}
	return planUnits, nil
}

func (d *solveOperatorUnPlanImpl) unplanOneIsland(
	solution Solution,
	numberOfStops int,
) (int, error) {
	planUnit := solution.PlannedPlanUnits().RandomElement()
	planStopsUnit := common.RandomElement(solution.Random(), planUnit.PlannedPlanStopsUnits())
	for _, solutionStop := range planStopsUnit.(*solutionPlanStopsUnitImpl).solutionStops {
		return d.unplanOneStopIsland(
			solutionStop,
			numberOfStops,
		)
	}

	return 0, nil
}

func (d *solveOperatorUnPlanImpl) unplanSomeStopsOfOneVehicle(
	solution Solution,
	chance float64,
) (int, error) {
	vehicles := common.Filter(solution.Vehicles(), func(vehicle SolutionVehicle) bool {
		return !vehicle.IsEmpty()
	})

	if len(vehicles) == 0 {
		return 0, nil
	}

	if chance == 0 {
		maxStops := 0
		for _, vehicle := range vehicles {
			if vehicle.NumberOfStops() > maxStops {
				maxStops = vehicle.NumberOfStops()
			}
		}
		weights := common.Map(vehicles, func(vehicle SolutionVehicle) float64 {
			return float64(maxStops - vehicle.NumberOfStops() + 1)
		})
		alias, err := common.NewAlias(weights)
		if err != nil {
			return 0, err
		}
		vehicle := vehicles[alias.Sample(solution.Random())]
		nrStops := vehicle.NumberOfStops()
		unplanned, err := vehicle.Unplan()
		if err != nil {
			return 0, err
		}
		if unplanned {
			return nrStops, nil
		}
	}

	vehicle := vehicles[solution.Random().Intn(len(vehicles))]
	stops := vehicle.SolutionStops()
	count := 0

	for _, stop := range stops {
		if stop.IsPlanned() && !stop.IsFixed() {
			if solution.Random().Float64() < chance {
				continue
			}
			unplanned, err := stop.PlanStopsUnit().UnPlan()
			if err != nil {
				return 0, err
			}
			if unplanned {
				count++
			}
		}
	}
	return count, nil
}

func (d *solveOperatorUnPlanImpl) unplanLocation(
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
