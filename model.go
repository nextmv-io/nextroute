// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nextmv-io/nextroute/common"
)

// Model defines routing problem.
type Model interface {
	ModelData
	SolutionObserved

	// AddConstraint adds a constraint to the model. The constraint is
	// checked at the specified violation.
	AddConstraint(constraint ModelConstraint) error

	// Constraints returns all constraints of the model.
	Constraints() ModelConstraints

	// ConstraintsCheckedAt returns all constraints of the model that
	// are checked at the specified time of having calculated the new
	// information for the changed solution.
	ConstraintsCheckedAt(violation CheckedAt) ModelConstraints

	// DistanceUnit returns the unit of distance used in the model. The
	// unit is used to convert distances to values and vice versa. This is
	// also used for reporting.
	DistanceUnit() common.DistanceUnit

	// DurationUnit returns the unit of duration used in the model. The
	// unit is used to convert durations to values and vice versa. This is
	// also used for reporting.
	DurationUnit() time.Duration

	// DurationToValue converts the specified duration to a value as it used
	// internally in the model.
	DurationToValue(duration time.Duration) float64

	// Epoch returns the epoch of the model. The epoch is used to convert
	// time.Time to float64 and vice versa. All float64 values are relative
	// to the epoch.
	Epoch() time.Time

	// Expressions returns all expressions of the model for which a solution
	// has to calculate values. The expressions are sorted by their index. The
	// constraints register their expressions with the model.
	Expressions() ModelExpressions

	// IsLocked returns true if the model is locked. The model is
	// locked after a solution has been created using the model.
	IsLocked() bool

	// NewPlanSequence creates a new plan sequence. A plan sequence is a plan
	// unit. A plan unit is a collection of stops which are always planned and
	// unplanned as a single unit. In this case they have to be planned as a
	// sequence on the same vehicle in the order of the stops provided as an
	// argument.
	NewPlanSequence(stops ModelStops) (ModelPlanStopsUnit, error)
	// NewPlanSingleStop creates a new plan unit. A plan single stop
	// is a plan unit of a single stop. A plan unit is a collection of
	// stops which are always planned and unplanned as a single unit.
	NewPlanSingleStop(stop ModelStop) (ModelPlanStopsUnit, error)
	// NewPlanMultipleStops creates a new plan of multiple [ModelStops]. A plan
	// of multiple stops is a [ModelPlanUnit] of more than one stop. A plan
	// unit is a collection of stops which are always planned and unplanned
	// as a single entity. When planned, they are always assigned to the same
	// vehicle. The function takes in a sequence represented by a
	// [DirectedAcyclicGraph] (DAG) which restricts the order in which the
	// stops can be planned on the vehicle. Using an empty DAG means that the
	// stops can be planned in any order, and they will always be assigned to
	// the same vehicle. Consider the stops [s1, s2, s3] and the sequence [s1
	// -> s2, s1 -> s3]. This means that we are restricting that the stop s1
	// must come before s2 and s3. However, we are not specifying the order of
	// s2 and s3. This means that we can plan s2 before s3 or s3 before s2.
	NewPlanMultipleStops(
		stops ModelStops,
		sequence DirectedAcyclicGraph,
	) (ModelPlanStopsUnit, error)

	// NewPlanAllPlanUnits creates a new plan units unit. A plan all plan
	// units unit is a collection of plan units which are always planned and
	// unplanned as a single unit. The sameVehicle argument specifies if the
	// plan units have to be planned on the same vehicle or not. If sameVehicle
	// is true, the plan units have to be planned on the same vehicle.
	// The plan units can only be part of one plan units unit.
	NewPlanAllPlanUnits(
		sameVehicle bool,
		planUnits ...ModelPlanUnit,
	) (ModelPlanUnitsUnit, error)

	// NewPlanOneOfPlanUnits creates a new plan units unit. A plan one of plan
	// units unit is a collection of plan units from which exactly one has to
	// be planned.
	NewPlanOneOfPlanUnits(planUnits ...ModelPlanUnit) (ModelPlanUnitsUnit, error)

	// NewStop creates a new stop. The stop is used to create plan units or can
	// be used to create a first or last stop of a vehicle.
	NewStop(location common.Location) (ModelStop, error)

	// NewVehicle creates a new vehicle. The vehicle is used to create
	// solutions. Every vehicle has a first and last stop - even if the vehicle
	// is empty.
	NewVehicle(
		vehicleType ModelVehicleType,
		start time.Time,
		first ModelStop,
		last ModelStop,
	) (ModelVehicle, error)
	// NewVehicleType creates a new vehicle type. The vehicle type is used to
	// create vehicles. The travelDuration defines the travel duration going
	// from one stop to another if the stops are planned on a vehicle of the
	// constructed type. The duration defines the duration of a stop that gets
	// planned on a vehicle of the constructed type.
	NewVehicleType(
		travelDuration TimeDependentDurationExpression,
		duration DurationExpression,
	) (ModelVehicleType, error)

	// NumberOfStops returns the number of stops in the model.
	NumberOfStops() int

	// Objective returns the objective of the model.
	Objective() ModelObjectiveSum

	// PlanUnits returns all plan units of the model. A plan unit
	// is a collection of stops which are always planned and unplanned as a
	// single unit.
	PlanUnits() ModelPlanUnits

	// PlanStopsUnits returns all plan units of the model that plan stops.
	PlanStopsUnits() ModelPlanStopsUnits

	// SequenceSampleSize returns the number of samples to take from all
	// possible permutations of the stops in a PlanUnit.
	SequenceSampleSize() int

	// SetSequenceSampleSize sets the number of samples to take from all
	// possible permutations of the stops in a PlanUnit.
	SetSequenceSampleSize(sequenceSampleSize int)

	// Random returns a random number generator.
	Random() *rand.Rand

	// SetRandom sets the random number generator of the model.
	SetRandom(random *rand.Rand)

	// Stops returns all stops of the model.
	Stops() ModelStops

	// Stop returns the stop with the specified index.
	Stop(index int) (ModelStop, error)

	// TimeFormat returns the time format used for reporting.
	TimeFormat() string

	// TimeToValue converts the specified time to a value as used
	// internally in the model.
	TimeToValue(time time.Time) float64

	// ValueToTime converts the specified value to a time.Time as used
	// by the user. It is assuming value represents time since
	// the [Model.Epoch()] in the unit [Model.DurationUnit()].
	ValueToTime(value float64) time.Time
	// Vehicles returns all vehicles of the model.
	Vehicles() ModelVehicles
	// VehicleTypes returns all vehicle types of the model.
	VehicleTypes() ModelVehicleTypes

	// Vehicle returns the vehicle with the specified index.
	Vehicle(index int) ModelVehicle

	// MaxTime returns the maximum end time (upper bound) for any stop. This
	// function uses the [Model.Epoch()] as a starting point and adds a large
	// number to provide a large enough upper bound.
	MaxTime() time.Time

	// MaxDuration returns the maximum duration (upper bound) for any stop.
	MaxDuration() time.Duration
}

