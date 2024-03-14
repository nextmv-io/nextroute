// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nextmv-io/nextroute/common"
)

// ModelStatistics provides statistics for a model.
type ModelStatistics interface {
	// FirstLocations returns the number of unique locations that are first
	// locations of a vehicle.
	FirstLocations() int

	// LastLocations returns the number of unique locations that are last
	// locations of a vehicle.
	LastLocations() int
	// Locations returns the number of unique locations excluding first and last
	// locations of a vehicle.
	Locations() int

	// PlanUnits returns the number of plan units.
	PlanUnits() int

	// Report returns a report of the statistics.
	Report() string

	// Stops returns the number of stops.
	Stops() int

	// VehicleTypes returns the number of vehicle types.
	VehicleTypes() int
	// Vehicles returns the number of vehicles.
	Vehicles() int
}

// VehicleStatistics provides statistics for a vehicle.
type VehicleStatistics interface {
	// FirstToLastSeconds returns the travel time from the first location to the
	// last location of a vehicle.
	FirstToLastSeconds() float64
	// FromFirstSeconds returns the travel time in seconds from the first
	// location to all stops as statistics.
	FromFirstSeconds() common.Statistics

	// Report returns a report of the statistics.
	Report() string

	// ToLastSeconds returns the travel time in seconds from all stops to the
	// last location as statistics.
	ToLastSeconds() common.Statistics
}

// NewModelStatistics returns a new model statistics implementation.
func NewModelStatistics(model Model) ModelStatistics {
	return modelStatisticsImpl{
		model: model,
	}
}

// NewVehicleStatistics returns a new vehicle statistics implementation.
func NewVehicleStatistics(vehicle ModelVehicle) VehicleStatistics {
	return vehicleStatisticsImpl{
		vehicle: vehicle,
	}
}

type vehicleStatisticsImpl struct {
	vehicle ModelVehicle
}

func (v vehicleStatisticsImpl) FirstToLastSeconds() float64 {
	return v.vehicle.VehicleType().TravelDurationExpression().Duration(
		v.vehicle.VehicleType(),
		v.vehicle.First(),
		v.vehicle.Last(),
	).Seconds()
}

func (v vehicleStatisticsImpl) FromFirstSeconds() common.Statistics {
	stops := make(ModelStops, 0)
	for _, planUnit := range v.vehicle.Model().PlanStopsUnits() {
		stops = append(stops, planUnit.Stops()...)
	}
	return common.NewStatistics(
		stops,
		func(stop ModelStop) float64 {
			return v.vehicle.VehicleType().TravelDurationExpression().Duration(
				v.vehicle.VehicleType(),
				v.vehicle.First(),
				stop,
			).Seconds()
		},
	)
}

func (v vehicleStatisticsImpl) ToLastSeconds() common.Statistics {
	stops := make(ModelStops, 0)
	for _, planUnit := range v.vehicle.Model().PlanStopsUnits() {
		stops = append(stops, planUnit.Stops()...)
	}
	return common.NewStatistics(
		stops,
		func(stop ModelStop) float64 {
			return v.vehicle.VehicleType().TravelDurationExpression().Duration(
				v.vehicle.VehicleType(),
				stop,
				v.vehicle.Last(),
			).Seconds()
		},
	)
}

func (v vehicleStatisticsImpl) Report() string {
	var sb strings.Builder
	line := strings.Repeat("-", 80)
	fmt.Fprintf(&sb, "%s\n", line)
	fmt.Fprintf(&sb, "Vehicle %s\n", v.vehicle.ID())
	fmt.Fprintf(&sb, "%s\n", line)

	firstToLastSeconds := v.FirstToLastSeconds()
	fmt.Fprintf(&sb, "First to last seconds       : %f\n",
		firstToLastSeconds)
	fmt.Fprintf(&sb, "First to all stops seconds  :\n")
	fmt.Fprintf(&sb, "%v", v.FromFirstSeconds().Report())
	if firstToLastSeconds > 0 {
		fmt.Fprintf(&sb, "Last to all stops seconds   :\n")
		fmt.Fprintf(&sb, "%v", v.ToLastSeconds().Report())
	}
	return sb.String()
}

type modelStatisticsImpl struct {
	model Model
}

func (m modelStatisticsImpl) PlanUnits() int {
	return len(m.model.PlanUnits())
}

func (m modelStatisticsImpl) VehicleTypes() int {
	return len(m.model.VehicleTypes())
}

func (m modelStatisticsImpl) Vehicles() int {
	return len(m.model.Vehicles())
}

// locationToString is an auxiliary function to map an interface to a
// comparable type.
func locationToString(location common.Location) string {
	return fmt.Sprintf("%f,%f", location.Longitude(), location.Latitude())
}

