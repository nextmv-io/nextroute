// © 2019-present nextmv.io inc

package nextroute

import (
	"errors"
)

// ModelVehicleType is a vehicle type. A vehicle type is a definition of a
// vehicle. It contains the process duration and travel duration expressions
// that are used to calculate the travel and process duration of a stop
// assignment to a vehicle of this type.
type ModelVehicleType interface {
	ModelData

	// TemporalValues calculates the temporal values if the vehicle
	// would depart at departure going from stop to stop. If from or to is
	// invalid, the returned travelDuration will be 0.
	TemporalValues(
		departure float64,
		from ModelStop,
		to ModelStop,
	) (travelDuration, arrival, start, end float64)

	// Index returns the index of the vehicle type.
	Index() int

	// Model returns the model of the vehicle type.
	Model() Model

	// ID returns the identifier of the vehicle.
	ID() string

	// DurationExpression returns the process duration expression of the
	// vehicle type. Is set in the factory method of the vehicle type
	// Model.NewVehicleType.
	DurationExpression() DurationExpression
	// SetDurationExpression modifies the process duration expression of
	// the vehicle type.
	SetDurationExpression(expression DurationExpression) error

	// SetID sets the identifier of the stop. This identifier is not used by
	// nextroute and therefore it does not have to be unique for nextroute
	// internally. However to make this ID useful for debugging and reporting it
	// should be made unique.
	SetID(string)

	// TravelDurationExpression returns the duration expression of the
	// vehicle type. Is set in the factory method of the vehicle type
	// Model.NewVehicleType.
	TravelDurationExpression() TimeDependentDurationExpression
	// SetTravelDurationExpression modifies the duration expression of the
	// vehicle type.
	SetTravelDurationExpression(expression TimeDependentDurationExpression) error

	// Vehicles returns the vehicles of this vehicle type.
	Vehicles() ModelVehicles
}

// ModelVehicleTypes is a slice of vehicle types.
type ModelVehicleTypes []ModelVehicleType

type vehicleTypeImpl struct {
	modelDataImpl
	model          Model
	travelDuration TimeDependentDurationExpression
	duration       DurationExpression
	id             string
	vehicles       ModelVehicles
	index          int
}

func (v *vehicleTypeImpl) Vehicles() ModelVehicles {
	vehicles := make(ModelVehicles, len(v.vehicles))
	copy(vehicles, v.vehicles)
	return vehicles
}

func (v *vehicleTypeImpl) Index() int {
	return v.index
}

func (v *vehicleTypeImpl) ID() string {
	return v.id
}

func (v *vehicleTypeImpl) SetID(id string) {
	v.id = id
}

func (v *vehicleTypeImpl) Model() Model {
	return v.model
}

func (v *vehicleTypeImpl) TravelDurationExpression() TimeDependentDurationExpression {
	return v.travelDuration
}

func (v *vehicleTypeImpl) DurationExpression() DurationExpression {
	return v.duration
}

func (v *vehicleTypeImpl) TemporalValues(
	departure float64,
	from ModelStop,
	to ModelStop,
) (travelDuration, arrival, start, end float64) {
	if from.Location().IsValid() && to.Location().IsValid() {
		travelDuration = v.travelDuration.ValueAtValue(
			departure,
			v,
			from,
			to,
		)
	}

	arrival = departure + travelDuration

	processDuration := v.duration.Value(
		v,
		from,
		to,
	)

	stopImpl := to.(*stopImpl)
	start = arrival
	earliestStart := stopImpl.ToEarliestStartValue(arrival)
	if earliestStart > start {
		start = earliestStart
	}
	end = start + processDuration

	return travelDuration, arrival, start, end
}

func (v *vehicleTypeImpl) SetTravelDurationExpression(e TimeDependentDurationExpression) error {
	if v.model.IsLocked() {
		return errors.New("cannot modify vehicle type (set travel duration expression) after model is locked")
	}

	if e == nil {
		return errors.New("cannot set travel duration expression to nil")
	}

	v.travelDuration = e
	return nil
}

func (v *vehicleTypeImpl) SetDurationExpression(e DurationExpression) error {
	if v.model.IsLocked() {
		return errors.New("cannot modify vehicle type (set process duration expression) after model is locked")
	}

	if e == nil {
		return errors.New("cannot set process duration expression to nil")
	}

	v.duration = e
	return nil
}
