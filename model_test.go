// © 2019-present nextmv.io inc

package nextroute_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestModel(t *testing.T) {
	pss := planSingleStops()
	ps := planPairSequences()
	vt := vehicleType("truck")
	v := vehicle("truck", depot())

	model, err := createModel(
		createInput(
			0,
			pss,
			ps,
			[]VehicleType{vt},
			[]Vehicle{v},
		),
	)
	if err != nil {
		t.Error(err)
	}
	if model.DurationUnit() != time.Second {
		t.Error("duration unit is not correct, expected second")
	}
	if model.DistanceUnit() != common.Meters {
		t.Error("distance unit is not correct, expected meters")
	}
	if len(model.VehicleTypes()) != 1 {
		t.Error("vehicle types length is not correct")
	}
	if len(model.Vehicles()) != 1 {
		t.Error("input length is not correct")
	}
	if len(model.PlanUnits()) != len(pss)+len(ps) {
		t.Error("plan planUnit length is not correct")
	}
	if len(model.Stops()) != len(pss)+len(ps)*2+1 {
		t.Errorf(
			"stops length is not correct,"+
				" expected %v from single stops,"+
				" expected %v from sequences,"+
				" expected 1 from depot, in total got %v",
			len(pss),
			len(ps)*2,
			len(model.Stops()),
		)
	}
}

func createModel(input Input) (nextroute.Model, error) {
	model, err := nextroute.NewModel()
	if err != nil {
		return nil, err
	}

	model.SetRandom(rand.New(rand.NewSource(input.Seed)))

	serviceDuration := nextroute.NewStopDurationExpression("serviceDuration", 0.0)

	for _, planSingleStop := range input.PlanSingleStops {
		location, err := common.NewLocation(
			planSingleStop.Stop.Location.Lon,
			planSingleStop.Stop.Location.Lat,
		)
		if err != nil {
			return nil, err
		}

		stop, err := model.NewStop(location)
		if err != nil {
			return nil, err
		}
		stop.SetID(planSingleStop.Stop.Name)

		serviceDuration.SetDuration(stop, planSingleStop.Stop.ServiceDuration)

		_, err = model.NewPlanSingleStop(stop)
		if err != nil {
			return nil, err
		}
	}

	for _, planSequence := range input.PlanSequences {
		stops := make(nextroute.ModelStops, len(planSequence.Stops))
		for idx, s := range planSequence.Stops {
			location, err := common.NewLocation(
				s.Location.Lon,
				s.Location.Lat,
			)
			if err != nil {
				return nil, err
			}
			stop, err := model.NewStop(location)
			if err != nil {
				return nil, err
			}
			stop.SetID(s.Name)

			serviceDuration.SetDuration(stop, s.ServiceDuration)

			stops[idx] = stop
		}
		_, err := model.NewPlanSequence(stops)
		if err != nil {
			return nil, err
		}
	}

	vehicleTypes := make(map[string]nextroute.ModelVehicleType)

	for _, vt := range input.VehicleTypes {
		if vt.Speed.MetersPerSecond < 0 {
			return nil,
				fmt.Errorf(
					"v type %s has invalid speed %f",
					vt.Name,
					vt.Speed.MetersPerSecond,
				)
		}

		vehicleType, err := model.NewVehicleType(
			nextroute.NewTimeIndependentDurationExpression(
				nextroute.NewTravelDurationExpression(
					nextroute.NewHaversineExpression(),
					common.NewSpeed(
						vt.Speed.MetersPerSecond,
						common.MetersPerSecond,
					),
				),
			),
			nextroute.NewDurationExpression(
				"travelDuration",
				serviceDuration,
				common.Second,
			),
		)
		if err != nil {
			return nil, err
		}

		vehicleType.SetID(vt.Name)

		vehicleTypes[vt.Name] = vehicleType
	}

	for _, v := range input.Vehicles {
		vehicleType, ok := vehicleTypes[v.Type]
		if !ok {
			return nil,
				fmt.Errorf(
					"v %s has invalid v type %s",
					v.Name,
					v.Type,
				)
		}
		var location common.Location
		if v.StartLocation.IsValid {
			location, err = common.NewLocation(
				v.StartLocation.Lon,
				v.StartLocation.Lat,
			)
			if err != nil {
				return nil, err
			}
		} else {
			location = common.NewInvalidLocation()
		}
		start, err := model.NewStop(location)
		if err != nil {
			return nil, err
		}

		end := start

		if v.EndLocation.Lat != v.StartLocation.Lat ||
			v.EndLocation.Lon != v.StartLocation.Lon {
			location, err := common.NewLocation(
				v.EndLocation.Lon,
				v.EndLocation.Lat,
			)
			if err != nil {
				return nil, err
			}
			end, err = model.NewStop(location)
			if err != nil {
				return nil, err
			}
		}

		startTime := model.Epoch()

		if v.StartTime != nil {
			startTime = *v.StartTime
		}

		vehicle, err := model.NewVehicle(
			vehicleType,
			startTime,
			start,
			end,
		)
		if err != nil {
			return nil, err
		}
		vehicle.SetID(v.Name)
	}

	return model, nil
}

