// © 2019-present nextmv.io inc

package common_test

import (
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

func TestBoundingBox(t *testing.T) {
	invalidBox := common.NewInvalidBoundingBox()
	if invalidBox.IsValid() {
		t.Errorf("Expected invalid bounding box")
	}
	if invalidBox.Maximum().IsValid() {
		t.Errorf("Expected invalid maximum for invalid bounding box")
	}
	if invalidBox.Minimum().IsValid() {
		t.Errorf("Expected invalid minimum for invalid bounding box")
	}
	location, _ := common.NewLocation(0, 0)

	box := common.NewBoundingBox(common.Locations{location})

	if !box.IsValid() {
		t.Errorf("Expected valid bounding box")
	}
	if !box.Maximum().IsValid() {
		t.Errorf("Expected valid maximum for valid bounding box")
	}
	if !box.Minimum().IsValid() {
		t.Errorf("Expected valid minimum for valid bounding box")
	}
	if box.Maximum().Latitude() != 0 {
		t.Errorf("Expected maximum latitude 0, got %v", box.Maximum().Latitude())
	}
	if box.Maximum().Longitude() != 0 {
		t.Errorf("Expected maximum longitude 0, got %v", box.Maximum().Longitude())
	}
	if box.Minimum().Latitude() != 0 {
		t.Errorf("Expected minimum latitude 0, got %v", box.Minimum().Latitude())
	}
	if box.Minimum().Longitude() != 0 {
		t.Errorf("Expected minimum longitude 0, got %v", box.Minimum().Longitude())
	}
	if box.Width().Value(common.Meters) != 0 {
		t.Errorf("Expected width 0, got %v", box.Width())
	}
	if box.Height().Value(common.Meters) != 0 {
		t.Errorf("Expected height 0, got %v", box.Height())
	}

	location2, _ := common.NewLocation(1, 1)

	box = common.NewBoundingBox(common.Locations{location, location2})

	if !box.IsValid() {
		t.Errorf("Expected valid bounding box")
	}
	if !box.Maximum().IsValid() {
		t.Errorf("Expected valid maximum for valid bounding box")
	}
	if !box.Minimum().IsValid() {
		t.Errorf("Expected valid minimum for valid bounding box")
	}
	if box.Maximum().Latitude() != 1 {
		t.Errorf("Expected maximum latitude 1, got %v", box.Maximum().Latitude())
	}
	if box.Maximum().Longitude() != 1 {
		t.Errorf("Expected maximum longitude 1, got %v", box.Maximum().Longitude())
	}
	if box.Minimum().Latitude() != 0 {
		t.Errorf("Expected minimum latitude 0, got %v", box.Minimum().Latitude())
	}
	if box.Minimum().Longitude() != 0 {
		t.Errorf("Expected minimum longitude 0, got %v", box.Minimum().Longitude())
	}
	if !common.WithinTolerance(box.Width().Value(common.Meters), 111194.92664455874, 0.000001) {
		t.Errorf("Expected width 111194.92664455874, got %v", box.Width().Value(common.Meters))
	}
	if !common.WithinTolerance(box.Height().Value(common.Meters), 111194.92664455874, 0.000001) {
		t.Errorf("Expected height 111194.92664455874, got %v", box.Height().Value(common.Meters))
	}

	location3, _ := common.NewLocation(-2, 2)
	location4, _ := common.NewLocation(2, -2)

	box = common.NewBoundingBox(common.Locations{location, location2, location3, location4})

	if !box.IsValid() {
		t.Errorf("Expected valid bounding box")
	}
	if !box.Maximum().IsValid() {
		t.Errorf("Expected valid maximum for valid bounding box")
	}
	if !box.Minimum().IsValid() {
		t.Errorf("Expected valid minimum for valid bounding box")
	}
	if box.Maximum().Latitude() != 2 {
		t.Errorf("Expected maximum latitude 2, got %v", box.Maximum().Latitude())
	}
	if box.Maximum().Longitude() != 2 {
		t.Errorf("Expected maximum longitude 2, got %v", box.Maximum().Longitude())
	}
	if box.Minimum().Latitude() != -2 {
		t.Errorf("Expected minimum latitude -2, got %v", box.Minimum().Latitude())
	}
	if box.Minimum().Longitude() != -2 {
		t.Errorf("Expected minimum longitude -2, got %v", box.Minimum().Longitude())
	}
	if !common.WithinTolerance(box.Width().Value(common.Meters), 444508.6487983116, 0.000001) {
		t.Errorf("Expected width 222389.85328911748, got %v", box.Width().Value(common.Meters))
	}
	if !common.WithinTolerance(box.Height().Value(common.Meters), 444779.70657823497, 0.000001) {
		t.Errorf("Expected height 222389.85328911748, got %v", box.Height().Value(common.Meters))
	}
}