// NewModel returns a new model.
func NewModel() (Model, error) {
	m := &modelImpl{
		modelDataImpl:                  newModelDataImpl(),
		constraintMap:                  make(map[CheckedAt]ModelConstraints),
		constraints:                    make(ModelConstraints, 0),
		constraintsWithStopUpdater:     make(ModelConstraints, 0),
		constraintsWithSolutionUpdater: make(ModelConstraints, 0),
		vehicles:                       make(ModelVehicles, 0),
		vehicleTypes:                   make(ModelVehicleTypes, 0),
		distanceUnit:                   common.Meters,
		durationUnit:                   time.Second,
		epoch:                          time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		expressions:                    make(map[int]ModelExpression),
		isLocked:                       false,
		objective:                      nil,
		objectivesWithStopUpdater:      make(ModelObjectives, 0),
		objectivesWithSolutionUpdater:  make(ModelObjectives, 0),
		random:                         rand.New(rand.NewSource(0)),
		timeFormat:                     time.UnixDate,
		stopVehicles:                   make(map[int]int),
		// TODO: 24 is a magic number, it expresses that for up to 4 stops in a
		// planunit without any relationship, we will still fully explore all
		// permutations. To find a better number we would have to run
		// experiments on pathologic cases.
		sequenceSampleSize: 24,
	}

	if m.epoch.Second() != 0 || m.epoch.Nanosecond() != 0 {
		return nil,
			fmt.Errorf("epoch %v is not on a minute boundary", m.epoch)
	}

	if m.durationUnit != time.Second {
		return nil,
			fmt.Errorf("duration unit %v is not supported", m.durationUnit)
	}

	m.objective = newModelObjectiveSum(m)

	for _, checkViolation := range CheckViolations {
		m.constraintMap[checkViolation] = make(ModelConstraints, 0)
	}

	return m, nil
}

