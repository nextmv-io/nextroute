// © 2019-present nextmv.io inc

package factory

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nextmv-io/nextroute"
	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/nextroute/schema"
)

// capacityObjective is a capacity objective.
type capacityObjective struct {
	Name   string
	Factor float64
	Offset float64
}

func convertStringToFloat(v string) (float64, error) {
	s, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}
	return s, nil
}

func parseCapacityObjectives(capacities string) ([]capacityObjective, error) {
	if capacities == "" {
		return []capacityObjective{}, nil
	}

	capacityTokens := strings.Split(capacities, ";")
	if len(capacityTokens)%3 != 0 {
		return nil, nmerror.NewInputDataError(fmt.Errorf(
			"capacity objectives must be provided as triplets 'name=default;factor=1.0;offset=0.0'"+
				" separated by ';', provided string '%s' which has %d tokens",
			capacities,
			len(capacityTokens),
		))
	}

	nrResources := len(capacityTokens) / 3
	capacityObjectives := make([]capacityObjective, nrResources)

	for i := 0; i < nrResources; i++ {
		nameDefinition := capacityTokens[i*3]
		nameTokens := strings.Split(nameDefinition, "=")
		if len(nameTokens) != 2 || nameTokens[0] != "name" {
			return nil, nmerror.NewInputDataError(fmt.Errorf(
				"capacity objective '%s' is not a valid name definition,"+
					" should be in the form 'name=resource'",
				nameDefinition,
			))
		}
		name := nameTokens[1]

		factorDefinition := capacityTokens[i*3+1]
		factorTokens := strings.Split(factorDefinition, "=")
		if len(factorTokens) != 2 || factorTokens[0] != "factor" {
			return nil, nmerror.NewInputDataError(fmt.Errorf(
				"capacity objective '%s' for '%s' is not a valid factor definition,"+
					" should be in the form 'factor=1.0'",
				factorDefinition,
				name,
			))
		}
		factor, err := convertStringToFloat(factorTokens[1])
		if err != nil {
			return nil, nmerror.NewInputDataError(fmt.Errorf(
				"capacity objective factor '%s' for '%s' is not a valid float, %w",
				factorTokens[1],
				name,
				err,
			))
		}

		offsetDefinition := capacityTokens[i*3+2]
		offsetTokens := strings.Split(offsetDefinition, "=")
		if len(offsetTokens) != 2 || offsetTokens[0] != "offset" {
			return nil, nmerror.NewInputDataError(fmt.Errorf(
				"capacity objective '%s' for '%s' is not a valid offset definition,"+
					" should be in the form 'offset=0.0'",
				offsetDefinition,
				name,
			))
		}
		offset, err := convertStringToFloat(offsetTokens[1])
		if err != nil {
			return nil, nmerror.NewInputDataError(fmt.Errorf(
				"capacity objective offset '%s' for '%s' is not a valid float, %w",
				offsetTokens[1],
				name,
				err,
			))
		}

		capacityObjectives[i] = capacityObjective{
			Name:   name,
			Factor: factor,
			Offset: offset,
		}
	}
	return capacityObjectives, nil
}

// addCapacityObjective uses the capacity penalty from the capacity objective
// to create a new objective and add it to the model.
func addCapacityObjective(
	input schema.Input,
	model nextroute.Model,
	options Options,
) (nextroute.Model, error) {
	if options.Objectives.Capacities == "" {
		return model, nil
	}

	capacityObjectives, err := parseCapacityObjectives(options.Objectives.Capacities)
	if err != nil {
		return nil, err
	}
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

	model, names, quantityExpressions, capacityExpressions, err := addMaximumObjectives(
		model,
		names,
		options,
		capacityObjectives,
	)
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

// addMaximumConstraint is an auxiliary function to return the actual
// expressions that the Model uses. Because it receives and returns a Model, it
// mutates it to add maximum as a constraint.
func addMaximumObjectives(
	model nextroute.Model,
	names map[string]bool,
	options Options,
	capacityObjectives []capacityObjective,
) (
	nextroute.Model,
	map[string]bool,
	map[string]nextroute.StopExpression,
	map[string]nextroute.VehicleTypeValueExpression,
	error,
) {
	disabledResources := map[string]bool{}
	for _, name := range options.Constraints.Disable.Capacities {
		disabledResources[name] = true
	}

	requirements := map[string]nextroute.StopExpression{}
	limits := map[string]nextroute.VehicleTypeValueExpression{}
	postedNames := map[string]bool{}

	for _, capacityObjective := range capacityObjectives {
		if capacityObjective.Factor == 0 {
			continue
		}
		if !options.Constraints.Disable.Capacity && !disabledResources[capacityObjective.Name] {
			if _, ok := names[capacityObjective.Name]; !ok {
				return nil, nil, nil, nil, nmerror.NewInputDataError(fmt.Errorf(
					"capacity objective '%s' does not match any resource, quantity, capacity or start level",
					capacityObjective.Name,
				))
			}
			continue
		}

		name := capacityObjective.Name
		if _, ok := names[name]; !ok {
			return nil, nil, nil, nil, nmerror.NewInputDataError(fmt.Errorf(
				"capacity objective '%s' does not match any resource, quantity, capacity or start level",
				name,
			))
		}
		postedNames[name] = true
		requirement := nextroute.NewStopExpression(name, 0.)
		limit := nextroute.NewVehicleTypeValueExpression(name, 0.)
		maximum, err := nextroute.NewMaximum(requirement, limit)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		maximum.(nextroute.Identifier).SetID("capacity_" + name)
		_, err = model.Objective().NewTerm(capacityObjective.Factor, maximum)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		err = maximum.SetPenaltyOffset(capacityObjective.Offset)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		requirements[name] = requirement
		limits[name] = limit
	}

	return model, postedNames, requirements, limits, nil
}
