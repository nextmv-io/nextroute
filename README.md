# Nextroute

<p align="center">
  <a href="https://nextmv.io"><img src="https://cdn.prod.website-files.com/60dee0fad10d14c8ab66dd74/660d508a253a526ba1c6c267_blog-banner-experimenting-haversine-osrm-nextmv-v2-p-2000.jpeg" alt="Nextmv" width="45%"></a>
</p>
<p align="center">
    <em>Nextmv: The home for all your optimization work</em>
</p>
<p align="center">
<a href="https://github.com/nextmv-io/nextroute/actions/workflows/go-test-lint.yml" target="_blank">
    <img src="https://github.com/nextmv-io/nextroute/actions/workflows/go-test-lint.yml/badge.svg?event=push&branch=develop" alt="Go Test Lint">
</a>
<a href="https://github.com/nextmv-io/nextroute/actions/workflows/python-test.yml" target="_blank">
    <img src="https://github.com/nextmv-io/nextroute/actions/workflows/python-test.yml/badge.svg?event=push&branch=develop" alt="Python Test">
</a>
<a href="https://github.com/nextmv-io/nextroute/actions/workflows/python-lint.yml" target="_blank">
    <img src="https://github.com/nextmv-io/nextroute/actions/workflows/python-lint.yml/badge.svg?event=workflow_dispatch&branch=develop" alt="Python Lint">
</a>
<a href="https://pkg.go.dev/github.com/nextmv-io/nextroute">
  <img src="https://pkg.go.dev/badge/github.com/nextmv-io/nextroute.svg" alt="Nextroute">
</a>
<a href="https://pypi.org/project/nextroute" target="_blank">
    <img src="https://img.shields.io/pypi/v/nextroute?color=%2334D058&label=nextroute" alt="Package version">
</a>
<a href="https://pypi.org/project/nextroute" target="_blank">
    <img src="https://img.shields.io/pypi/pyversions/nextroute.svg?color=%2334D058" alt="Supported Python versions">
</a>
</p>

