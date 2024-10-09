// © 2019-present nextmv.io inc

package factory

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	"github.com/nextmv-io/nextroute/schema"
	"github.com/nextmv-io/sdk/measure"
)

// addVehicles adds the vehicle types to the Model.
func addVehicles(
	input schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}

	var travelDuration nextroute.DurationExpression
	switch matrix := input.DurationMatrix.(type) {
	case [][]float64:
		travelDuration = travelDurationExpression(matrix)
	case map[string]any:
		var durationMatrices schema.DurationMatrices
		jsonData, err := json.Marshal(matrix)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonData, &durationMatrices)
		if err != nil {
			return nil, err
		}
		travelDuration, err = dependentTravelDurationExpression(durationMatrices, model)
		if err != nil {
			return nil, err
		}
	case []any:
		var durationMatrix [][]float64
		jsonData, err := json.Marshal(matrix)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonData, &durationMatrix)
		if err != nil {
			return nil, err
		}
		travelDuration = travelDurationExpression(durationMatrix)
	case nil:
	default:
		return nil, fmt.Errorf("invalid duration matrix type: %T", matrix)
	}

	durationGroupsExpression := NewDurationGroupsExpression(model.NumberOfStops(), len(input.Vehicles))
	distanceExpression := distanceExpression(input.DistanceMatrix)

	inputVehicleHasAlternateStops := false

	constraint, err := nextroute.NewAttributesConstraint()

	if err != nil {
		return nil, err
	}

	for idx, inputVehicle := range input.Vehicles {
		vehicleType, err := newVehicleType(
			inputVehicle,
			model,
			distanceExpression,
			travelDuration,
			durationGroupsExpression,
		)
		if err != nil {
			return nil, err
		}

		vehicle, err := newVehicle(inputVehicle, vehicleType, model, options)
		if err != nil {
			return nil, err
		}

		if inputVehicle.AlternateStops != nil {
			inputVehicleHasAlternateStops = true
			vehicle.First().SetMeasureIndex(len(input.Stops) + len(*input.AlternateStops) + idx*2)
			vehicle.Last().SetMeasureIndex(len(input.Stops) + len(*input.AlternateStops) + idx*2 + 1)

			err = constraint.SetVehicleTypeAttributes(
				vehicleType,
				[]string{alternateVehicleAttribute(idx)},
			)
			if err != nil {
				return nil, err
			}
			for _, alternateID := range *inputVehicle.AlternateStops {
				alternateStop, err := model.Stop(data.stopIDToIndex[alternateStopID(alternateID, inputVehicle)])
				if err != nil {
					return nil, err
				}
				err = constraint.SetStopAttributes(alternateStop, []string{alternateVehicleAttribute(idx)})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if inputVehicleHasAlternateStops {
		err = model.AddConstraint(constraint)
		if err != nil {
			return nil, err
		}
	}

	return model, nil
}

// newVehicleType returns the VehicleType that the Model needs.
func newVehicleType(
	vehicle schema.Vehicle,
	model nextroute.Model,
	distanceExpression nextroute.DistanceExpression,
	durationExpression nextroute.DurationExpression,
	durationGroupsExpression DurationGroupsExpression,
) (nextroute.ModelVehicleType, error) {
	if durationExpression == nil {
		s := common.NewSpeed(*vehicle.Speed, common.MetersPerSecond)
		durationExpression = nextroute.NewTravelDurationExpression(distanceExpression, s)
		durationExpression.SetName(fmt.Sprintf(
			"travelDuration(%s,%s,%s)",
			vehicle.ID,
			distanceExpression.Name(),
			s,
		))
	}

	var vehicleType nextroute.ModelVehicleType
	switch expression := durationExpression.(type) {
	case nextroute.TimeDependentDurationExpression:
		vt, err := model.NewVehicleType(
			expression,
			durationGroupsExpression,
		)
		if err != nil {
			return nil, err
		}
		vehicleType = vt
	default:
		vt, err := model.NewVehicleType(
			nextroute.NewTimeIndependentDurationExpression(durationExpression),
			durationGroupsExpression,
		)
		if err != nil {
			return nil, err
		}
		vehicleType = vt
	}

	vehicleType.SetID(vehicle.ID)
	vehicleType.SetData(vehicleTypeData{
		DistanceExpression: distanceExpression,
	})

	return vehicleType, nil
}

func newVehicle(
	inputVehicle schema.Vehicle,
	vehicleType nextroute.ModelVehicleType,
	model nextroute.Model,
	options Options,
) (nextroute.ModelVehicle, error) {
	startLocation := common.NewInvalidLocation()
	var err error
	if inputVehicle.StartLocation != nil {
		startLocation, err = common.NewLocation(
			inputVehicle.StartLocation.Lon,
			inputVehicle.StartLocation.Lat,
		)
		if err != nil {
			return nil, err
		}
	}
	start, err := model.NewStop(startLocation)
	if err != nil {
		return nil, err
	}
	start.SetID(inputVehicle.ID + "-start")

	endLocation := common.NewInvalidLocation()
	if inputVehicle.EndLocation != nil {
		endLocation, err = common.NewLocation(
			inputVehicle.EndLocation.Lon,
			inputVehicle.EndLocation.Lat,
		)
		if err != nil {
			return nil, err
		}
	}
	end, err := model.NewStop(endLocation)
	if err != nil {
		return nil, err
	}
	end.SetID(inputVehicle.ID + "-end")

	startTime := model.Epoch()
	if !options.Constraints.Disable.VehicleStartTime && inputVehicle.StartTime != nil {
		startTime = *inputVehicle.StartTime
	}

	vehicle, err := model.NewVehicle(
		vehicleType,
		startTime,
		start,
		end,
	)
	if err != nil {
		return nil, err
	}

	vehicle.SetID(inputVehicle.ID)
	vehicle.SetData(inputVehicle)

	return vehicle, nil
}

// travelDurationExpressions returns the expressions that define how vehicles
// travel from one stop to another and the time it takes them to process a stop
// (service it).
func travelDurationExpression(matrix [][]float64) nextroute.DurationExpression {
	var travelDuration nextroute.DurationExpression
	if matrix != nil {
		travelDuration = nextroute.NewDurationExpression(
			"travelDuration",
			nextroute.NewMeasureByIndexExpression(measure.Matrix(matrix)),
			common.Second,
		)
	}

	return travelDuration
}

func dependentTravelDurationExpression(
	durationMatrices schema.DurationMatrices,
	model nextroute.Model,
) (nextroute.DurationExpression, error) {
	if durationMatrices.DefaultMatrix != nil {
		defaultExpression := nextroute.NewDurationExpression(
			"default_duration_expression",
			nextroute.NewMeasureByIndexExpression(measure.Matrix(durationMatrices.DefaultMatrix)),
			common.Second,
		)

		timeExpression, err := nextroute.NewTimeDependentDurationExpression(model, defaultExpression)
		if err != nil {
			return nil, err
		}

		for i, tf := range durationMatrices.TimeFrames {
			if tf.ScalingFactor != nil {
				scaledExpression := nextroute.NewScaledDurationExpression(defaultExpression, *tf.ScalingFactor)
				if err := timeExpression.SetExpression(tf.StartTime, tf.EndTime, scaledExpression); err != nil {
					return nil, err
				}
			} else {
				trafficExpression := nextroute.NewDurationExpression(
					fmt.Sprintf("traffic_duration_expression_%d", i),
					nextroute.NewMeasureByIndexExpression(measure.Matrix(tf.Matrix)),
					common.Second,
				)
				if err := timeExpression.SetExpression(tf.StartTime, tf.EndTime, trafficExpression); err != nil {
					return nil, err
				}
			}
		}

		return timeExpression, nil
	}

	return nil, errors.New("no duration matrix provided")
}

// distanceExpression creates a distance expression for later use.
func distanceExpression(distanceMatrix *[][]float64) nextroute.DistanceExpression {
	distanceExpression := nextroute.NewHaversineExpression()
	if distanceMatrix != nil {
		distanceExpression = nextroute.NewDistanceExpression(
			"travelDistance",
			nextroute.NewMeasureByIndexExpression(measure.Matrix(*distanceMatrix)),
			common.Meters,
		)
	}
	return distanceExpression
}
