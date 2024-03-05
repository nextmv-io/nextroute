// © 2019-present nextmv.io inc

package observers

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/nextmv-io/nextroute"
)

// PerformanceObserver is an interface that is used to observe the performance
// of the model, and it's subsequent use.
type PerformanceObserver interface {
	nextroute.SolutionObserver

	// Duration returns the duration since the creation of the
	// PerformanceObserver.
	Duration() time.Duration

	// Report creates a report of the performance of the model, and it's
	// subsequent use.
	Report() string
}

// NewPerformanceObserver returns a new performance observer. A performance
// observer can be used to report the speed performance of moves.
//
// To use the observer, you need to register it with the model using the
// AddSolutionObserver method.
func NewPerformanceObserver(model nextroute.Model) PerformanceObserver {
	observer := performanceObserverImpl{
		model:          model,
		start:          time.Now(),
		constraintData: make(map[nextroute.ModelConstraint]constraintData),
		objectiveData: objectiveData{
			lastStartTimestamp: make(map[string]time.Time),
		},
		solutionData: solutionData{
			lastBestMoveStart: make(map[string]time.Time),
			lastCopyStart:     make(map[string]time.Time),
			lastMoveStart:     make(map[string]time.Time),
			lastNewStart:      make(map[string]time.Time),
		},
	}

	return &observer
}

type solutionData struct {
	lastBestMoveStart  map[string]time.Time
	lastMoveStart      map[string]time.Time
	lastNewStart       map[string]time.Time
	lastCopyStart      map[string]time.Time
	cumulativeMoves    time.Duration
	cumulativeNew      time.Duration
	copies             int
	cumulativeCopy     time.Duration
	cumulativeBestMove time.Duration
	bestMoves          int
	noBestMoves        int
	moves              int
	movesFailed        int
	new                int
}

type constraintData struct {
	lastStartTimestamp  map[string]time.Time
	checks              int
	cumulativeCheck     time.Duration
	cumulativeEstimate  time.Duration
	estimations         int
	estimatedViolations int
	violated            int
}

type objectiveData struct {
	lastStartTimestamp map[string]time.Time
	estimations        int
	cumulativeEstimate time.Duration
}

type performanceObserverImpl struct {
	constraintMutex sync.Mutex
	objectiveMutex  sync.Mutex
	moveMutex       sync.Mutex
	solutionMutex   sync.Mutex
	model           nextroute.Model
	start           time.Time
	constraintData  map[nextroute.ModelConstraint]constraintData
	objectiveData   objectiveData
	solutionData    solutionData
}

// OnSolutionConstraintChecked implements PerformanceObserver.
func (p *performanceObserverImpl) OnSolutionConstraintChecked(
	constraint nextroute.ModelConstraint,
	feasible bool,
) {
	p.constraintMutex.Lock()
	defer p.constraintMutex.Unlock()

	if data, ok := p.constraintData[constraint]; ok {
		data.cumulativeCheck += time.Since(data.lastStartTimestamp[p.routineName()])
		if !feasible {
			data.violated++
		}
		p.constraintData[constraint] = data
	}
}

// OnStopConstraintChecked implements PerformanceObserver.
func (p *performanceObserverImpl) OnStopConstraintChecked(
	_ nextroute.SolutionStop,
	constraint nextroute.ModelConstraint,
	feasible bool,
) {
	p.constraintMutex.Lock()
	defer p.constraintMutex.Unlock()

	if data, ok := p.constraintData[constraint]; ok {
		data.cumulativeCheck += time.Since(data.lastStartTimestamp[p.routineName()])
		if !feasible {
			data.violated++
		}
		p.constraintData[constraint] = data
	}
}

// OnVehicleConstraintChecked implements PerformanceObserver.
func (p *performanceObserverImpl) OnVehicleConstraintChecked(
	_ nextroute.SolutionVehicle,
	constraint nextroute.ModelConstraint,
	feasible bool,
) {
	p.constraintMutex.Lock()
	defer p.constraintMutex.Unlock()

	if data, ok := p.constraintData[constraint]; ok {
		data.cumulativeCheck += time.Since(data.lastStartTimestamp[p.routineName()])
		if !feasible {
			data.violated++
		}
		p.constraintData[constraint] = data
	}
}

func (p *performanceObserverImpl) OnBestMove(_ nextroute.Solution) {
	p.moveMutex.Lock()
	defer p.moveMutex.Unlock()

	p.solutionData.bestMoves++
	p.solutionData.lastBestMoveStart[p.routineName()] = time.Now()
}

func (p *performanceObserverImpl) OnBestMoveFound(move nextroute.SolutionMove) {
	p.moveMutex.Lock()
	defer p.moveMutex.Unlock()

	p.solutionData.cumulativeBestMove += time.Since(
		p.solutionData.lastBestMoveStart[p.routineName()],
	)

	if !move.IsExecutable() {
		p.solutionData.noBestMoves++
	}
}

func (p *performanceObserverImpl) OnPlanFailed(_ nextroute.SolutionMove, _ nextroute.ModelConstraint) {
	p.moveMutex.Lock()
	defer p.moveMutex.Unlock()
	p.solutionData.cumulativeMoves += time.Since(
		p.solutionData.lastMoveStart[p.routineName()],
	)
	p.solutionData.movesFailed++
}

