// © 2019-present nextmv.io inc

package nextroute

// CheckedAt is the type indicating when to check a constraint when
// a move it's consequences are being propagated.
type CheckedAt int64

const (
	// AtEachStop indicates that the constraint should be checked at each stop.
	// A constraint that is registered to be checked at each stop will be
	// checked as soon as all values for the expressions a stop are known. The
	// stops later in a vehicle will not been updated yet.
	AtEachStop CheckedAt = 0
	// AtEachVehicle indicates that the constraint should be checked at each
	// vehicle. A constraint that is registered to be checked at each vehicle
	// will be checked as soon as all values for the expressions of the stops
	// of a vehicle are known. The stops of other vehicles will not been
	// updated yet.
	AtEachVehicle = 1
	// AtEachSolution indicates that the constraint should be checked at each
	// solution. A constraint that is registered to be checked at each solution
	// will be checked as soon as all values for the expressions of the stops
	// of all vehicles are known, which is by definition all the stops.
	AtEachSolution = 2
	// Never indicates that the constraint should never be checked. A constraint
	// that is registered to be checked never relies completely on its estimate
	// of allowed moves to be correct. Also, not checking a constraint can
	// result in solutions that are not valid when un-planning stops.
	Never = 3
)

// CheckViolations is a list of all possible values for CheckedAt.
var CheckViolations = []CheckedAt{
	AtEachStop,
	AtEachVehicle,
	AtEachSolution,
	Never,
}

// String returns a string representation of the CheckedAt value.
func (checkViolation CheckedAt) String() string {
	switch checkViolation {
	case AtEachStop:
		return "Each Stop"
	case AtEachVehicle:
		return "Each ModelVehicle"
	case AtEachSolution:
		return "Each Solution"
	case Never:
		return "Never"
	default:
		panic("unknown check violation")
	}
}
