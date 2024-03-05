// © 2019-present nextmv.io inc

package factory

import (
	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/schema"
)

// modelModifier is a function that modifies a Model. The Model that is passed
// as an argument is mutated and returned. This allows developers to
// encapsulate Model-modifying logic individually in easy to digest functions.
type modelModifier func(schema.Input, nextroute.Model, Options) (nextroute.Model, error)

// NewModel is the implementation of NewModel.
func NewModel(
	input schema.Input,
	modelOptions Options,
) (nextroute.Model, error) {
	input = applyDefaults(input)
	err := validate(input, modelOptions)
	if err != nil {
		return nil, err
	}

	model, err := nextroute.NewModel()
	if err != nil {
		return nil, err
	}

	for _, modifier := range getModifiersFromOptions(modelOptions) {
		if model, err = modifier(input, model, modelOptions); err != nil {
			return nil, err
		}
	}

	return model, nil
}

func getModifiersFromOptions(options Options) []modelModifier {
	modifiers := []modelModifier{addStops, addAlternates, addVehicles}
	modifiers = appendConstraintModifiers(options, modifiers)
	modifiers = appendObjectiveModifiers(options, modifiers)
	modifiers = appendPropertiesModifiers(options, modifiers)
	modifiers = append(modifiers, addPlanUnits)

	if !options.Properties.Disable.InitialSolution {
		modifiers = append(modifiers, addInitialSolution)
	}

	return modifiers
}

func appendConstraintModifiers(
	options Options,
	modifiers []modelModifier,
) []modelModifier {
	if options.Constraints.Enable.Cluster {
		modifiers = append(modifiers, addClusterConstraint)
	}

	if !options.Constraints.Disable.Attributes {
		modifiers = append(modifiers, addAttributesConstraint)
	}

	if !options.Constraints.Disable.Capacity {
		modifiers = append(modifiers, addCapacityConstraint)
	}

	if !options.Constraints.Disable.DistanceLimit {
		modifiers = append(modifiers, addDistanceLimitConstraint)
	}

	if !options.Constraints.Disable.MaximumDuration {
		modifiers = append(modifiers, addMaximumDurationConstraint)
	}

	if !options.Constraints.Disable.Precedence {
		modifiers = append(modifiers, addPrecedenceInformation)
	}

	if !options.Constraints.Disable.Groups {
		modifiers = append(modifiers, addGroupInformation)
	}

	if !options.Constraints.Disable.VehicleEndTime {
		modifiers = append(modifiers, addVehicleEndTimeConstraint)
	}

	if !options.Constraints.Disable.StartTimeWindows {
		modifiers = append(modifiers, addWindowsConstraint)
	}

	if !options.Constraints.Disable.MaximumStops {
		modifiers = append(modifiers, addMaximumStopsConstraint)
	}

	if !options.Constraints.Disable.MaximumWaitStop {
		modifiers = append(modifiers, addMaximumWaitStopConstraint)
	}

	if !options.Constraints.Disable.MaximumWaitVehicle {
		modifiers = append(modifiers, addMaximumWaitVehicleConstraint)
	}

	if !options.Constraints.Disable.MixingItems {
		modifiers = append(modifiers, addNoMixConstraint)
	}

	return modifiers
}

func appendObjectiveModifiers(
	options Options,
	modifiers []modelModifier,
) []modelModifier {
	if options.Objectives.VehicleActivationPenalty > 0.0 {
		modifiers = append(modifiers, addActivationPenaltyObjective)
	}

	if options.Objectives.TravelDuration > 0.0 {
		modifiers = append(modifiers, addTravelDurationObjective)
	}

	if options.Objectives.VehiclesDuration > 0.0 {
		modifiers = append(modifiers, addVehiclesDurationObjective)
	}

	if options.Objectives.UnplannedPenalty > 0.0 {
		modifiers = append(modifiers, addUnplannedObjective)
	}

	if options.Objectives.EarlyArrivalPenalty > 0.0 {
		modifiers = append(modifiers, addEarlinessObjective)
	}

	if options.Objectives.LateArrivalPenalty > 0.0 {
		modifiers = append(modifiers, addLatenessObjective)
	}

	if options.Objectives.Cluster > 0.0 {
		modifiers = append(modifiers, addClusterObjective)
	}

	if options.Objectives.MinStops > 0.0 {
		modifiers = append(modifiers, addMinStopsObjective)
	}

	if len(options.Objectives.Capacities) > 0 {
		modifiers = append(modifiers, addCapacityObjective)
	}

	return modifiers
}

func appendPropertiesModifiers(
	options Options,
	modifiers []modelModifier,
) []modelModifier {
	if !options.Properties.Disable.Durations {
		modifiers = append(modifiers, addServiceDurations)
	}
	if !options.Properties.Disable.DurationGroups {
		modifiers = append(modifiers, addDurationGroups)
	}
	if !options.Properties.Disable.StopDurationMultipliers {
		modifiers = append(modifiers, addDurationMultipliers)
	}
	return modifiers
}
