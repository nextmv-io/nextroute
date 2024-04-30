// © 2019-present nextmv.io inc

package factory

import (
	"fmt"
	"slices"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/nextroute/schema"
)

func addPlanUnits(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}

	stop2Unit := make(map[int]nextroute.ModelPlanUnit, model.NumberOfStops())

	sequences := allSequences(data)

	for _, relations := range sequences {
		dag, err := buildDirectedAcyclicGraph(model, relations)
		if err != nil {
			return nil, err
		}
		modelPlanMultipleStops, err := model.NewPlanMultipleStops(dag.ModelStops(), dag)

		if err != nil {
			return nil, err
		}

		for _, stop := range dag.ModelStops() {
			stop2Unit[stop.Index()] = modelPlanMultipleStops
		}
	}

	for _, inputStop := range input.Stops {
		if _, ok := stop2Unit[data.stopIDToIndex[inputStop.ID]]; !ok {
			stop, err := model.Stop(data.stopIDToIndex[inputStop.ID])
			if err != nil {
				return nil, err
			}
			planSingleStop, err := model.NewPlanSingleStop(stop)
			if err != nil {
				return nil, err
			}
			stop2Unit[stop.Index()] = planSingleStop
		}
	}

	// travels through the vehicles and create the plan units for the alternate stops
	for _, vehicle := range input.Vehicles {
		if vehicle.AlternateStops == nil {
			continue
		}
		planUnits := make([]nextroute.ModelPlanUnit, len(*vehicle.AlternateStops))
		for idx, alternateID := range *vehicle.AlternateStops {
			stop, err := model.Stop(data.stopIDToIndex[alternateStopID(alternateID, vehicle)])
			if err != nil {
				return nil, err
			}
			planSingleStop, err := model.NewPlanSingleStop(stop)
			if err != nil {
				return nil, err
			}
			planUnits[idx] = planSingleStop
		}
		_, err := model.NewPlanOneOfPlanUnits(planUnits...)
		if err != nil {
			return nil, err
		}
	}

	for _, group := range data.groups {
		units := make([]nextroute.ModelPlanUnit, 0, len(group.stops))

		common.RangeMap(group.stops, func(id string, _ struct{}) bool {
			var stop nextroute.ModelStop
			stop, err = model.Stop(data.stopIDToIndex[id])
			if err != nil {
				return true
			}
			if _, ok := stop2Unit[stop.Index()]; !ok {
				err = nmerror.NewInputDataError(fmt.Errorf("stop %s is not part of a plan unit", id))
				return true
			}

			units = append(units, stop2Unit[stop.Index()])
			return false
		})
		if err != nil {
			return nil, err
		}
		uniquePlanUnits := common.UniqueDefined(units, func(t nextroute.ModelPlanUnit) int {
			return t.Index()
		})
		if len(uniquePlanUnits) > 1 {
			_, err := model.NewPlanAllPlanUnits(true, uniquePlanUnits...)
			if err != nil {
				return nil, err
			}
		}
	}
	return model, nil
}

type unitInformation struct {
	stops     map[string]struct{}
	sequences []sequence
}

