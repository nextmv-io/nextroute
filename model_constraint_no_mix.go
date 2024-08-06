// © 2019-present nextmv.io inc

package nextroute

import (
	"fmt"
)

// NoMixConstraint limits the order in which stops are assigned to a vehicle
// based upon the items the stops insert or remove from a vehicle.
type NoMixConstraint interface {
	Identifier
	ModelConstraint

	// Insert returns the mix items that are associated with a stop that
	// inserts items into a vehicle.
	Insert() map[ModelStop]MixItem
	// Remove returns the mix items that are associated with a stop that
	// removes items from a vehicle.
	Remove() map[ModelStop]MixItem
	// Value returns the value of the constraint for a stop. The value is
	// the amount of items on the vehicle at the moment of the stop. If
	// the stop is unplanned, the value has no semantic meaning.
	Value(solutionStop SolutionStop) MixItem
}

// MixItem is an item that is used to specify the type of mix.
// The type defines the type of each item. The count is the number units of
// the item are inserted or removed from a vehicle.
type MixItem struct {
	// Name is the name of the mix item.
	Name string `json:"name"`
	// Quantity is the number units of the mix items are inserted or removed from a
	// vehicle.
	Quantity int `json:"quantity"`
}

// NewNoMixConstraint returns a new NoMixConstraint.
func NewNoMixConstraint(
	deltas map[ModelStop]MixItem,
) (NoMixConstraint, error) {
	insert := make(map[ModelStop]MixItem, len(deltas)/2)
	remove := make(map[ModelStop]MixItem, len(deltas)/2)
	for stop, delta := range deltas {
		if delta.Quantity > 0 {
			insert[stop] = MixItem{
				Name:     delta.Name,
				Quantity: delta.Quantity,
			}
		} else {
			remove[stop] = MixItem{
				Name:     delta.Name,
				Quantity: -delta.Quantity,
			}
		}
	}

	return &noMixConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"no_mix",
			ModelExpressions{},
		),
		insert: insert,
		remove: remove,
	}, nil
}

func validate(
	insert map[ModelStop]MixItem,
	remove map[ModelStop]MixItem,
) error {
	if len(insert) == 0 && len(remove) == 0 {
		return nil
	}
	if len(insert)*len(remove) == 0 {
		return fmt.Errorf(
			"no-mix constraint, need both items that insert and items that remove or no items at all,"+
				" got %v items that insert and %v items that remove",
			len(insert),
			len(remove),
		)
	}
	deltaPerPlanUnit := make(map[ModelPlanStopsUnit]int)
	namePerPlanUnit := make(map[ModelPlanStopsUnit]string)
	stops := make(map[ModelStop]string, len(insert)+len(remove))
	for stop, i := range insert {
		if t, ok := stops[stop]; ok {
			return fmt.Errorf("no-mix constraint, stop %v has two items [%v, %v], a stop can only have one item",
				stop.ID(),
				t,
				i.Name,
			)
		}
		stops[stop] = i.Name
		deltaPerPlanUnit[stop.PlanStopsUnit()] += i.Quantity
		if t, ok := namePerPlanUnit[stop.PlanStopsUnit()]; ok {
			if t != i.Name {
				return fmt.Errorf(
					"no-mix constraint, items for stops in the same plan unit {%v}"+
						" must have the same name, have [%v, %v]",
					stop.PlanStopsUnit(),
					t,
					i.Name,
				)
			}
		}
		namePerPlanUnit[stop.PlanStopsUnit()] = i.Name
	}
	inRemove := make(map[ModelStop]string, len(remove))
	for stop, r := range remove {
		if t, ok := inRemove[stop]; ok {
			return fmt.Errorf("no-mix constraint, stop %v has two items [%v, %v], a stop can only have one item",
				stop.ID(),
				t,
				r.Name,
			)
		}
		inRemove[stop] = r.Name
		if t, ok := stops[stop]; ok {
			return fmt.Errorf("no-mix constraint, stop %v has two items [%v, %v], a stop can only have one item",
				stop.ID(),
				t,
				r.Name,
			)
		}
		stops[stop] = r.Name
		deltaPerPlanUnit[stop.PlanStopsUnit()] -= r.Quantity
		if t, ok := namePerPlanUnit[stop.PlanStopsUnit()]; ok {
			if t != r.Name {
				return fmt.Errorf(
					"no-mix constraint, items for stops in the same plan unit {%v}"+
						" must have the same name, have [%v, %v]",
					stop.PlanStopsUnit(),
					t,
					r.Name,
				)
			}
		}
		namePerPlanUnit[stop.PlanStopsUnit()] = r.Name
	}

	for modelPlanStopsUnit, d := range deltaPerPlanUnit {
		if d != 0 {
			divider := ""
			report := ""
			for idx, stop := range modelPlanStopsUnit.Stops() {
				if idx == 1 {
					divider = ","
				}
				if mixItem, ok := insert[stop]; ok {
					report += fmt.Sprintf("%v %v (+%v for %v)",
						divider,
						stop.ID(),
						mixItem.Quantity,
						mixItem.Name,
					)

					continue
				}
				if mixItem, ok := remove[stop]; ok {
					report += fmt.Sprintf("%v %v (-%v for %v)",
						divider,
						stop.ID(),
						mixItem.Quantity,
						mixItem.Name,
					)
					continue
				}
				report += fmt.Sprintf("%v %v",
					divider,
					stop.ID(),
				)
				report += fmt.Sprintf("%v: %v, ", stop.ID(), stops[stop])
			}
			return fmt.Errorf(
				"no-mix constraint, the sum of all quantities of an item of stops in a plan unit must be zero,"+
					" plan unit {%v} has a delta of %v",
				report,
				d,
			)
		}
	}
	return nil
}

