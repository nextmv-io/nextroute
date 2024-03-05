// © 2019-present nextmv.io inc

package factory

import (
	"reflect"
	"testing"

	"github.com/nextmv-io/nextroute/schema"
)

func Test_resources(t *testing.T) {
	sampleInt := 1
	sampleFloat := 1.0
	type args[T schema.Vehicle | schema.Stop] struct {
		entity T
		name   string
		sense  int
	}
	type test[T schema.Vehicle | schema.Stop] struct {
		want    map[string]float64
		name    string
		args    args[T]
		wantErr bool
	}
	testsVehicles := []test[schema.Vehicle]{
		{
			name: "map of attributes type - any",
			args: args[schema.Vehicle]{
				entity: schema.Vehicle{
					ID: "v1",
					Capacity: map[string]any{
						"a": 1,
						"b": 2,
					},
					StartLevel: map[string]any{
						"a": 0,
						"b": 0,
					},
				},
				name:  "Capacity",
				sense: 1,
			},
			want: map[string]float64{
				"a": 1,
				"b": 2,
			},
			wantErr: false,
		},
		{
			name: "map of attributes type - int",
			args: args[schema.Vehicle]{
				entity: schema.Vehicle{
					ID: "v1",
					Capacity: map[string]int{
						"a": 1,
						"b": 2,
					},
					StartLevel: map[string]int{
						"a": 0,
						"b": 0,
					},
				},
				name:  "Capacity",
				sense: 1,
			},
			want: map[string]float64{
				"a": 1,
				"b": 2,
			},
			wantErr: false,
		},
		{
			name: "map of attributes type - float",
			args: args[schema.Vehicle]{
				entity: schema.Vehicle{
					ID: "v1",
					Capacity: map[string]float64{
						"a": 1.0,
						"b": 2.0,
					},
					StartLevel: map[string]float64{
						"a": 0.0,
						"b": 0.0,
					},
				},
				name:  "Capacity",
				sense: 1,
			},
			want: map[string]float64{
				"a": 1.0,
				"b": 2.0,
			},
			wantErr: false,
		},
		{
			name: "wrong map of attributes type - string",
			args: args[schema.Vehicle]{
				entity: schema.Vehicle{
					ID: "v1",
					Capacity: map[string]string{
						"a": "1.0",
						"b": "2.0",
					},
					StartLevel: map[string]string{
						"a": "0.0",
						"b": "0.0",
					},
				},
				name:  "Capacity",
				sense: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "capacity type - float",
			args: args[schema.Vehicle]{
				entity: schema.Vehicle{
					ID:         "v1",
					Capacity:   sampleInt,
					StartLevel: 0,
				},
				name:  "Capacity",
				sense: 1,
			},
			want: map[string]float64{
				"default": sampleFloat,
			},
			wantErr: false,
		},
		{
			name: "wrong capacity type - string",
			args: args[schema.Vehicle]{
				entity: schema.Vehicle{
					ID:         "v1",
					Capacity:   "a wrong type",
					StartLevel: 0,
				},
				name:  "Capacity",
				sense: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "capacity type - float",
			args: args[schema.Vehicle]{
				entity: schema.Vehicle{
					ID:         "v1",
					Capacity:   6.2,
					StartLevel: 0,
				},
				name:  "Capacity",
				sense: 1,
			},
			want: map[string]float64{
				"default": 6.2,
			},
			wantErr: false,
		},
		{
			name: "wrong start level type - string",
			args: args[schema.Vehicle]{
				entity: schema.Vehicle{
					ID:         "v1",
					Capacity:   sampleInt,
					StartLevel: "a wrong type",
				},
				name:  "StartLevel",
				sense: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "start level type - float",
			args: args[schema.Vehicle]{
				entity: schema.Vehicle{
					ID:         "v1",
					Capacity:   sampleInt,
					StartLevel: 0.2,
				},
				name:  "StartLevel",
				sense: 1,
			},
			want: map[string]float64{
				"default": 0.2,
			},
			wantErr: false,
		},
		{
			name: "start level type - int",
			args: args[schema.Vehicle]{
				entity: schema.Vehicle{
					ID:         "v1",
					Capacity:   sampleInt,
					StartLevel: 1,
				},
				name:  "StartLevel",
				sense: 1,
			},
			want: map[string]float64{
				"default": 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range testsVehicles {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resources(tt.args.entity, tt.args.name, tt.args.sense)
			if (err != nil) != tt.wantErr {
				t.Errorf("vehicle requirements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("vehicle requirements() = %v, want %v", got, tt.want)
			}
		})
	}

	testsStops := []test[schema.Stop]{
		{
			name: "map of attributes",
			args: args[schema.Stop]{
				entity: schema.Stop{
					ID: "s1",
					Quantity: map[string]any{
						"a": 1,
						"b": 2,
					},
				},
				name:  "Quantity",
				sense: -1,
			},
			want: map[string]float64{
				"a": -1,
				"b": -2,
			},
			wantErr: false,
		},
		{
			name: "single attribute",
			args: args[schema.Stop]{
				entity: schema.Stop{
					ID:       "s1",
					Quantity: sampleInt,
				},
				name:  "Quantity",
				sense: -1,
			},
			want: map[string]float64{
				"default": -sampleFloat,
			},
			wantErr: false,
		},
		{
			name: "wrong type",
			args: args[schema.Stop]{
				entity: schema.Stop{
					ID:       "s1",
					Quantity: "a wrong type",
				},
				name:  "Quantity",
				sense: -1,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range testsStops {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resources(tt.args.entity, tt.args.name, tt.args.sense)
			if (err != nil) != tt.wantErr {
				t.Errorf("stop requirements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stop requirements() = %v, want %v", got, tt.want)
			}
		})
	}
}
