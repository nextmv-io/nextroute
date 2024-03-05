// © 2019-present nextmv.io inc

package nextroute

// Cost is type to indicate the cost of a function.
type Cost uint64

const (
	// Constant is a constant cost function.
	Constant Cost = 0
	// LinearVehicle is a linear cost function with respect to the number of
	// vehicles.
	LinearVehicle = 1
	// LinearStop is a linear cost function with respect to the number of stops.
	LinearStop = 2
	// QuadraticVehicle is a quadratic cost function with respect to the number
	// of vehicles.
	QuadraticVehicle = 3
	// QuadraticStop is a quadratic cost function with respect to the number of
	// stops.
	QuadraticStop = 4
	// ExponentialVehicle is an exponential cost function with respect to the
	// number of vehicles.
	ExponentialVehicle = 5
	// ExponentialStop is an exponential cost function with respect to the
	// number of stops.
	ExponentialStop = 6
	// CrazyExpensive is a function that is so expensive that it should never
	// be used.
	CrazyExpensive = 7
)

var costNames = map[Cost]string{
	Constant:           `O(1)`,
	LinearVehicle:      `O(ModelVehicle)`,
	LinearStop:         `O(Stop)`,
	QuadraticVehicle:   `O(ModelVehicle^2)`,
	QuadraticStop:      `O(Stop^2)`,
	ExponentialVehicle: `O(2^ModelVehicle`,
	ExponentialStop:    `O(2^Stop)`,
	CrazyExpensive:     `O(no)`,
}

// String returns the name of the cost.
func (cost Cost) String() string {
	return costNames[cost]
}

// Complexity is the interface for constraints that have a complexity.
type Complexity interface {
	// EstimationCost returns the cost of the Estimation function.
	EstimationCost() Cost
}
