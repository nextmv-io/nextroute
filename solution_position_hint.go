// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	noHint          *stopPositionHintImpl
	skipVehicleHint *stopPositionHintImpl
)

var (
	onceNoHint          sync.Once
	onceSkipVehicleHint sync.Once
)

// The following two variables can be used to avoid allocations.
var (
	constSkipVehiclePositionsHint = SkipVehiclePositionsHint()
	constNoPositionsHint          = NoPositionsHint()
)

// StopPositionsHint is an interface that can be used to give a hint to the
// solver about the next stop position. This can be used to speed up the
// solver. The solver will use the hint if it is available. Hints are generated
// by the estimate function of a constraint.
type StopPositionsHint interface {
	// HasNextStopPositions returns true if the hint contains next positions.
	HasNextStopPositions() bool

	// NextStopPositions returns the next positions.
	NextStopPositions() StopPositions

	// SkipVehicle returns true if the solver should skip the vehicle. The
	// solver will use the hint if it is available.
	SkipVehicle() bool
}

// NoPositionsHint returns a new StopPositionsHint that does not skip
// the vehicle and does not contain a next stop. The solver will try to find
// the next stop.
func NoPositionsHint() StopPositionsHint {
	return noPositionsHint()
}

func noPositionsHint() *stopPositionHintImpl {
	onceNoHint.Do(func() {
		noHint = &stopPositionHintImpl{
			skipVehicle: false,
		}
	})
	return noHint
}

// SkipVehiclePositionsHint returns a new StopPositionsHint that skips the
// vehicle.
func SkipVehiclePositionsHint() StopPositionsHint {
	return skipVehiclePositionsHint()
}

func skipVehiclePositionsHint() *stopPositionHintImpl {
	onceSkipVehicleHint.Do(func() {
		skipVehicleHint = &stopPositionHintImpl{
			skipVehicle: true,
		}
	})
	return skipVehicleHint
}

type stopPositionHintImpl struct {
	skipVehicle bool
}

func (n *stopPositionHintImpl) HasNextStopPositions() bool {
	return false
}

func (n *stopPositionHintImpl) NextStopPositions() StopPositions {
	return StopPositions{}
}

func (n *stopPositionHintImpl) SkipVehicle() bool {
	return n.skipVehicle
}

func newErrorOnNilHint(constraint ModelConstraint) error {
	name := reflect.TypeOf(constraint).Name()
	stringer, ok := constraint.(fmt.Stringer)
	if ok {
		name = stringer.String()
	}
	identifier, ok := constraint.(Identifier)
	if ok {
		name = identifier.ID()
	}
	return fmt.Errorf(
		"constraint %v returned nil hint in EstimateIsViolated, nil is not allowed use NoPositionsHint()",
		name,
	)
}
