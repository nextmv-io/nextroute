// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"

	"github.com/nextmv-io/nextroute/common"
)

// ConstantExpression is an expression that always returns the same value.
type ConstantExpression interface {
	ModelExpression

	// SetValue sets the value of the expression.
	SetValue(value float64) error
}

// DefaultExpression is an expression that has a default value if no other
// values are defined.
type DefaultExpression interface {
	ModelExpression
	// DefaultValue returns the default value of the expression.
	DefaultValue() float64
}

// FromStopExpression is an expression that has a value for each from stop.
type FromStopExpression interface {
	DefaultExpression

	// SetValue sets the value of the expression for the given from stop.
	SetValue(
		stop ModelStop,
		value float64,
	) error
}

// StopExpression is an expression that has a value for each to stop.
type StopExpression interface {
	DefaultExpression

	// SetValue sets the value of the expression for the given to stop.
	SetValue(
		stop ModelStop,
		value float64,
	) error
}

// VehicleTypeExpression is the base expression for
// VehicleTypeExpressions.
type VehicleTypeExpression interface {
	DefaultExpression
	ValueForVehicleType(ModelVehicleType) float64
}

// VehicleTypeValueExpression is a ModelExpression that returns a value per
// vehicle type and allows to set the value per vehicle type.
type VehicleTypeValueExpression interface {
	VehicleTypeExpression
	// SetValue sets the value of the expression for the given vehicle type.
	SetValue(
		vehicle ModelVehicleType,
		value float64,
	) error
}

// FromToExpression is an expression that has a value for each combination
// of from and to stop.
type FromToExpression interface {
	DefaultExpression

	// SetValue sets the value of the expression for the given
	// from and to stops.
	SetValue(
		from ModelStop,
		to ModelStop,
		value float64,
	) error
}

// VehicleFromToExpression is an expression that has a value for each
// combination of vehicle type, from and to stop.
type VehicleFromToExpression interface {
	DefaultExpression

	// SetValue sets the value of the expression for the given vehicle type,
	// from and to stops.
	SetValue(
		vehicle ModelVehicleType,
		from ModelStop,
		to ModelStop,
		value float64,
	) error
}

// NewConstantExpression returns an expression that always returns the same
// value.
func NewConstantExpression(
	name string,
	value float64,
) ConstantExpression {
	return &constantExpression{
		index: NewModelExpressionIndex(),
		name:  name,
		value: value,
	}
}

// NewFromStopExpression returns a FromStopExpression.
func NewFromStopExpression(
	name string,
	defaultValue float64,
) FromStopExpression {
	return &fromExpression{
		name:              name,
		index:             NewModelExpressionIndex(),
		defaultValue:      defaultValue,
		values:            []float64{},
		hasPositiveValues: defaultValue > 0,
		hasNegativeValues: defaultValue < 0,
	}
}

// NewStopExpression returns a StopExpression.
func NewStopExpression(
	name string,
	defaultValue float64,
) StopExpression {
	return &toExpression{
		name:              name,
		index:             NewModelExpressionIndex(),
		defaultValue:      defaultValue,
		values:            []float64{},
		hasPositiveValues: defaultValue > 0,
		hasNegativeValues: defaultValue < 0,
	}
}

// NewVehicleTypeValueExpression returns a VehicleTypeValueExpression.
func NewVehicleTypeValueExpression(
	name string,
	defaultValue float64,
) VehicleTypeValueExpression {
	return &vehicleTypeExpressionImpl{
		name:              name,
		index:             NewModelExpressionIndex(),
		defaultValue:      defaultValue,
		values:            []float64{},
		hasPositiveValues: defaultValue > 0,
		hasNegativeValues: defaultValue < 0,
	}
}

// NewVehicleTypeDistanceExpression returns a VehicleTypeDistanceExpression.
func NewVehicleTypeDistanceExpression(
	name string,
	defaultValue common.Distance,
) VehicleTypeDistanceExpression {
	return &vehicleTypeDistanceExpressionImpl{
		name:              name,
		index:             NewModelExpressionIndex(),
		defaultValue:      defaultValue,
		values:            []common.Distance{},
		hasPositiveValues: defaultValue.Value(common.Meters) > 0,
		hasNegativeValues: defaultValue.Value(common.Meters) < 0,
	}
}

