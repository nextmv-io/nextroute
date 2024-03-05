// © 2019-present nextmv.io inc

package factory

import (
	"fmt"
	"math"
	"reflect"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// addNoMixConstraint.
func addNoMixConstraint(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	mixingItems := make(map[string]map[nextroute.ModelStop]nextroute.MixItem)

	for index, inputStop := range input.Stops {
		if inputStop.MixingItems == nil {
			continue
		}

		stop, err := model.Stop(index)
		if err != nil {
			return nil, err
		}

		field := reflect.ValueOf(inputStop).FieldByName("MixingItems").Interface()
		if field == nil {
			continue
		}
		switch typeValue := field.(type) {
		case map[string]any:
			err = addMixingItems(mixingItems, stop, typeValue)
			if err != nil {
				return nil, err
			}
		case map[string]nextroute.MixItem:
			for t, mixingItem := range typeValue {
				if _, ok := mixingItems[t]; !ok {
					mixingItems[t] = make(map[nextroute.ModelStop]nextroute.MixItem)
				}
				mixingItems[t][stop] = mixingItem
			}
		default:
			return nil,
				fmt.Errorf("add no-mix constraint, stop %v unexpected type %T for mixing_items property",
					stop.ID(),
					typeValue,
				)
		}
	}

	for resource, items := range mixingItems {
		constraint, err := nextroute.NewNoMixConstraint(items)
		if err != nil {
			return nil, err
		}
		constraint.SetID(fmt.Sprintf("no_mix_%v", resource))
		err = model.AddConstraint(constraint)
		if err != nil {
			return nil, err
		}
	}

	return model, nil
}

func addMixingItems(
	items map[string]map[nextroute.ModelStop]nextroute.MixItem,
	stop nextroute.ModelStop,
	parsed map[string]any,
) error {
	for t, anyValue := range parsed {
		if mixingItem, isMixingItem := anyValue.(map[string]any); isMixingItem {
			if _, ok := mixingItem["quantity"]; !ok {
				return fmt.Errorf(
					"stop `%v`, mixing type `%v` is missing quantity property, item is defined as `%v`",
					stop.ID(),
					t,
					mixingItem,
				)
			}
			quantity := reflect.ValueOf(mixingItem["quantity"])
			if quantity.Kind() != reflect.Float64 && quantity.Kind() != reflect.Int {
				return fmt.Errorf(
					"stop `%v`, mixing type `%v` the quantity property is not an int, the type is %v and value is `%v`",
					stop.ID(),
					t,
					quantity.Kind(),
					quantity,
				)
			}
			if math.Mod(mixingItem["quantity"].(float64), 1) != 0 {
				return fmt.Errorf(
					"stop `%v`, mixing type `%v` the quantity property is not an int, the type is %v and value is `%v`",
					stop.ID(),
					t,
					quantity.Kind(),
					quantity,
				)
			}
			if _, ok := mixingItem["name"]; !ok {
				return fmt.Errorf(
					"stop `%v`, mixing type `%v` is missing name property, item is defined as `%v`",
					stop.ID(),
					t,
					mixingItem,
				)
			}
			name := reflect.ValueOf(mixingItem["name"])
			if name.Kind() != reflect.String {
				return fmt.Errorf(
					"stop `%v`, mixing type `%v` the name property is not a string, the type is %v and value is `%v`",
					stop.ID(),
					t,
					name.Kind(),
					name,
				)
			}
			if _, ok := items[t]; !ok {
				items[t] = make(map[nextroute.ModelStop]nextroute.MixItem)
			}

			stopsItems := items[t]
			stopsItems[stop] = nextroute.MixItem{
				Name:     mixingItem["name"].(string),
				Quantity: int(mixingItem["quantity"].(float64)),
			}
		} else {
			return fmt.Errorf("stop `%v`, mixing type `%v` has incorrect mix item definition `%v`",
				stop.ID(),
				t,
				anyValue,
			)
		}
	}
	return nil
}
