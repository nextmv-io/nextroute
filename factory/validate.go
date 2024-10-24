// © 2019-present nextmv.io inc

package factory

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/nextmv-io/nextroute/common"
	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/nextroute/schema"
)

// validate the input and return an error if invalid.
func validate(input schema.Input, modelOptions Options) error {
	allStopIDs := map[string]bool{}
	stopIDs := map[string]bool{}
	alternateStopIDs := map[string]bool{}

	for idx, stop := range input.Stops {
		if stop.ID == "" {
			return nmerror.NewInputDataError(fmt.Errorf("no id set for stop at index %v", idx))
		}
		allStopIDs[stop.ID] = true
		stopIDs[stop.ID] = true
	}
	duplicates := common.NotUniqueDefined(
		input.Stops,
		func(s schema.Stop) string {
			return s.ID
		},
	)
	if len(duplicates) != 0 {
		return nmerror.NewInputDataError(fmt.Errorf(
			"stop ID's are not unique, duplicate ID's are [`%v`]",
			strings.Join(
				common.Map(duplicates, func(s schema.Stop) string {
					return s.ID
				}),
				"`, `",
			)))
	}

	if input.AlternateStops != nil {
		alternateDuplicates := common.NotUniqueDefined(
			*input.AlternateStops,
			func(s schema.AlternateStop) string {
				return s.ID
			},
		)
		if len(alternateDuplicates) != 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"alternate stop ID's are not unique, duplicate ID's are [`%v`]",
				strings.Join(
					common.Map(alternateDuplicates, func(s schema.AlternateStop) string {
						return s.ID
					}),
					"`, `",
				)))
		}
		for idx, stop := range *input.AlternateStops {
			if stop.ID == "" {
				return nmerror.NewInputDataError(fmt.Errorf("empty id set for alternate stop at index %v", idx))
			}
			allStopIDs[stop.ID] = true
			alternateStopIDs[stop.ID] = true
		}
	}

	if err := validateVehicles(input, allStopIDs); err != nil {
		return err
	}
	if err := validateStops(input, allStopIDs, stopIDs, alternateStopIDs); err != nil {
		return err
	}
	if err := validateResources(input, modelOptions); err != nil {
		return err
	}
	return validateConstraints(input, modelOptions)
}

func identify(input schema.Input, i int) string {
	if i < len(input.Stops) {
		return input.Stops[i].ID
	}
	idx := i - len(input.Stops)
	if idx%2 == 0 {
		return fmt.Sprintf("start %s", input.Vehicles[idx/2].ID)
	}
	return fmt.Sprintf("end %s", input.Vehicles[idx/2].ID)
}

func location(input schema.Input, i int) common.Location {
	if i < len(input.Stops) {
		l, _ := common.NewLocation(
			input.Stops[i].Location.Lon,
			input.Stops[i].Location.Lat,
		)
		return l
	}
	idx := i - len(input.Stops)
	vehicle := input.Vehicles[idx/2]

	if idx%2 == 0 {
		if vehicle.StartLocation == nil {
			return common.NewInvalidLocation()
		}
		l, _ := common.NewLocation(
			vehicle.StartLocation.Lon,
			vehicle.StartLocation.Lat,
		)
		return l
	}
	if vehicle.EndLocation == nil {
		return common.NewInvalidLocation()
	}
	l, _ := common.NewLocation(
		vehicle.EndLocation.Lon,
		vehicle.EndLocation.Lat,
	)
	return l
}

