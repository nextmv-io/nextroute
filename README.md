# Nextroute

Welcome to Nextmv's **Nextroute**, a feature-rich Vehicle Routing Problem (VRP)
solver written in pure Go. Designed with a focus on maintainability,
feature-richness, and extensibility, Nextroute is built to handle real-world
applications across [all platforms that Go (cross)compiles
to](https://go.dev/doc/install/source#environment).

Our goal is not to compete on specific VRP type benchmarks, but to provide a
robust and versatile tool that can adapt to a variety of routing use-cases.
Whether you're optimizing the routes for a small fleet of delivery vans in a
city or managing complex logistics for a global supply chain, Nextroute is
equipped to help you find efficient solutions.

You can work with Nextroute in a variety of ways:

* Go package: Import the `nextroute` package in your Go project and use the
  solver directly.
* Python package: Use the `nextroute` Python package as an interface to the Go
  solver.

## Features

| Feature | Description |
| ------- | ----------- |
| [Alternate stops](https://www.nextmv.io/docs/vehicle-routing/features/alternate-stops) | Specify a set of alternate stops per vehicle for which only one should be serviced. |
| [Compatibility attributes](https://www.nextmv.io/docs/vehicle-routing/features/compatibility-attributes) | Specify which stops are compatible with which vehicles. |
| [Capacity](https://www.nextmv.io/docs/vehicle-routing/features/capacity) | Set capacities for vehicles and quantities (demanded or offered) at stops. |
| [Cluster constraint](https://www.nextmv.io/docs/vehicle-routing/features/cluster-constraint) | Enforce the creation of clustered routes. |
| [Cluster objective](https://www.nextmv.io/docs/vehicle-routing/features/cluster-objective) | Incentivize the creation of clustered routes. |
| [Custom constraints](https://www.nextmv.io/docs/vehicle-routing/features/custom-constraints) | Implement custom constraints with Nextmv SDK. |
| [Custom data](https://www.nextmv.io/docs/vehicle-routing/features/custom-data) | Add custom data that is preserved in the output. |
| [Custom matrices](https://www.nextmv.io/docs/vehicle-routing/features/custom-matrices) | Use custom matrices to achieve more precise drive time. |
| [Custom objectives](https://www.nextmv.io/docs/vehicle-routing/features/custom-objectives) | Implement custom objectives with Nextmv SDK. |
| [Custom operators](https://www.nextmv.io/docs/vehicle-routing/features/custom-operators) | Implement custom operators with Nextmv SDK. |
| [Custom output](https://www.nextmv.io/docs/vehicle-routing/features/custom-output) | Create a custom output for your app. |
| [Distance matrix](https://www.nextmv.io/docs/vehicle-routing/features/distance-matrix) | Specify a distance matrix in the input that provides the distance of going from location A to B. |
| [Duration matrix](https://www.nextmv.io/docs/vehicle-routing/features/duration-matrix) | Specify a duration matrix in the input that provides the duration of going from location A to B. |
| [Duration groups](https://www.nextmv.io/docs/vehicle-routing/features/duration-groups) | Specify a duration that is added every time a stop in the group is approached from a stop outside of the group. |
| [Early arrival time penalty](https://www.nextmv.io/docs/vehicle-routing/features/early-arrival-time-penalty) | Specify a penalty that is added to the objective when arriving before a stop's target arrival time. |
| [Late arrival time penalty](https://www.nextmv.io/docs/vehicle-routing/features/late-arrival-time-penalty) | Specify a penalty that is added to the objective when arriving after a stop's target arrival time. |
| [Map data in cloud](https://www.nextmv.io/docs/vehicle-routing/features/map-data) | Calculates duration and distance matrices using a hosted OSRM map service when running on Nextmv Cloud. Note that map data is a paid feature. |
| [Maximum route distance](https://www.nextmv.io/docs/vehicle-routing/features/max-distance) | Specify the maximum distance that a vehicle can travel. |
| [Maximum route duration](https://www.nextmv.io/docs/vehicle-routing/features/max-duration) | Specify the maximum duration that a vehicle can travel for. |
| [Maximum route stops](https://www.nextmv.io/docs/vehicle-routing/features/max-stops) | Specify the maximum stops that a vehicle can visit. |
| [Maximum wait time](https://www.nextmv.io/docs/vehicle-routing/features/max-wait) | Specify the maximum time a vehicle can wait when arriving before the start time window opens at a stop. |
| [Minimum route stops](https://www.nextmv.io/docs/vehicle-routing/features/min-stops) | Specify the minimum stops that a vehicle should visit (applying a penalty). |
| [Nextcheck](https://www.nextmv.io/docs/vehicle-routing/features/nextcheck) | Check which stops can be planned or why stops have been unplanned. |
| [Precedence](https://www.nextmv.io/docs/vehicle-routing/features/precedence) | Add pickups and deliveries or specify multiple pickups before deliveries and vice versa. |
| [Stop duration](https://www.nextmv.io/docs/vehicle-routing/features/stop-duration) | Specify the time it takes to service a stop. |
| [Stop duration multiplier](https://www.nextmv.io/docs/vehicle-routing/features/stop-duration-multiplier) | Specify a multiplier on time it takes a vehicle to service a stop. |
| [Stop groups](https://www.nextmv.io/docs/vehicle-routing/features/stop-groups) | Specify stops that must be assigned together on the same route, with no further requirements. |
| [Stop mixing](https://www.nextmv.io/docs/vehicle-routing/features/stop-mixing) | Specify properties of stops which can not be on the vehicle at the same time. |
| [Time windows](https://www.nextmv.io/docs/vehicle-routing/features/time-windows) | Specify the time window in which a stop must start service. |
| [Unplanned penalty](https://www.nextmv.io/docs/vehicle-routing/features/unplanned-penalty) | Specify a penalty that is added to the objective to leave a stop unplanned when all constraints cannot be fulfilled. |
| [Vehicle activation penalty](https://www.nextmv.io/docs/vehicle-routing/features/vehicle-activation-penalty) | Specify a penalty that is added to the objective for activating (using) a vehicle. |
| [Vehicle initial stops](https://www.nextmv.io/docs/vehicle-routing/features/vehicle-initial-stops) | Specify initial stops planned on a vehicle. |
| [Vehicle start/end location](https://www.nextmv.io/docs/vehicle-routing/features/vehicle-start-end-location) | Specify optional starting and ending locations for vehicles. |
| [Vehicle start/end time](https://www.nextmv.io/docs/vehicle-routing/features/vehicle-start-end-time) | Specify optional starting and ending time for a vehicle. |

## License

Please note that Nextroute is provided as _source-available_ software (not
_open-source_). For further information, please refer to the [LICENSE](./LICENSE.md)
file.

## Installation

* Go

    Install the Go package with the following command:

    ```bash
    go get github.com/nextmv-io/nextroute
    ```

* Python

    Install the Python package with the following command:

    ```bash
    pip install nextroute
    ```

## Usage

For further information on how to get started, features, deployment, etc.,
please refer to the [official
documentation](https://www.nextmv.io/docs/vehicle-routing).

### Go

A first run can be done with the following command. Stand at the root of the
repository and run:

```bash
go run cmd/main.go -runner.input.path cmd/input.json -solve.duration 5s
```

This will run the solver for 5 seconds and output the result to the console.

In order to start a _new project_, please refer to the sample app in the
[community-apps repository](https://github.com/nextmv-io/community-apps/tree/develop/go-nextroute).
If you have [Nextmv CLI](https://www.nextmv.io/docs/platform/installation#nextmv-cli)
installed, you can create a new project with the following command:

```bash
nextmv community clone -a go-nextroute
```

### Python

A first run can be done by executing the following script. Stand at the root of
the repository and execute it:

```python
import json

import nextroute

with open("cmd/input.json") as f:
    data = json.load(f)

input = nextroute.schema.Input.from_dict(data)
options = nextroute.Options(
    solve=nextroute.ParallelSolveOptions(
        duration=5,
    ),
)

output = nextroute.solve(input, options)
print(json.dumps(output.to_dict(), indent=2))
```

This will run the solver for 5 seconds and output the result to the console.

In order to start a _new project_, please refer to the sample app in the
[community-apps repository](https://github.com/nextmv-io/community-apps/tree/develop/python-nextroute).
If you have [Nextmv CLI](https://www.nextmv.io/docs/platform/installation#nextmv-cli)
installed, you can create a new project with the following command:

```bash
nextmv community clone -a python-nextroute
```

## Local benchmarking

To run the go benchmarks locally, you can use the following command:

```bash
go test -benchmem -timeout 20m -run=^$ -count 10 -bench "^Benchmark" ./...
```

In order to compare changes from a PR with the latest `develop` version, you can
use `benchstat`.

```bash
# on the develop branch (or any other branch)
go test -benchmem -timeout 20m -run=^$ -count 10 -bench "^Benchmark" ./...\
 | tee develop.txt
# on the new branch (or any other branch)
go test -benchmem -timeout 20m -run=^$ -count 10 -bench "^Benchmark" ./...\
 | tee new.txt
# compare the two
benchstat develop.txt new.txt
```

## Versioning

We try our best to version our software thoughtfully and only break APIs and
behaviors when we have a good reason to.

* Minor (`v1.^.0`) tags: new features, might be breaking.
* Patch (`v1.0.^`) tags: bug fixes.