// NewFromToExpression returns a FromToExpression.
func NewFromToExpression(
	name string,
	defaultValue float64,
) FromToExpression {
	return &fromToExpression{
		values:            map[int]map[int]float64{},
		name:              name,
		index:             NewModelExpressionIndex(),
		defaultValue:      defaultValue,
		hasPositiveValues: defaultValue > 0,
		hasNegativeValues: defaultValue < 0,
	}
}

// NewVehicleTypeFromToExpression returns a VehicleTypeFromToExpression.
func NewVehicleTypeFromToExpression(
	name string,
	defaultValue float64,
) VehicleFromToExpression {
	return &vehicleTypeFromToExpression{
		name:              name,
		index:             NewModelExpressionIndex(),
		defaultValue:      defaultValue,
		values:            map[int]map[int]map[int]float64{},
		hasPositiveValues: defaultValue > 0,
		hasNegativeValues: defaultValue < 0,
	}
}

// NewDistanceExpression returns a DistanceExpression.
func NewDistanceExpression(
	name string,
	modelExpression ModelExpression,
	unit common.DistanceUnit,
) DistanceExpression {
	return &distanceExpression{
		name:            name,
		modelExpression: modelExpression,
		unit:            unit,
		index:           NewModelExpressionIndex(),
	}
}

type distanceExpression struct {
	modelExpression ModelExpression
	name            string
	unit            common.DistanceUnit
	index           int
}

func (d *distanceExpression) HasNegativeValues() bool {
	return d.modelExpression.HasNegativeValues()
}

func (d *distanceExpression) HasPositiveValues() bool {
	return d.modelExpression.HasPositiveValues()
}

func (d *distanceExpression) String() string {
	return fmt.Sprintf("Distance[%v] '%v' %v %v",
		d.index,
		d.name,
		d.unit,
		d.modelExpression,
	)
}

func (d *distanceExpression) Index() int {
	return d.index
}

func (d *distanceExpression) Name() string {
	return d.name
}

func (d *distanceExpression) SetName(n string) {
	d.name = n
}

func (d *distanceExpression) Value(v ModelVehicleType, from, to ModelStop) float64 {
	return d.modelExpression.Value(v, from, to)
}

func (d *distanceExpression) Distance(v ModelVehicleType, from, to ModelStop) common.Distance {
	return common.NewDistance(d.Value(v, from, to), d.unit)
}

type fromExpression struct {
	name              string
	values            []float64
	index             int
	defaultValue      float64
	hasPositiveValues bool
	hasNegativeValues bool
}

func (s *fromExpression) HasNegativeValues() bool {
	return s.hasNegativeValues
}

func (s *fromExpression) HasPositiveValues() bool {
	return s.hasPositiveValues
}

func (s *fromExpression) String() string {
	return fmt.Sprintf("From[%v] '%v' default %v, entries %v",
		s.index,
		s.name,
		s.defaultValue,
		len(s.values),
	)
}

func (s *fromExpression) Index() int {
	return s.index
}

func (s *fromExpression) Name() string {
	return s.name
}

func (s *fromExpression) SetName(n string) {
	s.name = n
}

func (s *fromExpression) DefaultValue() float64 {
	return s.defaultValue
}

func (s *fromExpression) SetValue(
	stop ModelStop,
	value float64,
) error {
	if stop.Model().IsLocked() {
		return fmt.Errorf(
			fmt.Sprintf(
				"cannot set value of stop '%v' on '%v' after model is locked",
				stop,
				s,
			),
		)
	}
	s.hasNegativeValues = s.hasNegativeValues || value < 0
	s.hasPositiveValues = s.hasPositiveValues || value > 0
	index := stop.Index()
	s.values = expandSlice(s.values,
		s.defaultValue,
		index,
		stop.Model().NumberOfStops(),
	)
	s.values[index] = value

	return nil
}

func (s *fromExpression) Value(
	_ ModelVehicleType,
	from ModelStop,
	_ ModelStop,
) float64 {
	index := from.Index()
	if index >= len(s.values) || index < 0 {
		return s.defaultValue
	}
	return s.values[index]
}

