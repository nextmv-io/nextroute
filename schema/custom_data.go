// © 2019-present nextmv.io inc

// Package schema provides the input and output schema for
package schema

import (
	"encoding/json"
	"errors"
)

// ConvertCustomData converts the custom data into the given type. If the
// conversion fails, an error is returned.
func ConvertCustomData[T any](data any) (T, error) {
	// Marshal the data again in order to unmarshal it into the correct type.
	var b []byte
	var err error
	if rawCustomData, ok := data.(map[string]any); ok {
		// Typically, the custom data is a map.
		b, err = json.Marshal(rawCustomData)
		if err != nil {
			return *new(T), err
		}
	} else if rawCustomData, ok := data.([]any); ok {
		// Try slice, if not map.
		b, err = json.Marshal(rawCustomData)
		if err != nil {
			return *new(T), err
		}
	} else {
		return *new(T), errors.New("CustomData is not a map or slice")
	}

	// Unmarshal the custom data into the given custom type.
	value := new(T)
	if err := json.Unmarshal(b, value); err != nil {
		return *new(T), err
	}
	return *value, nil
}