func validateMatrix(
	input schema.Input,
	matrix [][]float64,
	asymmetryTolerance int,
	preFix string) error {
	size := len(input.Stops) + len(input.Vehicles)*2
	if len(matrix) != size {
		return nmerror.NewInputDataError(fmt.Errorf(
			"%s matrix length (%v)"+
				" does not match number of stops (%v) plus number of vehicles (%v) times 2",
			preFix,
			len(matrix),
			len(input.Stops),
			len(input.Vehicles),
		))
	}
	for i := 0; i < size; i++ {
		if len(matrix[i]) != size {
			return nmerror.NewInputDataError(fmt.Errorf(
				"%s matrix row %v length (%v)"+
					" does not match number of stops (%v) plus number of vehicles (%v) times 2",
				preFix,
				i,
				len(matrix[i]),
				len(input.Stops),
				len(input.Vehicles),
			))
		}
	}

	var asymmetries []string
	for i := 0; i < size; i++ {
		for j := i + 1; j < size; j++ {
			iLocation := location(input, i)
			jLocation := location(input, j)

			if !iLocation.IsValid() || !jLocation.IsValid() {
				continue
			}

			iID := identify(input, i)
			jID := identify(input, j)

			// Check if the matrix is negative
			if matrix[i][j] < 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"%s matrix has negative value %v for stops `%s` and `%s`",
					preFix,
					matrix[i][j],
					iID,
					jID,
				))
			}
			if matrix[j][i] < 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"%s matrix has negative value %v for stops `%s` and `%s`",
					preFix,
					matrix[j][i],
					iID,
					jID,
				))
			}
			// Check if cells with zero have the same location
			if matrix[i][j] == 0 {
				if !reflect.DeepEqual(iLocation, jLocation) {
					return nmerror.NewInputDataError(fmt.Errorf(
						"%s is zero for stop `%s`[%v] to `%s`[%v] at different locations",
						preFix,
						iID,
						iLocation,
						jID,
						jLocation,
					))
				}
			}
			// Check if the duration matrix is symmetric within tolerance
			diff := math.Abs(matrix[i][j]-matrix[j][i]) / ((matrix[i][j] + matrix[j][i]) / 2.0) * 100.0
			if diff > float64(asymmetryTolerance) {
				asymmetries = append(asymmetries, fmt.Sprintf(
					"`%s` to `%s` is %v, reverse is %v, difference is %.2f percent",
					iID,
					jID,
					matrix[i][j],
					matrix[j][i],
					diff),
				)
				if len(asymmetries) > 10 {
					return nmerror.NewInputDataError(fmt.Errorf(
						"%s matrix has too many asymmetries larger than %v percent, first 10 are `%s`",
						preFix,
						asymmetryTolerance,
						strings.Join(asymmetries, "`, `"),
					))
				}
			}
		}
	}
	if len(asymmetries) > 0 {
		return nmerror.NewInputDataError(fmt.Errorf(
			"%s matrix has too many asymmetries larger than %v percent, `%s`",
			preFix,
			asymmetryTolerance,
			strings.Join(asymmetries, "`, `"),
		))
	}
	return nil
}

func validateConstraints(input schema.Input, modelOptions Options) error {
	if !modelOptions.Validate.Disable.StartTime {
		hasStartTimeWindow := common.Has(
			input.Stops,
			true,
			func(s schema.Stop) bool {
				return s.StartTimeWindow != nil
			},
		)
		vehiclesHaveStartTime := common.Filter(
			input.Vehicles,
			func(v schema.Vehicle) bool {
				return v.StartTime != nil
			},
		)

		if hasStartTimeWindow && len(vehiclesHaveStartTime) != len(input.Vehicles) {
			return nmerror.NewInputDataError(fmt.Errorf(
				"there are stops with a start_time_window but not all vehicles have start_time," +
					" if intended use validate option to disable this start time validation" +
					" (`options.Model.Validate.Disable.StartTime = true`)",
			))
		}
	}

	if input.DistanceMatrix != nil && modelOptions.Validate.Enable.Matrix {
		distanceMatrix := *input.DistanceMatrix
		if err := validateMatrix(
			input,
			distanceMatrix,
			modelOptions.Validate.Enable.MatrixAsymmetryTolerance,
			"distance"); err != nil {
			return err
		}
	}

	if input.DurationMatrix == nil {
		return nil
	}
	switch matrix := input.DurationMatrix.(type) {
	case [][]float64:
		if modelOptions.Validate.Enable.Matrix {
			if err := validateMatrix(
				input,
				matrix,
				modelOptions.Validate.Enable.MatrixAsymmetryTolerance,
				"duration"); err != nil {
				return err
			}
		}
	case schema.TimeDependentMatrix:
		return validateTimeDependentMatrix(input, matrix, modelOptions, true)
	case []schema.TimeDependentMatrix:
		return validateTimeDependentMatricesAndIDs(input, matrix, modelOptions)
	case map[string]any:
		timeDependentMatrix, err := convertToTimeDependentMatrix(matrix)
		if err != nil {
			return err
		}
		return validateTimeDependentMatrix(input, timeDependentMatrix, modelOptions, true)
	case []any:
		// In this case we have a single matrix that can be a float64 matrix or a
		// multi duration matrix.
		return validateFloatOrMultiDurationMatrix(input, matrix, modelOptions)
	default:
		return nmerror.NewInputDataError(fmt.Errorf("invalid duration matrix type %T", matrix))
	}
	return nil
}

