// © 2019-present nextmv.io inc

package nextroute_test

import (
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
)

func TestTimeDependentDurationExpression_SetExpression(t *testing.T) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}

	defaultExpression := nextroute.NewConstantDurationExpression("default", 0*time.Hour)
	c1 := nextroute.NewConstantDurationExpression("c1", 1*time.Hour)
	c2 := nextroute.NewConstantDurationExpression("c2", 2*time.Hour)
	c3 := nextroute.NewConstantDurationExpression("c3", 3*time.Hour)
	c4 := nextroute.NewConstantDurationExpression("c4", 4*time.Hour)
	c5 := nextroute.NewConstantDurationExpression("c5", 5*time.Hour)

	timeDependentExpression, err := nextroute.NewTimeDependentDurationExpression(
		model,
		defaultExpression,
	)
	if err != nil {
		t.Fatal(err)
	}
	expression := timeDependentExpression.ExpressionAtTime(time.Now())
	if expression == nil {
		t.Error("expression should not be nil")
	}
	if expression != defaultExpression {
		t.Error("expression should be defaultExpression")
	}
	if timeDependentExpression.IsDependentOnTime() {
		t.Error("expression should not be dependent on time")
	}
	if len(timeDependentExpression.Expressions()) != 0 {
		t.Error("expression should have no time dependent expressions")
	}

	s1 := time.Date(2020, 1, 1, 9, 30, 0, 0, time.UTC)
	e1 := time.Date(2020, 1, 1, 10, 30, 0, 0, time.UTC)

	err = timeDependentExpression.SetExpression(
		s1,
		e1,
		c1,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !timeDependentExpression.IsDependentOnTime() {
		t.Error("expression should be dependent on time")
	}
	if len(timeDependentExpression.Expressions()) != 1 {
		t.Error("expression should have one time dependent expression")
	}
	if timeDependentExpression.ExpressionAtTime(s1.Add(-time.Second)) != defaultExpression {
		t.Error("expression before s1 should be defaultExpression")
	}
	if timeDependentExpression.ExpressionAtTime(s1) != c1 {
		t.Error("expression at s1 should be c1")
	}
	if timeDependentExpression.ExpressionAtTime(e1) != defaultExpression {
		t.Error("expression at e1 or later should be defaultExpression")
	}

	s2 := time.Date(2020, 1, 1, 11, 30, 0, 0, time.UTC)
	e2 := time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC)

	err = timeDependentExpression.SetExpression(
		s2,
		e2,
		c2,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !timeDependentExpression.IsDependentOnTime() {
		t.Error("expression should be dependent on time")
	}
	if len(timeDependentExpression.Expressions()) != 2 {
		t.Error("expression should have two time dependent expression")
	}
	if timeDependentExpression.ExpressionAtTime(s1.Add(-time.Second)) != defaultExpression {
		t.Error("expression before s1 should be defaultExpression")
	}
	if timeDependentExpression.ExpressionAtTime(s1) != c1 {
		t.Error("expression at s1 should be c1")
	}
	if timeDependentExpression.ExpressionAtTime(e1) != defaultExpression {
		t.Error("expression at e1 or later should be defaultExpression")
	}
	if timeDependentExpression.ExpressionAtTime(s2) != c2 {
		t.Error("expression at s2 should be c2")
	}
	if timeDependentExpression.ExpressionAtTime(e2) != defaultExpression {
		t.Error("expression at e2 should be defaultExpression")
	}

	s3 := time.Date(2020, 1, 1, 4, 30, 0, 0, time.UTC)
	e3 := time.Date(2020, 1, 1, 5, 30, 0, 0, time.UTC)

	err = timeDependentExpression.SetExpression(
		s3,
		e3,
		c3,
	)
	if err != nil {
		t.Fatal(err)
	}

	if timeDependentExpression.ExpressionAtTime(s1.Add(-time.Second)) != defaultExpression {
		t.Error("expression before s1 should be defaultExpression")
	}
	if timeDependentExpression.ExpressionAtTime(s1) != c1 {
		t.Error("expression at s1 should be c1")
	}
	if timeDependentExpression.ExpressionAtTime(e1) != defaultExpression {
		t.Error("expression at e1 or later should be defaultExpression")
	}
	if timeDependentExpression.ExpressionAtTime(s2) != c2 {
		t.Error("expression at s2 should be c2")
	}
	if timeDependentExpression.ExpressionAtTime(e2) != defaultExpression {
		t.Error("expression at e2 should be defaultExpression")
	}
	if timeDependentExpression.ExpressionAtTime(s3.Add(-time.Second)) != defaultExpression {
		t.Error("expression before s3 should be defaultExpression")
	}
	if timeDependentExpression.ExpressionAtTime(s3) != c3 {
		t.Error("expression at s3 should be c3")
	}
	if timeDependentExpression.ExpressionAtTime(e3) != defaultExpression {
		t.Error("expression at e3 should be defaultExpression")
	}
	s4 := time.Date(2020, 1, 1, 5, 0o0, 0, 0, time.UTC)
	e4 := time.Date(2020, 1, 1, 11, 0o0, 0, 0, time.UTC)
	err = timeDependentExpression.SetExpression(
		s4,
		e4,
		c4,
	)
	if err == nil {
		t.Fatal("should not be able to set expression that overlaps with existing expression")
	}

	s5 := time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC)
	e5 := time.Date(2020, 1, 1, 13, 0o0, 0, 0, time.UTC)
	err = timeDependentExpression.SetExpression(
		s5,
		e5,
		c5,
	)
	if err != nil {
		t.Fatal("should be able to set expression that starts at the end of an existing expression")
	}

	if len(timeDependentExpression.Expressions()) != 5 {
		t.Error("expression should have five time dependent expression")
	}

	s6 := time.Date(2020, 1, 1, 4, 0o0, 0, 0, time.UTC)
	e6 := time.Date(2020, 1, 1, 4, 30, 0, 0, time.UTC)
	err = timeDependentExpression.SetExpression(
		s6,
		e6,
		c5,
	)
	if err != nil {
		t.Fatal("should be able to set expression that ends at the start of an existing expression")
	}

	if len(timeDependentExpression.Expressions()) != 5 {
		t.Error("expression should have five time dependent expression")
	}
}

