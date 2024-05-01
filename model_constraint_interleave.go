// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"github.com/nextmv-io/nextroute/common"
)

// InterleaveConstraint is a constraint that disallows certain target to be
// interleaved.
type InterleaveConstraint interface {
	ModelConstraint
	// DisallowInterleaving disallows the given planUnits to be interleaved.
	DisallowInterleaving(source ModelPlanUnit, targets []ModelPlanUnit) error

	DisallowedInterleaves() []DisallowedInterleave

	// SourceDisallowedInterleaves returns the disallowed interleaves for the
	// given source.
	SourceDisallowedInterleaves(source ModelPlanUnit) []DisallowedInterleave

	// TargetDisallowedInterleaves returns the disallowed interleaves for the
	// given target.
	TargetDisallowedInterleaves(target ModelPlanUnit) []DisallowedInterleave
}

// NewInterleaveConstraint returns a new InterleaveConstraint.
func NewInterleaveConstraint() (InterleaveConstraint, error) {
	return &interleaveConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"interleave",
			ModelExpressions{},
		),
		disallowedInterleaves:       make([]DisallowedInterleave, 0),
		sourceDisallowedInterleaves: nil,
		targetDisallowedInterleaves: nil,
	}, nil
}

// DisallowedInterleave is a disallowed interleave between a target and a set of
// sources.
type DisallowedInterleave interface {
	// Target returns the target plan unit. This plan unit cannot be interleaved
	// with the sources.
	Target() ModelPlanUnit

	// Sources returns the sources that cannot be interleaved with the target.
	Sources() []ModelPlanUnit
}

type disallowedInterleaveImpl struct {
	target  ModelPlanUnit
	sources []ModelPlanUnit
}

func (d *disallowedInterleaveImpl) Target() ModelPlanUnit {
	return d.target
}

func (d *disallowedInterleaveImpl) Sources() []ModelPlanUnit {
	return d.sources
}

func newDisallowedInterleave(
	target ModelPlanUnit,
	sources []ModelPlanUnit,
) DisallowedInterleave {
	return &disallowedInterleaveImpl{
		target:  target,
		sources: common.DefensiveCopy(sources),
	}
}

type interleaveConstraintImpl struct {
	modelConstraintImpl
	disallowedInterleaves       []DisallowedInterleave
	sourceDisallowedInterleaves map[ModelPlanUnit][]DisallowedInterleave
	targetDisallowedInterleaves map[ModelPlanUnit][]DisallowedInterleave
}

func (l *interleaveConstraintImpl) SourceDisallowedInterleaves(source ModelPlanUnit) []DisallowedInterleave {
	if l.sourceDisallowedInterleaves != nil {
		if disallowedInterleaves, ok := l.sourceDisallowedInterleaves[source]; ok {
			return disallowedInterleaves
		}
		return []DisallowedInterleave{}
	}

	found := make([]DisallowedInterleave, 0)
	for _, disallowedInterleave := range l.disallowedInterleaves {
		for _, source := range disallowedInterleave.Sources() {
			if source == source {
				found = append(found, disallowedInterleave)
			}
		}
	}
	return found
}

func (l *interleaveConstraintImpl) TargetDisallowedInterleaves(target ModelPlanUnit) []DisallowedInterleave {
	if l.targetDisallowedInterleaves != nil {
		if disallowedInterleaves, ok := l.targetDisallowedInterleaves[target]; ok {
			return disallowedInterleaves
		}
		return []DisallowedInterleave{}
	}

	found := make([]DisallowedInterleave, 0)
	for _, disallowedInterleave := range l.disallowedInterleaves {
		if disallowedInterleave.Target() == target {
			found = append(found, disallowedInterleave)
		}
	}
	return found
}

func (l *interleaveConstraintImpl) DisallowedInterleaves() []DisallowedInterleave {
	return l.disallowedInterleaves
}

func addToMap(planUnit ModelPlanUnit, mapUnit map[ModelPlanUnit][]DisallowedInterleave, disallowedInterleave DisallowedInterleave) {
	if modelPlanUnitsUnit, ok := planUnit.(ModelPlanUnitsUnit); ok {
		for _, pu := range modelPlanUnitsUnit.PlanUnits() {
			addToMap(pu, mapUnit, disallowedInterleave)
		}
		return
	}

	if _, ok := mapUnit[planUnit]; !ok {
		mapUnit[planUnit] = []DisallowedInterleave{}
	}

	disallowedInterleaves := mapUnit[planUnit]
	disallowedInterleaves = append(disallowedInterleaves, disallowedInterleave)
	mapUnit[planUnit] = disallowedInterleaves
}

