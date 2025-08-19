// © 2019-present nextmv.io inc

package nextroute

import (
	"testing"
	"time"
)

func TestShouldTerminate(t *testing.T) {
	// Define the test cases
	basicProgression := []ProgressionEntry{
		{ElapsedSeconds: 2, Iterations: 100, Value: 150},
		{ElapsedSeconds: 5, Iterations: 200, Value: 100},
		{ElapsedSeconds: 8, Iterations: 300, Value: 70},
		{ElapsedSeconds: 10, Iterations: 400, Value: 50},
	}
	cases := []struct {
		opts            PlateauOptions
		progression     []ProgressionEntry
		iterations      int
		elapsed         float64
		shouldTerminate bool
	}{
		{ // no plateau tracking
			opts:            PlateauOptions{},
			progression:     basicProgression,
			iterations:      500,
			elapsed:         15,
			shouldTerminate: false,
		},
		{ // 5 seconds plateau (stagnation case, relative threshold)
			opts: PlateauOptions{
				Duration:          time.Duration(5) * time.Second,
				RelativeThreshold: 0.1,
			},
			progression:     basicProgression,
			iterations:      500,
			elapsed:         15,
			shouldTerminate: true,
		},
		{ // 5 seconds plateau (stagnation case, absolute threshold)
			opts: PlateauOptions{
				Duration:          time.Duration(5) * time.Second,
				AbsoluteThreshold: 10,
			},
			progression:     basicProgression,
			iterations:      500,
			elapsed:         15,
			shouldTerminate: true,
		},
		{ // 5 iterations plateau (stagnation case, relative threshold)
			opts: PlateauOptions{
				Iterations:        5,
				RelativeThreshold: 0.1,
			},
			progression:     basicProgression,
			iterations:      500,
			elapsed:         15,
			shouldTerminate: true,
		},
		{ // 5 iterations plateau (stagnation case, absolute threshold)
			opts: PlateauOptions{
				Iterations:        5,
				AbsoluteThreshold: 10,
			},
			progression:     basicProgression,
			iterations:      500,
			elapsed:         15,
			shouldTerminate: true,
		},
		{ // 5 seconds plateau (no-stagnation case, relative threshold)
			opts: PlateauOptions{
				Duration:          time.Duration(5) * time.Second,
				RelativeThreshold: 0.1,
			},
			progression:     basicProgression,
			iterations:      500,
			elapsed:         12, // catches two last entries
			shouldTerminate: false,
		},
		{ // 5 seconds plateau (no-stagnation case, absolute threshold)
			opts: PlateauOptions{
				Duration:          time.Duration(5) * time.Second,
				AbsoluteThreshold: 10,
			},
			progression:     basicProgression,
			iterations:      500,
			elapsed:         12, // catches two last entries
			shouldTerminate: false,
		},
		{ // 5 iterations plateau (no-stagnation case, relative threshold)
			opts: PlateauOptions{
				Iterations:        200,
				RelativeThreshold: 0.1,
			},
			progression:     basicProgression,
			iterations:      450,
			elapsed:         15,
			shouldTerminate: false,
		},
		{ // 5 iterations plateau (no-stagnation case, absolute threshold)
			opts: PlateauOptions{
				Iterations:        200,
				AbsoluteThreshold: 10,
			},
			progression:     basicProgression,
			iterations:      450,
			elapsed:         15,
			shouldTerminate: false,
		},
		{ // 5 seconds plateau (non-significant improvement case, relative threshold)
			opts: PlateauOptions{
				Duration:          time.Duration(5) * time.Second,
				RelativeThreshold: 0.3,
			},
			progression:     append(basicProgression, ProgressionEntry{ElapsedSeconds: 12, Iterations: 500, Value: 45}),
			iterations:      500,
			elapsed:         14, // catches two last entries
			shouldTerminate: true,
		},
		{ // 5 seconds plateau (non-significant improvement case, absolute threshold)
			opts: PlateauOptions{
				Duration:          time.Duration(5) * time.Second,
				AbsoluteThreshold: 10,
			},
			progression:     append(basicProgression, ProgressionEntry{ElapsedSeconds: 12, Iterations: 500, Value: 45}),
			iterations:      500,
			elapsed:         14, // catches two last entries
			shouldTerminate: true,
		},
	}

	// Run the tests
	for _, c := range cases {
		tracker := newPlateauTracker(c.opts)
		for _, p := range c.progression {
			tracker.onImprovement(p.ElapsedSeconds, p.Iterations, p.Value)
		}
		if got := tracker.ShouldTerminate(c.iterations, time.Duration(c.elapsed)*time.Second); got != c.shouldTerminate {
			t.Errorf("ShouldTerminate() = %v, want %v", got, c.shouldTerminate)
		}
	}
}
