import json

import nextroute

with open("../cmd/input.json") as f:
    data = json.load(f)

input = nextroute.schema.Input.from_dict(data)
options = nextroute.Options(
    solve=nextroute.ParallelSolveOptions(
        duration=2,
    ),
    format=nextroute.FormatOptions(
        disable=nextroute.DisableFormatOptions(
            progression=True,
        )
    ),
)
output = nextroute.solve(input, options)

with open("output1.json", "w") as f:
    json.dump(output.to_dict(), f, indent=2)
