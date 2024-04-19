// © 2019-present nextmv.io inc

package nextroute

// DirectPrecedencesConstraint is a constraint that limits the vehicles a plan unit
// can be added to. The Attribute constraint configures compatibility
// attributes for stops and vehicles separately. This is done by specifying
// a list of attributes for stops and vehicles, respectively. Stops that
// have configured attributes are only compatible with vehicles that match
// at least one of them. Stops that do not have any specified attributes are
// compatible with any vehicle. Vehicles that do not have any specified
// attributes are only compatible with stops without attributes.
type DirectPrecedencesConstraint interface {
	ModelConstraint
	DisallowSuccessors(ModelStop, ModelStops) error
}

// NewDirectPrecedencesConstraint returns a new DirectPrecedencesConstraint.
func NewDirectPrecedencesConstraint() (DirectPrecedencesConstraint, error) {
	return &directPrecedencesConstraintImpl{
		modelConstraintImpl: newModelConstraintImpl(
			"direct_precedences",
			ModelExpressions{},
		),
	}, nil
}

type directPrecedencesConstraintImpl struct {
	modelConstraintImpl
}

func (l *directPrecedencesConstraintImpl) Lock(model Model) error {

}

func (l *directPrecedencesConstraintImpl) DisallowSuccessors(
	stop ModelStop,
	successors ModelStops,
) error {
	return nil
}

func (l *directPrecedencesConstraintImpl) String() string {
	return l.name
}

func (l *directPrecedencesConstraintImpl) EstimationCost() Cost {
	return Constant
}

func (l *directPrecedencesConstraintImpl) EstimateIsViolated(
	move SolutionMoveStops,
) (isViolated bool, stopPositionsHint StopPositionsHint) {
	return false, nil
}
