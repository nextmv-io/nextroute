package nextroute

import "github.com/nextmv-io/sdk/nextroute"

// SolutionPlanUnit is a set of stops that are planned to be visited by
// a vehicle.
type SolutionPlanUnit interface {
	// IsFixed returns true if any of stops are fixed.
	IsFixed() bool

	// IsPlanned returns true if all the stops are planned.
	IsPlanned() bool

	// ModelPlanUnit returns the model plan unit associated with the
	// solution plan unit.
	ModelPlanUnit() ModelPlanUnit

	// PlannedPlanStopsUnits returns the plan stops units associated with the
	// invoking plan unit which are planned.
	PlannedPlanStopsUnits() SolutionPlanStopsUnits

	// Solution returns the solution this unit is part of.
	Solution() Solution

	// UnPlan un-plans the unit by removing the underlying solution stops
	// from the solution. Returns true if the unit was unplanned
	// successfully, false if the unit was not unplanned successfully. A
	// unit is not successful if it did not result in a change in the
	// solution without violating any hard constraints.
	UnPlan() (bool, error)
}

// SolutionPlanUnits is a slice of [SolutionPlanUnit].
type SolutionPlanUnits []SolutionPlanUnit

func copySolutionPlanUnit(
	solutionPlanUnit nextroute.SolutionPlanUnit,
	solution *solutionImpl,
) nextroute.SolutionPlanUnit {
	switch p := solutionPlanUnit.(type) {
	case nextroute.SolutionPlanStopsUnit:
		return copySolutionPlanStopsUnit(p, solution)
	case nextroute.SolutionPlanUnitsUnit:
		return copySolutionPlanUnitsUnit(p, solution)
	}
	panic("unknown solution plan unit type")
}

func copySolutionPlanStopsUnit(
	solutionPlanUnit nextroute.SolutionPlanStopsUnit,
	solution *solutionImpl,
) nextroute.SolutionPlanStopsUnit {
	solutionPlanUnitImpl := solutionPlanUnit.(*solutionPlanStopsUnitImpl)
	copyOfSolutionPlanUnit := &solutionPlanStopsUnitImpl{
		modelPlanStopsUnit: solutionPlanUnitImpl.modelPlanStopsUnit,
		solutionStops:      make([]solutionStopImpl, len(solutionPlanUnitImpl.solutionStops)),
	}
	for idx, solutionStop := range solutionPlanUnitImpl.solutionStops {
		copyOfSolutionPlanUnit.solutionStops[idx] = solutionStopImpl{
			index:    solutionStop.Index(),
			solution: solution,
		}
		solution.stopToPlanUnit[solutionStop.Index()] = copyOfSolutionPlanUnit
	}
	return copyOfSolutionPlanUnit
}

func copySolutionPlanUnitsUnit(
	solutionPlanUnit nextroute.SolutionPlanUnitsUnit,
	solution *solutionImpl,
) nextroute.SolutionPlanUnitsUnit {
	solutionPlanUnitImpl := solutionPlanUnit.(*solutionPlanUnitsUnitImpl)
	copyOfSolutionPlanUnit := &solutionPlanUnitsUnitImpl{
		modelPlanUnitsUnit: solutionPlanUnitImpl.modelPlanUnitsUnit,
		solutionPlanUnits:  make(nextroute.SolutionPlanUnits, len(solutionPlanUnitImpl.solutionPlanUnits)),
	}
	for idx, propositionSolutionPlanUnit := range solutionPlanUnitImpl.solutionPlanUnits {
		copyOfSolutionPlanUnit.solutionPlanUnits[idx] =
			copySolutionPlanUnit(propositionSolutionPlanUnit, solution)

		solution.propositionPlanUnits.add(copyOfSolutionPlanUnit.solutionPlanUnits[idx])
	}

	return copyOfSolutionPlanUnit
}
