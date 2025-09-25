// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"strings"
	"time"
)

// ModelVehicles is a slice of ModelVehicle.
type ModelVehicles []*ModelVehicle

// ModelVehicle is a vehicle in the model. A vehicle is a sequence of stops.
type ModelVehicle struct {
	start time.Time
	modelDataImpl
	vehicleType *ModelVehicleType
	id          string
	stops       ModelStops
	index       int
}

// NewModelVehicle returns a new ModelVehicle.
// Always use this function to create a new ModelVehicle.
func NewModelVehicle(
	index int,
	vehicleType *ModelVehicleType,
	start time.Time,
	first *ModelStop,
	last *ModelStop,
) (*ModelVehicle, error) {
	if first.HasPlanStopsUnit() {
		return nil,
			fmt.Errorf("first stop %s already has a plan unit", first)
	}

	if last.HasPlanStopsUnit() {
		return nil,
			fmt.Errorf("last stop %s already has a plan unit", last)
	}

	first.firstOrLast = true
	first.fixed = true
	last.fixed = true
	last.firstOrLast = true

	return &ModelVehicle{
		modelDataImpl: newModelDataImpl(),
		index:         index,
		vehicleType:   vehicleType,
		stops:         ModelStops{first, last},
		start:         start,
	}, nil
}

// VehicleType returns the vehicle type of the vehicle.
func (v *ModelVehicle) VehicleType() *ModelVehicleType {
	return v.vehicleType
}

// Index returns the index of the vehicle.
func (v *ModelVehicle) Index() int {
	return v.index
}

// First returns the first stop of the vehicle.
func (v *ModelVehicle) First() *ModelStop {
	return v.stops[0]
}

// Last returns the last stop of the vehicle.
func (v *ModelVehicle) Last() *ModelStop {
	return v.stops[len(v.stops)-1]
}

// Stops returns the stops of the vehicle that are provided as a start
// assignment. The first and last stop of the vehicle are not included in
// the returned slice.
func (v *ModelVehicle) Stops() ModelStops {
	result := make(ModelStops, len(v.stops)-2)
	if len(v.stops) > 2 {
		copy(result, v.stops[1:len(v.stops)-1])
	}
	return result
}

// AddStop adds a stop to the vehicle. The stop is added to the end of the
// vehicle, before the last stop. If fixed is true stop will be fixed and
// can not be unplanned.
func (v *ModelVehicle) AddStop(
	stop *ModelStop,
	fixed bool,
) error {
	message := "can not add a stop `%v` to vehicle `%v`, "
	if v.Model().IsLocked() {
		return fmt.Errorf(message+
			"the model is locked, this happens once a"+
			"solution has been created using this model",
			stop.ID(),
			v.ID(),
		)
	}
	if stop == nil {
		return fmt.Errorf("can not add a nil stop to vehicle `%v`",
			v.ID(),
		)
	}
	if stop.IsFirstOrLast() {
		return fmt.Errorf(message+
			"the stop is first or last",
			stop.ID(),
			v.ID(),
		)
	}
	if !stop.HasPlanStopsUnit() {
		return fmt.Errorf(message+
			"the stop does not have a plan unit",
			stop.ID(),
			v.ID(),
		)
	}
	if vIdx, stopAddedToVehicle := v.Model().stopVehicles[stop.Index()]; stopAddedToVehicle {
		return fmt.Errorf("can not add a stop `%v` to vehicle `%v` "+
			"the stop is already added to vehicle `%v`",
			stop.ID(),
			v.ID(),
			v.Model().Vehicles()[vIdx].ID(),
		)
	}

	stop.model.stopVehicles[stop.Index()] = v.Index()
	stop.fixed = fixed

	v.stops = append(v.stops, v.stops[len(v.stops)-1])
	v.stops[len(v.stops)-2] = stop

	return nil
}

// Model returns the model of the vehicle.
func (v *ModelVehicle) Model() *Model {
	return v.vehicleType.Model()
}

// Start returns the start time of the vehicle.
func (v *ModelVehicle) Start() time.Time {
	return v.start
}

// ID returns the identifier of the vehicle.
func (v *ModelVehicle) ID() string {
	return v.id
}

// SetID sets the identifier of the vehicle. This identifier is not used by
// nextroute, and therefore it does not have to be unique for nextroute
// internally. However, to make this ID useful for debugging and reporting it
// should be made unique.
func (v *ModelVehicle) SetID(id string) {
	v.id = id
}

func (v *ModelVehicle) String() string {
	var sb strings.Builder

	_, _ = fmt.Fprintf(
		&sb,
		"%v [%v]",
		v.id,
		v.index,
	)
	return sb.String()
}
