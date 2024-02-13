package nextroute

import "github.com/nextmv-io/sdk/nextroute"

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