func (p *performanceObserverImpl) OnPlanSucceeded(_ nextroute.SolutionMove) {
	p.moveMutex.Lock()
	defer p.moveMutex.Unlock()
	p.solutionData.cumulativeMoves += time.Since(
		p.solutionData.lastMoveStart[p.routineName()],
	)
}

func (p *performanceObserverImpl) OnPlan(_ nextroute.SolutionMove) {
	p.moveMutex.Lock()
	defer p.moveMutex.Unlock()
	p.solutionData.moves++

	p.solutionData.lastMoveStart[p.routineName()] = time.Now()
}

func (p *performanceObserverImpl) OnNewSolution(_ nextroute.Model) {
	p.solutionMutex.Lock()
	defer p.solutionMutex.Unlock()

	p.solutionData.new++
	p.solutionData.lastNewStart[p.routineName()] = time.Now()
}

func (p *performanceObserverImpl) OnNewSolutionCreated(_ nextroute.Solution) {
	p.solutionMutex.Lock()
	defer p.solutionMutex.Unlock()

	p.solutionData.cumulativeNew += time.Since(
		p.solutionData.lastNewStart[p.routineName()],
	)
}

func (p *performanceObserverImpl) OnCopySolution(_ nextroute.Solution) {
	p.solutionMutex.Lock()
	defer p.solutionMutex.Unlock()

	p.solutionData.copies++
	p.solutionData.lastCopyStart[p.routineName()] = time.Now()
}

func (p *performanceObserverImpl) OnCopiedSolution(_ nextroute.Solution) {
	p.solutionMutex.Lock()
	defer p.solutionMutex.Unlock()

	p.solutionData.cumulativeCopy += time.Since(
		p.solutionData.lastCopyStart[p.routineName()],
	)
}

func (p *performanceObserverImpl) Duration() time.Duration {
	return time.Since(p.start)
}

func (p *performanceObserverImpl) routineName() string {
	return string(bytes.Fields(debug.Stack())[1])
}

func (p *performanceObserverImpl) OnEstimateDeltaObjectiveScore() {
	p.objectiveMutex.Lock()
	defer p.objectiveMutex.Unlock()

	p.objectiveData.estimations++
	p.objectiveData.lastStartTimestamp[p.routineName()] = time.Now()
}

func (p *performanceObserverImpl) OnEstimatedDeltaObjectiveScore(_ float64) {
	p.objectiveMutex.Lock()
	defer p.objectiveMutex.Unlock()

	p.objectiveData.cumulativeEstimate += time.Since(p.objectiveData.lastStartTimestamp[p.routineName()])
}

func (p *performanceObserverImpl) OnCheckConstraint(
	constraint nextroute.ModelConstraint,
	_ nextroute.CheckedAt,
) {
	p.constraintMutex.Lock()
	defer p.constraintMutex.Unlock()

	if data, ok := p.constraintData[constraint]; ok {
		data.checks++
		data.lastStartTimestamp[p.routineName()] = time.Now()
		p.constraintData[constraint] = data
	} else {
		p.constraintData[constraint] = constraintData{
			checks:             1,
			violated:           0,
			estimations:        0,
			lastStartTimestamp: make(map[string]time.Time),
			cumulativeCheck:    time.Duration(0),
		}
		p.constraintData[constraint].lastStartTimestamp[p.routineName()] = time.Now()
	}
}

func (p *performanceObserverImpl) OnEstimateIsViolated(
	constraint nextroute.ModelConstraint,
) {
	p.constraintMutex.Lock()
	defer p.constraintMutex.Unlock()

	if data, ok := p.constraintData[constraint]; ok {
		data.estimations++
		data.lastStartTimestamp[p.routineName()] = time.Now()
		p.constraintData[constraint] = data
	} else {
		p.constraintData[constraint] = constraintData{
			checks:             0,
			violated:           0,
			estimations:        1,
			lastStartTimestamp: make(map[string]time.Time),
			cumulativeCheck:    time.Duration(0),
		}

		p.constraintData[constraint].lastStartTimestamp[p.routineName()] = time.Now()
	}
}

func (p *performanceObserverImpl) OnEstimatedIsViolated(
	_ nextroute.SolutionMove,
	constraint nextroute.ModelConstraint,
	violated bool,
	_ nextroute.StopPositionsHint,
) {
	p.constraintMutex.Lock()
	defer p.constraintMutex.Unlock()

	if data, ok := p.constraintData[constraint]; ok {
		data.cumulativeEstimate += time.Since(data.lastStartTimestamp[p.routineName()])
		if violated {
			data.estimatedViolations++
		}
		p.constraintData[constraint] = data
	}
}

