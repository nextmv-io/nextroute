package nextroute

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/nextmv-io/sdk/nextroute"
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

// NoPositionsHint returns a new StopPositionsHint that does not skip
// the vehicle and does not contain a next stop. The solver will try to find
// the next stop.
func NoPositionsHint() nextroute.StopPositionsHint {
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
// vehicle if skipVehicle is true. Is skipVehicle is false the solver will try
// to find the next stop.
func SkipVehiclePositionsHint() nextroute.StopPositionsHint {
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

func (n *stopPositionHintImpl) NextStopPositions() nextroute.StopPositions {
	return nextroute.StopPositions{}
}

func (n *stopPositionHintImpl) SkipVehicle() bool {
	return n.skipVehicle
}

func newErrorOnNilHint(constraint nextroute.ModelConstraint) error {
	name := reflect.TypeOf(constraint).Name()
	stringer, ok := constraint.(fmt.Stringer)
	if ok {
		name = stringer.String()
	}
	identifier, ok := constraint.(nextroute.Identifier)
	if ok {
		name = identifier.ID()
	}
	return fmt.Errorf(
		"constraint %v returned nil hint in EstimateIsViolated, nil is not allowed use NoPositionsHint()",
		name,
	)
}
