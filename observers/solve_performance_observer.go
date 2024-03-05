// © 2019-present nextmv.io inc

package observers

import (
	"fmt"
	"strings"
	"time"

	"github.com/nextmv-io/nextroute"
)

// SolvePerformanceObserver is an interface for observing the performance of an
// ALNS solver.
type SolvePerformanceObserver interface {
	// OperatorObservers returns the observers for the operators.
	OperatorObservers() []OperatorObserver

	// Report returns a report of the performance of the solver.
	Report() (string, error)
}

// OperatorObserver is an interface for observing the performance of an
// operator.
type OperatorObserver interface {
	// CumulativeTime returns the cumulative time spent executing the operator.
	CumulativeTime() time.Duration
	// Invocations returns the number of times the operator has been executed.
	Invocations() int
	// Report returns a report of the performance of the operator.
	Report() (string, error)
}

// NewSolvePerformanceObserver returns a new SolvePerformanceObserver.
func NewSolvePerformanceObserver(
	solver nextroute.Solver,
) SolvePerformanceObserver {
	performanceObserver := &solvePerformanceObserverImpl{
		operatorData: make(map[nextroute.SolveOperator]operatorDataImpl),
	}
	solver.SolveEvents().OperatorExecuting.Register(func(info nextroute.SolveInformation) {
		operators := info.SolveOperators()
		operator := operators[len(operators)-1]
		if _, ok := performanceObserver.operatorData[operator]; !ok {
			performanceObserver.operatorData[operator] = operatorDataImpl{
				name:           fmt.Sprintf("%T", operator),
				invocations:    0,
				cumulativeTime: 0,
			}
		}
		data := performanceObserver.operatorData[operator]
		data.invocations++
		data.lastStart = time.Now()
		performanceObserver.operatorData[operator] = data
	})
	solver.SolveEvents().OperatorExecuted.Register(func(info nextroute.SolveInformation) {
		operators := info.SolveOperators()
		operator := operators[len(operators)-1]
		data := performanceObserver.operatorData[operator]
		data.cumulativeTime += time.Since(data.lastStart)
		performanceObserver.operatorData[operator] = data
	})
	solver.SolveEvents().Done.Register(func(_ nextroute.SolveInformation) {
		fmt.Println(performanceObserver.Report())
	})
	return performanceObserver
}

type operatorDataImpl struct {
	lastStart      time.Time
	name           string
	invocations    int
	cumulativeTime time.Duration
}

func (o *operatorDataImpl) Report() (string, error) {
	var sb strings.Builder
	line := strings.Repeat("-", 80)
	_, err := fmt.Fprintf(&sb, "%s\n%s\n%s\n", line, o.name, line)
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(&sb, "Invocations                 : %d\n",
		o.invocations,
	)
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(&sb, "Duration                    : %v\n",
		o.cumulativeTime,
	)
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(&sb, "Average duration            : %v\n",
		o.cumulativeTime/time.Duration(o.invocations),
	)
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(&sb, "%s", line)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (o *operatorDataImpl) CumulativeTime() time.Duration {
	return o.cumulativeTime
}

func (o *operatorDataImpl) Invocations() int {
	return o.invocations
}

type solvePerformanceObserverImpl struct {
	operatorData map[nextroute.SolveOperator]operatorDataImpl
}

func (p *solvePerformanceObserverImpl) OperatorObservers() []OperatorObserver {
	observers := make([]OperatorObserver, 0)
	for _, data := range p.operatorData {
		observers = append(observers, &data)
	}
	return observers
}

func (p *solvePerformanceObserverImpl) Report() (string, error) {
	var sb strings.Builder
	line := strings.Repeat("-", 80)
	_, err := fmt.Fprintf(
		&sb,
		"%s\nSolver performance\n",
		line,
	)
	if err != nil {
		return "", err
	}
	for _, data := range p.operatorData {
		report, err := data.Report()
		if err != nil {
			return "", err
		}
		_, err = fmt.Fprintf(&sb, "%s\n", report)
		if err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}
