// © 2019-present nextmv.io inc

package factory

import (
	"reflect"

	"github.com/nextmv-io/nextroute/schema"
)

// applyDefaults modifies the input by applying default values to stops and
// vehicles as long as they are defined on the corresponding default structs.
func applyDefaults(input schema.Input) schema.Input {
	if input.Defaults == nil {
		return input
	}

	if input.Defaults.Vehicles != nil {
		for v, vehicle := range input.Vehicles {
			for _, field := range fields(input.Defaults.Vehicles) {
				vehicle = processDefault(input, vehicle, field)
			}
			input.Vehicles[v] = vehicle
		}
	}

	if input.Defaults.Stops != nil {
		for s, stop := range input.Stops {
			for _, field := range fields(input.Defaults.Stops) {
				stop = processDefault(input, stop, field)
			}
			input.Stops[s] = stop
		}
	}

	return input
}

// fields returns a list of field names from the given defaults' struct.
func fields[T schema.VehicleDefaults | schema.StopDefaults](defaults *T) []string {
	v := reflect.Indirect(reflect.ValueOf(defaults))
	fields := make([]string, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		fields[i] = v.Type().Field(i).Name
	}
	return fields
}

// processDefault sets a value on the entity if that value is present on the
// input's defaults based on the given name. If the field (given by the name)
// is already present on the entity, then nothing needs to happen. Here are two
// examples of how it works:
//
//	processDefault(i, vehicle, "Start") // Looks for i.Defaults.Vehicles.Start and sets it on vehicle.
//	processDefault(i, stop, "MaxWait") // Looks for i.Defaults.Stops.MaxWait and sets it on stop.
//
// This function is not tested because it is easier to test the function that
// calls it: applyDefaults. Because ultimately they test the same thing, it is
// not worth testing.
func processDefault[T schema.Vehicle | schema.Stop](i schema.Input, entity T, name string) T {
	// If the property on the entity is not nil, that means that it has been
	// individually set.
	field := reflect.Indirect(reflect.ValueOf(&entity)).FieldByName(name)
	if field.IsValid() && !field.IsZero() && !field.IsNil() {
		return entity
	}

	// Get the field from the defaults based on the entity type.
	var defaultField reflect.Value
	if _, ok := any(entity).(schema.Vehicle); ok {
		defaultField = reflect.Indirect(reflect.ValueOf(i.Defaults.Vehicles)).FieldByName(name)
	} else {
		defaultField = reflect.Indirect(reflect.ValueOf(i.Defaults.Stops)).FieldByName(name)
	}

	// If the default field is not nil, set it on the entity.
	if defaultField.IsValid() && !defaultField.IsZero() && !defaultField.IsNil() {
		reflect.Indirect(reflect.ValueOf(&entity)).FieldByName(name).Set(defaultField)
	}

	return entity
}
