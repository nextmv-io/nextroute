// © 2019-present nextmv.io inc

package nextroute

// NewStopBalanceObjective returns a new StopBalanceObjective.
func NewStopBalanceObjective() ModelObjective {
	return &BalancedObjective{}
}

// BalancedObjective is an objective that tries to balance the number of stops
// across all vehicles by minimizing the maximum number of stops assigned to a
// single vehicle.
type BalancedObjective struct{}

// EstimateDeltaValue implements the ModelObjective interface.
func (t *BalancedObjective) EstimateDeltaValue(
	move Move,
) float64 {
	solution := move.Solution()
	oldMax, newMax := t.maxStops(solution, move)
	return float64(newMax - oldMax)
}

// Value implements the ModelObjective interface.
func (t *BalancedObjective) Value(solution *Solution) float64 {
	maxBefore, _ := t.maxStops(solution, nil)
	return float64(maxBefore)
}

func (t *BalancedObjective) maxStops(solution *Solution, move SolutionMoveStops) (int, int) {
	maximum := 0
	maximumBefore := 0
	moveExists := move != nil
	var vehicle SolutionVehicle
	if moveExists {
		vehicle = move.Vehicle()
	}

	for _, v := range solution.vehicles {
		numberOfStops := v.NumberOfStops()
		if maximum < numberOfStops {
			maximum = numberOfStops
		}
		if maximumBefore < numberOfStops {
			maximumBefore = numberOfStops
		}
		if moveExists && v.Index() == vehicle.Index() {
			length := move.StopPositionsLength()
			if maximum < numberOfStops+length {
				maximum = numberOfStops + length
			}
		}
	}
	return maximumBefore, maximum
}

func (t *BalancedObjective) String() string {
	return "stop_balance"
}
