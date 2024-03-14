// © 2019-present nextmv.io inc

package nextroute_test

import (
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestSolutionInitialStops_Feasible(t *testing.T) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}

	somStartTime := time.Date(
		2023,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	dummyLocation, err := common.NewLocation(0.0, 0.0)
	if err != nil {
		t.Fatal(err)
	}

	latestEndExpression := nextroute.NewStopTimeExpression(
		"latestEnd",
		model.MaxTime(),
	)
	travelDurationExpression := nextroute.NewFromToExpression(
		"travelDurationExpression",
		0,
	)

	warehouse, err := model.NewStop(dummyLocation)
	if err != nil {
		t.Fatal(err)
	}
	warehouse.SetID("warehouse")

	latestEndExpression.SetTime(warehouse, somStartTime.Add(1*time.Hour))

	s1, err := model.NewStop(dummyLocation)
	if err != nil {
		t.Fatal(err)
	}
	_, err = model.NewPlanSingleStop(s1)
	if err != nil {
		t.Fatal(err)
	}

	// warehouse -> s1 -> warehouse will violate the latest end constraint
	// but the initial solution should be valid it is
	// warehouse -> s1 -> s2 -> warehouse
	err = travelDurationExpression.SetValue(s1, warehouse, (2 * time.Hour).Seconds())
	if err != nil {
		t.Fatal(err)
	}

	s2, err := model.NewStop(dummyLocation)
	if err != nil {
		t.Fatal(err)
	}
	_, err = model.NewPlanSingleStop(s2)
	if err != nil {
		t.Fatal(err)
	}

	constraint, err := nextroute.NewLatestEnd(latestEndExpression)
	if err != nil {
		t.Fatal(err)
	}
	err = model.AddConstraint(constraint)
	if err != nil {
		t.Fatal(err)
	}

	vt, err := model.NewVehicleType(
		nextroute.NewTimeIndependentDurationExpression(
			nextroute.NewDurationExpression(
				"travelDuration",
				travelDurationExpression,
				common.Second,
			),
		),
		nextroute.NewConstantDurationExpression(
			"stopDuration",
			time.Second*0,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	v, err := model.NewVehicle(
		vt,
		somStartTime,
		warehouse,
		warehouse,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = v.AddStop(s1, false)
	if err != nil {
		t.Fatal(err)
	}
	err = v.AddStop(s2, false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = nextroute.NewSolution(model)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSolutionInitialStops_InFeasible(t *testing.T) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}

	somStartTime := time.Date(
		2023,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	dummyLocation, err := common.NewLocation(0.0, 0.0)
	if err != nil {
		t.Fatal(err)
	}

	latestEndExpression := nextroute.NewStopTimeExpression(
		"latestEnd",
		model.MaxTime(),
	)
	travelDurationExpression := nextroute.NewFromToExpression(
		"travelDurationExpression",
		0,
	)

	warehouse, err := model.NewStop(dummyLocation)
	if err != nil {
		t.Fatal(err)
	}
	warehouse.SetID("warehouse")

	latestEndExpression.SetTime(warehouse, somStartTime.Add(1*time.Hour))

	s1, err := model.NewStop(dummyLocation)
	if err != nil {
		t.Fatal(err)
	}
	_, err = model.NewPlanSingleStop(s1)
	if err != nil {
		t.Fatal(err)
	}

	s2, err := model.NewStop(dummyLocation)
	if err != nil {
		t.Fatal(err)
	}
	_, err = model.NewPlanSingleStop(s2)
	if err != nil {
		t.Fatal(err)
	}
	// warehouse -> s1 -> warehouse and  warehouse -> s1 -> s2 -> warehouse
	// will violate the latest end constraint
	err = travelDurationExpression.SetValue(s1, warehouse, (2 * time.Hour).Seconds())
	if err != nil {
		t.Fatal(err)
	}
	err = travelDurationExpression.SetValue(s2, warehouse, (2 * time.Hour).Seconds())
	if err != nil {
		t.Fatal(err)
	}

	constraint, err := nextroute.NewLatestEnd(latestEndExpression)
	if err != nil {
		t.Fatal(err)
	}
	err = model.AddConstraint(constraint)
	if err != nil {
		t.Fatal(err)
	}

	vt, err := model.NewVehicleType(
		nextroute.NewTimeIndependentDurationExpression(
			nextroute.NewDurationExpression(
				"travelDuration",
				travelDurationExpression,
				common.Second,
			),
		),
		nextroute.NewConstantDurationExpression(
			"stopDuration",
			time.Second*0,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	v, err := model.NewVehicle(
		vt,
		somStartTime,
		warehouse,
		warehouse,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = v.AddStop(s1, false)
	if err != nil {
		t.Fatal(err)
	}
	err = v.AddStop(s2, true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = nextroute.NewSolution(model)
	if err == nil {
		t.Fatal("expected 'violates temporal constraints' error, got nil")
	}
}