type Location struct {
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	IsValid bool    `json:"is_valid"`
}
type Stop struct {
	Name            string        `json:"name"`
	Location        Location      `json:"location"`
	ServiceDuration time.Duration `json:"service_duration"`
}

type PlanSingleStop struct {
	Stop Stop `json:"stop"`
}

type PlanSequence struct {
	Stops []Stop `json:"stops"`
}

type Speed struct {
	MetersPerSecond float64 `json:"meters_per_second"`
}

type VehicleType struct {
	Name  string `json:"name"`
	Speed Speed  `json:"speed"`
}

type Vehicle struct {
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	StartLocation Location   `json:"start_location"`
	EndLocation   Location   `json:"end_location"`
}

type Input struct {
	Seed            int64            `json:"seed"`
	PlanSingleStops []PlanSingleStop `json:"plan_single_stops"`
	PlanSequences   []PlanSequence   `json:"plan_sequences"`
	VehicleTypes    []VehicleType    `json:"vehicle_types"`
	Vehicles        []Vehicle        `json:"vehicles"`
}

func planSingleStops() []PlanSingleStop {
	return []PlanSingleStop{
		{
			Stop: Stop{
				Name: "s1",
				Location: Location{
					Lon: -74.04866,
					Lat: 4.69018,
				},
			},
		},
		{
			Stop: Stop{
				Name: "s2",
				Location: Location{
					Lon: -74.044215,
					Lat: 4.693907,
				},
			},
		},
		{
			Stop: Stop{
				Name: "s3",
				Location: Location{
					Lon: -74.040,
					Lat: 4.696,
				},
			},
		},
	}
}

func planPairSequences() []PlanSequence {
	return []PlanSequence{
		{
			Stops: []Stop{
				{
					Name: "s1",
					Location: Location{
						Lon: -74.04866,
						Lat: 4.69018,
					},
				},
				{
					Name: "s2",
					Location: Location{
						Lon: -74.044215,
						Lat: 4.693907,
					},
				},
			},
		},
		{
			Stops: []Stop{
				{
					Name: "s3",
					Location: Location{
						Lon: -74.04866,
						Lat: 4.693907,
					},
				},
				{
					Name: "s4",
					Location: Location{
						Lon: -74.044215,
						Lat: 4.69018,
					},
				},
			},
		},
	}
}

func planTripleSequence() []PlanSequence {
	return []PlanSequence{
		{
			Stops: []Stop{
				{
					Name: "s1",
					Location: Location{
						Lon: -74.04866,
						Lat: 4.69018,
					},
				},
				{
					Name: "s2",
					Location: Location{
						Lon: -74.044215,
						Lat: 4.693907,
					},
				},
				{
					Name: "s3",
					Location: Location{
						Lon: -74.04866,
						Lat: 4.693907,
					},
				},
			},
		},
	}
}

func depot() Location {
	return Location{
		Lon:     -74.044219,
		Lat:     4.686293,
		IsValid: true,
	}
}

func vehicleTypes(names ...string) []VehicleType {
	vts := make([]VehicleType, len(names))
	for idx, name := range names {
		vts[idx] = vehicleType(name)
	}
	return vts
}

func vehicleType(name string) VehicleType {
	return VehicleType{
		Name: name,
		Speed: Speed{
			MetersPerSecond: 10.0,
		},
	}
}

func vehicle(vehicleType string, location Location) Vehicle {
	return Vehicle{
		Type:          vehicleType,
		StartLocation: location,
		EndLocation:   location,
	}
}

func vehicles(vehicleType string, location Location, count int) []Vehicle {
	vehicles := make([]Vehicle, count)
	for i := 0; i < count; i++ {
		vehicles[i] = vehicle(vehicleType, location)
	}
	return vehicles
}

func input(
	vehicleTypes []VehicleType,
	vehicles []Vehicle,
	planSingleStop []PlanSingleStop,
	planSequences []PlanSequence,
) Input {
	return createInput(
		0,
		planSingleStop,
		planSequences,
		vehicleTypes,
		vehicles,
	)
}

func singleVehiclePlanSingleStopsModel() Input {
	return input(
		vehicleTypes("truck"),
		vehicles("truck", depot(), 1),
		planSingleStops(),
		nil,
	)
}

func singleVehiclePlanSequenceModel() Input {
	return input(
		vehicleTypes("truck"),
		vehicles("truck", depot(), 1),
		nil,
		planPairSequences(),
	)
}

func createInput(
	seed int64,
	planSingleStop []PlanSingleStop,
	planSequences []PlanSequence,
	vehicleTypes []VehicleType,
	vehicles []Vehicle,
) Input {
	return Input{
		Seed:            seed,
		PlanSingleStops: planSingleStop,
		PlanSequences:   planSequences,
		VehicleTypes:    vehicleTypes,
		Vehicles:        vehicles,
	}
}
