package nextroute

type SolutionConstraintInterleaved interface {
	InterleaveConstraint() InterleaveConstraint
}

type plannedInformation struct {
	first SolutionStop
	last  SolutionStop
}

// SolutionConstraintInterleaved is a data structure attached to a solution that
// contains the information about the interleaved constraint.
type solutionConstraintInterleavedImpl struct {
	constraint                  InterleaveConstraint
	sourceDisallowedInterleaves map[ModelPlanUnit]DisallowedInterleave
	targetDisallowedInterleaves map[ModelPlanUnit]DisallowedInterleave
	plannedInformation          map[ModelPlanUnit]plannedInformation
}

func (s *solutionConstraintInterleavedImpl) InterleaveConstraint() InterleaveConstraint {
	return s.constraint
}

func newSolutionConstraintInterleaved(constraint InterleaveConstraint) SolutionConstraintInterleaved {
	impl := &solutionConstraintInterleavedImpl{
		constraint:                  constraint,
		sourceDisallowedInterleaves: make(map[ModelPlanUnit]DisallowedInterleave),
		targetDisallowedInterleaves: make(map[ModelPlanUnit]DisallowedInterleave),
		plannedInformation:          make(map[ModelPlanUnit]plannedInformation),
	}
	for _, disallowedInterleave := range constraint.DisallowedInterleaves() {
		impl.sourceDisallowedInterleaves[disallowedInterleave.Target()] = disallowedInterleave
		impl.plannedInformation[disallowedInterleave.Target()] = plannedInformation{}
		for _, source := range disallowedInterleave.Sources() {
			impl.targetDisallowedInterleaves[source] = disallowedInterleave
			impl.plannedInformation[source] = plannedInformation{
				first: nil,
				last:  nil,
			}
		}
	}
	return impl
}

// I want to add a stop before stop, is this allowed
// by the interleaved constraint?
func (s *solutionConstraintInterleavedImpl) disallowedSuccessors(
	stop SolutionStop,
	beforeStop SolutionStop,
) (bool, error) {
	// DisallowInterleaving({S1, S2}, {{G1, G2, G3},{K1,K2}})
	// F - G1 - G2 - A - B - G3 - L
	// F - G1 - G2 - S1? - B - A - G3 - L
	// ask can you tell me something about S1, it tells me something where the other sources are

	// F - S1 - G1 - G2 - B - A - G3 - S2? - L

	// F - S1 - G1 - G2 - S1 - B - A - G3 - L
	// F - K1? - S1 - G1 - G2 - B - A - G3 - L
	// F - K1 - S1 - K2? - G1 - G2 - B - A - G3 - L

	if !beforeStop.IsPlanned() {
		return false, nil
	}

	// we need to know if stop is a target
	modelStop := stop.ModelStop()
	if modelStop.HasPlanStopsUnit() {
		var stopPlanUnit ModelPlanUnit = modelStop.PlanStopsUnit()

		if planUnitsUnit, hasPlanUnitsUnit := stopPlanUnit.PlanUnitsUnit(); hasPlanUnitsUnit {
			stopPlanUnit = planUnitsUnit
		}

		// source side check
		//if disallowedInterleaves, ok := s.sourceDisallowedInterleaves[stopPlanUnit]; ok {
		//	stopPlanUnitInformation := s.plannedInformation[stopPlanUnit]
		//	if stopPlanUnitInformation.first != nil {
		//		if stopPlanUnitInformation.first.Position() <= beforeStop.Position() &&
		//			stopPlanUnitInformation.last.Position() >= beforeStop.Position() {
		//			return false, nil
		//		}
		//	}
		//	for _, source := range disallowedInterleaves.Sources() {
		//		solutionPlanUnit := solution.SolutionPlanUnit(source)
		//
		//	}
		//}
	}
	// it is possible that this stopPlanUnit is part of a planUnitsUnit
	// we need to know if any other planUnit of the planUnitsUnit is planned
	// if so, we need to check if the other planUnit is a target of the disallowedInterleave

	//beforeStopPlanUnit := beforeStop.PlanStopsUnit()

	return false, nil
}
