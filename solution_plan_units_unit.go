package nextroute

import (
	"fmt"

	"github.com/nextmv-io/sdk/common"
)

type solutionPlanUnitsUnitImpl struct {
	modelPlanUnitsUnit ModelPlanUnitsUnit
	solutionPlanUnits  SolutionPlanUnits
	sameVehicle        bool
}

func (p *solutionPlanUnitsUnitImpl) String() string {
	return fmt.Sprintf("solutionPlanUnitsUnit{%v, planned=%v}",
		p.modelPlanUnitsUnit,
		p.IsPlanned(),
	)
}

func (p *solutionPlanUnitsUnitImpl) PlannedPlanStopsUnits() SolutionPlanStopsUnits {
	if p.modelPlanUnitsUnit.PlanAll() {
		solutionPlanStopsUnits := make(SolutionPlanStopsUnits, 0, len(p.solutionPlanUnits))
		for _, solutionPlanUnit := range p.solutionPlanUnits {
			solutionPlanStopsUnits = append(solutionPlanStopsUnits,
				solutionPlanUnit.PlannedPlanStopsUnits()...,
			)
		}
		return solutionPlanStopsUnits
	}
	for _, solutionPlanUnit := range p.solutionPlanUnits {
		solutionPlanStopsUnits := solutionPlanUnit.PlannedPlanStopsUnits()
		if len(solutionPlanStopsUnits) > 0 {
			return solutionPlanStopsUnits
		}
	}
	return SolutionPlanStopsUnits{}
}

func (p *solutionPlanUnitsUnitImpl) SolutionPlanUnit(planUnit ModelPlanUnit) SolutionPlanUnit {
	for _, solutionPlanUnit := range p.solutionPlanUnits {
		if solutionPlanUnit.ModelPlanUnit().Index() == planUnit.Index() {
			return solutionPlanUnit
		}
	}
	panic(
		fmt.Errorf("solution plan unit for model plan unit %v [%v] not found in unit %v",
			planUnit,
			planUnit.Index(),
			p.Index(),
		),
	)
}

func (p *solutionPlanUnitsUnitImpl) SameVehicle() bool {
	return p.sameVehicle
}

func (p *solutionPlanUnitsUnitImpl) ModelPlanUnit() ModelPlanUnit {
	return p.modelPlanUnitsUnit
}

func (p *solutionPlanUnitsUnitImpl) ModelPlanUnitsUnit() ModelPlanUnitsUnit {
	return p.modelPlanUnitsUnit
}
func (p *solutionPlanUnitsUnitImpl) Index() int {
	return p.modelPlanUnitsUnit.Index()
}

func (p *solutionPlanUnitsUnitImpl) Solution() Solution {
	return p.solutionPlanUnits[0].Solution()
}

func (p *solutionPlanUnitsUnitImpl) SolutionPlanUnits() SolutionPlanUnits {
	return common.DefensiveCopy(p.solutionPlanUnits)
}

func (p *solutionPlanUnitsUnitImpl) IsPlanned() bool {
	if p.modelPlanUnitsUnit.PlanAll() {
		if len(p.solutionPlanUnits) == 0 {
			return false
		}
		for _, solutionPlanUnit := range p.solutionPlanUnits {
			if !solutionPlanUnit.IsPlanned() {
				return false
			}
		}
		return true
	}
	for _, solutionPlanUnit := range p.solutionPlanUnits {
		if solutionPlanUnit.IsPlanned() {
			return true
		}
	}
	return false
}

func (p *solutionPlanUnitsUnitImpl) IsFixed() bool {
	for _, solutionPlanUnit := range p.solutionPlanUnits {
		if solutionPlanUnit.IsFixed() {
			return true
		}
	}
	return false
}

func (p *solutionPlanUnitsUnitImpl) UnPlan() (bool, error) {
	if !p.IsPlanned() || p.IsFixed() {
		return false, nil
	}

	solution := p.Solution().(*solutionImpl)

	solution.plannedPlanUnits.remove(p)
	solution.unPlannedPlanUnits.add(p)

	for _, solutionPlanUnit := range p.solutionPlanUnits {
		if solutionPlanUnit.IsPlanned() {
			// TODO: what if one of a conjunction of plan units fails to unplan?
			success, err := solutionPlanUnit.UnPlan()
			if err != nil {
				success = false
			}
			if !success {
				solution.plannedPlanUnits.add(p)
				solution.unPlannedPlanUnits.remove(p)
			}
		}
	}
	return true, nil
}
