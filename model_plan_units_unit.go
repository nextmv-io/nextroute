package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
)

func newPlanUnitsUnit(
	index int,
	planUnits nextroute.ModelPlanUnits,
	planOneOf bool,
	sameVehicle bool,
) (nextroute.ModelPlanUnitsUnit, error) {
	if len(planUnits) == 0 {
		return nil,
			fmt.Errorf("plan units unit must have at least one plan unit")
	}

	uniquePlanUnits := common.UniqueDefined(planUnits, func(t nextroute.ModelPlanUnit) int {
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
		planUnits:     common.DefensiveCopy(planUnits),
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
		case nextroute.ModelPlanStopsUnit:
			err := planUnit.(*planMultipleStopsImpl).setPlanUnitsUnit(planUnitsUnit)
			if err != nil {
				return nil, err
			}
		case nextroute.ModelPlanUnitsUnit:
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

// planUnitsUnitImpl implements nextroute.ModelPlanUnitsUnit.
type planUnitsUnitImpl struct {
	modelDataImpl
	planUnits     nextroute.ModelPlanUnits
	planOneOf     bool
	index         int
	planUnitsUnit nextroute.ModelPlanUnitsUnit
	sameVehicle   bool
}

func (p *planUnitsUnitImpl) SameVehicle() bool {
	return p.sameVehicle
}

func (p *planUnitsUnitImpl) PlanUnitsUnit() (nextroute.ModelPlanUnitsUnit, bool) {
	return p.planUnitsUnit, p.planUnitsUnit != nil
}

func (p *planUnitsUnitImpl) setPlanUnitsUnit(planUnitsUnit nextroute.ModelPlanUnitsUnit) error {
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

func (p *planUnitsUnitImpl) PlanUnits() nextroute.ModelPlanUnits {
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
