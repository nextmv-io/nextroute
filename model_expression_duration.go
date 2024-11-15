// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"time"

	"github.com/nextmv-io/nextroute/common"
)

// StopDurationExpression is a ModelExpression that returns a duration per stop
// and allows to set the duration per stop.
type StopDurationExpression interface {
	DurationExpression
	// SetDuration sets the duration for the given stop.
	SetDuration(ModelStop, time.Duration)
}

// VehicleTypeDurationExpression is a ModelExpression that returns a duration
// per vehicle type and allows to set the duration per vehicle type.
type VehicleTypeDurationExpression interface {
	DurationExpression
	VehicleTypeExpression
	// SetDuration sets the duration for the given vehicle type.
	SetDuration(ModelVehicleType, time.Duration)
	DurationForVehicleType(ModelVehicleType) time.Duration
}

// DurationExpression is an expression that returns a duration.
type DurationExpression interface {
	ModelExpression
	// Duration returns the duration for the given vehicle type, start and
	// end stop.
	Duration(ModelVehicleType, ModelStop, ModelStop) time.Duration
}

// DistanceExpression is an expression that returns a distance.
type DistanceExpression interface {
	ModelExpression
	// Distance returns the distance for the given vehicle type, start and
	// end stop.
	Distance(ModelVehicleType, ModelStop, ModelStop) common.Distance
}

// VehicleTypeDistanceExpression is an expression that returns a distance per
// vehicle type and allows to set the duration per vehicle.
type VehicleTypeDistanceExpression interface {
	DistanceExpression
	VehicleTypeExpression

	SetDistance(ModelVehicleType, common.Distance) error

	DistanceForVehicleType(ModelVehicleType) common.Distance
}

// TravelDurationExpression is an expression that returns a duration based on
// a distance and a speed.
type TravelDurationExpression interface {
	DurationExpression
	// DistanceExpression returns the distance expression.
	DistanceExpression() DistanceExpression
	// Speed returns the speed.
	Speed() common.Speed
}

// NewDurationExpression returns a DurationExpression where values of the
// expression are interpreted as durations in units of the given duration unit.
func NewDurationExpression(
	name string,
	expression ModelExpression,
	unit common.DurationUnit,
) DurationExpression {
	return &scaledDurationExpressionImpl{
		index:      NewModelExpressionIndex(),
		expression: expression,
		multiplier: common.NewDuration(unit).Seconds(),
		name:       name,
	}
}

// NewScaledDurationExpression returns a new DurationExpression scaled by the
// given multiplier.
func NewScaledDurationExpression(
	expression DurationExpression,
	multiplier float64,
) DurationExpression {
	return &scaledDurationExpressionImpl{
		index:      NewModelExpressionIndex(),
		expression: expression,
		multiplier: multiplier,
		name:       fmt.Sprintf("%s * %v", expression.Name(), multiplier),
	}
}

// NewTravelDurationExpression returns a new TravelDurationExpression.
func NewTravelDurationExpression(
	distanceExpression DistanceExpression,
	speed common.Speed,
) TravelDurationExpression {
	return &travelDurationExpression{
		distanceExpression: distanceExpression,
		speed:              speed,
		index:              NewModelExpressionIndex(),
		name:               fmt.Sprintf("travelDuration(%s,%s)", distanceExpression.Name(), speed),
	}
}

// NewConstantDurationExpression returns a new ConstantDurationExpression.
func NewConstantDurationExpression(
	name string,
	duration time.Duration,
) DurationExpression {
	return &constantDurationExpressionImpl{
		index:    NewModelExpressionIndex(),
		name:     name,
		duration: duration,
	}
}

// NewStopDurationExpression returns a new StopDurationExpression.
func NewStopDurationExpression(
	name string,
	duration time.Duration,
) StopDurationExpression {
	return &stopDurationExpressionImpl{
		index:             NewModelExpressionIndex(),
		name:              name,
		defaultValue:      duration.Seconds(),
		values:            map[int]float64{},
		hasNegativeValues: duration < 0,
		hasPositiveValues: duration > 0,
	}
}

// NewVehicleTypeDurationExpression returns a new VehicleTypeDurationExpression.
func NewVehicleTypeDurationExpression(
	name string,
	duration time.Duration,
) VehicleTypeDurationExpression {
	return &vehicleTypeDurationExpressionImpl{
		index:             NewModelExpressionIndex(),
		name:              name,
		defaultValue:      duration.Seconds(),
		values:            map[int]float64{},
		hasNegativeValues: duration < 0,
		hasPositiveValues: duration > 0,
	}
}

