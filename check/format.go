package check

import (
	"context"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/factory"
	sdkNextRoute "github.com/nextmv-io/sdk/nextroute"
	sdkCheck "github.com/nextmv-io/sdk/nextroute/check"
	runSchema "github.com/nextmv-io/sdk/run/schema"
)

// Format formats a solution in a basic format using factory.ToSolutionOutput
// to format each solution and also allows to check the solutions and add the
// check to the output of each solution.
func Format(
	ctx context.Context,
	options any,
	checkOptions sdkCheck.Options,
	progressioner sdkNextRoute.Progressioner,
	solutions ...sdkNextRoute.Solution,
) (runSchema.Output, error) {
	return nextroute.Format(
		ctx,
		options,
		progressioner,
		func(solution sdkNextRoute.Solution) any {
			solutionOutput := factory.ToSolutionOutput(solution)
			if checkOptions.Duration > 0 &&
				sdkCheck.ToVerbosity(checkOptions.Verbosity) != sdkCheck.Off {
				solutionCheckOutput, err := SolutionCheck(
					solution,
					checkOptions,
				)
				if err == nil {
					solutionOutput.Check = &solutionCheckOutput
				}
			}
			return solutionOutput
		},
		solutions...,
	), nil
}
