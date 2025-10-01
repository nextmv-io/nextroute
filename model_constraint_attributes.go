// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"slices"

	"github.com/nextmv-io/nextroute/common"
)

// AttributesConstraint is a constraint that limits the vehicles a plan unit
// can be added to. The Attribute constraint configures compatibility
// attributes for stops and vehicles separately. This is done by specifying
// a list of attributes for stops and vehicles, respectively. Stops that
// have configured attributes are only compatible with vehicles that match
// at least one of them. Stops that do not have any specified attributes are
// compatible with any vehicle. Vehicles that do not have any specified
// attributes are only compatible with stops without attributes.
type AttributesConstraint struct {
	stopAttributes        map[int][]string
	vehicleTypeAttributes map[int][]string
	compatible            []bool
	vehicleTypes          int
}

// NewAttributesConstraint returns a new AttributesConstraint.
func NewAttributesConstraint() (*AttributesConstraint, error) {
	return &AttributesConstraint{
		stopAttributes:        make(map[int][]string),
		vehicleTypeAttributes: make(map[int][]string),
	}, nil
}

// Lock locks the constraint to the given model. After locking, no more
// attributes can be set for stops and vehicle types.
func (l *AttributesConstraint) Lock(model *Model) error {
	vehicleTypeAttributes := make(map[int]map[string]bool)
	vehicleTypes := model.VehicleTypes()
	l.vehicleTypes = len(vehicleTypes)
	for _, vehicleType := range vehicleTypes {
		vehicleTypeAttributes[vehicleType.Index()] = make(map[string]bool)
		for _, attribute := range l.vehicleTypeAttributes[vehicleType.Index()] {
			vehicleTypeAttributes[vehicleType.Index()][attribute] = true
		}
	}

	// Determine which stops are individually compatible with which vehicle
	// types.
	stopVehicleCompatible := make([]bool, model.NumberOfStops()*len(vehicleTypes))
	for _, stop := range model.stops {
		for _, vehicleType := range vehicleTypes {
			idx := l.mapTwoIndices(stop.Index(), vehicleType.Index())
			stopVehicleCompatible[idx] = len(l.stopAttributes[stop.Index()]) == 0
			for _, stopAttribute := range l.stopAttributes[stop.Index()] {
				if _, ok := vehicleTypeAttributes[vehicleType.Index()][stopAttribute]; ok {
					stopVehicleCompatible[idx] = true
					break
				}
			}
		}
	}

	// Determine which plan unit is compatible with which vehicle type by
	// gathering all the stops in the plan unit and checking if they are
	// compatible with the vehicle type.
	l.compatible = make([]bool, len(model.planUnits)*len(vehicleTypes))
	for _, planUnit := range model.PlanStopsUnits() {
		stops := planUnit.Stops()
		for _, vehicleType := range vehicleTypes {
			compatible := true
			for _, stop := range stops {
				idx := l.mapTwoIndices(stop.Index(), vehicleType.Index())
				compatible = compatible && stopVehicleCompatible[idx]
			}
			idx := l.mapTwoIndices(planUnit.Index(), vehicleType.Index())
			l.compatible[idx] = compatible
		}
	}

	return nil
}

// String returns the string representation of the constraint.
func (l *AttributesConstraint) String() string {
	return "attributes"
}

// StopAttributes returns the attributes for the given stop. The attributes
// are specified as a list of strings.
func (l *AttributesConstraint) StopAttributes(stop *ModelStop) []string {
	if attributes, hasAttributes := l.stopAttributes[stop.Index()]; hasAttributes {
		return slices.Clone(attributes)
	}
	return []string{}
}

// VehicleTypeAttributes returns the attributes for the given vehicle type.
// The attributes are specified as a list of strings.
func (l *AttributesConstraint) VehicleTypeAttributes(vehicle *ModelVehicleType) []string {
	if attributes, hasAttributes := l.vehicleTypeAttributes[vehicle.Index()]; hasAttributes {
		return slices.Clone(attributes)
	}
	return []string{}
}

// SetStopAttributes sets the attributes for the given stop. The attributes
// are specified as a list of strings. The attributes are not interpreted
// in any way. They are only used to determine compatibility between stops
// and vehicle types.
func (l *AttributesConstraint) SetStopAttributes(
	stop *ModelStop,
	stopAttributes []string,
) error {
	if stop.Model().IsLocked() {
		return fmt.Errorf(lockErrorMessage, "set stop attributes")
	}
	l.stopAttributes[stop.Index()] = common.Unique(stopAttributes)
	return nil
}

// SetVehicleTypeAttributes sets the attributes for the given vehicle type.
// The attributes are specified as a list of strings. The attributes are not
// interpreted in any way. They are only used to determine compatibility
// between stops and vehicle types.
func (l *AttributesConstraint) SetVehicleTypeAttributes(
	vehicleType *ModelVehicleType,
	vehicleAttributes []string,
) error {
	if vehicleType.Model().IsLocked() {
		return fmt.Errorf(lockErrorMessage, "set vehicle type attributes")
	}
	l.vehicleTypeAttributes[vehicleType.Index()] = common.Unique(vehicleAttributes)
	return nil
}

// EstimationCost returns the cost of the constraint for estimation purposes.
func (l *AttributesConstraint) EstimationCost() Cost {
	return Constant
}

// EstimateIsViolated estimates whether the given move violates the constraint.
func (l *AttributesConstraint) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	moveImpl := move.(*solutionMoveStopsImpl)
	planUnitIdx := moveImpl.planUnit.modelPlanStopsUnit.Index()
	vehicleType := moveImpl.vehicle().ModelVehicle().VehicleType()
	idx := l.mapTwoIndices(planUnitIdx, vehicleType.Index())
	compatible := l.compatible[idx]
	if compatible {
		return false, NoPositionsHint()
	}
	return true, SkipVehiclePositionsHint()
}

func (l *AttributesConstraint) mapTwoIndices(i, j int) int {
	return i*l.vehicleTypes + j
}
