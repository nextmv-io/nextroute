// © 2019-present nextmv.io inc

package nextroute_test

import (
	"testing"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func createTestModel() (nextroute.Model, error) {
	model, err := nextroute.NewModel()
	if err != nil {
		return nil, err
	}
	coordinates := [][2]float64{
		{35.358333, 138.761944}, // Mount Fuji
		{35.199444, 139.016944}, // Hakone Open-Air Museum
		{35.658333, 139.745278}, // Tokyo Tower
		{35.689444, 139.7025},   // Shinjuku Gyoen National Garden
		{34.999167, 135.782222}, // Kiyomizu-dera Temple
		{34.988333, 135.771667}, // Fushimi Inari-taisha Shrine
		{34.695833, 135.508611}, // Osaka Castle
		{34.668056, 135.4975},   // Dotombori
		{34.891111, 135.193056}, // Hiroshima Peace Memorial Park
		{34.886944, 135.192222}, // Hiroshima Museum of Art
	}

	for _, c := range coordinates {
		location, err := common.NewLocation(c[1], c[0])
		if err != nil {
			return nil, err
		}
		_, err = model.NewStop(location)
		if err != nil {
			return nil, err
		}
	}
	return model, nil
}
func TestModelStopsDistanceQueries_New(t *testing.T) {
	model, err := createTestModel()

	if err != nil {
		t.Fatal(err)
	}

	modelStopsDistanceQueries, err := nextroute.NewModelStopsDistanceQueries(model.Stops())

	if err != nil {
		t.Fatal(err)
	}

	modelStops := modelStopsDistanceQueries.ModelStops()
	if len(modelStops) != len(model.Stops()) {
		t.Errorf("got %v, want %v", len(modelStops), len((model.Stops())))
	}
}

func TestModelStopsDistanceQueries_WithinDistanceStops(t *testing.T) {
	model, err := createTestModel()

	if err != nil {
		t.Fatal(err)
	}

	modelStopsDistanceQueries, err := nextroute.NewModelStopsDistanceQueries(model.Stops())

	if err != nil {
		t.Fatal(err)
	}

	invalidLocation := common.NewInvalidLocation()

	someStop, err := model.NewStop(invalidLocation)

	if err != nil {
		t.Fatal(err)
	}

	result, err := modelStopsDistanceQueries.WithinDistanceStops(
		someStop,
		common.NewDistance(100, common.Kilometers),
	)

	if err == nil {
		t.Errorf("expected error, got %v", result)
	}

	stops := model.Stops()

	result, err = modelStopsDistanceQueries.WithinDistanceStops(
		stops[0],
		common.NewDistance(100, common.Kilometers),
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 stops, got %v", len(result))
	}

	for _, s := range result {
		if s.Index() == stops[0].Index() {
			t.Errorf("expected %v not to be in result", s.ID())
		}
		d, _ := common.Haversine(s.Location(), stops[0].Location())
		if d.Value(common.Kilometers) > 100 {
			t.Errorf("expected %v to be within 100km, got %v", s.ID(), d.Value(common.Kilometers))
		}
		if s != model.Stops()[1] && s != model.Stops()[2] && s != model.Stops()[3] {
			t.Errorf("expected %v not to be in result", s)
		}
	}

	result, err = modelStopsDistanceQueries.WithinDistanceStops(
		stops[0],
		common.NewDistance(0, common.Kilometers),
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 stops, got %v", len(result))
	}
}
func TestModelStopsDistanceQueries_NearestStops(t *testing.T) {
	model, err := createTestModel()

	if err != nil {
		t.Fatal(err)
	}

	modelStopsDistanceQueries, err := nextroute.NewModelStopsDistanceQueries(model.Stops())

	if err != nil {
		t.Fatal(err)
	}

	invalidLocation := common.NewInvalidLocation()

	someStop, err := model.NewStop(invalidLocation)

	if err != nil {
		t.Fatal(err)
	}

	result, err := modelStopsDistanceQueries.NearestStops(someStop, 3)

	if err == nil {
		t.Errorf("expected error, got %v", result)
	}

	result, err = modelStopsDistanceQueries.NearestStops(model.Stops()[0], -1)

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 stops, got %v", len(result))
	}

	result, err = modelStopsDistanceQueries.NearestStops(model.Stops()[0], 3)

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 stops, got %v", len(result))
	}

	for _, s := range result {
		if s.Index() == model.Stops()[0].Index() {
			t.Errorf("expected %v not to be in result", s.ID())
		}
		if s != model.Stops()[1] && s != model.Stops()[2] && s != model.Stops()[3] {
			t.Errorf("expected %v not to be in result", s)
		}
	}
}