type noMixConstraintImpl struct {
	modelConstraintImpl
	insert map[ModelStop]MixItem
	remove map[ModelStop]MixItem
}

type noMixSolutionStopData struct {
	content  MixItem
	tour     int
	removing bool
}

func (l *noMixSolutionStopData) Copy() Copier {
	return &noMixSolutionStopData{
		content: MixItem{
			Name:     l.content.Name,
			Quantity: l.content.Quantity,
		},
		tour:     l.tour,
		removing: l.removing,
	}
}

func (l *noMixConstraintImpl) Lock(_ Model) error {
	return validate(l.insert, l.remove)
}

func (l *noMixConstraintImpl) Value(solutionStop SolutionStop) MixItem {
	if !solutionStop.IsPlanned() {
		return MixItem{
			Name:     "",
			Quantity: 0,
		}
	}
	noMixSolutionStopData := solutionStop.ConstraintData(l).(*noMixSolutionStopData)

	return noMixSolutionStopData.content
}

func (l *noMixConstraintImpl) UpdateConstraintStopData(
	solutionStop SolutionStop,
) (Copier, error) {
	solutionStopImp := solutionStop

	if solutionStopImp.IsFirst() {
		return &noMixSolutionStopData{
			content: MixItem{
				Name:     "",
				Quantity: 0,
			},
			tour:     0,
			removing: false,
		}, nil
	}

	previousNoMixData := solutionStopImp.Previous().ConstraintData(l).(*noMixSolutionStopData)

	insertMixIngredient, hasInsertMixIngredient := l.insert[solutionStop.ModelStop()]
	if hasInsertMixIngredient {
		if previousNoMixData.content.Name != insertMixIngredient.Name && previousNoMixData.content.Quantity != 0 {
			return nil, fmt.Errorf(
				"cannot insert stop %v ingredient %v quantity %v because "+
					"previous stop content is %v of %v and removing is %v",
				solutionStopImp.ModelStop().Index(),
				insertMixIngredient.Name,
				insertMixIngredient.Quantity,
				previousNoMixData.content.Quantity,
				previousNoMixData.content.Name,
				previousNoMixData.removing,
			)
		}
		tour := previousNoMixData.tour
		if previousNoMixData.content.Quantity == 0 {
			tour++
		}
		return &noMixSolutionStopData{
			content: MixItem{
				Name:     insertMixIngredient.Name,
				Quantity: previousNoMixData.content.Quantity + insertMixIngredient.Quantity,
			},
			tour:     tour,
			removing: false,
		}, nil
	}

	removeMixIngredient, hasRemoveMixIngredient := l.remove[solutionStop.ModelStop()]
	if hasRemoveMixIngredient {
		if previousNoMixData.content.Name != removeMixIngredient.Name ||
			previousNoMixData.content.Quantity < removeMixIngredient.Quantity {
			return nil, fmt.Errorf(
				"cannot remove stop %v with content %v and quantity %v because previous stop has content %v and quantity %v",
				solutionStopImp.ModelStop().Index(),
				removeMixIngredient.Name,
				removeMixIngredient.Quantity,
				previousNoMixData.content.Name,
				previousNoMixData.content.Quantity,
			)
		}

		removing := true

		if previousNoMixData.content.Quantity == removeMixIngredient.Quantity {
			removing = false
		}

		return &noMixSolutionStopData{
			content: MixItem{
				Name:     previousNoMixData.content.Name,
				Quantity: previousNoMixData.content.Quantity - removeMixIngredient.Quantity,
			},
			tour:     previousNoMixData.tour,
			removing: removing,
		}, nil
	}

	ingredientName := previousNoMixData.content.Name
	if previousNoMixData.content.Quantity == 0 {
		ingredientName = ""
	}

	return &noMixSolutionStopData{
		content: MixItem{
			Name:     ingredientName,
			Quantity: previousNoMixData.content.Quantity,
		},
		tour:     previousNoMixData.tour,
		removing: previousNoMixData.removing,
	}, nil
}