func validateFloatOrMultiDurationMatrix(input schema.Input, matrix []any, modelOptions Options) error {
	if floatMatrix, ok := common.TryAssertFloat64Matrix(matrix); ok {
		if modelOptions.Validate.Enable.Matrix {
			return validateMatrix(
				input,
				floatMatrix,
				modelOptions.Validate.Enable.MatrixAsymmetryTolerance,
				"duration")
		}
		return nil
	}

	timeDependentMatrices, err := convertToTimeDependentMatrices(matrix)
	if err != nil {
		return err
	}
	return validateTimeDependentMatricesAndIDs(input, timeDependentMatrices, modelOptions)
}

// Converts a single or multiple time-dependent matrices from a slice of interfaces to []schema.TimeDependentMatrix.
func convertToTimeDependentMatrices(data []any) ([]schema.TimeDependentMatrix, error) {
	var result []schema.TimeDependentMatrix

	for _, item := range data {
		if matrixMap, ok := item.(map[string]any); ok {
			matrix, err := convertToTimeDependentMatrix(matrixMap)
			if err != nil {
				return nil, err
			}
			result = append(result, matrix)
		} else {
			return nil, fmt.Errorf("invalid time-dependent matrix format")
		}
	}

	return result, nil
}

func validateTimeDependentMatricesAndIDs(
	input schema.Input,
	timeDependentMatrices []schema.TimeDependentMatrix,
	modelOptions Options,
) error {
	vIDs := make(map[string]bool)
	for _, durationMatrix := range timeDependentMatrices {
		if durationMatrix.VehicleIDs == nil || len(durationMatrix.VehicleIDs) == 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"vehicle ids are not set for duration matrix",
			))
		}
		for _, vID := range durationMatrix.VehicleIDs {
			if _, ok := vIDs[vID]; ok {
				return nmerror.NewInputDataError(fmt.Errorf(
					"duplicate vehicle id in duration matrices: %s",
					durationMatrix.VehicleIDs,
				))
			}
			vIDs[vID] = true
		}
		if err := validateTimeDependentMatrix(input, durationMatrix, modelOptions, false); err != nil {
			return err
		}
	}

	// Make sure all vehicles in the input have a duration matrix defined.
	for _, vehicle := range input.Vehicles {
		if _, ok := vIDs[vehicle.ID]; !ok {
			return nmerror.NewInputDataError(fmt.Errorf(
				"vehicle id %s is not defined in duration matrices",
				vehicle.ID,
			))
		}
	}

	// Make sure there is no definition in the input that does not exist as a vehicle.
	if len(vIDs) != len(input.Vehicles) {
		return nmerror.NewInputDataError(fmt.Errorf(
			"vehicle ids in duration matrices do not match vehicle ids in input: %v",
			vIDs,
		))
	}
	return nil
}

