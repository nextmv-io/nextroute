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
	oldMax := t.Max(solution, nil)
	NewMax := t.Max(solution, move)
	return float64(NewMax - oldMax)
}

func (t *balanceObjectiveImpl) Value(solution Solution) float64 {
	return float64(t.Max(solution, nil))
}

func (t *balanceObjectiveImpl) Max(solution Solution, move SolutionMoveStops) int {
	max := 0
	moveExists := move != nil
	var vehicle SolutionVehicle
	if moveExists {
		vehicle = move.Vehicle()
	}

	for _, v := range solution.Vehicles() {
		if max < v.NumberOfStops() {
			max = v.NumberOfStops()
		}
		if moveExists && v.Index() == vehicle.Index() {
			if max < v.NumberOfStops()+move.StopPositionsLength() {
				max = v.NumberOfStops() + move.StopPositionsLength()
			}
		}
	}
	return max
}

func (t *balanceObjectiveImpl) String() string {
	return "stop_balance"
}
