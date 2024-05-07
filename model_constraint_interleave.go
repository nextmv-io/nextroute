// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"github.com/nextmv-io/nextroute/common"
	"math"
)

// InterleaveConstraint is a constraint that disallows certain target to be
// interleaved.
type InterleaveConstraint interface {
	ModelConstraint
	// DisallowInterleaving disallows the given planUnits to be interleaved.
	DisallowInterleaving(source ModelPlanUnit, targets []ModelPlanUnit) error

	// DisallowedInterleaves returns the disallowed interleaves.
	DisallowedInterleaves() []DisallowedInterleave

	// SourceDisallowedInterleaves returns the disallowed interleaves for the
	// given source.
	SourceDisallowedInterleaves(source ModelPlanUnit) []DisallowedInterleave

	// TargetDisallowedInterleaves returns the disallowed interleaves for the
	// given target.
	TargetDisallowedInterleaves(target ModelPlanUnit) []DisallowedInterleave

	// UpdateConstraintStopData is called when a stop has been added to a
	// solution for each stop in a vehicle after the added stop.
	UpdateConstraintStopData(s SolutionStop) (Copier, error)
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

type interleaveSolutionStopSpan struct {
	first int
	last  int
}

type interleaveConstraintData struct {
	solutionPlanStopUnits map[int]interleaveSolutionStopSpan
}

func (i *interleaveConstraintData) Copy() Copier {
	solutionPlanStopUnits := make(map[int]interleaveSolutionStopSpan, len(i.solutionPlanStopUnits))

	for k := range i.solutionPlanStopUnits {
		solutionPlanStopUnits[k] = interleaveSolutionStopSpan{
			first: i.solutionPlanStopUnits[k].first,
			last:  i.solutionPlanStopUnits[k].last,
		}
	}

	return &interleaveConstraintData{
		solutionPlanStopUnits: solutionPlanStopUnits,
	}
}

func (l *interleaveConstraintImpl) UpdateConstraintStopData(s SolutionStop) (Copier, error) {
	if s.IsLast() {
		data := interleaveConstraintData{
			solutionPlanStopUnits: make(map[int]interleaveSolutionStopSpan),
		}

		modelPlanStopsUnits := map[SolutionPlanStopsUnit]struct{}{}
		stop := s.Vehicle().First().Next()
		for !stop.IsLast() {

			planStopsUnit := stop.PlanStopsUnit()

			stop = stop.Next()

			if _, ok := modelPlanStopsUnits[planStopsUnit]; ok {
				continue
			}

			if modelPlanUnitsUnit, ok := planStopsUnit.ModelPlanStopsUnit().PlanUnitsUnit(); ok {
				modelPlanStopsUnits[planStopsUnit] = struct{}{}

				planUnitsUnitFirst := math.MaxInt64
				planUnitsUnitLast := math.MinInt64

				for _, planUnit := range modelPlanUnitsUnit.PlanUnits() {
					if modelPlanStopsUnit, ok := planUnit.(ModelPlanStopsUnit); ok {
						solutionPlanStopsUnit := s.Solution().SolutionPlanStopsUnit(modelPlanStopsUnit).(SolutionPlanStopsUnit)
						if solutionPlanStopsUnit.IsPlanned() {
							first := solutionPlanStopsUnit.SolutionStops()[0].Position()
							last := solutionPlanStopsUnit.SolutionStops()[len(solutionPlanStopsUnit.SolutionStops())-1].Position()

							data.solutionPlanStopUnits[modelPlanStopsUnit.Index()] = interleaveSolutionStopSpan{
								first: first,
								last:  last,
							}

							if first < planUnitsUnitFirst {
								planUnitsUnitFirst = first
							}
							if last > planUnitsUnitLast {
								planUnitsUnitLast = last
							}
						}
					}
				}
				data.solutionPlanStopUnits[modelPlanUnitsUnit.Index()] = interleaveSolutionStopSpan{
					first: planUnitsUnitFirst,
					last:  planUnitsUnitLast,
				}
			}

		}
		return &data, nil
	}
	return nil, nil
}

func (l *interleaveConstraintImpl) SourceDisallowedInterleaves(
	source ModelPlanUnit,
) []DisallowedInterleave {
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

func (l *interleaveConstraintImpl) TargetDisallowedInterleaves(
	target ModelPlanUnit,
) []DisallowedInterleave {
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

func addToMap(
	planUnit ModelPlanUnit,
	mapUnit map[ModelPlanUnit][]DisallowedInterleave,
	disallowedInterleave DisallowedInterleave,
) {
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
		return fmt.Errorf("target cannot be nil")
	}

	if sources == nil {
		return fmt.Errorf("sources cannot be nil")
	}

	if len(sources) == 0 {
		return nil
	}

	if _, hasPlanUnitsUnit := target.PlanUnitsUnit(); hasPlanUnitsUnit {
		return fmt.Errorf("target cannot be a plan unit part of a PlanUnitsUnit")
	}

	if modelPlanStopsUnit, ok := target.(ModelPlanStopsUnit); ok {
		if modelPlanStopsUnit.Stops()[0].Model().IsLocked() {
			return fmt.Errorf(lockErrorMessage, "DisallowInterleaving")
		}
	}

	for idx, source := range sources {
		if source == nil {
			return fmt.Errorf("source[%v] cannot be nil", idx)
		}
		if _, hasPlanUnitsUnit := source.PlanUnitsUnit(); hasPlanUnitsUnit {
			return fmt.Errorf("source at index %v cannot be a plan unit part of a PlanUnitsUnit", idx)
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

func isViolatedPositions(sourceFirstPosition, sourceLastPosition, targetFirstPosition, targetLastPosition int) bool {
	//        S===S
	//     T=========T
	if sourceFirstPosition > targetFirstPosition &&
		sourceLastPosition < targetLastPosition {
		return true
	}

	//   S=====S
	//     T=========T
	if sourceFirstPosition < targetFirstPosition &&
		sourceLastPosition > targetFirstPosition &&
		sourceLastPosition < targetLastPosition {
		return true
	}
	//            S=====S
	//     T=========T
	if sourceFirstPosition > targetFirstPosition &&
		sourceFirstPosition < targetLastPosition &&
		sourceLastPosition > targetLastPosition {
		return true
	}

	return false
}
func (l *interleaveConstraintImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	solution := move.Solution()

	moveImpl := move.(*solutionMoveStopsImpl)
	vehicle := moveImpl.vehicle()
	data := vehicle.last().ConstraintData(l).(*interleaveConstraintData)

	solutionMoveStops := move.(*solutionMoveStopsImpl)

	generator := newSolutionStopGenerator(*solutionMoveStops, true, true)
	defer generator.release()

	newPositions := make(map[SolutionStop]int)
	oldPositions := make([]SolutionStop, 0, vehicle.NumberOfStops()+2)

	position := 0
	for solutionStop, ok := generator.next(); ok; solutionStop, ok = generator.next() {
		newPositions[solutionStop] = position
		if solutionStop.IsPlanned() {
			oldPositions = append(oldPositions, solutionStop)
		}
		position += 1
	}

	newPlanUnitSpanFirstPosition := move.Previous().Position() + 1
	newPlanUnitSpanLastPosition := move.Next().Position() + len(move.PlanStopsUnit().SolutionStops()) - 1

	// Check if the plan unit we are moving is a plan units unit, if it is we
	// need to check if the plan units that are already are planned of this unit
	// spans the new positions of the plan unit we are moving, if so we are good
	if modelPlanUnitsUnit, hasModelPlanUnitsUnit :=
		move.PlanStopsUnit().ModelPlanUnit().PlanUnitsUnit(); hasModelPlanUnitsUnit {

		if _, ok := data.solutionPlanStopUnits[modelPlanUnitsUnit.Index()]; ok {
			firstSolutionStop := oldPositions[data.solutionPlanStopUnits[modelPlanUnitsUnit.Index()].first]
			if newPositions[firstSolutionStop] < newPlanUnitSpanFirstPosition {
				newPlanUnitSpanFirstPosition = newPositions[firstSolutionStop]
			}

			lastSolutionStop := oldPositions[data.solutionPlanStopUnits[modelPlanUnitsUnit.Index()].last]
			if newPositions[lastSolutionStop] > newPlanUnitSpanLastPosition {
				newPlanUnitSpanLastPosition = newPositions[lastSolutionStop]
			}
		}
	}

	// Check if the plan unit we are moving is a target
	if targetDisallowedInterleaves, isTargetPlanUnit :=
		l.targetDisallowedInterleaves[move.PlanStopsUnit().ModelPlanUnit()]; isTargetPlanUnit {
		for _, disallowedInterleave := range targetDisallowedInterleaves {
			for _, sourcePlanUnit := range disallowedInterleave.Sources() {
				sourceSolutionPlanUnit := move.Solution().SolutionPlanUnit(sourcePlanUnit)
				if sourceSolutionPlanUnit.IsPlanned() {
					var sourceSpanFirst, sourceSpanLast SolutionStop
					solutionPlanStopsUnits := sourceSolutionPlanUnit.PlannedPlanStopsUnits()
					for _, solutionPlanStopsUnit := range solutionPlanStopsUnits {
						if solutionPlanStopsUnit.SolutionStops()[0].Vehicle() != move.Vehicle() {
							continue
						}

						sourceSpanFirst, sourceSpanLast = determineFirstLastSolutionStops(
							sourceSpanFirst,
							sourceSpanLast,
							solutionPlanStopsUnit,
						)
						newSourceSpanFirstPosition := newPositions[sourceSpanFirst]
						newSourceSpanLastPosition := newPositions[sourceSpanLast]

						if isViolatedPositions(
							newSourceSpanFirstPosition,
							newSourceSpanLastPosition,
							newPlanUnitSpanFirstPosition,
							newPlanUnitSpanLastPosition,
						) {
							return true, noPositionsHint()
						}
					}
				}
			}
		}
	}

	// check if plan unit is a source
	if sourceDisallowedInterleaves, isSourcePlanUnit :=
		l.sourceDisallowedInterleaves[move.PlanStopsUnit().ModelPlanUnit()]; isSourcePlanUnit {
		for _, disallowedInterleave := range sourceDisallowedInterleaves {
			targetSolutionPlanUnit := solution.SolutionPlanUnit(disallowedInterleave.Target())
			if targetSolutionPlanUnit.IsPlanned() {
				var targetSpanFirst, targetSpanLast SolutionStop
				for _, plannedSolutionStops := range targetSolutionPlanUnit.PlannedPlanStopsUnits() {
					if plannedSolutionStops.SolutionStops()[0].Vehicle() != move.Vehicle() {
						continue
					}
					targetSpanFirst, targetSpanLast = determineFirstLastSolutionStops(
						targetSpanFirst,
						targetSpanLast,
						plannedSolutionStops,
					)
					newTargetSpanFirstPosition := newPositions[targetSpanFirst]
					newTargetSpanLastPosition := newPositions[targetSpanLast]

					if isViolatedPositions(
						newPlanUnitSpanFirstPosition,
						newPlanUnitSpanLastPosition,
						newTargetSpanFirstPosition,
						newTargetSpanLastPosition,
					) {
						return true, noPositionsHint()
					}
				}
			}
		}
	}
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