func (l *noMixConstraintImpl) Insert() map[ModelStop]MixItem {
	insert := make(map[ModelStop]MixItem, len(l.insert))
	for stop, mixItem := range l.insert {
		insert[stop] = mixItem
	}
	return insert
}

func (l *noMixConstraintImpl) Remove() map[ModelStop]MixItem {
	remove := make(map[ModelStop]MixItem, len(l.remove))
	for stop, mixItem := range l.remove {
		remove[stop] = mixItem
	}
	return remove
}

func (l *noMixConstraintImpl) String() string {
	return l.name
}

func (l *noMixConstraintImpl) ID() string {
	return l.name
}

func (l *noMixConstraintImpl) SetID(id string) {
	l.name = id
}

func (l *noMixConstraintImpl) EstimationCost() Cost {
	return LinearStop
}

func (l *noMixConstraintImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	moveImpl := move.(*solutionMoveStopsImpl)
	_, hasRemoveMixItem := l.remove[moveImpl.stopPositions[0].Stop().ModelStop()]
	if hasRemoveMixItem {
		return true, constNoPositionsHint
	}

	previousStopImp := moveImpl.stopPositions[0].Previous()
	previousNoMixData := previousStopImp.ConstraintData(l).(*noMixSolutionStopData)
	contentName := previousNoMixData.content.Name
	contentQuantity := previousNoMixData.content.Quantity

	deltaQuantity := 0

	insertMixItem, hasInsertMixItem := l.insert[moveImpl.stopPositions[0].Stop().ModelStop()]
	if hasInsertMixItem {
		if contentName != insertMixItem.Name && previousNoMixData.content.Quantity != 0 {
			return true, constNoPositionsHint
		}
		deltaQuantity += insertMixItem.Quantity
	}

	if !hasRemoveMixItem && !hasInsertMixItem {
		// If the stop is not associated with any mix item, then the constraint
		// cannot be violated (as it is not mixing any new item between existing
		// ones). Note that the content name of all stops of a move is the same,
		// so it is enough to check the first stop.
		return false, constNoPositionsHint
	}

	tour := previousNoMixData.tour

	if previousNoMixData.content.Quantity == 0 {
		contentName = insertMixItem.Name
		tour++
	}

	for idx := 1; idx < len(moveImpl.stopPositions); idx++ {
		previousStopImp = moveImpl.stopPositions[idx].Previous()
		if previousStopImp.IsPlanned() {
			previousNoMixData = previousStopImp.ConstraintData(l).(*noMixSolutionStopData)
			if previousNoMixData.tour != tour || previousNoMixData.content.Name != contentName {
				return true, constNoPositionsHint
			}
		}
		insertMixItem, hasInsertMixItem = l.insert[moveImpl.stopPositions[idx].Stop().ModelStop()]
		if hasInsertMixItem {
			if contentName != insertMixItem.Name {
				return true, constNoPositionsHint
			}
			deltaQuantity += insertMixItem.Quantity
			continue
		}
		removeMixItem, hasRemoveMixItem := l.remove[moveImpl.stopPositions[idx].Stop().ModelStop()]
		if hasRemoveMixItem {
			if contentName != removeMixItem.Name || contentQuantity+deltaQuantity < removeMixItem.Quantity {
				return true, constNoPositionsHint
			}
			deltaQuantity -= removeMixItem.Quantity
		}
	}
	return false, constNoPositionsHint
}
