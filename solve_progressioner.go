// © 2019-present nextmv.io inc

package nextroute

// Progressioner is an interface that can be implemented by a solver to indicate
// that is can return the progression of the solver.
type Progressioner interface {
	// Progression returns the progression of the solver.
	Progression() []ProgressionEntry
}

// ProgressionEntry is a single entry in the progression of the solver.
type ProgressionEntry struct {
	ElapsedSeconds float64 `json:"elapsed_seconds"`
	Value          float64 `json:"value"`
	Iterations     int     `json:"iterations"`
}
