// © 2019-present nextmv.io inc

// Package schema provides the input and output schema for nextroute.
package schema_test

import (
	"testing"

	"github.com/nextmv-io/nextroute/schema"
)

type customMap struct {
	Foo string `json:"foo"`
}

type customSlice []string

func TestConvertCustomData(t *testing.T) {
	data := map[string]any{
		"custom_objective": map[string]any{
			"foo": "bar",
		},
		"custom_constraint": []any{
			"foo",
			"bar",
		},
	}

	customObjective, err := schema.ConvertCustomData[customMap](data["custom_objective"])
	if err != nil {
		t.Error(err)
	}
	if customObjective.Foo != "bar" {
		t.Errorf("expected %s, got %s", data["foo"], customObjective.Foo)
	}

	customConstraint, err := schema.ConvertCustomData[customSlice](data["custom_constraint"])
	if err != nil {
		t.Error(err)
	}
	if len(customConstraint) != 2 {
		t.Errorf("expected %d, got %d", 2, len(customConstraint))
	}
	for i, v := range customConstraint {
		if v != data["custom_constraint"].([]any)[i] {
			t.Errorf("expected %s, got %s", data["custom_constraint"].([]any)[i], v)
		}
	}
}
