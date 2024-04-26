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
}

// NewInterleaveConstraint returns a new InterleaveConstraint.
func NewInterleaveConstraint() (InterleaveConstraint, error) {
	return &interleaveConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"interleave",
			ModelExpressions{},
		),
	}, nil
}

type DisallowedInterleave interface {
	Target() ModelPlanUnit

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
	disallowedInterleaves []DisallowedInterleave
}

func (l *interleaveConstraintImpl) DisallowedInterleaves() []DisallowedInterleave {
	return l.disallowedInterleaves
}

func (l *interleaveConstraintImpl) Lock(model Model) error {

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

func (l *interleaveConstraintImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	for _, stopPosition := range move.StopPositions() {
		disallowed, _ := move.Solution().(*solutionImpl).interleaveConstraint.(*solutionConstraintInterleavedImpl).disallowedSuccessors(
			stopPosition.Stop(),
			stopPosition.Next(),
		)
		if disallowed {
			return true, noPositionsHint()
		}
	}

	return false, noPositionsHint()
}
