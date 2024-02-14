package nextroute

import (
	"github.com/nextmv-io/sdk/events"
)

// SolveEvents is a struct that contains events that are fired during a solve
// invocation.
type SolveEvents struct {
	// ContextDone is fired when the context is done for any reason.
	ContextDone *events.BaseEvent1[SolveInformation]

	// Done is fired when the solver is done.
	Done *events.BaseEvent1[SolveInformation]

	// Iterated is fired when the solver has iterated.
	Iterated *events.BaseEvent1[SolveInformation]
	// Iterating is fired when the solver is iterating.
	Iterating *events.BaseEvent1[SolveInformation]

	// NewBestSolution is fired when a new best solution is found.
	NewBestSolution *events.BaseEvent1[SolveInformation]

	// OperatorExecuted is fired when a solve-operator has been executed.
	OperatorExecuted *events.BaseEvent1[SolveInformation]
	// OperatorExecuting is fired when a solve-operator is executing.
	OperatorExecuting *events.BaseEvent1[SolveInformation]

	// Reset is fired when the solver is reset.
	Reset *events.BaseEvent2[Solution, SolveInformation]

	// Start is fired when the solver is started.
	Start *events.BaseEvent1[SolveInformation]
}

// NewSolveEvents creates a new instance of SolveEvents.
func NewSolveEvents() SolveEvents {
	return SolveEvents{
		OperatorExecuting: &events.BaseEvent1[SolveInformation]{},
		OperatorExecuted:  &events.BaseEvent1[SolveInformation]{},
		NewBestSolution:   &events.BaseEvent1[SolveInformation]{},
		Iterating:         &events.BaseEvent1[SolveInformation]{},
		Iterated:          &events.BaseEvent1[SolveInformation]{},
		ContextDone:       &events.BaseEvent1[SolveInformation]{},
		Start:             &events.BaseEvent1[SolveInformation]{},
		Reset:             &events.BaseEvent2[Solution, SolveInformation]{},
		Done:              &events.BaseEvent1[SolveInformation]{},
	}
}
