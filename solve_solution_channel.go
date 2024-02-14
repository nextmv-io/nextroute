package nextroute

// SolutionChannel is a channel of solutions.
type SolutionChannel <-chan Solution

// All returns all solutions in the channel.
func (solutions SolutionChannel) All() []Solution {
	solutionArray := make([]Solution, 0)
	for s := range solutions {
		solutionArray = append(solutionArray, s)
	}
	return solutionArray
}

// Last returns the last solution in the channel.
func (solutions SolutionChannel) Last() Solution {
	var solution Solution
	for s := range solutions {
		solution = s
	}
	return solution
}