func validateStop(idx int, stop schema.Stop, stopIDs map[string]bool) error {
	if stop.ID == "" {
		return nmerror.NewInputDataError(fmt.Errorf("no id set for stop at index %v", idx))
	}

	if stop.StartTimeWindow != nil {
		_, err := convertTimeWindow(stop.StartTimeWindow, stop.ID)
		if err != nil {
			return err
		}
	}

	if stop.MaxWait != nil {
		maxWait := *stop.MaxWait
		if maxWait < 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"stop `%s` max wait must be non-negative, it is `%v` seconds",
				stop.ID,
				maxWait,
			))
		}
	}

	if stop.Duration != nil {
		duration := *stop.Duration
		if duration < 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"stop `%s` duration must be non-negative, it is `%v` seconds",
				stop.ID,
				duration,
			))
		}
	}

	if stop.UnplannedPenalty != nil {
		unplannedPenalty := *stop.UnplannedPenalty
		if unplannedPenalty < 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"stop `%s` unplanned penalty must be non-negative, it is `%v`",
				stop.ID,
				unplannedPenalty,
			))
		}
	}

	if stop.EarlyArrivalTimePenalty != nil {
		earlyArrivalTimePenalty := *stop.EarlyArrivalTimePenalty
		if earlyArrivalTimePenalty < 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"stop `%s` early arrival time penalty must be non-negative, it is `%v`",
				stop.ID,
				earlyArrivalTimePenalty,
			))
		}
	}

	if stop.LateArrivalTimePenalty != nil {
		lateArrivalTimePenalty := *stop.LateArrivalTimePenalty
		if lateArrivalTimePenalty < 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"stop `%s` late arrival time penalty must be non-negative, it is `%v`",
				stop.ID,
				lateArrivalTimePenalty,
			))
		}
	}

	if stop.CompatibilityAttributes != nil {
		compatibilityAttributes := *stop.CompatibilityAttributes
		duplicateAttributes := common.NotUnique(compatibilityAttributes)
		if len(duplicateAttributes) != 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"stop `%s` has duplicate compatibility attributes, duplicates are [`%s`]",
				stop.ID,
				strings.Join(duplicateAttributes, "`, `"),
			))
		}
	}

	if reflect.DeepEqual(stop.Location, schema.Location{}) {
		return nmerror.NewInputDataError(fmt.Errorf("stop `%s` has no location", stop.ID))
	}

	if _, err := common.NewLocation(
		stop.Location.Lon,
		stop.Location.Lat,
	); err != nil {
		return nmerror.NewInputDataError(fmt.Errorf(
			"stop `%s` location is invalid: %w",
			stop.ID,
			err,
		))
	}

	precedes, err := precedence(stop, "Precedes")
	if err != nil {
		return err
	}

	succeeds, err := precedence(stop, "Succeeds")
	if err != nil {
		return err
	}

	for _, p := range precedes {
		if !stopIDs[p.id] {
			return nmerror.NewInputDataError(fmt.Errorf(
				"stop `%s` precedes references unknown stop %s",
				stop.ID,
				p.id,
			))
		}

		if stop.Precedes == stop.ID {
			return nmerror.NewInputDataError(fmt.Errorf(
				"stop `%s` precedes itself",
				stop.ID,
			))
		}
	}
	for _, s := range succeeds {
		if !stopIDs[s.id] {
			return nmerror.NewInputDataError(fmt.Errorf(
				"stop `%s` succeeds references unknown stop %s",
				stop.ID,
				s.id,
			))
		}

		if stop.Succeeds == stop.ID {
			return nmerror.NewInputDataError(fmt.Errorf(
				"stop `%s` succeeds itself",
				stop.ID,
			))
		}
	}

	return nil
}

func validateAlternateStop(idx int, stop schema.AlternateStop) error {
	if stop.ID == "" {
		return nmerror.NewInputDataError(fmt.Errorf("no id set for alternate stop at index %v", idx))
	}

	if stop.StartTimeWindow != nil {
		_, err := convertTimeWindow(stop.StartTimeWindow, stop.ID)
		if err != nil {
			return err
		}
	}

	if stop.MaxWait != nil {
		maxWait := *stop.MaxWait
		if maxWait < 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"alternate stop `%s` max wait must be non-negative, it is `%v` seconds",
				stop.ID,
				maxWait,
			))
		}
	}

	if stop.Duration != nil {
		duration := *stop.Duration
		if duration < 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"alternate stop `%s` duration must be non-negative, it is `%v` seconds",
				stop.ID,
				duration,
			))
		}
	}

	if stop.UnplannedPenalty != nil {
		unplannedPenalty := *stop.UnplannedPenalty
		if unplannedPenalty < 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"alternate stop `%s` unplanned penalty must be non-negative, it is `%v`",
				stop.ID,
				unplannedPenalty,
			))
		}
	}

	if stop.EarlyArrivalTimePenalty != nil {
		earlyArrivalTimePenalty := *stop.EarlyArrivalTimePenalty
		if earlyArrivalTimePenalty < 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"alternate stop `%s` early arrival time penalty must be non-negative, it is `%v`",
				stop.ID,
				earlyArrivalTimePenalty,
			))
		}
	}

	if stop.LateArrivalTimePenalty != nil {
		lateArrivalTimePenalty := *stop.LateArrivalTimePenalty
		if lateArrivalTimePenalty < 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"alternate stop `%s` late arrival time penalty must be non-negative, it is `%v`",
				stop.ID,
				lateArrivalTimePenalty,
			))
		}
	}

	if reflect.DeepEqual(stop.Location, schema.Location{}) {
		return nmerror.NewInputDataError(fmt.Errorf("alternate stop `%s` has no location", stop.ID))
	}

	if _, err := common.NewLocation(
		stop.Location.Lon,
		stop.Location.Lat,
	); err != nil {
		return nmerror.NewInputDataError(fmt.Errorf(
			"stop `%s` location is invalid: %w",
			stop.ID,
			err,
		))
	}

	return nil
}

