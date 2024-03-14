// © 2019-present nextmv.io inc

package factory

import (
	"fmt"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/nextroute/schema"
)

// addServiceDurations sets the time it takes them to service a stop.
func addServiceDurations(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	durationExpressions := common.UniqueDefined(
		common.Map(
			model.VehicleTypes(),
			func(vt nextroute.ModelVehicleType) nextroute.DurationExpression {
				return vt.DurationExpression()
			}),
		func(e nextroute.DurationExpression) int {
			return e.Index()
		},
	)

	err := addServiceDurationsStops(input, model, durationExpressions)
	if err != nil {
		return nil, err
	}

	err = addServiceDurationsAlternateStops(input, model, durationExpressions)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func addServiceDurationsStops(
	input schema.Input,
	model nextroute.Model,
	durationExpressions []nextroute.DurationExpression) error {
	for s, inputStop := range input.Stops {
		if inputStop.Duration == nil || *inputStop.Duration == 0 {
			continue
		}

		stop, err := model.Stop(s)
		if err != nil {
			return err
		}

		for _, durationExpression := range durationExpressions {
			durationGroupsExpression, ok := durationExpression.(DurationGroupsExpression)
			if !ok {
				return fmt.Errorf("process duration expression %s is not a duration group expression",
					durationExpression.Name(),
				)
			}

			durationGroupsExpression.SetStopDuration(
				stop, time.Duration(*inputStop.Duration)*time.Second,
			)
		}
	}
	return nil
}

func addServiceDurationsAlternateStops(
	input schema.Input,
	model nextroute.Model,
	durationExpressions []nextroute.DurationExpression) error {
	if input.AlternateStops == nil {
		return nil
	}

	data, err := getModelData(model)
	if err != nil {
		return err
	}

	for _, inputVehicle := range input.Vehicles {
		if inputVehicle.AlternateStops == nil {
			continue
		}

		for _, alternateID := range *inputVehicle.AlternateStops {
			stop, err := model.Stop(data.stopIDToIndex[alternateStopID(alternateID, inputVehicle)])
			if err != nil {
				return err
			}

			alternateInputStop := stop.Data().(alternateInputStop)

			if alternateInputStop.stop.Duration == nil {
				continue
			}

			for _, durationExpression := range durationExpressions {
				durationGroupsExpression, ok := durationExpression.(DurationGroupsExpression)
				if !ok {
					return fmt.Errorf("process duration expression %s is not a duration group expression",
						durationExpression.Name(),
					)
				}

				durationGroupsExpression.SetStopDuration(
					stop, time.Duration(*alternateInputStop.stop.Duration)*time.Second,
				)
			}
		}
	}

	return nil
}

func groupToStops(ids []string, model nextroute.Model) (nextroute.ModelStops, error) {
	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}

	modelStops := make(nextroute.ModelStops, len(ids))
	for idx, id := range ids {
		index, ok := data.stopIDToIndex[id]
		if !ok {
			return nil, nmerror.NewInputDataError(fmt.Errorf("group contains id %s that is not known in the list of stops", id))
		}
		modelStops[idx], err = model.Stop(index)
		if err != nil {
			return nil, nmerror.NewInputDataError(fmt.Errorf("group contains id %s that is not known in the list of stops", id))
		}
	}
	return modelStops, nil
}

// addDurationGroups sets the time it takes them to service a stop.
func addDurationGroups(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	if input.DurationGroups == nil {
		return model, nil
	}
	durationExpressions := common.UniqueDefined(
		common.Map(
			model.VehicleTypes(),
			func(vt nextroute.ModelVehicleType) nextroute.DurationExpression {
				return vt.DurationExpression()
			}),
		func(e nextroute.DurationExpression) int {
			return e.Index()
		},
	)

	for _, durationGroup := range *input.DurationGroups {
		if durationGroup.Duration == 0 {
			continue
		}
		modelStops, err := groupToStops(durationGroup.Group, model)
		if err != nil {
			return nil, err
		}
		for _, durationExpression := range durationExpressions {
			stopDurationExpression, ok := durationExpression.(DurationGroupsExpression)
			if !ok {
				return nil,
					fmt.Errorf("process duration expression %s is not a duration group expression",
						durationExpression.Name(),
					)
			}

			err := stopDurationExpression.AddGroup(
				modelStops,
				time.Duration(durationGroup.Duration)*time.Second,
			)
			if err != nil {
				return nil, err
			}
		}
	}
	return model, nil
}

// addServiceDurations sets the time it takes them to service a stop.
func addDurationMultipliers(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	// containerType struct with a field called durationExpression and a field called
	// inputVehicle
	type containerType struct {
		durationExpression nextroute.DurationExpression
		multiplier         float64
	}

	container :=
		common.Map(
			model.VehicleTypes(),
			func(vt nextroute.ModelVehicleType) containerType {
				multiplier := 1.0
				if input.Vehicles[vt.Index()].StopDurationMultiplier != nil {
					multiplier = *input.Vehicles[vt.Index()].StopDurationMultiplier
				}
				return containerType{
					durationExpression: vt.DurationExpression(),
					multiplier:         multiplier,
				}
			})

	for _, element := range container {
		durationGroupsExpression, ok := element.durationExpression.(DurationGroupsExpression)
		if !ok {
			return nil,
				fmt.Errorf("process duration expression %s is not a duration group expression",
					element.durationExpression.Name(),
				)
		}

		// multiply the durations
		for stop, value := range durationGroupsExpression.Durations() {
			if value == 0 {
				continue
			}
			durationGroupsExpression.SetStopDuration(
				stop, time.Second*time.Duration(value.Seconds()*element.multiplier),
			)
		}

		// multiply the group durations
		groups := durationGroupsExpression.Groups()
		for _, group := range groups {
			if group.Duration == 0 {
				continue
			}
			err := durationGroupsExpression.SetGroupDuration(
				group.Stops,
				time.Second*time.Duration(group.Duration.Seconds()*element.multiplier),
			)
			if err != nil {
				return nil, err
			}
		}
	}

	return model, nil
}
