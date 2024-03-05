// © 2019-present nextmv.io inc

package nextroute

// ModelPlanUnit is a plan unit. It is a unit defining what should be planned .
// For example, a unit can be a pickup and a delivery stop that are required to
// be planned together on the same vehicle.
type ModelPlanUnit interface {
	ModelData

	// Index returns the index of the invoking unit.
	Index() int

	// IsFixed returns true if the PlanUnit is fixed.
	IsFixed() bool

	// PlanUnitsUnit returns the [ModelPlanUnitsUnit] associated with the unit
	// with a bool indicating if it actually has one. A plan unit is associated
	// with at most one plan units unit. Can be nil if the unit is not part of a
	// plan units unit in which case the second return argument will be false.
	PlanUnitsUnit() (ModelPlanUnitsUnit, bool)
}

// ModelPlanUnits is a slice of plan units .
type ModelPlanUnits []ModelPlanUnit
