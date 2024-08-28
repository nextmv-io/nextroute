// © 2019-present nextmv.io inc

package nextroute

// SolveEvents is a struct that contains events that are fired during a solve
// invocation.
type SolveEvents struct {
	// ContextDone is fired when the context is done for any reason.
	ContextDone *BaseEvent1[SolveInformation]

	// Done is fired when the solver is done.
	Done *BaseEvent1[SolveInformation]

	// Iterated is fired when the solver has iterated.
	Iterated *BaseEvent1[SolveInformation]
	// Iterating is fired when the solver is iterating.
	Iterating *BaseEvent1[SolveInformation]

	// NewBestSolution is fired when a new best solution is found.
	NewBestSolution *BaseEvent1[SolveInformation]

	// OperatorExecuted is fired when a solve-operator has been executed.
	OperatorExecuted *BaseEvent1[SolveInformation]
	// OperatorExecuting is fired when a solve-operator is executing.
	OperatorExecuting *BaseEvent1[SolveInformation]

	// Reset is fired when the solver is reset.
	Reset *BaseEvent2[Solution, SolveInformation]

	// Start is fired when the solver is started.
	Start *BaseEvent1[SolveInformation]
}

// NewSolveEvents creates a new instance of Solve.
func NewSolveEvents() SolveEvents {
	return SolveEvents{
		OperatorExecuting: &BaseEvent1[SolveInformation]{},
		OperatorExecuted:  &BaseEvent1[SolveInformation]{},
		NewBestSolution:   &BaseEvent1[SolveInformation]{},
		Iterating:         &BaseEvent1[SolveInformation]{},
		Iterated:          &BaseEvent1[SolveInformation]{},
		ContextDone:       &BaseEvent1[SolveInformation]{},
		Start:             &BaseEvent1[SolveInformation]{},
		Reset:             &BaseEvent2[Solution, SolveInformation]{},
		Done:              &BaseEvent1[SolveInformation]{},
	}
}

// BaseEvent1 is a base event type that can be used to implement events
// with one payload.
type BaseEvent1[T any] struct {
	handlers []Handler1[T]
}

// Register adds an event handler for this event.
func (e *BaseEvent1[T]) Register(handler Handler1[T]) {
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload.
func (e *BaseEvent1[T]) Trigger(payload T) {
	for _, handler := range e.handlers {
		handler(payload)
	}
}

// Handler1 is a function that handles an event with one payload.
type Handler1[T any] func(payload T)

// BaseEvent2 is a base event type that can be used to implement events
// with two payloads.
type BaseEvent2[S any, T any] struct {
	handlers []Handler2[S, T]
}

// Register adds an event handler for this event.
func (e *BaseEvent2[S, T]) Register(handler Handler2[S, T]) {
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload.
func (e *BaseEvent2[S, T]) Trigger(payload1 S, payload2 T) {
	for _, handler := range e.handlers {
		handler(payload1, payload2)
	}
}

// Handler2 is a function that handles an event with two payloads.
type Handler2[S any, T any] func(payload1 S, payload2 T)

// BaseEvent3 is a base event type that can be used to implement events
// with three payloads.
type BaseEvent3[S any, T any, U any] struct {
	handlers []Handler3[S, T, U]
}

// Register adds an event handler for this event.
func (e *BaseEvent3[S, T, U]) Register(handler Handler3[S, T, U]) {
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload.
func (e *BaseEvent3[S, T, U]) Trigger(payload1 S, payload2 T, payload3 U) {
	for _, handler := range e.handlers {
		handler(payload1, payload2, payload3)
	}
}

// Handler3 is a function that handles an event with three payloads.
type Handler3[S any, T any, U any] func(payload1 S, payload2 T, payload3 U)

// BaseEvent4 is a base event type that can be used to implement events
// with four payloads.
type BaseEvent4[S any, T any, U any, V any] struct {
	handlers []Handler4[S, T, U, V]
}

// Register adds an event handler for this event.
func (e *BaseEvent4[S, T, U, V]) Register(handler Handler4[S, T, U, V]) {
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload.
func (e *BaseEvent4[S, T, U, V]) Trigger(payload1 S, payload2 T, payload3 U, payload4 V) {
	for _, handler := range e.handlers {
		handler(payload1, payload2, payload3, payload4)
	}
}

// Handler4 is a function that handles an event with four payloads.
type Handler4[S any, T any, U any, V any] func(payload1 S, payload2 T, payload3 U, payload4 V)