type scaledDurationExpressionImpl struct {
	expression ModelExpression
	name       string
	index      int
	multiplier float64
}

func (s *scaledDurationExpressionImpl) HasNegativeValues() bool {
	if s.multiplier < 0 {
		return s.expression.HasPositiveValues()
	}
	return s.expression.HasNegativeValues()
}

func (s *scaledDurationExpressionImpl) HasPositiveValues() bool {
	if s.multiplier < 0 {
		return s.expression.HasNegativeValues()
	}
	return s.expression.HasPositiveValues()
}

func (s *scaledDurationExpressionImpl) String() string {
	return fmt.Sprintf("Scaled[%v] %v * %v",
		s.index,
		s.multiplier,
		s.expression,
	)
}

func (s *scaledDurationExpressionImpl) Name() string {
	return s.name
}

func (s *scaledDurationExpressionImpl) SetName(n string) {
	s.name = n
}

func (s *scaledDurationExpressionImpl) ScaledExpression() ModelExpression {
	return s.expression
}

func (s *scaledDurationExpressionImpl) Multiplier() float64 {
	return s.multiplier
}

func (s *scaledDurationExpressionImpl) Index() int {
	return s.index
}

func (s *scaledDurationExpressionImpl) Value(
	vehicleType ModelVehicleType,
	from ModelStop,
	to ModelStop,
) float64 {
	return s.expression.Value(
		vehicleType,
		from,
		to,
	) * s.multiplier
}

func (s *scaledDurationExpressionImpl) Duration(
	vehicleType ModelVehicleType,
	from ModelStop,
	to ModelStop,
) time.Duration {
	return time.Duration(
		s.Value(vehicleType, from, to),
	) * time.Second
}

type stopDurationExpressionImpl struct {
	values            map[int]float64
	name              string
	index             int
	defaultValue      float64
	hasNegativeValues bool
	hasPositiveValues bool
}

func (s *stopDurationExpressionImpl) HasNegativeValues() bool {
	return s.hasNegativeValues
}

func (s *stopDurationExpressionImpl) HasPositiveValues() bool {
	return s.hasPositiveValues
}

func (s *stopDurationExpressionImpl) String() string {
	return fmt.Sprintf("stop_duration[%v] '%s', default %v, entries %v",
		s.index,
		s.name,
		s.defaultValue,
		len(s.values),
	)
}

func (s *stopDurationExpressionImpl) Index() int {
	return s.index
}

func (s *stopDurationExpressionImpl) Name() string {
	return s.name
}

func (s *stopDurationExpressionImpl) SetName(n string) {
	s.name = n
}

func (s *stopDurationExpressionImpl) Value(
	_ ModelVehicleType,
	_ ModelStop,
	stop ModelStop,
) float64 {
	if value, ok := s.values[stop.Index()]; ok {
		return value
	}
	return s.defaultValue
}

func (s *stopDurationExpressionImpl) Duration(
	_ ModelVehicleType,
	_ ModelStop,
	stop ModelStop,
) time.Duration {
	return stop.Model().DurationUnit() *
		time.Duration(s.Value(nil, nil, stop))
}

func (s *stopDurationExpressionImpl) SetDuration(
	stop ModelStop,
	duration time.Duration,
) {
	if stop == nil {
		panic("stop is nil")
	}
	if stop.Model().IsLocked() {
		panic(
			fmt.Sprintf(
				"cannot set value on '%v' after model is locked",
				s,
			),
		)
	}
	s.hasNegativeValues = duration < 0
	s.hasPositiveValues = duration > 0

	s.values[stop.Index()] = duration.Seconds()
}

type vehicleTypeDurationExpressionImpl struct {
	values            map[int]float64
	name              string
	index             int
	defaultValue      float64
	hasNegativeValues bool
	hasPositiveValues bool
}

func (v *vehicleTypeDurationExpressionImpl) HasNegativeValues() bool {
	return v.hasNegativeValues
}

func (v *vehicleTypeDurationExpressionImpl) HasPositiveValues() bool {
	return v.hasPositiveValues
}

func (v *vehicleTypeDurationExpressionImpl) String() string {
	return fmt.Sprintf("vehicle_type_duration[%v] '%v', default %v, entries %v",
		v.index,
		v.name,
		v.defaultValue,
		len(v.values),
	)
}

func (v *vehicleTypeDurationExpressionImpl) Index() int {
	return v.index
}

func (v *vehicleTypeDurationExpressionImpl) Name() string {
	return v.name
}