func validateTimeDependentMatrix(
	input schema.Input,
	durationMatrices schema.TimeDependentMatrix,
	modelOptions Options,
	isSingleMatrix bool,
) error {
	if modelOptions.Validate.Enable.Matrix {
		if err := validateMatrix(
			input,
			durationMatrices.DefaultMatrix,
			modelOptions.Validate.Enable.MatrixAsymmetryTolerance,
			"time_dependent_duration"); err != nil {
			return err
		}
	}
	if isSingleMatrix {
		if len(durationMatrices.VehicleIDs) != 0 {
			return nmerror.NewInputDataError(fmt.Errorf(
				"single matrix has vehicle ids set, it must be empty"))
		}
	}
	for i, tf := range durationMatrices.MatrixTimeFrames {
		if tf.Matrix == nil && tf.ScalingFactor == nil {
			return nmerror.NewInputDataError(fmt.Errorf(
				"duration for time frame %d is missing both matrix and scaling factor", i))
		}

		if tf.Matrix != nil && tf.ScalingFactor != nil {
			return nmerror.NewInputDataError(fmt.Errorf(
				"duration for time frame %d has both matrix and scaling factor, only one is allowed", i))
		}

		if tf.Matrix != nil && modelOptions.Validate.Enable.Matrix {
			if err := validateMatrix(
				input,
				tf.Matrix,
				modelOptions.Validate.Enable.MatrixAsymmetryTolerance,
				fmt.Sprintf("time_dependent_duration for time frame %d", i),
			); err != nil {
				return err
			}
		}

		if tf.ScalingFactor != nil {
			if *tf.ScalingFactor <= 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"time_dependent_duration for time frame %d has invalid scaling factor %v", i, *tf.ScalingFactor))
			}
		}

		if tf.StartTime.IsZero() {
			return nmerror.NewInputDataError(fmt.Errorf(
				"time_dependent_duration for time frame %d has no start time", i))
		}
		if tf.EndTime.IsZero() {
			return nmerror.NewInputDataError(fmt.Errorf(
				"time_dependent_duration for time frame %d has no end time", i))
		}
		if tf.StartTime.After(tf.EndTime) || tf.StartTime.Equal(tf.EndTime) {
			return nmerror.NewInputDataError(fmt.Errorf(
				"time_dependent_duration for time frame %d has invalid start and end time, "+
					"start time is after or equal to end time", i))
		}
	}
	return nil
}

func validateStops(
	input schema.Input,
	allStopIDs map[string]bool,
	stopIDs map[string]bool,
	alternateStopIDs map[string]bool) error {
	if len(input.Stops) == 0 {
		return errors.New("no stops provided")
	}

	for idx, stop := range input.Stops {
		err := validateStop(idx, stop, allStopIDs)
		if err != nil {
			return err
		}
	}

	if input.AlternateStops != nil {
		for idx, stop := range *input.AlternateStops {
			err := validateAlternateStop(idx, stop)
			if err != nil {
				return err
			}
		}
	}

	if input.StopGroups != nil {
		stopGroups := *input.StopGroups
		for i, stopGroup := range stopGroups {
			duplicateStops := common.NotUnique(stopGroup)
			if len(duplicateStops) != 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"stop group at index %d has duplicate stops, duplicates are [`%s`]",
					i,
					strings.Join(duplicateStops, "`, `"),
				))
			}
			for _, id := range stopGroup {
				if alternateStopIDs[id] {
					return nmerror.NewInputDataError(fmt.Errorf("stop group at index %d references an alternate stop `%s`,"+
						" alternate stops can not be used in stop groups",
						i,
						id,
					))
				}
				if !stopIDs[id] {
					return nmerror.NewInputDataError(fmt.Errorf("stop group at index %d references an unknown stop `%s`",
						i,
						id,
					))
				}
			}
		}
	}

	return nil
}