type toExpression struct {
	name              string
	values            []float64
	index             int
	defaultValue      float64
	hasPositiveValues bool
	hasNegativeValues bool
}

func (s *toExpression) HasNegativeValues() bool {
	return s.hasNegativeValues
}

func (s *toExpression) HasPositiveValues() bool {
	return s.hasPositiveValues
}

func (s *toExpression) String() string {
	return fmt.Sprintf("To[%v] '%v' default %v, entries %v",
		s.index,
		s.name,
		s.defaultValue,
		len(s.values),
	)
}

func (s *toExpression) Index() int {
	return s.index
}

func (s *toExpression) Name() string {
	return s.name
}

func (s *toExpression) SetName(n string) {
	s.name = n
}

func (s *toExpression) DefaultValue() float64 {
	return s.defaultValue
}

func (s *toExpression) SetValue(
	stop ModelStop,
	value float64,
) error {
	if stop.Model().IsLocked() {
		return fmt.Errorf(
			"cannot set value of stop '%v' on '%v' after model is locked",
			stop,
			s,
		)
	}
	s.hasNegativeValues = s.hasNegativeValues || value < 0
	s.hasPositiveValues = s.hasPositiveValues || value > 0
	index := stop.Index()
	s.values = expandSlice(s.values,
		s.defaultValue,
		index,
		stop.Model().NumberOfStops(),
	)
	s.values[index] = value

	return nil
}

// expandSlice will first check if the slice is already long enough
// (requiredLength), and if so, it will return the slice.
// If the slice is not long enough, it will create a new slice of length
// maxLength and copy the values from the original slice into the new slice. It
// will then fill the remaining values with the defaultValue.
func expandSlice[T any](slice []T, defaultValue T, requiredLength, maxLength int) []T {
	if requiredLength < len(slice) {
		return slice
	}
	values := make([]T, maxLength)
	copy(values, slice)
	for i := len(slice); i < len(values); i++ {
		values[i] = defaultValue
	}
	return values
}

func (s *toExpression) Value(
	_ ModelVehicleType,
	_ ModelStop,
	to ModelStop,
) float64 {
	index := to.Index()
	if index >= len(s.values) || index < 0 {
		return s.defaultValue
	}
	return s.values[index]
}

type vehicleTypeExpressionImpl struct {
	values            []float64
	name              string
	index             int
	defaultValue      float64
	hasNegativeValues bool
	hasPositiveValues bool
}

func (v *vehicleTypeExpressionImpl) HasNegativeValues() bool {
	return v.hasNegativeValues
}

func (v *vehicleTypeExpressionImpl) HasPositiveValues() bool {
	return v.hasPositiveValues
}

func (v *vehicleTypeExpressionImpl) String() string {
	return fmt.Sprintf("VehicleType[%v] '%v' default %v, entries %v",
		v.index,
		v.name,
		v.defaultValue,
		len(v.values),
	)
}

func (v *vehicleTypeExpressionImpl) Index() int {
	return v.index
}

func (v *vehicleTypeExpressionImpl) Name() string {
	return v.name
}

func (v *vehicleTypeExpressionImpl) SetName(n string) {
	v.name = n
}

func (v *vehicleTypeExpressionImpl) DefaultValue() float64 {
	return v.defaultValue
}

func (v *vehicleTypeExpressionImpl) SetValue(
	vehicle ModelVehicleType,
	value float64,
) error {
	if vehicle.Model().IsLocked() {
		return fmt.Errorf(
			"cannot set value of vehicle '%v' on '%v' after model is locked",
			vehicle,
			v,
		)
	}
	v.hasNegativeValues = v.hasNegativeValues || value < 0
	v.hasPositiveValues = v.hasPositiveValues || value > 0
	index := vehicle.Index()
	v.values = expandSlice(v.values, v.defaultValue, index, len(vehicle.Model().(*modelImpl).vehicleTypes))
	v.values[index] = value

	return nil
}

func (v *vehicleTypeExpressionImpl) ValueForVehicleType(
	vehicleType ModelVehicleType,
) float64 {
	return v.Value(vehicleType, nil, nil)
}

func (v *vehicleTypeExpressionImpl) Value(
	vehicleType ModelVehicleType,
	_, _ ModelStop,
) float64 {
	if vehicleType == nil {
		panic("vehicle type is nil for vehicle type expression")
	}
	index := vehicleType.Index()

	if index >= len(v.values) {
		return v.defaultValue
	}
	return v.values[index]
}

