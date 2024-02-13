package nextroute

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nextmv-io/sdk/common"
	"github.com/nextmv-io/sdk/nextroute"
	"golang.org/x/exp/slices"
)

// NewModel returns a new model.
func NewModel() (nextroute.Model, error) {
	m := &modelImpl{
		modelDataImpl:                  newModelDataImpl(),
		constraintMap:                  make(map[nextroute.CheckedAt]nextroute.ModelConstraints),
		constraints:                    make(nextroute.ModelConstraints, 0),
		constraintsWithStopUpdater:     make(nextroute.ModelConstraints, 0),
		constraintsWithSolutionUpdater: make(nextroute.ModelConstraints, 0),
		vehicles:                       make(nextroute.ModelVehicles, 0),
		vehicleTypes:                   make(nextroute.ModelVehicleTypes, 0),
		distanceUnit:                   common.Meters,
		durationUnit:                   time.Second,
		epoch:                          time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		expressions:                    make(map[int]nextroute.ModelExpression),
		isLocked:                       false,
		objective:                      nil,
		objectivesWithStopUpdater:      make(nextroute.ModelObjectives, 0),
		objectivesWithSolutionUpdater:  make(nextroute.ModelObjectives, 0),
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

	for _, checkViolation := range nextroute.CheckViolations {
		m.constraintMap[checkViolation] = make(nextroute.ModelConstraints, 0)
	}

	return m, nil
}

type modelImpl struct {
	epoch time.Time
	modelDataImpl
	objective                  nextroute.ModelObjectiveSum
	stopVehicles               map[int]int
	random                     *rand.Rand
	expressions                map[int]nextroute.ModelExpression
	constraintMap              map[nextroute.CheckedAt]nextroute.ModelConstraints
	timeFormat                 string
	constraints                nextroute.ModelConstraints
	vehicleTypes               nextroute.ModelVehicleTypes
	constraintsWithStopUpdater nextroute.ModelConstraints
	planUnits                  nextroute.ModelPlanUnits
	solutionObservedImpl
	stops                          nextroute.ModelStops
	vehicles                       nextroute.ModelVehicles
	constraintsWithSolutionUpdater nextroute.ModelConstraints
	objectivesWithStopUpdater      nextroute.ModelObjectives
	objectivesWithSolutionUpdater  nextroute.ModelObjectives
	distanceUnit                   common.DistanceUnit
	durationUnit                   time.Duration
	sequenceSampleSize             int
	mutex                          sync.RWMutex
	isLocked                       bool
}

func (m *modelImpl) Vehicles() nextroute.ModelVehicles {
	return common.DefensiveCopy(m.vehicles)
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

func (m *modelImpl) Expressions() nextroute.ModelExpressions {
	expressions := make(nextroute.ModelExpressions, 0, len(m.expressions))
	for _, expression := range m.expressions {
		expressions = append(expressions, expression)
	}
	slices.SortStableFunc(expressions, func(i, j nextroute.ModelExpression) int {
		return i.Index() - j.Index()
	})

	return expressions
}

func (m *modelImpl) NewVehicle(
	vehicleType nextroute.ModelVehicleType,
	start time.Time,
	first nextroute.ModelStop,
	last nextroute.ModelStop,
) (nextroute.ModelVehicle, error) {
	if m.isLocked {
		return nil,
			fmt.Errorf("model is isLocked, a model is" +
				" isLocked once a" +
				" solution has been created using this model",
			)
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
	travelDuration nextroute.TimeDependentDurationExpression,
	processDuration nextroute.DurationExpression,
) (nextroute.ModelVehicleType, error) {
	if m.isLocked {
		return nil,
			fmt.Errorf("model is isLocked, a model is" +
				" isLocked once a" +
				" solution has been created using this model",
			)
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

func (m *modelImpl) addExpression(expression nextroute.ModelExpression) {
	if existingExpression, ok := m.expressions[expression.Index()]; ok {
		if existingExpression.Name() != expression.Name() {
			panic(fmt.Sprintf(
				"expression index %d already exists with name %s,"+
					" expression indices must be unique,"+
					" did you forget to use NewModelExpressionIndex() on"+
					" a custom expression",
				expression.Index(),
				existingExpression.Name(),
			))
		}
	} else {
		m.expressions[expression.Index()] = expression
	}
}

func (m *modelImpl) setConstraintEstimationOrder() {
	sort.SliceStable(m.constraints, func(i, j int) bool {
		ci := m.constraints[i]
		cj := m.constraints[j]
		if complexityOfI, ok := ci.(nextroute.Complexity); ok {
			if complexityOfJ, ok := cj.(nextroute.Complexity); ok {
				return complexityOfI.EstimationCost() <
					complexityOfJ.EstimationCost()
			}
			return true
		}

		if _, ok := cj.(nextroute.Complexity); ok {
			return false
		}

		return i < j
	})
}

func (m *modelImpl) addToCheckAt(checkAt nextroute.CheckedAt, constraint nextroute.ModelConstraint) {
	if _, ok := m.constraintMap[checkAt]; !ok {
		m.constraintMap[checkAt] = make(nextroute.ModelConstraints, 0, 1)
	}
	m.constraintMap[checkAt] = append(m.constraintMap[checkAt], constraint)
}

func (m *modelImpl) AddConstraint(constraint nextroute.ModelConstraint) error {
	if m.IsLocked() {
		return fmt.Errorf("model is isLocked, a model is" +
			" isLocked once a" +
			" solution has been created using this model",
		)
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
	if _, ok := constraint.(nextroute.ConstraintDataUpdater); ok {
		return fmt.Errorf(
			"nextroute.ConstraintDataUpdater has been deprecated, "+
				"please use nextroute.ConstraintStopDataUpdater instead, "+
				"rename UpdateConstraintData to UpdateConstraintStopData for %s",
			reflect.TypeOf(constraint).String(),
		)
	}
	if _, ok := constraint.(nextroute.SolutionStopViolationCheck); ok {
		m.addToCheckAt(nextroute.AtEachStop, constraint)
	}
	if _, ok := constraint.(nextroute.SolutionVehicleViolationCheck); ok {
		m.addToCheckAt(nextroute.AtEachVehicle, constraint)
	}
	if _, ok := constraint.(nextroute.SolutionViolationCheck); ok {
		m.addToCheckAt(nextroute.AtEachSolution, constraint)
	}

	m.constraints = append(m.constraints, constraint)

	if registered, ok := constraint.(nextroute.RegisteredModelExpressions); ok {
		for _, expression := range registered.ModelExpressions() {
			m.addExpression(expression)
		}
	}

	if _, ok := constraint.(nextroute.ConstraintStopDataUpdater); ok {
		m.constraintsWithStopUpdater = append(
			m.constraintsWithStopUpdater,
			constraint,
		)
	}
	if _, ok := constraint.(nextroute.ConstraintSolutionDataUpdater); ok {
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

func (m *modelImpl) Constraints() nextroute.ModelConstraints {
	return common.DefensiveCopy(m.constraints)
}

func (m *modelImpl) ConstraintsCheckedAt(violation nextroute.CheckedAt) nextroute.ModelConstraints {
	if constraints, ok := m.constraintMap[violation]; ok {
		return common.DefensiveCopy(constraints)
	}
	return make(nextroute.ModelConstraints, 0)
}

func (m *modelImpl) Random() *rand.Rand {
	return m.random
}

func (m *modelImpl) Objective() nextroute.ModelObjectiveSum {
	return m.objective
}

func (m *modelImpl) NewPlanOneOfPlanUnits(
	planUnits ...nextroute.ModelPlanUnit,
) (nextroute.ModelPlanUnitsUnit, error) {
	if m.IsLocked() {
		return nil,
			fmt.Errorf("model is locked, can not create a plan," +
				" a model is locked once a" +
				" solution has been created using this model",
			)
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
	planUnits ...nextroute.ModelPlanUnit,
) (nextroute.ModelPlanUnitsUnit, error) {
	if m.IsLocked() {
		return nil,
			fmt.Errorf("model is locked, can not create a plan all plan," +
				" a model is locked once a" +
				" solution has been created using this model",
			)
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

func (m *modelImpl) NewPlanSingleStop(stop nextroute.ModelStop) (nextroute.ModelPlanStopsUnit, error) {
	if m.IsLocked() {
		return nil,
			fmt.Errorf("model is locked, can not create a plan one of plan unit," +
				" a model is locked once a" +
				" solution has been created using this model",
			)
	}

	planSingleStop, err := newPlanSingleStop(len(m.planUnits), stop)
	if err != nil {
		return nil, err
	}

	m.planUnits = append(m.planUnits, planSingleStop)

	return planSingleStop, nil
}

func (m *modelImpl) NewPlanSequence(stops nextroute.ModelStops) (nextroute.ModelPlanStopsUnit, error) {
	if m.IsLocked() {
		return nil,
			fmt.Errorf("model is locked, can not create a plan sequence," +
				" a model is locked once a" +
				" solution has been created using this model",
			)
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
	stops nextroute.ModelStops,
	sequence nextroute.DirectedAcyclicGraph,
) (nextroute.ModelPlanStopsUnit, error) {
	if m.IsLocked() {
		return nil,
			fmt.Errorf("model is locked, can not create multiple stops plan," +
				" a model is locked once a" +
				" solution has been created using this model",
			)
	}

	planUnit, err := newPlanMultipleStops(len(m.planUnits), stops, sequence)
	if err != nil {
		return nil, err
	}

	m.planUnits = append(m.planUnits, planUnit)

	return planUnit, nil
}

func (m *modelImpl) PlanUnits() nextroute.ModelPlanUnits {
	return common.DefensiveCopy(m.planUnits)
}

func (m *modelImpl) PlanStopsUnits() nextroute.ModelPlanStopsUnits {
	planStopsUnits := make(nextroute.ModelPlanStopsUnits, 0, len(m.planUnits))
	for _, planUnit := range m.planUnits {
		if planStopsUnit, ok := planUnit.(nextroute.ModelPlanStopsUnit); ok {
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
		if locker, ok := constraint.(nextroute.Locker); ok {
			err := locker.Lock(m)
			if err != nil {
				return err
			}
		}
	}
	for _, term := range m.objective.Terms() {
		if locker, ok := term.Objective().(nextroute.Locker); ok {
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
			func(idx int) nextroute.ModelPlanStopsUnit {
				return m.stops[idx].PlanStopsUnit()
			},
		), func(planUnit nextroute.ModelPlanStopsUnit) int {
			return planUnit.Index()
		},
	)
	for _, planUnit := range planUnits {
		vehicleIndex := -1

		modelStops := planUnit.Stops()
		modelStopsInVehicle := make([]nextroute.ModelStop, 0, len(modelStops))
		modelStopsNotInVehicle := make([]nextroute.ModelStop, 0, len(modelStops))
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
						func(modelStop nextroute.ModelStop) []string {
							return []string{modelStop.ID()}
						}),
					", ",
				),
				m.Vehicles()[m.stopVehicles[modelStopsInVehicle[0].Index()]].ID(),
				strings.Join(
					common.MapSlice(
						modelStopsNotInVehicle,
						func(modelStop nextroute.ModelStop) []string {
							return []string{modelStop.ID()}
						}),
					", ",
				),
			)
		}

		vehicle := m.Vehicles()[vehicleIndex]
		sequence := make(nextroute.ModelStops, 0, len(modelStops))
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
						func(modelStop nextroute.ModelStop) []string {
							return []string{modelStop.ID()}
						}),
					", ",
				),
				vehicle.ID(),
			)
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

func (m *modelImpl) VehicleTypes() nextroute.ModelVehicleTypes {
	return common.DefensiveCopy(m.vehicleTypes)
}

func (m *modelImpl) Vehicle(index int) nextroute.ModelVehicle {
	return m.vehicles[index]
}

func (m *modelImpl) NewStop(
	location common.Location,
) (nextroute.ModelStop, error) {
	if m.isLocked {
		return nil,
			fmt.Errorf("model is isLocked, a model is" +
				" isLocked once a solution has been created using" +
				" this model")
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

func (m *modelImpl) Stop(index int) (nextroute.ModelStop, error) {
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

func (m *modelImpl) Stops() nextroute.ModelStops {
	return common.DefensiveCopy(m.stops)
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