type modelImpl struct {
	epoch time.Time
	modelDataImpl
	objective                  ModelObjectiveSum
	stopVehicles               map[int]int
	random                     *rand.Rand
	expressions                map[int]ModelExpression
	constraintMap              map[CheckedAt]ModelConstraints
	timeFormat                 string
	constraints                ModelConstraints
	vehicleTypes               ModelVehicleTypes
	constraintsWithStopUpdater ModelConstraints
	planUnits                  ModelPlanUnits
	solutionObservedImpl
	stops                          ModelStops
	vehicles                       ModelVehicles
	constraintsWithSolutionUpdater ModelConstraints
	objectivesWithStopUpdater      ModelObjectives
	objectivesWithSolutionUpdater  ModelObjectives
	distanceUnit                   common.DistanceUnit
	durationUnit                   time.Duration
	sequenceSampleSize             int
	mutex                          sync.RWMutex
	isLocked                       bool
	disallowedSuccessors           [][]bool
	hasDirectSuccessors            bool
}

func (m *modelImpl) Vehicles() ModelVehicles {
	return slices.Clone(m.vehicles)
}

func (m *modelImpl) SetRandom(random *rand.Rand) {
	m.random = random
}

func (m *modelImpl) SequenceSampleSize() int {
	return m.sequenceSampleSize
}

func (m *modelImpl) SetSequenceSampleSize(sequenceSampleSize int) {
	m.sequenceSampleSize = sequenceSampleSize
}

func (m *modelImpl) SetTimeFormat(timeFormat string) {
	m.timeFormat = timeFormat
}

func (m *modelImpl) Expressions() ModelExpressions {
	expressions := make(ModelExpressions, 0, len(m.expressions))
	for _, expression := range m.expressions {
		expressions = append(expressions, expression)
	}
	slices.SortStableFunc(expressions, func(i, j ModelExpression) int {
		return i.Index() - j.Index()
	})

	return expressions
}

func (m *modelImpl) NewVehicle(
	vehicleType ModelVehicleType,
	start time.Time,
	first ModelStop,
	last ModelStop,
) (ModelVehicle, error) {
	if m.isLocked {
		return nil,
			fmt.Errorf(lockErrorMessage, "vehicle")
	}

	vehicle, err := newModelVehicle(
		len(m.vehicles),
		vehicleType,
		start,
		first,
		last,
	)

	if err != nil {
		return nil, err
	}

	m.vehicles = append(m.vehicles, vehicle)

	vehicleType.(*vehicleTypeImpl).vehicles = append(
		vehicleType.(*vehicleTypeImpl).vehicles,
		vehicle,
	)

	return vehicle, nil
}

func (m *modelImpl) NewVehicleType(
	travelDuration TimeDependentDurationExpression,
	processDuration DurationExpression,
) (ModelVehicleType, error) {
	if m.isLocked {
		return nil,
			fmt.Errorf(lockErrorMessage, "vehicle type")
	}
	vehicle := &vehicleTypeImpl{
		index:          len(m.vehicleTypes),
		model:          m,
		travelDuration: travelDuration,
		duration:       processDuration,
	}
	m.vehicleTypes = append(m.vehicleTypes, vehicle)

	return vehicle, nil
}

func (m *modelImpl) addExpression(expression ModelExpression) error {
	if existingExpression, ok := m.expressions[expression.Index()]; ok {
		if existingExpression.Name() != expression.Name() {
			return fmt.Errorf(
				"expression index %d already exists with name %s,"+
					" expression indices must be unique,"+
					" did you forget to use NewModelExpressionIndex() on"+
					" a custom expression",
				expression.Index(),
				existingExpression.Name(),
			)
		}
	} else {
		m.expressions[expression.Index()] = expression
	}
	return nil
}