func TestTimeDependentDurationExpression_ValueAtTime1(t *testing.T) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}

	defaultExpression := nextroute.NewConstantDurationExpression("default", 1*time.Hour)
	c1 := nextroute.NewConstantDurationExpression("c1", 30*time.Minute)

	timeDependentExpression, err := nextroute.NewTimeDependentDurationExpression(
		model,
		defaultExpression,
	)
	if err != nil {
		t.Fatal(err)
	}

	s1 := time.Date(2020, 1, 1, 8, 30, 0, 0, time.UTC)
	e1 := time.Date(2020, 1, 1, 8, 45, 0, 0, time.UTC)

	value := timeDependentExpression.ValueAtTime(s1, nil, nil, nil)
	if value != 3600 {
		t.Fatal("value should be 3600, 1 hour from defaultExpression")
	}

	err = timeDependentExpression.SetExpression(
		s1,
		e1,
		c1,
	)
	if err != nil {
		t.Fatal(err)
	}

	value = timeDependentExpression.ValueAtTime(s1, nil, nil, nil)

	if value != 2700 {
		t.Error("value should be 2700, 1800 from defaultExpression and 900 from c1")
	}

	value = timeDependentExpression.ValueAtTime(e1, nil, nil, nil)
	if value != 3600 {
		t.Fatalf(
			"value should be 3600, 1 hour from defaultExpression, it is %v", value)
	}
}

func TestTimeDependentDurationExpression_ValueAtTime2(t *testing.T) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}

	defaultExpression := nextroute.NewConstantDurationExpression("default", 6*time.Hour)
	c1 := nextroute.NewConstantDurationExpression("c1", 12*time.Hour)
	c2 := nextroute.NewConstantDurationExpression("c2", 3*time.Hour)

	timeDependentExpression, err := nextroute.NewTimeDependentDurationExpression(
		model,
		defaultExpression,
	)
	if err != nil {
		t.Fatal(err)
	}

	s1 := time.Date(2020, 1, 1, 8, 0o0, 0, 0, time.UTC)
	e1 := time.Date(2020, 1, 1, 9, 0o0, 0, 0, time.UTC)

	err = timeDependentExpression.SetExpression(
		s1,
		e1,
		c1,
	)
	if err != nil {
		t.Fatal(err)
	}

	s2 := time.Date(2020, 1, 1, 9, 0o0, 0, 0, time.UTC)
	e2 := time.Date(2020, 1, 1, 11, 0o0, 0, 0, time.UTC)

	err = timeDependentExpression.SetExpression(
		s2,
		e2,
		c2,
	)
	if err != nil {
		t.Fatal(err)
	}

	s3 := time.Date(2020, 1, 1, 12, 0o0, 0, 0, time.UTC)
	e3 := time.Date(2020, 1, 1, 12, 0o0, 0, 0, time.UTC)

	err = timeDependentExpression.SetExpression(
		s3,
		e3,
		c2,
	)
	if err != nil {
		t.Fatal(err)
	}

	value := timeDependentExpression.ValueAtTime(s1.Add(-1*time.Hour), nil, nil, nil)

	if value != 16200 {
		t.Error("value should be 16200")
	}

	value = timeDependentExpression.ValueAtTime(s2, nil, nil, nil)

	if int(value) != 14400 {
		t.Error("value should be 14400 but is: ", value)
	}

	value = timeDependentExpression.ValueAtTime(e2, nil, nil, nil)

	if value != 21600 {
		t.Error("value should be 21600 but is: ", value)
	}

	handlePanic := func() {
		r := recover()

		if r != nil {
			t.Fatalf("should not panic, but did: %v", r)
		}
	}

	defer handlePanic()

	value = timeDependentExpression.ValueAtTime(s2.Add(-30*time.Minute), nil, nil, nil)

	if int(value) != 15300 {
		t.Error("value should be 15300 but is: ", value)
	}
}
