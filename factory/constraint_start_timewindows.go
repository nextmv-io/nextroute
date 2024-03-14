// © 2019-present nextmv.io inc

package factory

import (
	"fmt"
	"reflect"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/nextroute/schema"
)

// addWindowsConstraint adds the time windows to the Model.
func addWindowsConstraint(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	latestStartExpression, model, err := latestStartExpression(model)
	if err != nil {
		return nil, err
	}

	stopsHaveTimeWindows, err := addWindowsStops(input, model, latestStartExpression)
	if err != nil {
		return nil, err
	}
	alternateStopsHaveTimeWindows, err := addWindowsAlternateStops(input, model, latestStartExpression)
	if err != nil {
		return nil, err
	}

	if !stopsHaveTimeWindows && !alternateStopsHaveTimeWindows {
		return model, nil
	}

	model, err = addLatestStartConstraint(model, latestStartExpression)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func addWindowsStops(
	input schema.Input,
	model nextroute.Model,
	latestStartExpression nextroute.StopTimeExpression,
) (bool, error) {
	hasTimeWindow := false
	for index, inputStop := range input.Stops {
		if inputStop.StartTimeWindow == nil {
			continue
		}

		stop, err := model.Stop(index)
		if err != nil {
			return false, err
		}

		windows, err := convertTimeWindow(inputStop.StartTimeWindow, inputStop.ID)
		if err != nil {
			return false, err
		}

		if len(windows) == 0 {
			continue
		}

		err = stop.SetWindows(windows)
		if err != nil {
			return false, err
		}
		latestStartExpression.SetTime(stop, windows[len(windows)-1][1])

		hasTimeWindow = true
	}

	return hasTimeWindow, nil
}

func addWindowsAlternateStops(
	input schema.Input,
	model nextroute.Model,
	latestStartExpression nextroute.StopTimeExpression,
) (bool, error) {
	if input.AlternateStops == nil {
		return false, nil
	}

	if common.AllTrue(
		*input.AlternateStops,
		func(stop schema.AlternateStop) bool {
			return stop.StartTimeWindow == nil
		},
	) {
		return false, nil
	}

	data, err := getModelData(model)

	if err != nil {
		return false, err
	}

	hasTimeWindow := false

	for _, vehicle := range input.Vehicles {
		if vehicle.AlternateStops == nil {
			continue
		}

		for _, alternateID := range *vehicle.AlternateStops {
			stop, err := model.Stop(data.stopIDToIndex[alternateStopID(alternateID, vehicle)])
			if err != nil {
				return false, err
			}

			alternateInputStop := stop.Data().(alternateInputStop)

			if alternateInputStop.stop.StartTimeWindow == nil {
				continue
			}

			hasTimeWindow = true

			windows, err := convertTimeWindow(
				alternateInputStop.stop.StartTimeWindow,
				alternateInputStop.stop.ID,
			)

			if err != nil {
				return false, err
			}

			if len(windows) == 0 {
				continue
			}

			err = stop.SetWindows(windows)

			if err != nil {
				return false, err
			}
			latestStartExpression.SetTime(stop, windows[len(windows)-1][1])
		}
	}
	return hasTimeWindow, nil
}

// convertTimeWindow converts various inputs to a slice of 2-tuples of time.Time.
// If the input is not in a convertible format, an error is returned.
func convertTimeWindow(window any, stopID string) ([][2]time.Time, error) {
	if reflect.TypeOf(window).Kind() == reflect.Ptr && !reflect.ValueOf(window).IsNil() {
		window = reflect.ValueOf(window).Elem().Interface()
	}

	windowValue := reflect.ValueOf(window)
	hostKind := windowValue.Kind()
	if hostKind != reflect.Slice && hostKind != reflect.Array {
		return nil, nmerror.NewInputDataError(fmt.Errorf("window %v of stop %s is not a slice", window, stopID))
	}

	if windowValue.Len() == 0 {
		return nil, nmerror.NewInputDataError(fmt.Errorf("window of stop %s is empty", stopID))
	}

	if windowValue.Index(0).IsNil() {
		return nil, nmerror.NewInputDataError(fmt.Errorf("window of stop %s is nil", stopID))
	}

	subKind := reflect.TypeOf(windowValue.Index(0).Interface()).Kind()
	if subKind == reflect.Slice || subKind == reflect.Array {
		internalWindow := make([][2]time.Time, windowValue.Len())
		for i := 0; i < windowValue.Len(); i++ {
			if windowValue.Index(i).IsNil() {
				return nil, nmerror.NewInputDataError(
					fmt.Errorf("window %v at index %d of stop %s is nil", window, i, stopID),
				)
			}
			subWindowValue := reflect.ValueOf(windowValue.Index(i).Interface())
			win, err := convertSingleTimeWindow(subWindowValue, stopID)
			if err != nil {
				return nil, err
			}
			internalWindow[i] = win
		}
		return internalWindow, nil
	}

	win, err := convertSingleTimeWindow(windowValue, stopID)
	if err != nil {
		return nil, err
	}
	return [][2]time.Time{win}, nil
}

// convertSingleTimeWindow converts a single input time window to the internal
// representation.
func convertSingleTimeWindow(windowValue reflect.Value, stopID string) ([2]time.Time, error) {
	convertTime := func(t any, stop string) (time.Time, error) {
		if tt, ok := t.(time.Time); ok {
			return tt, nil
		}
		ts, ok := t.(string)
		if !ok {
			return time.Time{}, nmerror.NewInputDataError(
				fmt.Errorf(
					"time %v of stop %s is not a string (got type %v)", t, stop, reflect.TypeOf(t),
				),
			)
		}
		tt, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return time.Time{}, nmerror.NewInputDataError(
				fmt.Errorf(
					"time %s of stop %s is not a valid RFC3339 time", ts, stop),
			)
		}
		return tt, nil
	}

	if windowValue.Len() != 2 {
		return [2]time.Time{}, nmerror.NewInputDataError(
			fmt.Errorf(
				"window %v of stop %s is not of length 2", windowValue, stopID),
		)
	}

	start, err := convertTime(windowValue.Index(0).Interface(), stopID)
	if err != nil {
		return [2]time.Time{}, err
	}
	end, err := convertTime(windowValue.Index(1).Interface(), stopID)
	if err != nil {
		return [2]time.Time{}, err
	}

	if start.After(end) || start.Equal(end) {
		return [2]time.Time{}, nmerror.NewInputDataError(
			fmt.Errorf("start %v is after end %v", start, end),
		)
	}

	return [2]time.Time{start, end}, nil
}