type vehicleTypeDistanceExpressionImpl struct {
	values            []common.Distance
	name              string
	index             int
	defaultValue      common.Distance
	hasNegativeValues bool
	hasPositiveValues bool
}

func (v *vehicleTypeDistanceExpressionImpl) HasNegativeValues() bool {
	return v.hasNegativeValues
}

func (v *vehicleTypeDistanceExpressionImpl) HasPositiveValues() bool {
	return v.hasPositiveValues
}

func (v *vehicleTypeDistanceExpressionImpl) String() string {
	return fmt.Sprintf("VehicleType[%v] '%v' default %v, entries %v",
		v.index,
		v.name,
		v.defaultValue,
		len(v.values),
	)
}

func (v *vehicleTypeDistanceExpressionImpl) Index() int {
	return v.index
}

func (v *vehicleTypeDistanceExpressionImpl) Name() string {
	return v.name
}

func (v *vehicleTypeDistanceExpressionImpl) SetName(n string) {
	v.name = n
}

func (v *vehicleTypeDistanceExpressionImpl) DefaultValue() float64 {
	return v.defaultValue.Value(common.Meters)
}

func (v *vehicleTypeDistanceExpressionImpl) SetDistance(
	vehicle ModelVehicleType,
	value common.Distance,
) error {
	if vehicle.Model().IsLocked() {
		return fmt.Errorf(
			"cannot set value of vehicle '%v' on '%v' after model is locked",
			vehicle,
			v,
		)
	}
	v.hasNegativeValues = v.hasNegativeValues || value.Value(common.Meters) < 0
	v.hasPositiveValues = v.hasPositiveValues || value.Value(common.Meters) > 0
	index := vehicle.Index()
	v.values = expandSlice(v.values, v.defaultValue, index, len(vehicle.Model().(*modelImpl).vehicleTypes))
	v.values[index] = value
	return nil
}

func (v *vehicleTypeDistanceExpressionImpl) ValueForVehicleType(
	vehicleType ModelVehicleType,
) float64 {
	return v.Value(vehicleType, nil, nil)
}

func (v *vehicleTypeDistanceExpressionImpl) DistanceForVehicleType(
	vehicleType ModelVehicleType,
) common.Distance {
	return v.Distance(vehicleType, nil, nil)
}

func (v *vehicleTypeDistanceExpressionImpl) Distance(
	vehicleType ModelVehicleType, _, _ ModelStop,
) common.Distance {
	value := v.Value(vehicleType, nil, nil)
	return common.NewDistance(value, common.Meters)
}

func (v *vehicleTypeDistanceExpressionImpl) Value(
	vehicleType ModelVehicleType,
	_, _ ModelStop,
) float64 {
	if vehicleType == nil {
		panic("vehicle type is nil for vehicle type expression")
	}
	index := vehicleType.Index()
	if index >= len(v.values) {
		return v.defaultValue.Value(common.Meters)
	}
	return v.values[index].Value(common.Meters)
}

type fromToExpression struct {
	values            map[int]map[int]float64
	name              string
	index             int
	defaultValue      float64
	hasPositiveValues bool
	hasNegativeValues bool
}

func (m *fromToExpression) HasNegativeValues() bool {
	return m.hasNegativeValues
}

func (m *fromToExpression) HasPositiveValues() bool {
	return m.hasPositiveValues
}

func (m *fromToExpression) String() string {
	entries := 0
	common.RangeMap(m.values, func(_ int, tos map[int]float64) bool {
		entries += len(tos)
		return false
	})

	return fmt.Sprintf("FromTo[%v] '%v' default %v, entries %v",
		m.index,
		m.name,
		m.defaultValue,
		entries,
	)
}

func (m *fromToExpression) Index() int {
	return m.index
}

func (m *fromToExpression) Name() string {
	return m.name
}

func (m *fromToExpression) SetName(n string) {
	m.name = n
}

func (m *fromToExpression) DefaultValue() float64 {
	return m.defaultValue
}

