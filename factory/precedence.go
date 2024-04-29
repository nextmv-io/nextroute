// © 2019-present nextmv.io inc

package factory

import (
	"fmt"
	"reflect"

	"github.com/nextmv-io/nextroute"
	nmerror "github.com/nextmv-io/nextroute/common/errors"
	"github.com/nextmv-io/nextroute/schema"
)

// addPrecedenceInformation adds information to the Model data, when a stop's
// "precedes" or "succeeds" field is not nil.
func addPrecedenceInformation(
	input schema.Input,
	model nextroute.Model,
	_ Options,
) (nextroute.Model, error) {
	present := false
	var sequences []sequence
	stopIDToIndex := map[string]int{}
	for s, stop := range input.Stops {
		stopIDToIndex[stop.ID] = s
		if stop.Precedes == nil && stop.Succeeds == nil {
			continue
		}

		stopSequences, err := getSequences(stop)
		if err != nil {
			return nil, err
		}

		sequences = append(sequences, stopSequences...)
		present = true
	}

	if !present {
		return model, nil
	}

	data, err := getModelData(model)
	if err != nil {
		return nil, err
	}

	data.sequences = sequences

	model.SetData(data)

	return model, nil
}

// precedence processes the "Precedes" or "Succeeds" field of a stop. It return
// the precedence (succeeds or precedes) as a slice of strings, even for a
// single string.
func precedence(stop schema.Stop, name string) ([]precedenceData, error) {
	field := reflect.ValueOf(stop).FieldByName(name).Interface()
	var precedence []precedenceData
	if field == nil {
		return precedence, nil
	}

	// field can be a string, a slice of strings or a slice of structs of type
	// precedenceData.
	switch field := field.(type) {
	case string:
		precedence = append(precedence, precedenceData{id: field})
		return precedence, nil
	case []any:
		for i, element := range field {
			switch element := element.(type) {
			case string:
				precedence = append(precedence, precedenceData{id: element})
			case map[string]any:
				if id, ok := element["id"].(string); ok {
					direct, _ := element["direct"].(bool)
					precedence = append(precedence, precedenceData{id: id, direct: direct})
				} else {
					return nil,
						nmerror.NewInputDataError(fmt.Errorf(
							"could not obtain %s from stop %s, "+
								"element %v in slice is missing field id",
							name,
							stop.ID,
							i,
						))
				}
			default:
				return nil,
					nmerror.NewInputDataError(fmt.Errorf(
						"could not obtain %s from stop %s, "+
							"element %v in slice is neither string nor struct, got %v",
						name,
						stop.ID,
						i,
						element,
					))
			}
		}
		return precedence, nil
	default:
		return nil,
			fmt.Errorf(
				"could not obtain %s from stop %s, "+
					"it is not of type string or slice of string or slice of structs with fields id and direct, got %v",
				name,
				stop.ID,
				field,
			)
	}
}

type precedenceData struct {
	id     string
	direct bool
}

// getSequences returns all the sequences for a stop, based on the "precedes"
// and "succeeds" fields.
func getSequences(stop schema.Stop) ([]sequence, error) {
	var sequences []sequence
	if stop.Precedes != nil {
		precedes, err := precedence(stop, "Precedes")
		if err != nil {
			return nil, err
		}

		predecessorSequences := make([]sequence, len(precedes))
		for i, p := range precedes {
			predecessorSequences[i] = sequence{
				predecessor: stop.ID,
				successor:   p.id,
				direct:      p.direct,
			}
		}
		sequences = append(sequences, predecessorSequences...)
	}

	if stop.Succeeds != nil {
		succeeds, err := precedence(stop, "Succeeds")
		if err != nil {
			return nil, err
		}

		successorSequences := make([]sequence, len(succeeds))
		for i, s := range succeeds {
			successorSequences[i] = sequence{
				predecessor: s.id,
				successor:   stop.ID,
				direct:      s.direct,
			}
		}

		sequences = append(sequences, successorSequences...)
	}

	return sequences, nil
}