func (m modelStatisticsImpl) LastLocations() int {
	return len(
		common.UniqueDefined(
			common.Map(
				m.model.Vehicles(),
				func(vehicle ModelVehicle) common.Location {
					return vehicle.Last().Location()
				},
			),
			locationToString,
		),
	)
}

func (m modelStatisticsImpl) Locations() int {
	stops := make(ModelStops, 0)
	for _, planUnit := range m.model.PlanStopsUnits() {
		stops = append(stops, planUnit.Stops()...)
	}
	return len(
		common.UniqueDefined(
			common.Map(
				stops,
				func(stop ModelStop) common.Location {
					return stop.Location()
				},
			),
			locationToString,
		),
	)
}

func (m modelStatisticsImpl) FirstLocations() int {
	return len(
		common.UniqueDefined(
			common.Map(
				m.model.Vehicles(),
				func(vehicle ModelVehicle) common.Location {
					return vehicle.First().Location()
				},
			),
			locationToString,
		),
	)
}

func (m modelStatisticsImpl) Stops() int {
	return len(m.model.Stops())
}

func (m modelStatisticsImpl) Report() string {
	var sb strings.Builder
	line := strings.Repeat("-", 80)
	fmt.Fprintf(&sb, "%s\nModel statistics\n%s\n",
		line,
		line)
	fmt.Fprintf(&sb, "Stops                       : %d\n",
		m.Stops())
	fmt.Fprintf(&sb, "Plan units                  : %d\n",
		m.PlanUnits())
	fmt.Fprintf(&sb, "Unique locations            : %d\n",
		m.Locations())
	fmt.Fprintf(&sb, "%s\n", line)
	fmt.Fprintf(&sb, "Vehicle types               : %d\n",
		m.VehicleTypes())
	fmt.Fprintf(&sb, "Vehicles                    : %d\n",
		m.Vehicles())
	fmt.Fprintf(&sb, "Unique first locations      : %d\n",
		m.FirstLocations())
	fmt.Fprintf(&sb, "Unique last locations       : %d\n",
		m.LastLocations())
	fmt.Fprintf(&sb, "%s\nExpressions\n%s\n",
		line,
		line)

	for _, expression := range m.model.Expressions() {
		fmt.Fprintf(&sb, "Name                        : %s\n",
			expression.Name())
		fmt.Fprintf(&sb, "Definition                  : %v\n",
			expression)
		fmt.Fprintf(&sb, "%s\n", line)
	}
	fmt.Fprintf(&sb, "Constraints\n%s\n",
		line)

	for _, constraint := range m.model.Constraints() {
		fmt.Fprintf(&sb, "Name                        : %s\n",
			reflect.TypeOf(constraint).String())
		fmt.Fprintf(&sb, "Definition                  : %v\n",
			constraint)
		fmt.Fprintf(&sb, "%s\n", line)
	}

	fmt.Fprintf(&sb, "Vehicles\n%s\n",
		line)

	uniqueVehicles := common.GroupBy(m.model.Vehicles(),
		func(t ModelVehicle) string {
			return fmt.Sprintf("%v-%v-%v-%v",
				t.VehicleType().Index(),
				t.First().Index(),
				t.Last().Index(),
				t.Start(),
			)
		},
	)
	common.RangeMap(uniqueVehicles, func(_ string, uniqueVehicle []ModelVehicle) bool {
		fmt.Fprintf(&sb, "Vehicle type index          : %v\n",
			uniqueVehicle[0].VehicleType().Index())
		fmt.Fprintf(&sb, "Travel duration expression  : %v\n",
			uniqueVehicle[0].VehicleType().TravelDurationExpression().Name())
		fmt.Fprintf(&sb, "Process duration expression : %v\n",
			uniqueVehicle[0].VehicleType().DurationExpression().Name())
		fmt.Fprintf(&sb, "Starts at time              : %v\n",
			uniqueVehicle[0].Start())
		fmt.Fprintf(&sb, "Start stop index            : %v {lat: %v, lon: %v}\n",
			uniqueVehicle[0].First().Index(),
			uniqueVehicle[0].First().Location().Latitude(),
			uniqueVehicle[0].First().Location().Longitude())
		fmt.Fprintf(&sb, "End stop index              : %v {lat: %v, lon: %v}\n",
			uniqueVehicle[0].Last().Index(),
			uniqueVehicle[0].Last().Location().Latitude(),
			uniqueVehicle[0].Last().Location().Longitude())
		statistics := NewVehicleStatistics(uniqueVehicle[0])
		fmt.Fprintf(&sb, "%v", statistics.Report())
		fmt.Fprintf(&sb, "%s\n", line)
		return false
	})
	return sb.String()
}
