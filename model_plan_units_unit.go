// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"slices"

	"github.com/nextmv-io/nextroute/common"
)

// ModelPlanUnitsUnit is a set of plan units. A plan unit is a set of stops
// that must be visited together.
type ModelPlanUnitsUnit interface {
	ModelPlanUnit

	// PlanUnits returns the plan units in the invoking unit.
	PlanUnits() ModelPlanUnits

	// PlanOneOf returns true if the plan unit only has to plan exactly one of
	// the associated plan units. If PlanOneOf returns true, then PlanAll will
	// return false and vice versa.
	PlanOneOf() bool

	// PlanAll returns true if the plan unit has to plan all the associated
	// plan units. If PlanAll returns true, then PlanOneOf will return false
	// and vice versa.
	PlanAll() bool

	// SameVehicle returns true if all the plan units in this unit have to be
	// planned on the same vehicle. If this unit is a conjunction, then
	// this will return true if all the plan units in this unit have to be
	// planned on the same vehicle. If this unit is a disjunction, then
	// this has no semantic meaning.
	SameVehicle() bool
}

// ModelPlanUnitsUnits is a slice of model plan units units .
type ModelPlanUnitsUnits []ModelPlanUnitsUnit

func newPlanUnitsUnit(
	index int,
	planUnits ModelPlanUnits,
	planOneOf bool,
	sameVehicle bool,
) (ModelPlanUnitsUnit, error) {
	if len(planUnits) == 0 {
		return nil,
			fmt.Errorf("plan units unit must have at least one plan unit")
	}

	uniquePlanUnits := common.UniqueDefined(planUnits, func(t ModelPlanUnit) int {
		return t.Index()
	})

	if len(uniquePlanUnits) != len(planUnits) {
		return nil,
			fmt.Errorf("plan units unit cannot have duplicate plan units")
	}

	planUnitsUnit := &planUnitsUnitImpl{
		modelDataImpl: newModelDataImpl(),
		index:         index,
		planOneOf:     planOneOf,
		planUnits:     slices.Clone(planUnits),
		sameVehicle:   sameVehicle,
	}

	for _, planUnit := range planUnits {
		if planUnit == nil {
			return nil, fmt.Errorf("plan unit cannot be nil")
		}
		if _, isElementOfPlanUnitsUnit := planUnit.PlanUnitsUnit(); isElementOfPlanUnitsUnit {
			return nil, fmt.Errorf("plan unit cannot be a member of two plan units units")
		}
		switch planUnit.(type) {
		case ModelPlanStopsUnit:
			err := planUnit.(*planMultipleStopsImpl).setPlanUnitsUnit(planUnitsUnit)
			if err != nil {
				return nil, err
			}
		case ModelPlanUnitsUnit:
			err := planUnit.(*planUnitsUnitImpl).setPlanUnitsUnit(planUnitsUnit)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("plan unit cannot be a plan units unit")
		}
	}

	return planUnitsUnit, nil
}

// planUnitsUnitImpl implements ModelPlanUnitsUnit.
type planUnitsUnitImpl struct {
	modelDataImpl
	planUnits     ModelPlanUnits
	planOneOf     bool
	index         int
	planUnitsUnit ModelPlanUnitsUnit
	sameVehicle   bool
}

func (p *planUnitsUnitImpl) SameVehicle() bool {
	return p.sameVehicle
}

func (p *planUnitsUnitImpl) PlanUnitsUnit() (ModelPlanUnitsUnit, bool) {
	return p.planUnitsUnit, p.planUnitsUnit != nil
}

func (p *planUnitsUnitImpl) setPlanUnitsUnit(planUnitsUnit ModelPlanUnitsUnit) error {
	if p.planUnitsUnit != nil {
		return fmt.Errorf("plan unit %v already has a plan units unit", p)
	}

	p.planUnitsUnit = planUnitsUnit
	return nil
}

func (p *planUnitsUnitImpl) String() string {
	return fmt.Sprintf("plan_units_unit{%v}", p.planUnits)
}

func (p *planUnitsUnitImpl) Index() int {
	return p.index
}

func (p *planUnitsUnitImpl) PlanUnits() ModelPlanUnits {
	return p.planUnits
}

func (p *planUnitsUnitImpl) PlanOneOf() bool {
	return p.planOneOf
}

func (p *planUnitsUnitImpl) PlanAll() bool {
	return !p.planOneOf
}
func (p *planUnitsUnitImpl) IsFixed() bool {
	for _, planUnit := range p.planUnits {
		if planUnit.IsFixed() {
			return true
		}
	}
	return false
}
