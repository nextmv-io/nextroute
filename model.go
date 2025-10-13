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

// NewModel returns a new model.
func NewModel() (*Model, error) {
	m := &Model{
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
		expressions:                    []ModelExpression{},
		expressionsIndexMap:            map[int]int{},
		isLocked:                       false,
		objective:                      nil,
		objectivesWithStopUpdater:      make(ModelObjectives, 0),
		objectivesWithVehicleUpdater:   make(ModelObjectives, 0),
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

// Model defines routing problem.
type Model struct {
	epoch time.Time
	modelDataImpl
	objective    *ModelObjectiveSum
	stopVehicles map[int]int
	random       *rand.Rand
	expressions  []ModelExpression
	// expressionsIndexMap maps expression index to its position
	expressionsIndexMap        map[int]int
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
	objectivesWithVehicleUpdater   ModelObjectives
	objectivesWithSolutionUpdater  ModelObjectives
	distanceUnit                   common.DistanceUnit
	durationUnit                   time.Duration
	sequenceSampleSize             int
	mutex                          sync.RWMutex
	isLocked                       bool
	disallowedSuccessors           [][]bool
	hasDirectSuccessors            bool
}

// Vehicles returns all vehicles of the model.
func (m *Model) Vehicles() ModelVehicles {
	return slices.Clone(m.vehicles)
}

// SetRandom sets the random number generator of the model.
func (m *Model) SetRandom(random *rand.Rand) {
	m.random = random
}

// SequenceSampleSize returns the number of samples to take from all
// possible permutations of the stops in a PlanUnit.
func (m *Model) SequenceSampleSize() int {
	return m.sequenceSampleSize
}

// SetSequenceSampleSize sets the number of samples to take from all
// possible permutations of the stops in a PlanUnit.
func (m *Model) SetSequenceSampleSize(sequenceSampleSize int) {
	m.sequenceSampleSize = sequenceSampleSize
}

// SetTimeFormat sets the time format used for reporting.
func (m *Model) SetTimeFormat(timeFormat string) {
	m.timeFormat = timeFormat
}

// Expressions returns all expressions of the model for which a solution
// has to calculate values. The expressions are sorted by their index. The
// constraints register their expressions with the model.
func (m *Model) Expressions() ModelExpressions {
	expressions := make(ModelExpressions, 0, len(m.expressions))
	for _, expression := range m.expressions {
		expressions = append(expressions, expression)
	}
	slices.SortStableFunc(expressions, func(i, j ModelExpression) int {
		return i.Index() - j.Index()
	})

	return expressions
}

// NewVehicle creates a new vehicle. The vehicle is used to create
// solutions. Every vehicle has a first and last stop - even if the vehicle
// is empty.
func (m *Model) NewVehicle(
	vehicleType *ModelVehicleType,
	start time.Time,
	first *ModelStop,
	last *ModelStop,
) (*ModelVehicle, error) {
	if m.isLocked {
		return nil,
			fmt.Errorf(lockErrorMessage, "vehicle")
	}

	vehicle, err := NewModelVehicle(
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

	vehicleType.vehicles = append(
		vehicleType.vehicles,
		vehicle,
	)

	return vehicle, nil
}

// NewVehicleType creates a new vehicle type. The vehicle type is used to
// create vehicles. The travelDuration defines the travel duration going
// from one stop to another if the stops are planned on a vehicle of the
// constructed type. The duration defines the duration of a stop that gets
// planned on a vehicle of the constructed type.
func (m *Model) NewVehicleType(
	travelDuration TimeDependentDurationExpression,
	processDuration DurationExpression,
) (*ModelVehicleType, error) {
	if m.isLocked {
		return nil,
			fmt.Errorf(lockErrorMessage, "vehicle type")
	}
	vehicle := &ModelVehicleType{
		index:          len(m.vehicleTypes),
		model:          m,
		travelDuration: travelDuration,
		duration:       processDuration,
	}
	m.vehicleTypes = append(m.vehicleTypes, vehicle)

	return vehicle, nil
}

func (m *Model) addExpression(expression ModelExpression) error {
	if eIdx, ok := m.expressionsIndexMap[expression.Index()]; ok {
		existingExpression := m.expressions[eIdx]
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
		m.expressions = append(m.expressions, expression)
		m.expressionsIndexMap[expression.Index()] = len(m.expressions) - 1
	}
	return nil
}

func (m *Model) setConstraintEstimationOrder() {
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

func (m *Model) addToCheckAt(checkAt CheckedAt, constraint ModelConstraint) {
	if _, ok := m.constraintMap[checkAt]; !ok {
		m.constraintMap[checkAt] = make(ModelConstraints, 0, 1)
	}
	m.constraintMap[checkAt] = append(m.constraintMap[checkAt], constraint)
}

// AddConstraint adds a constraint to the model. The constraint is
// checked at the specified violation.
func (m *Model) AddConstraint(constraint ModelConstraint) error {
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

// Epoch returns the epoch of the model. The epoch is used to convert
// time.Time to float64 and vice versa. All float64 values are relative
// to the epoch.
func (m *Model) Epoch() time.Time {
	return m.epoch
}

// Constraints returns all constraints of the model.
func (m *Model) Constraints() ModelConstraints {
	return slices.Clone(m.constraints)
}

// ConstraintsCheckedAt returns all constraints of the model that
// are checked at the specified time of having calculated the new
// information for the changed solution.
func (m *Model) ConstraintsCheckedAt(violation CheckedAt) ModelConstraints {
	if constraints, ok := m.constraintMap[violation]; ok {
		return slices.Clone(constraints)
	}
	return make(ModelConstraints, 0)
}

// Random returns a random number generator.
func (m *Model) Random() *rand.Rand {
	return m.random
}

// Objective returns the objective of the model.
func (m *Model) Objective() *ModelObjectiveSum {
	return m.objective
}

const lockErrorMessage = "model is locked, can not create a %s," +
	" a model is locked once a solution has been created using this model"

// NewPlanOneOfPlanUnits creates a new plan units unit. A plan one of plan
// units unit is a collection of plan units from which exactly one has to
// be planned.
func (m *Model) NewPlanOneOfPlanUnits(
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

// NewPlanAllPlanUnits creates a new plan units unit. A plan all plan
// units unit is a collection of plan units which are always planned and
// unplanned as a single unit. The sameVehicle argument specifies if the
// plan units have to be planned on the same vehicle or not. If sameVehicle
// is true, the plan units have to be planned on the same vehicle.
// The plan units can only be part of one plan units unit.
func (m *Model) NewPlanAllPlanUnits(
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

// NewPlanSingleStop creates a new plan unit. A plan single stop
// is a plan unit of a single stop. A plan unit is a collection of
// stops which are always planned and unplanned as a single unit.
func (m *Model) NewPlanSingleStop(stop *ModelStop) (ModelPlanStopsUnit, error) {
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

// NewPlanSequence creates a new plan sequence. A plan sequence is a plan
// unit. A plan unit is a collection of stops which are always planned and
// unplanned as a single unit. In this case they have to be planned as a
// sequence on the same vehicle in the order of the stops provided as an
// argument.
func (m *Model) NewPlanSequence(stops ModelStops) (ModelPlanStopsUnit, error) {
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
func (m *Model) NewPlanMultipleStops(
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

// PlanUnits returns all plan units of the model. A plan unit
// is a collection of stops which are always planned and unplanned as a
// single unit.
func (m *Model) PlanUnits() ModelPlanUnits {
	return slices.Clone(m.planUnits)
}

// PlanStopsUnits returns all plan units of the model that plan stops.
func (m *Model) PlanStopsUnits() ModelPlanStopsUnits {
	planStopsUnits := make(ModelPlanStopsUnits, 0, len(m.planUnits))
	for _, planUnit := range m.planUnits {
		if planStopsUnit, ok := planUnit.(ModelPlanStopsUnit); ok {
			planStopsUnits = append(planStopsUnits, planStopsUnit)
		}
	}
	return planStopsUnits
}

// TimeFormat returns the time format used for reporting.
func (m *Model) TimeFormat() string {
	return m.timeFormat
}

// DistanceUnit returns the unit of distance used in the model. The
// unit is used to convert distances to values and vice versa. This is
// also used for reporting.
func (m *Model) DistanceUnit() common.DistanceUnit {
	return m.distanceUnit
}

// DurationUnit returns the unit of duration used in the model. The
// unit is used to convert durations to values and vice versa. This is
// also used for reporting.
func (m *Model) DurationUnit() time.Duration {
	return m.durationUnit
}

// DurationToValue converts the specified duration to a value as it used
// internally in the model.
func (m *Model) DurationToValue(duration time.Duration) float64 {
	return duration.Seconds()
}

// TimeToValue converts the specified time to a value as used
// internally in the model.
func (m *Model) TimeToValue(time time.Time) float64 {
	return m.DurationToValue(time.Sub(m.epoch))
}

// ValueToTime converts the specified value to a time.Time as used
// by the user. It is assuming value represents time since
// the [Model.Epoch()] in the unit [Model.DurationUnit()].
func (m *Model) ValueToTime(value float64) time.Time {
	return m.epoch.Add(time.Duration(value) * m.durationUnit)
}

func (m *Model) lock() error {
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
		modelStopsInVehicle := make([]*ModelStop, 0, len(modelStops))
		modelStopsNotInVehicle := make([]*ModelStop, 0, len(modelStops))
		for _, modelStop := range modelStops {
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

			err := modelStop.validate()
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
						func(modelStop *ModelStop) []string {
							return []string{modelStop.ID()}
						}),
					", ",
				),
				m.Vehicles()[m.stopVehicles[modelStopsInVehicle[0].Index()]].ID(),
				strings.Join(
					common.MapSlice(
						modelStopsNotInVehicle,
						func(modelStop *ModelStop) []string {
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
						func(modelStop *ModelStop) []string {
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

// IsLocked returns true if the model is locked. The model is
// locked after a solution has been created using the model.
func (m *Model) IsLocked() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isLocked
}

// VehicleTypes returns all vehicle types of the model.
func (m *Model) VehicleTypes() ModelVehicleTypes {
	return slices.Clone(m.vehicleTypes)
}

// Vehicle returns the vehicle with the specified index.
func (m *Model) Vehicle(index int) *ModelVehicle {
	return m.vehicles[index]
}

// NewStop creates a new stop. The stop is used to create plan units or can
// be used to create a first or last stop of a vehicle.
func (m *Model) NewStop(
	location common.Location,
) (*ModelStop, error) {
	if m.isLocked {
		return nil,
			fmt.Errorf(lockErrorMessage, "stop")
	}

	stop := &ModelStop{
		index:        len(m.stops),
		measureIndex: len(m.stops),
		model:        m,
		location:     location,
	}
	m.stops = append(m.stops, stop)
	return stop, nil
}

// Stop returns the stop with the specified index.
func (m *Model) Stop(index int) (*ModelStop, error) {
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

// Stops returns all stops of the model.
func (m *Model) Stops() ModelStops {
	return slices.Clone(m.stops)
}

// NumberOfStops returns the number of stops in the model.
func (m *Model) NumberOfStops() int {
	return len(m.stops)
}

// MaxTime returns the maximum end time (upper bound) for any stop. This
// function uses the [Model.Epoch()] as a starting point and adds a large
// number to provide a large enough upper bound.
func (m *Model) MaxTime() time.Time {
	return m.epoch.Add(time.Duration(24*365*200) * time.Hour)
}

// MaxDuration returns the maximum duration (upper bound) for any stop.
func (m *Model) MaxDuration() time.Duration {
	return m.MaxTime().Sub(m.epoch)
}

func (m *Model) hasDisallowedSuccessors() bool {
	return m.disallowedSuccessors != nil
}
