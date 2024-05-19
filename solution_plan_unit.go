// © 2019-present nextmv.io inc

package nextroute

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
	solutionPlanUnit SolutionPlanUnit,
	solution *solutionImpl,
) SolutionPlanUnit {
	switch p := solutionPlanUnit.(type) {
	case SolutionPlanStopsUnit:
		return copySolutionPlanStopsUnit(p, solution)
	case SolutionPlanUnitsUnit:
		return copySolutionPlanUnitsUnit(p, solution)
	}
	panic("unknown solution plan unit type")
}

func copySolutionPlanStopsUnit(
	solutionPlanUnit SolutionPlanStopsUnit,
	solution *solutionImpl,
) SolutionPlanStopsUnit {
	solutionPlanUnitImpl := solutionPlanUnit.(*solutionPlanStopsUnitImpl)
	copyOfSolutionPlanUnit := &solutionPlanStopsUnitImpl{
		modelPlanStopsUnit: solutionPlanUnitImpl.modelPlanStopsUnit,
		solutionStops:      make([]SolutionStop, len(solutionPlanUnitImpl.solutionStops)),
	}
	for idx, solutionStop := range solutionPlanUnitImpl.solutionStops {
		copyOfSolutionPlanUnit.solutionStops[idx] = SolutionStop{
			index:    solutionStop.Index(),
			solution: solution,
		}
		solution.stopToPlanUnit[solutionStop.Index()] = copyOfSolutionPlanUnit
	}
	return copyOfSolutionPlanUnit
}

func copySolutionPlanUnitsUnit(
	solutionPlanUnit SolutionPlanUnitsUnit,
	solution *solutionImpl,
) SolutionPlanUnitsUnit {
	solutionPlanUnitImpl := solutionPlanUnit.(*solutionPlanUnitsUnitImpl)
	copyOfSolutionPlanUnit := &solutionPlanUnitsUnitImpl{
		modelPlanUnitsUnit: solutionPlanUnitImpl.modelPlanUnitsUnit,
		solutionPlanUnits:  make(SolutionPlanUnits, len(solutionPlanUnitImpl.solutionPlanUnits)),
	}
	for idx, propositionSolutionPlanUnit := range solutionPlanUnitImpl.solutionPlanUnits {
		copyOfSolutionPlanUnit.solutionPlanUnits[idx] =
			copySolutionPlanUnit(propositionSolutionPlanUnit, solution)

		solution.propositionPlanUnits.add(copyOfSolutionPlanUnit.solutionPlanUnits[idx])
	}

	return copyOfSolutionPlanUnit
}
