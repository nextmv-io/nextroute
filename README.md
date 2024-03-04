# nextroute

The Nextmv **nextroute** solver optimizes the routes of a fleet of vehicles under various constraints.

## License

Please note that nextroute is provided as _source-available_ software (not
_open-source_). For further information, please refer to the [LICENSE](./LICENSE)
file.

## Usage

A first run can be done with the following command:

```bash
go run cmd/main.go -runner.input.path cmd/input.json -solve.duration 5s
```

This will run the solver for 5 seconds and output the result to the console.

In order to start a _new project_, please refer to the sample app in the
[community-apps repository](https://github.com/nextmv-io/community-apps/tree/develop/nextroute).
If you have [Nextmv CLI](https://docs.nextmv.io/docs/platform/installation#nextmv-cli)
installed, you can create a new project with the following command:

```bash
nextmv community clone -a nextroute
```

For further information on how to get started, features, deployment, etc.,
please refer to the [official documentation](https://docs.nextmv.io/docs/vehicle-routing).