func (m *modelImpl) setConstraintEstimationOrder() {
	sort.SliceStable(m.constraints, func(i, j int) bool {
		ci := m.constraints[i]
		cj := m.constraints[j]
		if complexityOfI, ok := ci.(Complexity); ok {
			if complexityOfJ, ok := cj.(Complexity); ok {
				return complexityOfI.EstimationCost() <
					complexityOfJ.EstimationCost()
			}
			return true
		}

		if _, ok := cj.(Complexity); ok {
			return false
		}

		return i < j
	})
}

func (m *modelImpl) addToCheckAt(checkAt CheckedAt, constraint ModelConstraint) {
	if _, ok := m.constraintMap[checkAt]; !ok {
		m.constraintMap[checkAt] = make(ModelConstraints, 0, 1)
	}
	m.constraintMap[checkAt] = append(m.constraintMap[checkAt], constraint)
}

func (m *modelImpl) AddConstraint(constraint ModelConstraint) error {
	if m.IsLocked() {
		return fmt.Errorf(lockErrorMessage, "constraint")
	}
	for _, existingConstraint := range m.constraints {
		if &existingConstraint == &constraint {
			return fmt.Errorf(
				"constraint '%s' with the same address already added, "+
					"constraint addresses must be unique",
				reflect.TypeOf(constraint).String(),
			)
		}
	}
	if _, ok := constraint.(ConstraintDataUpdater); ok {
		return fmt.Errorf(
			"ConstraintDataUpdater has been deprecated, "+
				"please use ConstraintStopDataUpdater instead, "+
				"rename UpdateConstraintData to UpdateConstraintStopData for %s",
			reflect.TypeOf(constraint).String(),
		)
	}
	if _, ok := constraint.(SolutionStopViolationCheck); ok {
		m.addToCheckAt(AtEachStop, constraint)
	}
	if _, ok := constraint.(SolutionVehicleViolationCheck); ok {
		m.addToCheckAt(AtEachVehicle, constraint)
	}
	if _, ok := constraint.(SolutionViolationCheck); ok {
		m.addToCheckAt(AtEachSolution, constraint)
	}

	m.constraints = append(m.constraints, constraint)

	if registered, ok := constraint.(RegisteredModelExpressions); ok {
		for _, expression := range registered.ModelExpressions() {
			err := m.addExpression(expression)
			if err != nil {
				return err
			}
		}
	}

	if _, ok := constraint.(ConstraintStopDataUpdater); ok {
		m.constraintsWithStopUpdater = append(
			m.constraintsWithStopUpdater,
			constraint,
		)
	}
	if _, ok := constraint.(ConstraintSolutionDataUpdater); ok {
		m.constraintsWithSolutionUpdater = append(
			m.constraintsWithSolutionUpdater,
			constraint,
		)
	}

	return nil
}

func (m *modelImpl) Epoch() time.Time {
	return m.epoch
}

func (m *modelImpl) Constraints() ModelConstraints {
	return slices.Clone(m.constraints)
}

func (m *modelImpl) ConstraintsCheckedAt(violation CheckedAt) ModelConstraints {
	if constraints, ok := m.constraintMap[violation]; ok {
		return slices.Clone(constraints)
	}
	return make(ModelConstraints, 0)
}

func (m *modelImpl) Random() *rand.Rand {
	return m.random
}

func (m *modelImpl) Objective() ModelObjectiveSum {
	return m.objective
}

const lockErrorMessage = "model is locked, can not create a %s," +
	" a model is locked once a solution has been created using this model"

func (m *modelImpl) NewPlanOneOfPlanUnits(
	planUnits ...ModelPlanUnit,
) (ModelPlanUnitsUnit, error) {
	if m.IsLocked() {
		return nil,
			fmt.Errorf(lockErrorMessage, "one plan")
	}
	plan, err := newPlanUnitsUnit(
		len(m.planUnits),
		planUnits,
		true,
		false,
	)
	if err != nil {
		return nil, err
	}

	m.planUnits = append(m.planUnits, plan)

	return plan, nil
}

func (m *modelImpl) NewPlanAllPlanUnits(
	sameVehicle bool,
	planUnits ...ModelPlanUnit,
) (ModelPlanUnitsUnit, error) {
	if m.IsLocked() {
		return nil,
			fmt.Errorf(lockErrorMessage, "all plan")
	}
	plan, err := newPlanUnitsUnit(
		len(m.planUnits),
		planUnits,
		false,
		sameVehicle,
	)
	if err != nil {
		return nil, err
	}

	m.planUnits = append(m.planUnits, plan)

	return plan, nil
}

