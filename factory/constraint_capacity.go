// © 2019-present nextmv.io inc

package factory

import (
	"fmt"
	"reflect"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/nextroute/schema"
)

// addCapacityConstraint uses the stop's quantity and vehicle's capacity
// to create new stop and vehicleType expressions, respectively. It uses these
// expressions to add a new maximum constraint to the model.
func addCapacityConstraint(
	input schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	quantities, names, quantitiesPresent, err := stopQuantities(input.Stops)
	if err != nil {
		return nil, err
	}

	alternateQuantitiesPresent := false

	if input.AlternateStops != nil {
		quantities, names, alternateQuantitiesPresent, err = alternateStopQuantities(input, model, quantities, names)
		if err != nil {
			return nil, err
		}
	}

	capacities, names, capacitiesPresent, err := capacities(input.Vehicles, names)
	if err != nil {
		return nil, err
	}

	startLevels, names, initialLevelsPresent, err := startLevels(input.Vehicles, names)
	if err != nil {
		return nil, err
	}

	if !quantitiesPresent && !alternateQuantitiesPresent && !capacitiesPresent {
		if initialLevelsPresent {
			return nil, nmerror.NewInputDataError(fmt.Errorf(
				"start levels present in vehicles (%t) but no capacity present in vehicles (%t)",
				initialLevelsPresent,
				capacitiesPresent,
			))
		}
		return model, nil
	}

	if (quantitiesPresent || alternateQuantitiesPresent) && !capacitiesPresent {
		return nil, nmerror.NewInputDataError(fmt.Errorf(
			"quantity present in stops (%t) or in alternate stops (%t) but no capacities in vehicles",
			quantitiesPresent,
			alternateQuantitiesPresent,
		))
	}

	for _, disableName := range options.Constraints.Disable.Capacities {
		if _, ok := names[disableName]; !ok {
			return nil, nmerror.NewInputDataError(fmt.Errorf(
				"cannot disable capacity constraint for resource '%s', "+
					"it does not exist in the input",
				disableName,
			))
		}
		delete(names, disableName)
	}

	model, quantityExpressions, capacityExpressions, err := addMaximumConstraint(model, names)
	if err != nil {
		return nil, err
	}

	err = setExpressionValues(
		names,
		quantities,
		capacities,
		startLevels,
		model.Stops(),
		model.Vehicles(),
		quantityExpressions,
		capacityExpressions,
	)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// stopQuantities returns the resource quantities for the stops. It also
// returns the names of the resources. The int flag indicates if there are
// resource quantities present in the stops, as indicated by the presence of
// the Quantity field.
func stopQuantities(stops []schema.Stop) (
	map[int]map[string]float64,
	map[string]bool,
	bool,
	error,
) {
	requirements := make(map[int]map[string]float64, len(stops))
	names := map[string]bool{}
	present := false
	for s, stop := range stops {
		if stop.Quantity != nil {
			present = true
			resources, err := resources(stop, "Quantity", -1)
			if err != nil {
				return nil, nil, present, err
			}
			requirements[s] = resources
			for k := range resources {
				names[k] = true
			}
		}
	}

	return requirements, names, present, nil
}

func alternateStopQuantities(
	input schema.Input,
	model nextroute.Model,
	requirements map[int]map[string]float64,
	names map[string]bool,
) (
	map[int]map[string]float64,
	map[string]bool,
	bool,
	error,
) {
	if input.AlternateStops == nil {
		return requirements, names, false, nil
	}

	if common.AllTrue(
		*input.AlternateStops,
		func(stop schema.AlternateStop) bool {
			return stop.Quantity == nil
		},
	) {
		return requirements, names, false, nil
	}

	data, err := getModelData(model)

	if err != nil {
		return requirements, names, false, err
	}

	for _, inputVehicle := range input.Vehicles {
		if inputVehicle.AlternateStops == nil {
			continue
		}

		for _, alternateID := range *inputVehicle.AlternateStops {
			stop, err := model.Stop(data.stopIDToIndex[alternateStopID(alternateID, inputVehicle)])
			if err != nil {
				return nil, nil, false, err
			}

			alternateInputStop := stop.Data().(alternateInputStop)

			if alternateInputStop.stop.Quantity == nil {
				continue
			}

			resources, err := resources(alternateInputStop.stop, "Quantity", -1)
			if err != nil {
				return nil, nil, false, err
			}

			for k := range resources {
				names[k] = true
			}

			requirements[stop.Index()] = resources
		}
	}

	return requirements, names, true, nil
}

// startLevels returns the resource start levels for the vehicles. It also
// appends names to the list of resource names.  The int flag indicates if there
// are resource levels present in the vehicles, as indicated by the presence of
// the StartLevel field.
func startLevels(vehicles []schema.Vehicle, names map[string]bool) (
	map[int]map[string]float64,
	map[string]bool,
	bool,
	error,
) {
	levels := make(map[int]map[string]float64, len(vehicles))
	present := false
	for v, vehicle := range vehicles {
		if vehicle.Capacity != nil {
			present = true
			resources, err := resources(vehicle, "StartLevel", 1)
			if err != nil {
				return nil, nil, present, err
			}
			levels[v] = resources
			for k := range resources {
				names[k] = true
			}
		}
	}

	return levels, names, present, nil
}

// capacities returns the resource capacities for the vehicles. It also appends
// names to the list of resource names.  The int flag indicates if there are
// resource capacities present in the vehicles, as indicated by the presence of
// the Capacity field.
func capacities(vehicles []schema.Vehicle, names map[string]bool) (
	map[int]map[string]float64,
	map[string]bool,
	bool,
	error,
) {
	limits := make(map[int]map[string]float64, len(vehicles))
	present := false
	for v, vehicle := range vehicles {
		if vehicle.Capacity != nil {
			present = true
			resources, err := resources(vehicle, "Capacity", 1)
			if err != nil {
				return nil, nil, present, err
			}
			limits[v] = resources
			for k := range resources {
				names[k] = true
			}
		}
	}

	return limits, names, present, nil
}

// resources processes the resources from an entity. The function analyzes if
// there are multiple resources (characterized by a map) or a single one (just
// an int). In any case, a map is always returned. The name argument specifies
// the field of the struct, for example: "Capacity". Sense manipulates the sign
// of the requirement. For example, for a stop the sense should be -1 to flip
// the sense's meaning. In nextroute, a positive (+) quantity consumes a
// resource and a negative (-) quantity adds to the level of resource. For a
// vehicle we shouldn't need to flip the sense, so it should be 1.
func resources[T schema.Vehicle | schema.Stop | schema.AlternateStop](
	entity T,
	name string,
	sense int,
) (map[string]float64, error) {
	field := reflect.ValueOf(entity).FieldByName(name).Interface()
	requirements := map[string]float64{}
	if field == nil {
		return requirements, nil
	}
	id := reflect.ValueOf(entity).FieldByName("ID")

	switch typeValue := field.(type) {
	case map[string]any:
		return stringAnyMap(typeValue, name, sense, id)
	case map[string]int:
		return stringAnyMap(typeValue, name, sense, id)
	case map[string]float64:
		return typeValue, nil
	case float64:
		requirements["default"] = typeValue * float64(sense)
		return requirements, nil
	case int:
		value, ok := convertToFloat(field)
		if ok {
			requirements["default"] = value * float64(sense)
			return requirements, nil
		}
	}

	return nil,
		nmerror.NewInputDataError(fmt.Errorf(
			"could not obtain %s requirement from entity %s, "+
				"it is neither a \n"+
				"- map of string to int\n"+
				"- map of string to float\n"+
				"- or int or float\n"+
				"got: %v",
			name,
			id,
			field,
		))
}

type anyOrInt interface {
	int | any
}

func stringAnyMap[T anyOrInt](
	parsed map[string]T,
	name string,
	sense int,
	id reflect.Value,
) (map[string]float64, error) {
	requirements := make(map[string]float64)
	for key, v := range parsed {
		value := 0.0
		if anyValue, ok := any(v).(int); ok {
			value = float64(anyValue)
		} else {
			value, ok = convertToFloat(v)
			if !ok {
				return nil,
					nmerror.NewInputDataError(fmt.Errorf(
						"could not obtain %s requirement from entity %s, "+
							"expected map of string to int, got %v",
						name,
						id,
						v,
					))
			}
		}

		requirements[key] = value * float64(sense)
	}

	return requirements, nil
}

// addMaximumConstraint is an auxiliary function to return the actual
// expressions that the Model uses. Because it receives and returns a Model, it
// mutates it to add maximum as a constraint.
func addMaximumConstraint(
	model nextroute.Model,
	names map[string]bool,
) (
	nextroute.Model,
	map[string]nextroute.StopExpression,
	map[string]nextroute.VehicleTypeValueExpression,
	error,
) {
	requirements := map[string]nextroute.StopExpression{}
	limits := map[string]nextroute.VehicleTypeValueExpression{}
	for name := range names {
		requirement := nextroute.NewStopExpression(name, 0.)
		limit := nextroute.NewVehicleTypeValueExpression(name, 0.)
		maximum, err := nextroute.NewMaximum(requirement, limit)
		if err != nil {
			return nil, nil, nil, err
		}
		maximum.(nextroute.Identifier).SetID("capacity_" + name)
		err = model.AddConstraint(maximum)
		if err != nil {
			return nil, nil, nil, err
		}
		requirements[name] = requirement
		limits[name] = limit
	}

	return model, requirements, limits, nil
}

// setExpressionValues is an auxiliary function that sets the values of the
// expressions based on the quantities and capacities.
func setExpressionValues(
	names map[string]bool,
	quantities, capacities, startLevels map[int]map[string]float64,
	stops nextroute.ModelStops,
	vehicles nextroute.ModelVehicles,
	quantityExpressions map[string]nextroute.StopExpression,
	capacityExpressions map[string]nextroute.VehicleTypeValueExpression,
) error {
	for s, stop := range stops {
		for name := range names {
			if value, ok := quantities[s][name]; ok {
				if value == 0 {
					continue
				}
				err := quantityExpressions[name].SetValue(stop, value)
				if err != nil {
					return err
				}
			}
		}
	}

	for v, vehicle := range vehicles {
		for name := range names {
			if value, ok := capacities[v][name]; ok {
				err := capacityExpressions[name].SetValue(vehicle.VehicleType(), value)
				if err != nil {
					return err
				}
			}
			if level, ok := startLevels[v][name]; ok {
				err := quantityExpressions[name].SetValue(vehicle.First(), level)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func convertToFloat(unknown any) (float64, bool) {
	floatType := reflect.TypeOf(float64(0))
	v := reflect.ValueOf(unknown)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, false
	}
	fv := v.Convert(floatType)

	return fv.Float(), true
}