func (v *vehicleTypeDurationExpressionImpl) SetName(n string) {
	v.name = n
}

func (v *vehicleTypeDurationExpressionImpl) DefaultValue() float64 {
	return v.defaultValue
}

func (v *vehicleTypeDurationExpressionImpl) Value(
	vehicleType ModelVehicleType,
	_ ModelStop,
	_ ModelStop,
) float64 {
	if value, ok := v.values[vehicleType.Index()]; ok {
		return value
	}
	return v.defaultValue
}

func (v *vehicleTypeDurationExpressionImpl) ValueForVehicleType(
	vehicleType ModelVehicleType,
) float64 {
	return v.Duration(vehicleType, nil, nil).Seconds()
}

func (v *vehicleTypeDurationExpressionImpl) DurationForVehicleType(
	vehicleType ModelVehicleType,
) time.Duration {
	return v.Duration(vehicleType, nil, nil)
}

func (v *vehicleTypeDurationExpressionImpl) Duration(
	vehicleType ModelVehicleType,
	_ ModelStop,
	_ ModelStop,
) time.Duration {
	return vehicleType.Model().DurationUnit() *
		time.Duration(v.Value(vehicleType, nil, nil))
}

func (v *vehicleTypeDurationExpressionImpl) SetDuration(
	vehicleType ModelVehicleType,
	duration time.Duration,
) {
	if vehicleType == nil {
		panic("vehicleType is nil")
	}

	if vehicleType.Model().IsLocked() {
		panic(
			fmt.Sprintf(
				"cannot set value on '%v' after model is locked",
				v,
			),
		)
	}

	v.hasNegativeValues = duration < 0
	v.hasPositiveValues = duration > 0
	v.values[vehicleType.Index()] = duration.Seconds()
}

type constantDurationExpressionImpl struct {
	name     string
	index    int
	duration time.Duration
}

func (c *constantDurationExpressionImpl) HasNegativeValues() bool {
	return c.duration < 0
}

func (c *constantDurationExpressionImpl) HasPositiveValues() bool {
	return c.duration > 0
}

func (c *constantDurationExpressionImpl) String() string {
	return fmt.Sprintf("constant_duration[%v] '%v' %v",
		c.index,
		c.name,
		c.duration,
	)
}

func (c *constantDurationExpressionImpl) Index() int {
	return c.index
}

func (c *constantDurationExpressionImpl) Name() string {
	return c.name
}

func (c *constantDurationExpressionImpl) SetName(n string) {
	c.name = n
}

func (c *constantDurationExpressionImpl) Value(
	_ ModelVehicleType,
	_, _ ModelStop,
) float64 {
	return c.duration.Seconds()
}

func (c *constantDurationExpressionImpl) Duration(
	_ ModelVehicleType,
	_ ModelStop,
	_ ModelStop,
) time.Duration {
	return c.duration
}

type travelDurationExpression struct {
	distanceExpression DistanceExpression
	speed              common.Speed
	name               string
	index              int
}

func (d *travelDurationExpression) HasNegativeValues() bool {
	return d.distanceExpression.HasNegativeValues()
}

func (d *travelDurationExpression) HasPositiveValues() bool {
	return d.distanceExpression.HasPositiveValues()
}

func (d *travelDurationExpression) String() string {
	return fmt.Sprintf("travel_duration[%v] speed %v, %v",
		d.index,
		d.speed,
		d.distanceExpression,
	)
}

func (d *travelDurationExpression) Index() int {
	return d.index
}

func (d *travelDurationExpression) Name() string {
	return d.name
}

func (d *travelDurationExpression) SetName(n string) {
	d.name = n
}

func (d *travelDurationExpression) DistanceExpression() DistanceExpression {
	return d.distanceExpression
}

func (d *travelDurationExpression) Speed() common.Speed {
	return d.speed
}

func (d *travelDurationExpression) Duration(
	vehicle ModelVehicleType,
	from ModelStop,
	to ModelStop,
) time.Duration {
	return time.Second *
		time.Duration(
			d.distanceExpression.Distance(
				vehicle,
				from,
				to,
			).Value(common.Meters)/
				d.speed.Value(common.MetersPerSecond),
		)
}

func (d *travelDurationExpression) Value(
	vehicle ModelVehicleType,
	from ModelStop,
	to ModelStop,
) float64 {
	return common.DurationValue(
		common.NewDistance(
			d.distanceExpression.Value(vehicle, from, to),
			vehicle.Model().DistanceUnit(),
		),
		d.speed,
		vehicle.Model().DurationUnit(),
	)
}