func (m *modelImpl) NewPlanSingleStop(stop ModelStop) (ModelPlanStopsUnit, error) {
	if m.IsLocked() {
		return nil,
			fmt.Errorf(lockErrorMessage, "plan single stop")
	}

	planSingleStop, err := newPlanSingleStop(len(m.planUnits), stop)
	if err != nil {
		return nil, err
	}

	m.planUnits = append(m.planUnits, planSingleStop)

	return planSingleStop, nil
}

func (m *modelImpl) NewPlanSequence(stops ModelStops) (ModelPlanStopsUnit, error) {
	if m.IsLocked() {
		return nil,
			fmt.Errorf(lockErrorMessage, "plan sequence")
	}

	directedAcyclicGraph := NewDirectedAcyclicGraph()

	for i := 1; i < len(stops); i++ {
		if err := directedAcyclicGraph.AddArc(stops[i-1], stops[i]); err != nil {
			return nil, err
		}
	}

	return m.NewPlanMultipleStops(stops, directedAcyclicGraph)
}

func (m *modelImpl) NewPlanMultipleStops(
	stops ModelStops,
	sequence DirectedAcyclicGraph,
) (ModelPlanStopsUnit, error) {
	if m.IsLocked() {
		return nil,
			fmt.Errorf(lockErrorMessage, "plan multiple stops")
	}

	planUnit, err := newPlanMultipleStops(len(m.planUnits), stops, sequence)
	if err != nil {
		return nil, err
	}

	m.planUnits = append(m.planUnits, planUnit)

	return planUnit, nil
}

func (m *modelImpl) PlanUnits() ModelPlanUnits {
	return slices.Clone(m.planUnits)
}

func (m *modelImpl) PlanStopsUnits() ModelPlanStopsUnits {
	planStopsUnits := make(ModelPlanStopsUnits, 0, len(m.planUnits))
	for _, planUnit := range m.planUnits {
		if planStopsUnit, ok := planUnit.(ModelPlanStopsUnit); ok {
			planStopsUnits = append(planStopsUnits, planStopsUnit)
		}
	}
	return planStopsUnits
}

func (m *modelImpl) TimeFormat() string {
	return m.timeFormat
}

func (m *modelImpl) DistanceUnit() common.DistanceUnit {
	return m.distanceUnit
}

func (m *modelImpl) DurationUnit() time.Duration {
	return m.durationUnit
}

func (m *modelImpl) DurationToValue(duration time.Duration) float64 {
	return duration.Seconds()
}

func (m *modelImpl) TimeToValue(time time.Time) float64 {
	return m.DurationToValue(time.Sub(m.epoch))
}

func (m *modelImpl) ValueToTime(value float64) time.Time {
	return m.epoch.Add(time.Duration(value) * m.durationUnit)
}

