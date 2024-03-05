// © 2019-present nextmv.io inc

package nextroute

// SolutionInfo contains solutions and error if one raised.
type SolutionInfo struct {
	Solution
	Error error
}

// SolutionChannel is a channel of solutions.
type SolutionChannel <-chan SolutionInfo

// All returns all solutions in the channel.
func (solutions SolutionChannel) All() ([]Solution, error) {
	solutionArray := make([]Solution, 0)
	for s := range solutions {
		if s.Error != nil {
			return nil, s.Error
		}
		solutionArray = append(solutionArray, s)
	}
	return solutionArray, nil
}

// Last returns the last solution in the channel.
func (solutions SolutionChannel) Last() (Solution, error) {
	var solution Solution
	for s := range solutions {
		if s.Error != nil {
			return nil, s.Error
		}
		solution = s
	}
	return solution, nil
}
