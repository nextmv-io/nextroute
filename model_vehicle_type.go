package nextroute

import (
	"errors"

	"github.com/nextmv-io/sdk/nextroute"
)

type vehicleTypeImpl struct {
	modelDataImpl
	model          nextroute.Model
	travelDuration nextroute.TimeDependentDurationExpression
	duration       nextroute.DurationExpression
	id             string
	vehicles       nextroute.ModelVehicles
	index          int
}

func (v *vehicleTypeImpl) Vehicles() nextroute.ModelVehicles {
	vehicles := make(nextroute.ModelVehicles, len(v.vehicles))
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

func (v *vehicleTypeImpl) Model() nextroute.Model {
	return v.model
}

func (v *vehicleTypeImpl) TravelDurationExpression() nextroute.TimeDependentDurationExpression {
	return v.travelDuration
}

func (v *vehicleTypeImpl) DurationExpression() nextroute.DurationExpression {
	return v.duration
}

func (v *vehicleTypeImpl) TemporalValues(
	departure float64,
	from nextroute.ModelStop,
	to nextroute.ModelStop,
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

func (v *vehicleTypeImpl) SetTravelDurationExpression(e nextroute.TimeDependentDurationExpression) error {
	if v.model.IsLocked() {
		return errors.New("cannot modify vehicle type (set travel duration expression) after model is locked")
	}

	if e == nil {
		return errors.New("cannot set travel duration expression to nil")
	}

	v.travelDuration = e
	return nil
}

func (v *vehicleTypeImpl) SetDurationExpression(e nextroute.DurationExpression) error {
	if v.model.IsLocked() {
		return errors.New("cannot modify vehicle type (set process duration expression) after model is locked")
	}

	if e == nil {
		return errors.New("cannot set process duration expression to nil")
	}

	v.duration = e
	return nil
}