func (m *modelImpl) lock() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.isLocked {
		return nil
	}

	m.setConstraintEstimationOrder()
	for _, constraint := range m.constraints {
		if locker, ok := constraint.(Locker); ok {
			err := locker.Lock(m)
			if err != nil {
				return err
			}
		}
	}
	for _, term := range m.objective.Terms() {
		if locker, ok := term.Objective().(Locker); ok {
			err := locker.Lock(m)
			if err != nil {
				return err
			}
		}
	}
	// Check if all stops pre-assigned to vehicles are complete, that is
	// if any stops of a plan unit are pre-assigned to a vehicle, then all
	// stops of that plan unit must be pre-assigned to the same vehicle in
	// an order that is allowed by the plan unit.
	planUnits := common.UniqueDefined(
		common.Map(
			common.Keys(m.stopVehicles),
			func(idx int) ModelPlanStopsUnit {
				return m.stops[idx].PlanStopsUnit()
			},
		), func(planUnit ModelPlanStopsUnit) int {
			return planUnit.Index()
		},
	)
	for _, planUnit := range planUnits {
		vehicleIndex := -1

		modelStops := planUnit.Stops()
		modelStopsInVehicle := make([]ModelStop, 0, len(modelStops))
		modelStopsNotInVehicle := make([]ModelStop, 0, len(modelStops))
		for _, modelStop := range modelStops {
			modelStopImpl := modelStop.(*stopImpl)
			if index, inVehicle := m.stopVehicles[modelStop.Index()]; inVehicle {
				if vehicleIndex == -1 {
					vehicleIndex = index
				}
				if vehicleIndex != index {
					return fmt.Errorf(
						"stop `%v` is in initial_stops of vehicle `%v`"+
							" while other stops of the plan unit are in initial_stops of vehicle `%v`",
						modelStop.ID(),
						m.Vehicles()[index].ID(),
						m.Vehicles()[vehicleIndex].ID(),
					)
				}
				modelStopsInVehicle = append(modelStopsInVehicle, modelStop)
			} else {
				modelStopsNotInVehicle = append(modelStopsNotInVehicle, modelStop)
			}

			err := modelStopImpl.validate()
			if err != nil {
				return err
			}
		}

		// Check if all stops of the plan unit are on the same vehicle
		if len(modelStopsNotInVehicle) > 0 {
			return fmt.Errorf("a plan unit has stops "+
				"that are added as initial stops [%v] for vehicle `%v`, "+
				"either all or no stops of a plan unit must be added as initial stops, "+
				"missing stops [%v] in initial stops of vehicle",
				strings.Join(
					common.MapSlice(
						modelStopsInVehicle,
						func(modelStop ModelStop) []string {
							return []string{modelStop.ID()}
						}),
					", ",
				),
				m.Vehicles()[m.stopVehicles[modelStopsInVehicle[0].Index()]].ID(),
				strings.Join(
					common.MapSlice(
						modelStopsNotInVehicle,
						func(modelStop ModelStop) []string {
							return []string{modelStop.ID()}
						}),
					", ",
				),
			)
		}

		vehicle := m.Vehicles()[vehicleIndex]
		sequence := make(ModelStops, 0, len(modelStops))
		for _, stop := range vehicle.Stops() {
			if stop.PlanStopsUnit().Index() == planUnit.Index() {
				sequence = append(sequence, stop)
			}
		}
		allowed, err := planUnit.DirectedAcyclicGraph().IsAllowed(sequence)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf(
				"stops [%v] in this order, in start assignment of vehicle `%v` "+
					"violate the DAG (successor, predecessor) constraints of the plan unit",
				strings.Join(
					common.MapSlice(
						sequence,
						func(modelStop ModelStop) []string {
							return []string{modelStop.ID()}
						}),
					", ",
				),
				vehicle.ID(),
			)
		}
	}

	// Loop all planunit combinations and check whether they must be neighbors.
	for _, planUnit := range m.PlanStopsUnits() {
		if len(planUnit.DirectedAcyclicGraph().(*directedAcyclicGraphImpl).outboundDirectArcs) > 0 {
			m.hasDirectSuccessors = true
			break
		}
	}

	m.stopVehicles = make(map[int]int)
	m.isLocked = true

	return nil
}

func (m *modelImpl) IsLocked() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isLocked
}

func (m *modelImpl) VehicleTypes() ModelVehicleTypes {
	return slices.Clone(m.vehicleTypes)
}

func (m *modelImpl) Vehicle(index int) ModelVehicle {
	return m.vehicles[index]
}

func (m *modelImpl) NewStop(
	location common.Location,
) (ModelStop, error) {
	if m.isLocked {
		return nil,
			fmt.Errorf(lockErrorMessage, "stop")
	}

	stop := &stopImpl{
		index:        len(m.stops),
		measureIndex: len(m.stops),
		model:        m,
		location:     location,
	}
	m.stops = append(m.stops, stop)
	return stop, nil
}

func (m *modelImpl) Stop(index int) (ModelStop, error) {
	if index < 0 || index >= len(m.stops) {
		return nil,
			fmt.Errorf(
				"stop index %d is out of range [0, %d]",
				index,
				len(m.stops)-1,
			)
	}
	return m.stops[index], nil
}

func (m *modelImpl) Stops() ModelStops {
	return slices.Clone(m.stops)
}

func (m *modelImpl) NumberOfStops() int {
	return len(m.stops)
}

func (m *modelImpl) MaxTime() time.Time {
	return m.epoch.Add(time.Duration(24*365*200) * time.Hour)
}

func (m *modelImpl) MaxDuration() time.Duration {
	return m.MaxTime().Sub(m.epoch)
}

func (m *modelImpl) hasDisallowedSuccessors() bool {
	return m.disallowedSuccessors != nil
}