func (m *fromToExpression) SetValue(
	from ModelStop,
	to ModelStop,
	value float64,
) error {
	if from.Model().IsLocked() {
		return fmt.Errorf(
			"cannot set value on '%v' after model is locked",
			m,
		)
	}

	if _, ok := m.values[from.Index()]; !ok {
		m.values[from.Index()] = map[int]float64{}
	}

	m.hasNegativeValues = m.hasNegativeValues || value < 0
	m.hasPositiveValues = m.hasPositiveValues || value > 0
	m.values[from.Index()][to.Index()] = value

	return nil
}

func (m *fromToExpression) Value(
	_ ModelVehicleType,
	from ModelStop,
	to ModelStop,
) float64 {
	if _, ok := m.values[from.Index()]; ok {
		if value, ok := m.values[from.Index()][to.Index()]; ok {
			return value
		}
	}
	return m.defaultValue
}

type constantExpression struct {
	name  string
	index int
	value float64
}

func (c *constantExpression) HasNegativeValues() bool {
	return c.value < 0
}

func (c *constantExpression) HasPositiveValues() bool {
	return c.value > 0
}

func (c *constantExpression) String() string {
	return fmt.Sprintf("Constant[%v] '%v' value %v",
		c.index,
		c.name,
		c.value,
	)
}

func (c *constantExpression) Index() int {
	return c.index
}

func (c *constantExpression) Name() string {
	return c.name
}

func (c *constantExpression) SetName(n string) {
	c.name = n
}

func (c *constantExpression) SetValue(value float64) error {
	c.value = value
	return nil
}

func (c *constantExpression) Value(
	_ ModelVehicleType,
	_ ModelStop,
	_ ModelStop,
) float64 {
	return c.value
}

type vehicleTypeFromToExpression struct {
	values            map[int]map[int]map[int]float64
	name              string
	index             int
	defaultValue      float64
	hasNegativeValues bool
	hasPositiveValues bool
}

func (m *vehicleTypeFromToExpression) HasNegativeValues() bool {
	return m.hasNegativeValues
}

func (m *vehicleTypeFromToExpression) HasPositiveValues() bool {
	return m.hasPositiveValues
}

func (m *vehicleTypeFromToExpression) String() string {
	entries := 0
	common.RangeMap(m.values, func(_ int, froms map[int]map[int]float64) bool {
		common.RangeMap(froms, func(_ int, tos map[int]float64) bool {
			entries += len(tos)
			return false
		})
		return false
	})
	return fmt.Sprintf("VehicleTypeFromTo[%v] '%v' default %v, entries %v",
		m.index,
		m.name,
		m.defaultValue,
		entries,
	)
}

func (m *vehicleTypeFromToExpression) Index() int {
	return m.index
}

func (m *vehicleTypeFromToExpression) Name() string {
	return m.name
}

func (m *vehicleTypeFromToExpression) SetName(n string) {
	m.name = n
}

func (m *vehicleTypeFromToExpression) DefaultValue() float64 {
	return m.defaultValue
}

func (m *vehicleTypeFromToExpression) SetValue(
	vehicle ModelVehicleType,
	from ModelStop,
	to ModelStop,
	value float64,
) error {
	if from.Model().IsLocked() {
		return fmt.Errorf(
			"cannot set value on '%v' after model is locked",
			m,
		)
	}

	if _, ok := m.values[vehicle.Index()]; !ok {
		m.values[vehicle.Index()] = map[int]map[int]float64{}
	}
	if _, ok := m.values[vehicle.Index()][from.Index()]; !ok {
		m.values[vehicle.Index()][from.Index()] = map[int]float64{}
	}

	m.hasNegativeValues = m.hasNegativeValues || value < 0
	m.hasPositiveValues = m.hasPositiveValues || value > 0
	m.values[vehicle.Index()][from.Index()][to.Index()] = value

	return nil
}

func (m *vehicleTypeFromToExpression) Value(
	vehicle ModelVehicleType,
	from ModelStop,
	to ModelStop,
) float64 {
	if _, ok := m.values[vehicle.Index()]; ok {
		if _, ok := m.values[vehicle.Index()][from.Index()]; ok {
			if value, ok := m.values[vehicle.Index()][from.Index()][to.Index()]; ok {
				return value
			}
		}
	}
	return m.defaultValue
}
