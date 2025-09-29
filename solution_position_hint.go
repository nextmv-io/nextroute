// © 2019-present nextmv.io inc

package nextroute

// StopPositionsHint is an interface that can be used to give a hint to the
// solver about the next stop position. This can be used to speed up the
// solver. The solver will use the hint if it is available. Hints are generated
// by the estimate function of a constraint.
type StopPositionsHint struct {
	skipVehicle bool
}

// NoPositionsHint returns a new StopPositionsHint that does not skip
// the vehicle and does not contain a next stop. The solver will try to find
// the next stop.
func NoPositionsHint() StopPositionsHint {
	return StopPositionsHint{}
}

// SkipVehiclePositionsHint returns a new StopPositionsHint that skips the
// vehicle.
func SkipVehiclePositionsHint() StopPositionsHint {
	return StopPositionsHint{
		skipVehicle: true,
	}
}

// HasNextStopPositions returns true if the hint contains next positions.
func (n StopPositionsHint) HasNextStopPositions() bool {
	return false
}

// NextStopPositions returns the next positions.
func (n StopPositionsHint) NextStopPositions() StopPositions {
	return StopPositions{}
}

// SkipVehicle returns true if the solver should skip the vehicle. The
// solver will use the hint if it is available.
func (n StopPositionsHint) SkipVehicle() bool {
	return n.skipVehicle
}