func (p *performanceObserverImpl) Report() string {
	var sb strings.Builder
	line := strings.Repeat("-", 80)
	fmt.Fprintf(&sb, "%s\nTotal\n%s\n",
		line,
		line)
	fmt.Fprintf(&sb, "Duration                    : %v\n",
		p.Duration(),
	)

	routines := 0

	for _, data := range p.constraintData {
		if routines < len(data.lastStartTimestamp) {
			routines = len(data.lastStartTimestamp)
		}
	}
	fmt.Fprintf(&sb, "Go routines seen            : %v\n",
		routines,
	)

	fmt.Fprintf(&sb, "%s\n", line)
	fmt.Fprintf(&sb, "Solution\n%s\n", line)
	fmt.Fprintf(&sb, "New count                   : %v\n",
		p.solutionData.new,
	)
	fmt.Fprintf(&sb, "Total new duration          : %v\n",
		p.solutionData.cumulativeNew,
	)
	duration := 0 * time.Second
	if p.solutionData.new > 0 {
		duration = p.solutionData.cumulativeNew /
			time.Duration(p.solutionData.new)
	}
	fmt.Fprintf(&sb, "Average new duration        : %v\n",
		duration,
	)
	fmt.Fprintf(&sb, "Copies count                : %v\n",
		p.solutionData.copies,
	)
	fmt.Fprintf(&sb, "Total copies duration       : %v\n",
		p.solutionData.cumulativeCopy,
	)
	duration = 0 * time.Second
	if p.solutionData.copies > 0 {
		duration = p.solutionData.cumulativeCopy /
			time.Duration(p.solutionData.copies)
	}
	fmt.Fprintf(&sb, "Average copies duration     : %v\n",
		duration,
	)
	fmt.Fprintf(&sb, "Best move requested         : %v\n",
		p.solutionData.bestMoves,
	)
	fmt.Fprintf(&sb, "No best move found          : %v\n",
		p.solutionData.noBestMoves,
	)
	fmt.Fprintf(&sb, "Total best move duration    : %v\n",
		p.solutionData.cumulativeBestMove,
	)
	duration = 0 * time.Second
	if p.solutionData.bestMoves > 0 {
		duration = p.solutionData.cumulativeBestMove /
			time.Duration(p.solutionData.bestMoves)
	}
	fmt.Fprintf(&sb, "Average best move duration  : %v\n",
		duration,
	)

	fmt.Fprintf(&sb, "Moves executed              : %v\n",
		p.solutionData.moves,
	)
	fmt.Fprintf(&sb, "Total moves duration        : %v\n",
		p.solutionData.cumulativeMoves,
	)
	duration = 0 * time.Second
	if p.solutionData.moves > 0 {
		duration = p.solutionData.cumulativeMoves /
			time.Duration(p.solutionData.moves)
	}
	fmt.Fprintf(&sb, "Average moves duration      : %v\n",
		duration,
	)
	fmt.Fprintf(&sb, "Moves failed                : %v\n",
		p.solutionData.movesFailed,
	)
	fmt.Fprintf(&sb, "%s\n", line)
	fmt.Fprintf(&sb, "Objective: %v\n",
		p.model.Objective(),
	)
	fmt.Fprintf(&sb, "%s\n", line)
	fmt.Fprintf(&sb, "Estimates count             : %v\n",
		p.objectiveData.estimations,
	)
	fmt.Fprintf(&sb, "Total estimates duration    : %v\n",
		p.objectiveData.cumulativeEstimate,
	)
	duration = 0 * time.Second
	if p.objectiveData.estimations > 0 {
		duration = p.objectiveData.cumulativeEstimate /
			time.Duration(p.objectiveData.estimations)
	}
	fmt.Fprintf(&sb, "Average estimates duration  : %v\n",
		duration,
	)
	constraints := p.model.Constraints()
	for _, constraint := range constraints {
		fmt.Fprintf(&sb, "%s\n", line)
		fmt.Fprintf(&sb, "Name: %s (address=%v)\n",
			reflect.TypeOf(constraint).String(),
			&constraint,
		)
		fmt.Fprintf(&sb, "%s\n", line)

		if data, ok := p.constraintData[constraint]; ok {
			fmt.Fprintf(&sb, "Estimates count             : %v\n",
				data.estimations,
			)
			fmt.Fprintf(&sb, "Estimates violation count   : %v\n",
				data.estimatedViolations,
			)
			fmt.Fprintf(&sb, "Total estimates duration    : %v\n",
				data.cumulativeEstimate,
			)
			duration = 0 * time.Second
			if data.estimations > 0 {
				duration = data.cumulativeEstimate /
					time.Duration(data.estimations)
			}
			fmt.Fprintf(&sb, "Average estimates duration  : %v\n",
				duration,
			)
			fmt.Fprintf(&sb, "Checks count                : %v\n",
				data.checks,
			)
			fmt.Fprintf(&sb, "Total check duration        : %v\n",
				data.cumulativeCheck,
			)
			duration = 0 * time.Second
			if data.checks > 0 {
				duration = data.cumulativeCheck / time.Duration(data.checks)
			}
			fmt.Fprintf(&sb, "Average check duration      : %v\n",
				duration,
			)
			fmt.Fprintf(&sb, "Violated count              : %v\n",
				data.violated,
			)
		} else {
			fmt.Fprintf(&sb, "No data\n")
		}
	}
	return sb.String()
}
