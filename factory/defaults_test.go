// © 2019-present nextmv.io inc

package factory

import (
	"reflect"
	"testing"
	"time"

	"github.com/nextmv-io/nextroute/schema"
)

func Test_applyDefaults(t *testing.T) {
	location1 := schema.Location{Lon: 1, Lat: 2}
	location2 := schema.Location{Lon: 3, Lat: 4}
	v1 := 10
	v2 := 20
	v3 := 30
	activationPenalty1 := 100
	activationPenalty2 := 200
	sampleTime1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	sampleTime2 := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	defaults := schema.Defaults{
		Vehicles: &schema.VehicleDefaults{
			Capacity:          v1,
			StartLevel:        v3,
			StartLocation:     &location1,
			EndLocation:       &location1,
			StartTime:         &sampleTime1,
			EndTime:           &sampleTime1,
			ActivationPenalty: &activationPenalty1,
		},
		Stops: &schema.StopDefaults{
			Quantity: v1,
			StartTimeWindow: &[]time.Time{
				sampleTime1,
				sampleTime1,
			},
			MaxWait:           &v1,
			TargetArrivalTime: &sampleTime1,
		},
	}

	type args struct {
		i schema.Input
	}
	tests := []struct {
		args args
		want schema.Input
		name string
	}{
		{
			name: "defaults are applied",
			args: args{
				i: schema.Input{
					Defaults: &defaults,
					Vehicles: []schema.Vehicle{
						{ID: "v1"},
					},
					Stops: []schema.Stop{
						{ID: "s1"},
					},
				},
			},
			want: schema.Input{
				Defaults: &defaults,
				Vehicles: []schema.Vehicle{
					{
						ID:                "v1",
						Capacity:          v1,
						StartLevel:        v3,
						StartLocation:     &location1,
						EndLocation:       &location1,
						StartTime:         &sampleTime1,
						EndTime:           &sampleTime1,
						ActivationPenalty: &activationPenalty1,
					},
				},
				Stops: []schema.Stop{
					{
						ID:       "s1",
						Quantity: v1,
						StartTimeWindow: &[]time.Time{
							sampleTime1,
							sampleTime1,
						},
						MaxWait:           &v1,
						TargetArrivalTime: &sampleTime1,
					},
				},
			},
		},
		{
			name: "defaults are overridden",
			args: args{
				i: schema.Input{
					Defaults: &defaults,
					Vehicles: []schema.Vehicle{
						{
							ID:                "v1",
							Capacity:          v2,
							StartLevel:        v2,
							StartLocation:     &location2,
							EndLocation:       &location2,
							StartTime:         &sampleTime2,
							EndTime:           &sampleTime2,
							ActivationPenalty: &activationPenalty2,
						},
					},
					Stops: []schema.Stop{
						{
							ID:       "s1",
							Quantity: v2,
							StartTimeWindow: &[]time.Time{
								sampleTime2,
								sampleTime2,
							},
							MaxWait:           &v2,
							TargetArrivalTime: &sampleTime2,
						},
					},
				},
			},
			want: schema.Input{
				Defaults: &defaults,
				Vehicles: []schema.Vehicle{
					{
						ID:                "v1",
						Capacity:          v2,
						StartLevel:        v2,
						StartLocation:     &location2,
						EndLocation:       &location2,
						StartTime:         &sampleTime2,
						EndTime:           &sampleTime2,
						ActivationPenalty: &activationPenalty2,
					},
				},
				Stops: []schema.Stop{
					{
						ID:       "s1",
						Quantity: v2,
						StartTimeWindow: &[]time.Time{
							sampleTime2,
							sampleTime2,
						},
						MaxWait:           &v2,
						TargetArrivalTime: &sampleTime2,
					},
				},
			},
		},
		{
			name: "defaults are not applied because they are not present",
			args: args{
				i: schema.Input{
					Defaults: nil,
					Vehicles: []schema.Vehicle{
						{ID: "v1"},
					},
					Stops: []schema.Stop{
						{ID: "s1"},
					},
				},
			},
			want: schema.Input{
				Defaults: nil,
				Vehicles: []schema.Vehicle{
					{ID: "v1"},
				},
				Stops: []schema.Stop{
					{ID: "s1"},
				},
			},
		},
		{
			name: "vehicle defaults are applied and stop defaults are skipped",
			args: args{
				i: schema.Input{
					Defaults: &schema.Defaults{
						Vehicles: &schema.VehicleDefaults{
							Capacity:          v1,
							StartLevel:        v2,
							StartLocation:     &location1,
							EndLocation:       &location1,
							StartTime:         &sampleTime1,
							EndTime:           &sampleTime1,
							ActivationPenalty: &activationPenalty1,
						},
					},
					Vehicles: []schema.Vehicle{
						{ID: "v1"},
					},
					Stops: []schema.Stop{
						{ID: "s1"},
					},
				},
			},
			want: schema.Input{
				Defaults: &schema.Defaults{
					Vehicles: &schema.VehicleDefaults{
						Capacity:          v1,
						StartLevel:        v2,
						StartLocation:     &location1,
						EndLocation:       &location1,
						StartTime:         &sampleTime1,
						EndTime:           &sampleTime1,
						ActivationPenalty: &activationPenalty1,
					},
				},
				Vehicles: []schema.Vehicle{
					{
						ID:                "v1",
						Capacity:          v1,
						StartLevel:        v2,
						StartLocation:     &location1,
						EndLocation:       &location1,
						StartTime:         &sampleTime1,
						EndTime:           &sampleTime1,
						ActivationPenalty: &activationPenalty1,
					},
				},
				Stops: []schema.Stop{
					{ID: "s1"},
				},
			},
		},
		{
			name: "stop defaults are applied and vehicle defaults are skipped",
			args: args{
				i: schema.Input{
					Defaults: &schema.Defaults{
						Stops: &schema.StopDefaults{
							Quantity: v1,
							StartTimeWindow: &[]time.Time{
								sampleTime1,
								sampleTime1,
							},
							MaxWait:           &v1,
							TargetArrivalTime: &sampleTime1,
						},
					},
					Vehicles: []schema.Vehicle{
						{ID: "v1"},
					},
					Stops: []schema.Stop{
						{ID: "s1"},
					},
				},
			},
			want: schema.Input{
				Defaults: &schema.Defaults{
					Stops: &schema.StopDefaults{
						Quantity: v1,
						StartTimeWindow: &[]time.Time{
							sampleTime1,
							sampleTime1,
						},
						MaxWait:           &v1,
						TargetArrivalTime: &sampleTime1,
					},
				},
				Vehicles: []schema.Vehicle{
					{ID: "v1"},
				},
				Stops: []schema.Stop{
					{
						ID:       "s1",
						Quantity: v1,
						StartTimeWindow: &[]time.Time{
							sampleTime1,
							sampleTime1,
						},
						MaxWait:           &v1,
						TargetArrivalTime: &sampleTime1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applyDefaults(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyDefaults() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_fields(t *testing.T) {
	type args[T schema.VehicleDefaults | schema.StopDefaults] struct {
		defaults *T
	}
	type test[T schema.VehicleDefaults | schema.StopDefaults] struct {
		name string
		args args[T]
		want []string
	}
	vehicleTests := []test[schema.VehicleDefaults]{
		{
			name: "vehicle fields",
			args: args[schema.VehicleDefaults]{
				defaults: &schema.VehicleDefaults{},
			},
			want: []string{
				"Capacity",
				"StartLevel",
				"StartLocation",
				"EndLocation",
				"Speed",
				"StartTime",
				"EndTime",
				"MinStops",
				"MinStopsPenalty",
				"MaxStops",
				"MaxDistance",
				"MaxDuration",
				"MaxWait",
				"CompatibilityAttributes",
				"ActivationPenalty",
				"AlternateStops",
			},
		},
	}
	for _, tt := range vehicleTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fields(tt.args.defaults); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fields() on vehicleTests = %v, want %v", got, tt.want)
			}
		})
	}

	stopTests := []test[schema.StopDefaults]{
		{
			name: "stop fields",
			args: args[schema.StopDefaults]{
				defaults: &schema.StopDefaults{},
			},
			want: []string{
				"UnplannedPenalty",
				"Quantity",
				"StartTimeWindow",
				"MaxWait",
				"Duration",
				"TargetArrivalTime",
				"EarlyArrivalTimePenalty",
				"LateArrivalTimePenalty",
				"CompatibilityAttributes",
			},
		},
	}
	for _, tt := range stopTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fields(tt.args.defaults); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fields() on stopTests = %v, want %v", got, tt.want)
			}
		})
	}
}
