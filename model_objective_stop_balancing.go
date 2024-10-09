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
	return (NewMax - oldMax)
}

func (t *balanceObjectiveImpl) Value(solution Solution) float64 {
	return t.Max(solution, nil)
}

func (t *balanceObjectiveImpl) Max(solution Solution, move SolutionMoveStops) float64 {
	max := 0.0
	moveExists := move != nil
	var vehicle SolutionVehicle
	if moveExists {
		vehicle = move.Vehicle()
	}

	for _, v := range solution.Vehicles() {
		if max < float64(v.NumberOfStops()) {
			max = float64(v.NumberOfStops())
		}
		if moveExists && v.Index() == vehicle.Index() {
			if max < float64(v.NumberOfStops()+move.StopPositionsLength()) {
				max = float64(v.NumberOfStops() + move.StopPositionsLength())
			}
		}
	}
	return max
}

func (t *balanceObjectiveImpl) String() string {
	return "stop_balance"
}
