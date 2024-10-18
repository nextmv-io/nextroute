// © 2019-present nextmv.io inc

package nextroute

// NewStopBalanceObjective returns a new StopBalanceObjective.
func NewStopBalanceObjective() ModelObjective {
	return &balanceObjectiveImpl{}
}

type balanceObjectiveImpl struct {
}

func (t *balanceObjectiveImpl) EstimateDeltaValue(
	move Move,
) float64 {
	solution := move.Solution()
	oldMax, newMax := t.maxStops(solution, move)
	return float64(newMax - oldMax)
}

func (t *balanceObjectiveImpl) Value(solution Solution) float64 {
	maxBefore, _ := t.maxStops(solution, nil)
	return float64(maxBefore)
}

func (t *balanceObjectiveImpl) maxStops(solution Solution, move SolutionMoveStops) (int, int) {
	max := 0
	maxBefore := 0
	moveExists := move != nil
	var vehicle SolutionVehicle
	if moveExists {
		vehicle = move.Vehicle()
	}

	for _, v := range solution.(*solutionImpl).vehicles {
		numberOfStops := v.NumberOfStops()
		if max < numberOfStops {
			max = numberOfStops
		}
		if maxBefore < numberOfStops {
			maxBefore = numberOfStops
		}
		if moveExists && v.Index() == vehicle.Index() {
			length := move.StopPositionsLength()
			if max < numberOfStops+length {
				max = numberOfStops + length
			}
		}
	}
	return maxBefore, max
}

func (t *balanceObjectiveImpl) String() string {
	return "stop_balance"
}