Welcome to Nextmv's **Nextroute**, a feature-rich Vehicle Routing Problem (VRP)
solver written in pure Go and supported in Python. Designed with a focus on
maintainability, feature-richness, and extensibility, Nextroute is built to
handle real-world applications across [all platforms that Go (cross)compiles
to](https://go.dev/doc/install/source#environment).

Our goal is not to compete on specific VRP type benchmarks, but to provide a
robust and versatile tool that can adapt to a variety of routing use-cases.
Whether you're optimizing the routes for a small fleet of delivery vans in a
city or managing complex logistics for a global supply chain, Nextroute is
equipped to help you find efficient solutions.

> [!IMPORTANT]  
> Please note that Nextroute is provided as _source-available_ software
> (not _open-source_). For further information, please refer to the
> [LICENSE](./LICENSE) file.

📖 To learn more about Nextroute, visit the [docs][docs].

## Installation

### Go

Install the Go package with the following command:

```bash
go get github.com/nextmv-io/nextroute
```

### Python

The package is hosted on [PyPI][nextroute-pypi]. Requires Python `>=3.10`.
Install using the Python package manager of your choice:

* `uv`

    ```bash
    uv add nextroute
    ```

* `pip`

    ```bash
    pip install nextroute
    ```

* `pipx`

    ```bash
    pipx install nextroute
    ```

## Features

| Feature | Description |
| ------- | ----------- |
| [Alternate stops](https://www.nextmv.io/docs/nextroute/features/alternate-stops) | Specify a set of alternate stops per vehicle for which only one should be serviced. |
| [Compatibility attributes](https://www.nextmv.io/docs/nextroute/features/compatibility-attributes) | Specify which stops are compatible with which vehicles. |
| [Capacity](https://www.nextmv.io/docs/nextroute/features/capacity) | Set capacities for vehicles and quantities (demanded or offered) at stops. |
| [Cluster constraint](https://www.nextmv.io/docs/nextroute/features/cluster-constraint) | Enforce the creation of clustered routes. |
| [Cluster objective](https://www.nextmv.io/docs/nextroute/features/cluster-objective) | Incentivize the creation of clustered routes. |
| [Custom constraints](https://www.nextmv.io/docs/nextroute/features/custom-constraints) | Implement custom constraints with Nextmv SDK. |
| [Custom data](https://www.nextmv.io/docs/nextroute/features/custom-data) | Add custom data that is preserved in the output. |
| [Custom matrices](https://www.nextmv.io/docs/nextroute/features/custom-matrices) | Use custom matrices to achieve more precise drive time. |
| [Custom objectives](https://www.nextmv.io/docs/nextroute/features/custom-objectives) | Implement custom objectives with Nextmv SDK. |
| [Custom operators](https://www.nextmv.io/docs/nextroute/features/custom-operators) | Implement custom operators with Nextmv SDK. |
| [Custom output](https://www.nextmv.io/docs/nextroute/features/custom-output) | Create a custom output for your app. |
| [Distance matrix](https://www.nextmv.io/docs/nextroute/features/distance-matrix) | Specify a distance matrix in the input that provides the distance of going from location A to B. |
| [Duration matrix](https://www.nextmv.io/docs/nextroute/features/duration-matrix) | Specify a duration matrix in the input that provides the duration of going from location A to B. |
| [Duration groups](https://www.nextmv.io/docs/nextroute/features/duration-groups) | Specify a duration that is added every time a stop in the group is approached from a stop outside of the group. |
| [Early arrival time penalty](https://www.nextmv.io/docs/nextroute/features/early-arrival-time-penalty) | Specify a penalty that is added to the objective when arriving before a stop's target arrival time. |
| [Late arrival time penalty](https://www.nextmv.io/docs/nextroute/features/late-arrival-time-penalty) | Specify a penalty that is added to the objective when arriving after a stop's target arrival time. |
| [Map data in cloud](https://www.nextmv.io/docs/nextroute/features/map-data) | Calculates duration and distance matrices using a hosted OSRM map service when running on Nextmv Cloud. Note that map data is a paid feature. |
| [Maximum route distance](https://www.nextmv.io/docs/nextroute/features/max-distance) | Specify the maximum distance that a vehicle can travel. |
| [Maximum route duration](https://www.nextmv.io/docs/nextroute/features/max-duration) | Specify the maximum duration that a vehicle can travel for. |
| [Maximum route stops](https://www.nextmv.io/docs/nextroute/features/max-stops) | Specify the maximum stops that a vehicle can visit. |
| [Maximum wait time](https://www.nextmv.io/docs/nextroute/features/max-wait) | Specify the maximum time a vehicle can wait when arriving before the start time window opens at a stop. |
| [Minimum route stops](https://www.nextmv.io/docs/nextroute/features/min-stops) | Specify the minimum stops that a vehicle should visit (applying a penalty). |
| [Nextcheck](https://www.nextmv.io/docs/nextroute/features/nextcheck) | Check which stops can be planned or why stops have been unplanned. |
| [Precedence](https://www.nextmv.io/docs/nextroute/features/precedence) | Add pickups and deliveries or specify multiple pickups before deliveries and vice versa. |
| [Stop duration](https://www.nextmv.io/docs/nextroute/features/stop-duration) | Specify the time it takes to service a stop. |
| [Stop duration multiplier](https://www.nextmv.io/docs/nextroute/features/stop-duration-multiplier) | Specify a multiplier on time it takes a vehicle to service a stop. |
| [Stop groups](https://www.nextmv.io/docs/nextroute/features/stop-groups) | Specify stops that must be assigned together on the same route, with no further requirements. |
| [Stop mixing](https://www.nextmv.io/docs/nextroute/features/stop-mixing) | Specify properties of stops which can not be on the vehicle at the same time. |
| [Time windows](https://www.nextmv.io/docs/nextroute/features/time-windows) | Specify the time window in which a stop must start service. |
| [Unplanned penalty](https://www.nextmv.io/docs/nextroute/features/unplanned-penalty) | Specify a penalty that is added to the objective to leave a stop unplanned when all constraints cannot be fulfilled. |
| [Vehicle activation penalty](https://www.nextmv.io/docs/nextroute/features/vehicle-activation-penalty) | Specify a penalty that is added to the objective for activating (using) a vehicle. |
| [Vehicle initial stops](https://www.nextmv.io/docs/nextroute/features/vehicle-initial-stops) | Specify initial stops planned on a vehicle. |
| [Vehicle start/end location](https://www.nextmv.io/docs/nextroute/features/vehicle-start-end-location) | Specify optional starting and ending locations for vehicles. |
| [Vehicle start/end time](https://www.nextmv.io/docs/nextroute/features/vehicle-start-end-time) | Specify optional starting and ending time for a vehicle. |

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

[docs]: https://docs.nextmv.io/nextroute
[nextroute-pypi]: https://pypi.org/project/nextroute/
