// © 2019-present nextmv.io inc

package factory

import (
	"errors"

	"github.com/nextmv-io/nextroute"
)

// modelData represents custom data at the Model level that can be used across
// different modifier functions.
type modelData struct {
	// Expression and constraint to represent the latest end time at a stop,
	// including the vehicle's ending location.
	latestEndExpression   nextroute.StopTimeExpression
	latestStartConstraint nextroute.LatestStart
	latestStartExpression nextroute.StopTimeExpression
	latestEndConstraint   nextroute.LatestEnd
	targetTime            nextroute.StopTimeExpression

	// Stop ID -> index in the input stops array.
	stopIDToIndex map[string]int
	// Precedence relationships between stops.
	sequences []sequence
	// Groups of stops that must be assigned to a vehicle as a group or not be
	// assigned.
	groups []group
}

// vehicleTypeData represents custom data for a VehicleType that can be used
// across different modifier functions.
type vehicleTypeData struct {
	DistanceExpression nextroute.DistanceExpression
}

// group represents a group of stops that must be assigned to a vehicle as a
// group or not be assigned. The order of the stops in the group is not
// important unless defined by precedence relationships.
type group struct {
	stops map[string]struct{}
}

// sequence represents two stops that must be part of the same planUnit. The
// predecessor must be visited before the successor; the direct field indicates
// if the successor must be the direct successor of the predecessor.
type sequence struct {
	predecessor string
	successor   string
	direct      bool
}

// latestStartExpression returns the StopTimeExpression that represents the
// latest start time at a stop. If the expression hasn't been added to the Model
// data, it is added. On the other hand, if the expression already exists, it
// is returned, as opposed to being created again.
func latestStartExpression(model nextroute.Model) (
	nextroute.StopTimeExpression,
	nextroute.Model,
	error,
) {
	data, err := getModelData(model)
	if err != nil {
		return nil, nil, err
	}

	if data.latestStartExpression != nil {
		return data.latestStartExpression, model, nil
	}

	expression := nextroute.NewStopTimeExpression("latest_start", model.MaxTime())
	data.latestStartExpression = expression
	model.SetData(data)

	return expression, model, nil
}

// addLatestEndConstraint checks if there is a LatestEnd constraint already
// added to the model. If there is, it does nothing. If there isn't, it adds it
// to the model.
func addLatestStartConstraint(
	model nextroute.Model,
	expression nextroute.StopTimeExpression,
) (nextroute.Model, error) {
	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}

	if data.latestStartConstraint != nil {
		return model, nil
	}

	constraint, err := nextroute.NewLatestStart(expression)
	if err != nil {
		return nil, err
	}

	err = model.AddConstraint(constraint)
	if err != nil {
		return nil, err
	}

	data.latestStartConstraint = constraint
	model.SetData(data)

	return model, nil
}

// latestEndExpression returns the StopTimeExpression that represents the
// latest end time at a stop. If the expression hasn't been added to the Model
// data, it is added. On the other hand, if the expression already exists, it
// is returned, as opposed to being created again.
func latestEndExpression(model nextroute.Model) (
	nextroute.StopTimeExpression,
	nextroute.Model,
	error,
) {
	data, err := getModelData(model)
	if err != nil {
		return nil, nil, err
	}

	if data.latestEndExpression != nil {
		return data.latestEndExpression, model, nil
	}

	expression := nextroute.NewStopTimeExpression("latest_end", model.MaxTime())
	data.latestEndExpression = expression
	model.SetData(data)

	return expression, model, nil
}

// addLatestEndConstraint checks if there is a LatestEnd constraint already
// added to the model. If there is, it does nothing. If there isn't, it adds it
// to the model.
func addLatestEndConstraint(
	model nextroute.Model,
	expression nextroute.StopTimeExpression,
) (nextroute.Model, error) {
	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}

	if data.latestEndConstraint != nil {
		return model, nil
	}

	constraint, err := nextroute.NewLatestEnd(expression)
	if err != nil {
		return nil, err
	}

	err = model.AddConstraint(constraint)
	if err != nil {
		return nil, err
	}

	data.latestEndConstraint = constraint
	model.SetData(data)

	return model, nil
}

// targetTimeExpression returns the StopTimeExpression that represents the
// target time at a stop. If the expression hasn't been added to the Model
// data, it is added. On the other hand, if the expression already exists, it
// is returned, as opposed to being created again.
func targetTimeExpression(model nextroute.Model) (
	nextroute.StopTimeExpression,
	nextroute.Model,
	error,
) {
	data, err := getModelData(model)
	if err != nil {
		return nil, nil, err
	}

	if data.targetTime != nil {
		return data.targetTime, model, nil
	}

	expression := nextroute.NewStopTimeExpression("target_time", model.MaxTime())
	data.targetTime = expression
	model.SetData(data)

	return expression, model, nil
}

// getModelData safely accesses the custom Model data and parses it to the type
// that is used across the model modifier functions.
func getModelData(model nextroute.Model) (modelData, error) {
	data := modelData{
		groups:        make([]group, 0),
		stopIDToIndex: make(map[string]int),
	}

	var err error
	if model.Data() != nil {
		d, ok := model.Data().(modelData)
		if !ok {
			err = errors.New("result of Model.Data() not of type modelData")
		}
		data = d
	}

	return data, err
}
