// © 2019-present nextmv.io inc

package nextroute

import (
	"time"
)

type plateauTracker struct {
	// progression is the value progression of the solver. This is tracked
	// separately of any other progression tracking to avoid conflicts.
	progression []ProgressionEntry
	// durationIndex is the current index of the first progression entry within
	// the duration cutoff.
	durationIndex int
	// iterationsIndex is the current index of the first progression entry
	// within the iterations cutoff.
	iterationsIndex int
	// options are the options for the plateau tracker.
	options PlateauOptions
}

func newPlateauTracker(options PlateauOptions) *plateauTracker {
	return &plateauTracker{
		progression:     make([]ProgressionEntry, 0),
		durationIndex:   0,
		iterationsIndex: 0,
		options:         options,
	}
}

// plateauTrackingActivated returns true if the plateau tracking should be
// activated based on the provided options.
func plateauTrackingActivated(options PlateauOptions) bool {
	// We need to be testing within some duration or iteration cutoff.
	return (options.Duration > 0 || options.Iterations > 0) &&
		// We need to have some threshold configured (negative threshold is deactivating the corresponding check).
		(options.AbsoluteThreshold >= 0 || options.RelativeThreshold >= 0)
}

// onImprovement is called to update the plateau tracker whenever a new
// improvement is found.
func (t *plateauTracker) onImprovement(elapsed float64, iterations int, value float64) {
	if t == nil {
		return
	}
	// Add the new progression entry.
	t.progression = append(t.progression, ProgressionEntry{
		ElapsedSeconds: elapsed,
		Value:          value,
		Iterations:     iterations,
	})
}

// ShouldTerminate returns true if the solver should terminate due to a detected
// plateau.
func (t *plateauTracker) ShouldTerminate(iterations int, elapsed time.Duration) bool {
	if t == nil {
		return false
	}

	currentValue := t.progression[len(t.progression)-1].Value

	// Check if no significantly improving solutions were found during the
	// configured duration.
	if t.options.Duration > 0 {
		cutoffSeconds := t.options.Duration.Seconds()
		elapsedSeconds := elapsed.Seconds()
		// Move the duration index to the first entry within the cutoff.
		for t.durationIndex < len(t.progression) &&
			(elapsedSeconds-t.progression[t.durationIndex].ElapsedSeconds) > cutoffSeconds {
			t.durationIndex++
		}
		// If the duration index is at the end of the progression, no
		// improvement was found within the cutoff.
		if t.durationIndex >= len(t.progression) {
			return true
		}
		// Compare the current value to the first value within the cutoff. It
		// needs to be (significantly) better to avoid termination.
		// Note: The cutoff value will always be larger than (or equal to) the
		// current value, as we are minimizing.
		cutoffValue := t.progression[t.durationIndex].Value
		if t.options.AbsoluteThreshold >= 0 &&
			cutoffValue-currentValue < t.options.AbsoluteThreshold {
			return true
		}
		if t.options.RelativeThreshold >= 0 &&
			currentValue > 0 && // Relative threshold is only supported for positive values.
			(cutoffValue-currentValue)/currentValue < t.options.RelativeThreshold {
			return true
		}
	}

	// Check if no significantly improving solutions were found during the
	// configured iterations.
	if t.options.Iterations > 0 {
		// Move the iterations index to the first entry within the cutoff.
		for t.iterationsIndex < len(t.progression) &&
			iterations-t.progression[t.iterationsIndex].Iterations > t.options.Iterations {
			t.iterationsIndex++
		}
		// If the iterations index is at the end of the progression, no
		// improvement was found within the cutoff.
		if t.iterationsIndex >= len(t.progression) {
			return true
		}
		// Compare the current value to the first value within the cutoff. It
		// needs to be (significantly) better to avoid termination.
		// Note: The cutoff value will always be larger than (or equal to) the
		// current value, as we are minimizing.
		cutoffValue := t.progression[t.iterationsIndex].Value
		if t.options.AbsoluteThreshold >= 0 &&
			cutoffValue-currentValue < t.options.AbsoluteThreshold {
			return true
		}
		if t.options.RelativeThreshold >= 0 &&
			currentValue > 0 && // Relative threshold is only supported for positive values.
			(cutoffValue-currentValue)/currentValue < t.options.RelativeThreshold {
			return true
		}
	}

	// No plateau detected.
	return false
}