// allSequences returns all the sequences of stops that are related to each
// other due to the precedence (pickups & deliveries) constraint.
func allSequences(data modelData) [][]sequence {
	dataSequences := slices.Clone(data.sequences)
	if len(dataSequences) == 0 {
		return [][]sequence{}
	}

	inUnit := make(map[string]int)
	units := make([]unitInformation, 0)
	for _, dataSequence := range dataSequences {
		// Check if any of the stops are already in a unit.
		unitIndex1 := -1
		unitIndex2 := -1
		if uIdx, ok := inUnit[dataSequence.predecessor]; ok {
			unitIndex1 = uIdx
		}
		if uIdx, ok := inUnit[dataSequence.successor]; ok {
			unitIndex2 = uIdx
		}

		// only predecessor is already in a unit.
		if unitIndex1 != -1 && unitIndex2 == -1 {
			units, inUnit = toExistingUnit(units, unitIndex1, dataSequence, inUnit)
			continue
			// only successor is already in a unit.
		} else if unitIndex1 == -1 && unitIndex2 != -1 {
			units, inUnit = toExistingUnit(units, unitIndex2, dataSequence, inUnit)
			continue
			// both predecessor and successor are already in a unit and not the same unit.
		} else if unitIndex1 != -1 && unitIndex2 != -1 && unitIndex1 != unitIndex2 {
			units, inUnit = mergeUnits(unitIndex1, unitIndex2, units, inUnit, dataSequence)
			continue
			// both predecessor and successor are already in a unit and the same unit.
		} else if unitIndex1 != -1 && unitIndex2 != -1 && unitIndex1 == unitIndex2 {
			units[unitIndex1].sequences = append(units[unitIndex1].sequences, dataSequence)
			continue
		}

		// neither predecessor nor successor are already in a unit.
		inUnit[dataSequence.predecessor] = len(units)
		inUnit[dataSequence.successor] = len(units)
		units = append(
			units,
			unitInformation{
				sequences: []sequence{dataSequence},
				stops: map[string]struct{}{
					dataSequence.predecessor: {},
					dataSequence.successor:   {},
				},
			},
		)
	}

	sequences := make([][]sequence, 0, len(data.sequences))
	for _, unit := range units {
		sequences = append(sequences, unit.sequences)
	}

	return sequences
}

func mergeUnits(
	unitIndex1 int,
	unitIndex2 int,
	units []unitInformation,
	inUnit map[string]int,
	datasequence sequence,
) ([]unitInformation, map[string]int) {
	if unitIndex1 > unitIndex2 {
		unitIndex1, unitIndex2 = unitIndex2, unitIndex1
	}
	oldUnit := units[unitIndex2]
	units = append(units[:unitIndex2], units[unitIndex2+1:]...)
	for i, ui := range units {
		for s := range ui.stops {
			inUnit[s] = i
		}
	}
	for stop := range oldUnit.stops {
		inUnit[stop] = unitIndex1
		units[unitIndex1].stops[stop] = struct{}{}
	}
	units[unitIndex1].sequences = append(units[unitIndex1].sequences, oldUnit.sequences...)
	units[unitIndex1].sequences = append(units[unitIndex1].sequences, datasequence)
	return units, inUnit
}

func toExistingUnit(
	units []unitInformation,
	unitIndex int,
	dataSequence sequence,
	inUnit map[string]int,
) ([]unitInformation, map[string]int) {
	units[unitIndex].sequences = append(units[unitIndex].sequences, dataSequence)
	units[unitIndex].stops[dataSequence.predecessor] = struct{}{}
	units[unitIndex].stops[dataSequence.successor] = struct{}{}
	inUnit[dataSequence.predecessor] = unitIndex
	inUnit[dataSequence.successor] = unitIndex
	return units, inUnit
}

// buildDirectedAcyclicGraph return the Directed Acyclic Graph
// that make up a ModelPlanUnit.
func buildDirectedAcyclicGraph(
	model nextroute.Model,
	sequences []sequence,
) (
	nextroute.DirectedAcyclicGraph,
	error,
) {
	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}
	dag := nextroute.NewDirectedAcyclicGraph()
	for _, sequence := range sequences {
		origin, err := model.Stop(data.stopIDToIndex[sequence.predecessor])
		if err != nil {
			return nil, err
		}
		destination, err := model.Stop(data.stopIDToIndex[sequence.successor])
		if err != nil {
			return nil, err
		}

		if sequence.direct {
			if err = dag.AddDirectArc(origin, destination); err != nil {
				return nil, err
			}
		} else if err = dag.AddArc(origin, destination); err != nil {
			return nil, err
		}
	}
	return dag, nil
}