func validateVehicles(input schema.Input, stopIDs map[string]bool) error {
	if len(input.Vehicles) == 0 {
		return errors.New("no vehicles provided")
	}
	duplicates := common.NotUniqueDefined(
		input.Vehicles,
		func(v schema.Vehicle) string {
			return v.ID
		},
	)
	if len(duplicates) != 0 {
		return nmerror.NewInputDataError(fmt.Errorf(
			"vehicle ID's are not unique, duplicate ID's are [`%v`]",
			strings.Join(
				common.Map(duplicates, func(v schema.Vehicle) string {
					return v.ID
				}),
				"`, `",
			)))
	}

	for idx, vehicle := range input.Vehicles {
		if vehicle.ID == "" {
			return nmerror.NewInputDataError(fmt.Errorf("no id set for vehicle at index %v", idx))
		}

		if input.DurationMatrix == nil && vehicle.Speed == nil {
			return nmerror.NewInputDataError(fmt.Errorf(
				"vehicle `%s` no duration matrix and no speed set,"+
					" requires speed to determine duration based on distance",
				vehicle.ID,
			))
		}

		if vehicle.StartLocation != nil {
			startLocation := *vehicle.StartLocation
			if _, err := common.NewLocation(
				startLocation.Lon,
				startLocation.Lat,
			); err != nil {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` start location is invalid: %w",
					vehicle.ID,
					err,
				))
			}
		}
		if vehicle.EndLocation != nil {
			endLocation := *vehicle.EndLocation
			if _, err := common.NewLocation(
				endLocation.Lon,
				endLocation.Lat,
			); err != nil {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` end location is invalid: %w",
					vehicle.ID,
					err,
				))
			}
		}
		if vehicle.Speed != nil {
			speed := *vehicle.Speed
			if speed <= 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` speed must be greater than 0, it is %v meters per second",
					vehicle.ID,
					speed,
				))
			}
		}

		if vehicle.StartTime != nil {
			startTime := *vehicle.StartTime
			if vehicle.EndTime != nil {
				endTime := *vehicle.EndTime
				if startTime.After(endTime) {
					return nmerror.NewInputDataError(fmt.Errorf(
						"vehicle `%s` start time `%v` is %v after end time `%v`",
						vehicle.ID,
						startTime,
						startTime.Sub(endTime),
						endTime,
					))
				}
			}
		}

		if vehicle.MaxStops != nil {
			maxStops := *vehicle.MaxStops
			if maxStops < 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` maximum stops must be non-negative, it is %v",
					vehicle.ID,
					maxStops,
				))
			}
		}

		if vehicle.MaxDistance != nil {
			maxDistance := *vehicle.MaxDistance
			if maxDistance < 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` maximum distance must be non-negative, it is %v meters",
					vehicle.ID,
					maxDistance,
				))
			}
		}

		if vehicle.MaxDuration != nil {
			maxDuration := *vehicle.MaxDuration
			if maxDuration < 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s`maximum duration must be non-negative, it is %v seconds",
					vehicle.ID,
					maxDuration,
				))
			}
		}

		if vehicle.MaxWait != nil {
			maxWait := *vehicle.MaxWait
			if maxWait < 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` maximum wait must be non-negative, it is %v seconds",
					vehicle.ID,
					maxWait,
				))
			}
		}

		if vehicle.CompatibilityAttributes != nil {
			compatibilityAttributes := *vehicle.CompatibilityAttributes
			duplicateAttributes := common.NotUnique(compatibilityAttributes)
			if len(duplicateAttributes) != 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` has duplicate compatibility attributes, duplicates are [`%s`]",
					vehicle.ID,
					strings.Join(duplicateAttributes, "`, `"),
				))
			}
		}

		if vehicle.InitialStops != nil {
			initialStops := *vehicle.InitialStops
			for i, initialStop := range initialStops {
				if initialStop.ID == "" {
					return nmerror.NewInputDataError(fmt.Errorf(
						"vehicle `%s` no id set for initial stop at index %v",
						vehicle.ID,
						i))
				}
			}
			duplicateInitialStops := common.NotUniqueDefined(
				initialStops,
				func(s schema.InitialStop) string {
					return s.ID
				},
			)
			if len(duplicateInitialStops) != 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` has duplicate initial stops, duplicates are [`%s`]",
					vehicle.ID,
					strings.Join(
						common.Map(duplicateInitialStops, func(s schema.InitialStop) string {
							return s.ID
						}),
						"`, `",
					),
				))
			}
		}

		if vehicle.InitialStops != nil {
			initialStops := *vehicle.InitialStops
			for _, stop := range initialStops {
				if _, ok := stopIDs[stop.ID]; !ok {
					return nmerror.NewInputDataError(fmt.Errorf(
						"vehicle `%s` initial stop `%s` does not exist",
						vehicle.ID,
						stop.ID,
					))
				}
			}
			if vehicle.AlternateStops != nil {
				alternateInitialStops := common.Intersect(
					common.Map(initialStops, func(s schema.InitialStop) string {
						return s.ID
					}),
					*vehicle.AlternateStops,
				)

				if len(alternateInitialStops) > 1 {
					return nmerror.NewInputDataError(fmt.Errorf(
						"vehicle `%s` has multiple initial stops that are alternate stops, only one allowed, initial stops are [`%s`]",
						vehicle.ID,
						strings.Join(alternateInitialStops, "`, `"),
					))
				}
			}
		}
	}

	return nil
}

type resourceInfo struct {
	allStartLevelsZero       bool
	allStartLevelsAtCapacity bool
	anyStops                 bool
	allStopsNegative         bool
	allStopsPositive         bool
}

func validateResources(input schema.Input, modelOptions Options) error {
	resourcesInfo := map[string]*resourceInfo{}

	for _, vehicle := range input.Vehicles {
		resourceCapacities, err := resources(vehicle, "Capacity", 1)
		if err != nil {
			return err
		}

		for name, resourceCapacity := range resourceCapacities {
			if resourceCapacity < 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` capacity must be positive, resource `%s` has negative capacity %f",
					vehicle.ID,
					name,
					resourceCapacity,
				))
			}

			if _, ok := resourcesInfo[name]; !ok {
				resourcesInfo[name] = &resourceInfo{
					allStartLevelsZero:       true,
					allStartLevelsAtCapacity: true,
					anyStops:                 false,
					allStopsNegative:         true,
					allStopsPositive:         true,
				}
			}
		}

		levels, err := resources(vehicle, "StartLevel", 1)
		if err != nil {
			return err
		}
		for name, level := range levels {
			if _, ok := resourceCapacities[name]; !ok {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` start level for resource `%s` is set but resource is not defined",
					vehicle.ID,
					name,
				))
			}
			if level < 0 {
				return nmerror.NewInputDataError(fmt.Errorf(
					"vehicle `%s` start level must be positive, resource `%s` has negative start level %f",
					vehicle.ID,
					name,
					level,
				))
			}

			if resourceCapacity, ok := resourceCapacities[name]; ok {
				if level > resourceCapacity {
					return nmerror.NewInputDataError(fmt.Errorf(
						"vehicle `%s` start level must be less or equal to capacity,"+
							" resource `%s` has capacity %f and start level %f",
						vehicle.ID,
						name,
						resourceCapacity,
						level,
					))
				}
			}
		}

		for name := range resourceCapacities {
			if level, ok := levels[name]; ok {
				if level != 0 {
					resourcesInfo[name].allStartLevelsZero = false
				}
				if level != resourceCapacities[name] {
					resourcesInfo[name].allStartLevelsAtCapacity = false
				}
			} else {
				resourcesInfo[name].allStartLevelsAtCapacity = false
			}
		}
	}

	for _, stop := range input.Stops {
		quantity, err := resources(stop, "Quantity", 1)
		if err != nil {
			return err
		}

		for name, value := range quantity {
			if _, ok := resourcesInfo[name]; !ok {
				return nmerror.NewInputDataError(fmt.Errorf(
					"stop `%s` quantity %f for resource `%s` is set,"+
						" but capacity for resource is not defined on any vehicle",
					stop.ID,
					value,
					name,
				))
			}
			resourcesInfo[name].anyStops = true
			resourcesInfo[name].allStopsNegative =
				resourcesInfo[name].allStopsNegative && value < 0
			resourcesInfo[name].allStopsPositive =
				resourcesInfo[name].allStopsPositive && value > 0
		}
	}

	if input.AlternateStops != nil {
		for _, stop := range *input.AlternateStops {
			quantity, err := resources(stop, "Quantity", 1)
			if err != nil {
				return err
			}

			for name, value := range quantity {
				if _, ok := resourcesInfo[name]; !ok {
					return nmerror.NewInputDataError(fmt.Errorf(
						"alternate stop `%s` quantity %v for resource `%s` is set,"+
							" but capacity for resource is not defined on any vehicle",
						stop.ID,
						value,
						name,
					))
				}
				resourcesInfo[name].anyStops = true
				resourcesInfo[name].allStopsNegative =
					resourcesInfo[name].allStopsNegative && value < 0
				resourcesInfo[name].allStopsPositive =
					resourcesInfo[name].allStopsPositive && value > 0
			}
		}
	}

	if !modelOptions.Validate.Disable.Resources {
		for name, info := range resourcesInfo {
			if info.anyStops && info.allStopsPositive && info.allStartLevelsZero {
				return nmerror.NewInputDataError(fmt.Errorf(
					"resource `%s` is starting without any capacity being"+
						" used. All your stops have a positive quantity and"+
						" are considered as dropoff stops. You need to have"+
						" at least one pickup stop (negative quantity) or a"+
						" start level > 0 to plan a stop with a positive"+
						" quantity",
					name,
				))
			}

			if info.anyStops && info.allStopsNegative && info.allStartLevelsAtCapacity {
				return nmerror.NewInputDataError(fmt.Errorf(
					"resource `%s` is starting with all of the capacity"+
						" being used. All your stops have a negative quantity"+
						" and are considered as pickup stops. You need to have"+
						" at least one dropoff stop (positive quantity) or a"+
						" start level < max capacity to plan a stop with a"+
						" negative quantity",
					name,
				))
			}
		}
	}

	return nil
}

// Converts a time-dependent matrix from a JSON map to a schema.TimeDependentMatrix.
func convertToTimeDependentMatrix(data map[string]any) (schema.TimeDependentMatrix, error) {
	var result schema.TimeDependentMatrix

	if dMatrix, ok := data["default_matrix"].([]any); ok {
		if fMatrix, ok := common.TryAssertFloat64Matrix(dMatrix); ok {
			result.DefaultMatrix = fMatrix
		} else {
			return result, fmt.Errorf("invalid or missing default_matrix")
		}
	} else {
		return result, fmt.Errorf("invalid or missing default_matrix")
	}

	if vIDs, ok := data["vehicle_ids"].([]any); ok {
		if vehicleIDs, ok := common.TryAssertStringSlice(vIDs); ok {
			result.VehicleIDs = vehicleIDs
		}
	}

	if timeFrames, ok := data["matrix_time_frames"].([]any); ok {
		result.MatrixTimeFrames = make([]schema.MatrixTimeFrame, len(timeFrames))
		for i, tf := range timeFrames {
			timeFrame, ok := tf.(map[string]any)
			if !ok {
				return result, fmt.Errorf("invalid time frame at index %d", i)
			}

			var mtf schema.MatrixTimeFrame

			if mMatrix, ok := timeFrame["matrix"].([]any); ok {
				if fMatrix, ok := common.TryAssertFloat64Matrix(mMatrix); ok {
					mtf.Matrix = fMatrix
				}
			}

			if scalingFactor, ok := timeFrame["scaling_factor"].(float64); ok {
				mtf.ScalingFactor = &scalingFactor
			}

			if startTime, ok := timeFrame["start_time"].(string); ok {
				t, err := time.Parse(time.RFC3339, startTime)
				if err != nil {
					return result, fmt.Errorf("invalid start_time at index %d: %v", i, err)
				}
				mtf.StartTime = t
			} else {
				return result, fmt.Errorf("missing or invalid start_time at index %d", i)
			}

			if endTime, ok := timeFrame["end_time"].(string); ok {
				t, err := time.Parse(time.RFC3339, endTime)
				if err != nil {
					return result, fmt.Errorf("invalid end_time at index %d: %v", i, err)
				}
				mtf.EndTime = t
			} else {
				return result, fmt.Errorf("missing or invalid end_time at index %d", i)
			}

			result.MatrixTimeFrames[i] = mtf
		}
	}

	return result, nil
}
