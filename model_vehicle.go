// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"strings"
	"time"
)

// ModelVehicle is a vehicle in the model. A vehicle is a sequence of stops.
type ModelVehicle interface {
	ModelData

	// AddStop adds a stop to the vehicle. The stop is added to the end of the
	// vehicle, before the last stop. If fixed is true stop will be fixed and
	// can not be unplanned.
	AddStop(stop ModelStop, fixed bool) error

	// First returns the first stop of the vehicle.
	First() ModelStop

	// ID returns the identifier of the vehicle.
	ID() string
	// Index returns the index of the vehicle.
	Index() int

	// Last returns the last stop of the vehicle.
	Last() ModelStop

	// Model returns the model of the vehicle.
	Model() Model

	// SetID sets the identifier of the vehicle. This identifier is not used by
	// nextroute, and therefore it does not have to be unique for nextroute
	// internally. However, to make this ID useful for debugging and reporting it
	// should be made unique.
	SetID(string)
	// Start returns the start time of the vehicle.
	Start() time.Time
	// Stops returns the stops of the vehicle that are provided as a start
	// assignment. The first and last stop of the vehicle are not included in
	// the returned slice.
	Stops() ModelStops

	// VehicleType returns the vehicle type of the vehicle.
	VehicleType() ModelVehicleType
}

// ModelVehicles is a slice of ModelVehicle.
type ModelVehicles []ModelVehicle

type modelVehicleImpl struct {
	start time.Time
	modelDataImpl
	vehicleType ModelVehicleType
	id          string
	stops       ModelStops
	index       int
}

func newModelVehicle(
	index int,
	vehicleType ModelVehicleType,
	start time.Time,
	first ModelStop,
	last ModelStop,
) (ModelVehicle, error) {
	if first.HasPlanStopsUnit() {
		return nil,
			fmt.Errorf("first stop %s already has a plan unit", first)
	}

	if last.HasPlanStopsUnit() {
		return nil,
			fmt.Errorf("last stop %s already has a plan unit", last)
	}

	first.(*stopImpl).firstOrLast = true
	first.(*stopImpl).fixed = true
	last.(*stopImpl).fixed = true
	last.(*stopImpl).firstOrLast = true

	return &modelVehicleImpl{
		modelDataImpl: newModelDataImpl(),
		index:         index,
		vehicleType:   vehicleType,
		stops:         ModelStops{first, last},
		start:         start,
	}, nil
}

func (v *modelVehicleImpl) VehicleType() ModelVehicleType {
	return v.vehicleType
}

func (v *modelVehicleImpl) Index() int {
	return v.index
}

func (v *modelVehicleImpl) First() ModelStop {
	return v.stops[0]
}

func (v *modelVehicleImpl) Last() ModelStop {
	return v.stops[len(v.stops)-1]
}

func (v *modelVehicleImpl) Stops() ModelStops {
	result := make(ModelStops, len(v.stops)-2)
	if len(v.stops) > 2 {
		copy(result, v.stops[1:len(v.stops)-1])
	}
	return result
}

func (v *modelVehicleImpl) AddStop(
	stop ModelStop,
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
	if vIdx, stopAddedToVehicle := v.Model().(*modelImpl).stopVehicles[stop.Index()]; stopAddedToVehicle {
		return fmt.Errorf("can not add a stop `%v` to vehicle `%v` "+
			"the stop is already added to vehicle `%v`",
			stop.ID(),
			v.ID(),
			v.Model().Vehicles()[vIdx].ID(),
		)
	}

	stop.(*stopImpl).model.stopVehicles[stop.Index()] = v.Index()
	stop.(*stopImpl).fixed = fixed

	v.stops = append(v.stops, v.stops[len(v.stops)-1])
	v.stops[len(v.stops)-2] = stop.(*stopImpl)

	return nil
}

func (v *modelVehicleImpl) Model() Model {
	return v.vehicleType.Model()
}

func (v *modelVehicleImpl) Start() time.Time {
	return v.start
}

func (v *modelVehicleImpl) ID() string {
	return v.id
}

func (v *modelVehicleImpl) SetID(id string) {
	v.id = id
}

func (v *modelVehicleImpl) String() string {
	var sb strings.Builder

	_, _ = fmt.Fprintf(
		&sb,
		"%v [%v]",
		v.id,
		v.index,
	)
	return sb.String()
}
