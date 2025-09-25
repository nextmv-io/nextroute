// © 2019-present nextmv.io inc

package nextroute

import (
	"errors"
)

// ModelVehicleType is a vehicle type. A vehicle type is a definition of a
// vehicle. It contains the process duration and travel duration expressions
// that are used to calculate the travel and process duration of a stop
// assignment to a vehicle of this type.
type ModelVehicleType struct {
	modelDataImpl
	model          *Model
	travelDuration TimeDependentDurationExpression
	duration       DurationExpression
	distance       DistanceExpression
	id             string
	vehicles       ModelVehicles
	index          int
}

// ModelVehicleTypes is a slice of vehicle types.
type ModelVehicleTypes []*ModelVehicleType

// Vehicles returns the vehicles of this vehicle type.
func (v *ModelVehicleType) Vehicles() ModelVehicles {
	vehicles := make(ModelVehicles, len(v.vehicles))
	copy(vehicles, v.vehicles)
	return vehicles
}

// Index returns the index of the vehicle type.
func (v *ModelVehicleType) Index() int {
	return v.index
}

// ID returns the identifier of the vehicle.
func (v *ModelVehicleType) ID() string {
	return v.id
}

// SetID sets the identifier of the vehicle type.
func (v *ModelVehicleType) SetID(id string) {
	v.id = id
}

// Model returns the model of the vehicle type.
func (v *ModelVehicleType) Model() *Model {
	return v.model
}

// TravelDurationExpression returns the duration expression of the
// vehicle type. Is set in the factory method of the vehicle type
// Model.NewVehicleType.
func (v *ModelVehicleType) TravelDurationExpression() TimeDependentDurationExpression {
	return v.travelDuration
}

// DurationExpression returns the process duration expression of the
// vehicle type. Is set in the factory method of the vehicle type
// Model.NewVehicleType.
func (v *ModelVehicleType) DurationExpression() DurationExpression {
	return v.duration
}

// DistanceExpression returns the distance expression of the vehicle type.
func (v *ModelVehicleType) DistanceExpression() DistanceExpression {
	return v.distance
}

// TemporalValues calculates the temporal values if the vehicle
// would depart at departure going from stop to stop. If from or to is
// invalid, the returned travelDuration will be 0.
func (v *ModelVehicleType) TemporalValues(
	departure float64,
	from *ModelStop,
	to *ModelStop,
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

	start = arrival
	earliestStart := to.ToEarliestStartValue(arrival)
	if earliestStart > start {
		start = earliestStart
	}
	end = start + processDuration

	return travelDuration, arrival, start, end
}

// SetTravelDurationExpression modifies the duration expression of the
// vehicle type.
func (v *ModelVehicleType) SetTravelDurationExpression(e TimeDependentDurationExpression) error {
	if v.model.IsLocked() {
		return errors.New("cannot modify vehicle type (set travel duration expression) after model is locked")
	}

	if e == nil {
		return errors.New("cannot set travel duration expression to nil")
	}

	v.travelDuration = e
	return nil
}

// SetDurationExpression modifies the process duration expression of
// the vehicle type.
func (v *ModelVehicleType) SetDurationExpression(e DurationExpression) error {
	if v.model.IsLocked() {
		return errors.New("cannot modify vehicle type (set process duration expression) after model is locked")
	}

	if e == nil {
		return errors.New("cannot set process duration expression to nil")
	}

	v.duration = e
	return nil
}

// SetDistanceExpression modifies the distance expression of the vehicle type.
func (v *ModelVehicleType) SetDistanceExpression(expression DistanceExpression) error {
	if v.model.IsLocked() {
		return errors.New("cannot modify vehicle type (set distance expression) after model is locked")
	}

	if expression == nil {
		return errors.New("cannot set distance expression to nil")
	}

	v.distance = expression
	return nil
}