func (l *interleaveConstraintImpl) Lock(model Model) error {
	l.sourceDisallowedInterleaves = make(map[ModelPlanUnit][]DisallowedInterleave)
	l.targetDisallowedInterleaves = make(map[ModelPlanUnit][]DisallowedInterleave)

	for _, disallowedInterleave := range l.disallowedInterleaves {
		addToMap(disallowedInterleave.Target(), l.targetDisallowedInterleaves, disallowedInterleave)

		for _, source := range disallowedInterleave.Sources() {
			addToMap(source, l.sourceDisallowedInterleaves, disallowedInterleave)
		}
	}
	return nil
}

func verifyPlanUnitAllOnSameVehicle(planUnit ModelPlanUnit, preFix string) error {
	if modelPlanUnitsUnit, isModelPlanUnitsUnit := planUnit.(ModelPlanUnitsUnit); isModelPlanUnitsUnit {
		if modelPlanUnitsUnit.PlanAll() && !modelPlanUnitsUnit.SameVehicle() {
			return fmt.Errorf(
				"%s, all plan units in a conjunction must be on the same vehicle",
				preFix,
			)
		}
		for _, planUnit := range modelPlanUnitsUnit.PlanUnits() {
			if err := verifyPlanUnitAllOnSameVehicle(planUnit, preFix); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *interleaveConstraintImpl) DisallowInterleaving(target ModelPlanUnit, sources []ModelPlanUnit) error {
	if target == nil {
		return fmt.Errorf("source cannot be nil")
	}

	// TODO: cover all cases
	if modelPlanStopsUnit, ok := target.(ModelPlanStopsUnit); ok {
		if modelPlanStopsUnit.Stops()[0].Model().IsLocked() {
			return fmt.Errorf(lockErrorMessage, "DisallowInterleaving")
		}
	}

	if sources == nil {
		return fmt.Errorf("sources cannot be nil")
	}
	if len(sources) == 0 {
		return nil
	}
	for idx, source := range sources {
		if source == nil {
			return fmt.Errorf("source[%v] cannot be nil", idx)
		}
		if source == target {
			return fmt.Errorf("target is also in a source")
		}
	}
	uniqueSources := common.UniqueDefined(sources, func(t ModelPlanUnit) int {
		return t.Index()
	})
	if len(uniqueSources) != len(sources) {
		return fmt.Errorf("sources cannot have duplicate plan units")
	}
	// check the type of planUnit and
	err := verifyPlanUnitAllOnSameVehicle(target, "target")
	if err != nil {
		return err
	}
	for idx, unit := range sources {
		err = verifyPlanUnitAllOnSameVehicle(unit, fmt.Sprintf("sources[%v]", idx))
		if err != nil {
			return err
		}
	}

	index := common.FindIndex(l.disallowedInterleaves, func(disallowedInterleave DisallowedInterleave) bool {
		if disallowedInterleave.Target() == target {
			return true
		}
		return false
	})
	if index < 0 {
		l.disallowedInterleaves = append(l.disallowedInterleaves, newDisallowedInterleave(target, sources))
	} else {
		l.disallowedInterleaves[index].(*disallowedInterleaveImpl).sources = append(
			l.disallowedInterleaves[index].(*disallowedInterleaveImpl).sources,
			sources...,
		)
		l.disallowedInterleaves[index].(*disallowedInterleaveImpl).sources = common.UniqueDefined(
			l.disallowedInterleaves[index].(*disallowedInterleaveImpl).sources,
			func(t ModelPlanUnit) int {
				return t.Index()
			},
		)
	}

	return nil
}

func (l *interleaveConstraintImpl) String() string {
	return l.name
}

func (l *interleaveConstraintImpl) EstimationCost() Cost {
	return LinearStop
}

//func nameMove(move Move) string {
//	result := ""
//	for idx, stopPosition := range move.StopPositions() {
//		if idx > 0 {
//			result += " .. "
//		}
//		result += stopPosition.Previous().ModelStop().ID() + " - ["
//		result += stopPosition.Stop().ModelStop().ID() + "] - "
//		result += stopPosition.Next().ModelStop().ID()
//	}
//	return result
//}
//
//func namePlanStopsUnit(planStopsUnit ModelPlanStopsUnit) string {
//	result := "["
//	for idx, stop := range planStopsUnit.Stops() {
//		if idx > 0 {
//			result += ", "
//		}
//		result += stop.ID()
//	}
//	result += "]"
//	return result
//}

func (l *interleaveConstraintImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	solution := move.Solution()
	//fmt.Println("\U0001FAE3 Move: ",
	//	nameMove(move),
	//	"for unit",
	//	namePlanStopsUnit(move.PlanStopsUnit().ModelPlanUnit().(ModelPlanStopsUnit)),
	//)

	// FirstPlanUnit and LastPlanUnit are the first and last stops of the plan unit of
	// the move. In case the move planUnit is a part of a PlanUnitsUnit, the first and last
	// stops should include the already planned stops of the other plan units of the PlanUnitsUnit.
	firstPlanUnit := move.Previous()
	lastPlanUnit := move.Next()

	if modelPlanUnitsUnit, hasModelPlanUnitsUnit := move.PlanStopsUnit().ModelPlanUnit().PlanUnitsUnit(); hasModelPlanUnitsUnit {
		// fmt.Println("  ☐ Move is for a plan unit that is part of a PlanUnitsUnit")
		// first and last stop of the already planned plan-units of PlanUnitsUnit
		var first, last SolutionStop
		for _, planUnit := range modelPlanUnitsUnit.PlanUnits() {
			if planUnit.Index() == move.PlanStopsUnit().ModelPlanUnit().Index() {
				continue
			}
			solutionPlanUnit := solution.SolutionPlanUnit(planUnit)
			if solutionPlanUnit.IsPlanned() {
				for _, solutionPlanStopsUnit := range solutionPlanUnit.PlannedPlanStopsUnits() {
					if solutionPlanStopsUnit.SolutionStops()[0].Vehicle() != move.Vehicle() {
						continue
					}
					// fmt.Println("    ☐ Getting first and last of planned unit",
					//	namePlanStopsUnit(solutionPlanStopsUnit.ModelPlanStopsUnit()),
					// )
					first, last = determineFirstLastSolutionStops(first, last, solutionPlanStopsUnit)
					if first != nil &&
						move.Previous().Position() >= first.Position() &&
						move.Next().Position() <= last.Position() {
						// fmt.Println("    ✅ Move is in the middle of already planned plan units of PlanUnitsUnit")
						return false, noPositionsHint()
					}
				}
			}
		}
		if first != nil && first.Position() <= firstPlanUnit.Position() {
			firstPlanUnit = first
		}
		if last != nil && last.Position() >= lastPlanUnit.Position() {
			lastPlanUnit = last
		}
	}

	// check if the target is disallowed to be interleaved with the source
	if targetDisallowedInterleaves, isTargetPlanUnit := l.targetDisallowedInterleaves[move.PlanStopsUnit().ModelPlanUnit()]; isTargetPlanUnit {
		// fmt.Println("  ☐ Move is for a plan unit that is a target for a source")
		var first, last SolutionStop

		for _, disallowedInterleave := range targetDisallowedInterleaves {
			for _, sourcePlanUnit := range disallowedInterleave.Sources() {
				sourceSolutionPlanUnit := move.Solution().SolutionPlanUnit(sourcePlanUnit)
				if sourceSolutionPlanUnit.IsPlanned() {
					solutionPlanStopsUnits := sourceSolutionPlanUnit.PlannedPlanStopsUnits()
					for _, solutionPlanStopsUnit := range solutionPlanStopsUnits {
						if solutionPlanStopsUnit.SolutionStops()[0].Vehicle() != move.Vehicle() {
							continue
						}
						// fmt.Println("    ☐ Planned source",
						//	namePlanStopsUnit(solutionPlanStopsUnit.ModelPlanStopsUnit()),
						//	"can not interleave with target",
						//	namePlanStopsUnit(move.PlanStopsUnit().ModelPlanUnit().(ModelPlanStopsUnit)),
						// )
						first, last = determineFirstLastSolutionStops(first, last, solutionPlanStopsUnit)
					}
				}
			}
		}

		if first != nil {
			if lastPlanUnit.Position() <= first.Position() || firstPlanUnit.Position() >= last.Position() {
				// fmt.Println("    ✅ Sources do not overlap with to be planned target")
				return false, noPositionsHint()
			}
		}

		// fmt.Println("    ❌ A source would be interleaved  with to be planned target")
		return true, noPositionsHint()
	}
	// fmt.Println("✅ No violations for move")
	return false, noPositionsHint()
}

// determineFirstLastSolutionStops determines the first and last solution stops
// of the given solution plan stops unit.
func determineFirstLastSolutionStops(
	first, last SolutionStop,
	solutionPlanStopUnit SolutionPlanStopsUnit,
) (SolutionStop, SolutionStop) {
	if !solutionPlanStopUnit.IsPlanned() {
		return nil, nil
	}
	solutionStops := solutionPlanStopUnit.SolutionStops()
	f := solutionStops[0]
	if f != nil && (first == nil || f.Position() < first.Position()) {
		first = f
	}
	l := solutionStops[len(solutionStops)-1]
	if l != nil && (last == nil || l.Position() > last.Position()) {
		last = l
	}
	return first, last
}
