// © 2019-present nextmv.io inc

package nextroute

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/nextmv-io/nextroute/common"
	"github.com/nextmv-io/sdk/run"
	"github.com/nextmv-io/sdk/run/schema"
	"github.com/nextmv-io/sdk/run/statistics"
)

// FormatOptions are the options that influence the format of the output.
type FormatOptions struct {
	Disable struct {
		Progression bool `json:"progression" usage:"disable the progression series"`
	} `json:"disable"`
}

// Format formats a solution in basic format using the map function
// toSolutionOutputFn to map a solution to a user specific format.
func Format(
	ctx context.Context,
	options any,
	progressioner Progressioner,
	toSolutionOutputFn func(Solution) any,
	inputSolutions ...Solution,
) schema.Output {
	solutions := common.Filter(
		inputSolutions,
		func(solution Solution) bool {
			return solution != nil
		},
	)
	output := schema.NewOutput(
		options,
		common.Map(solutions, toSolutionOutputFn)...,
	)

	output.Statistics = statistics.NewStatistics()

	if start, ok := ctx.Value(run.Start).(time.Time); ok {
		duration := time.Since(start).Seconds()
		output.Statistics.Run = &statistics.Run{
			Duration: &duration,
		}
	}

	if data, ok := ctx.Value(run.Data).(*sync.Map); ok {
		if iterations, ok := data.Load(Iterations); ok {
			if iterations, ok := iterations.(int); ok {
				output.Statistics.Run.Iterations = &iterations
			}
		}
	}

	if len(solutions) == 0 {
		return output
	}

	solution := solutions[len(solutions)-1]

	if progressioner == nil {
		return output
	}

	progressionValues := progressioner.Progression()

	if len(progressionValues) == 0 {
		return output
	}

	seriesData := common.Map(
		progressionValues,
		func(progressionEntry ProgressionEntry) statistics.DataPoint {
			return statistics.DataPoint{
				X: statistics.Float64(progressionEntry.ElapsedSeconds),
				Y: statistics.Float64(progressionEntry.Value),
			}
		},
	)
	iterationsSeriesData := common.Map(
		progressionValues,
		func(progressionEntry ProgressionEntry) statistics.DataPoint {
			return statistics.DataPoint{
				X: statistics.Float64(progressionEntry.ElapsedSeconds),
				Y: statistics.Float64(progressionEntry.Iterations),
			}
		},
	)

	lastProgressionElement := progressionValues[len(progressionValues)-1]
	lastProgressionValue := statistics.Float64(lastProgressionElement.Value)

	output.Statistics.Result = &statistics.Result{
		Duration: &lastProgressionElement.ElapsedSeconds,
		Value:    &lastProgressionValue,
	}

	r := reflect.ValueOf(options)
	f := reflect.Indirect(r).FieldByName("Format")
	if f.IsValid() && f.CanInterface() {
		if format, ok := f.Interface().(FormatOptions); ok {
			if format.Disable.Progression {
				return output
			}
		}
	}

	output.Statistics.SeriesData = &statistics.SeriesData{
		Value: statistics.Series{
			Name:       fmt.Sprintf("%v", solution.Model().Objective()),
			DataPoints: seriesData,
		},
	}
	output.Statistics.SeriesData.Custom = append(output.Statistics.SeriesData.Custom, statistics.Series{
		Name:       "iterations",
		DataPoints: iterationsSeriesData,
	})

	return output
}
