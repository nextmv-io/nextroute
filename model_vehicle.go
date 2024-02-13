package nextroute

import (
	"fmt"
	"strings"
	"time"

	"github.com/nextmv-io/sdk/nextroute"
)

type modelVehicleImpl struct {
	start time.Time
	modelDataImpl
	vehicleType nextroute.ModelVehicleType
	id          string
	stops       nextroute.ModelStops
	index       int
}

func newModelVehicle(
	index int,
	vehicleType nextroute.ModelVehicleType,
	start time.Time,
	first nextroute.ModelStop,
	last nextroute.ModelStop,
) (nextroute.ModelVehicle, error) {
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
		stops:         nextroute.ModelStops{first, last},
		start:         start,
	}, nil
}

func (v *modelVehicleImpl) VehicleType() nextroute.ModelVehicleType {
	return v.vehicleType
}

func (v *modelVehicleImpl) Index() int {
	return v.index
}

func (v *modelVehicleImpl) First() nextroute.ModelStop {
	return v.stops[0]
}

func (v *modelVehicleImpl) Last() nextroute.ModelStop {
	return v.stops[len(v.stops)-1]
}

func (v *modelVehicleImpl) Stops() nextroute.ModelStops {
	result := make(nextroute.ModelStops, len(v.stops)-2)
	if len(v.stops) > 2 {
		copy(result, v.stops[1:len(v.stops)-1])
	}
	return result
}

func (v *modelVehicleImpl) AddStop(
	stop nextroute.ModelStop,
	fixed bool,
) error {
	if v.Model().IsLocked() {
		return fmt.Errorf("can not add a stop `%v` to vehicle `%v`, "+
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
		return fmt.Errorf("can not add a stop `%v` to vehicle `%v`, "+
			"the stop is first or last",
			stop.ID(),
			v.ID(),
		)
	}
	if !stop.HasPlanStopsUnit() {
		return fmt.Errorf("can not add a stop `%v` to vehicle `%v`, "+
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

func (v *modelVehicleImpl) Model() nextroute.Model {
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
