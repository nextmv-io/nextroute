// © 2019-present nextmv.io inc

package common_test

import (
	"testing"

	"github.com/nextmv-io/nextroute/common"
)

type TestItem interface {
	Value() float64
}

type TestItemImpl struct {
	value float64
}

func (t TestItemImpl) Value() float64 {
	return t.value
}

func TestSmallest(t *testing.T) {
	items := []TestItem{
		TestItemImpl{value: 1},
		TestItemImpl{value: 2},
		TestItemImpl{value: 3},
		TestItemImpl{value: 4},
		TestItemImpl{value: 5},
	}
	f := func(item TestItem) float64 {
		return item.Value()
	}
	result := common.NSmallest(items, f, 3)
	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %v", len(result))
	}
	if result[0].Value() != 1 {
		t.Errorf("Expected 1, got %v", result[0].Value())
	}
	if result[1].Value() != 2 {
		t.Errorf("Expected 2, got %v", result[1].Value())
	}
	if result[2].Value() != 3 {
		t.Errorf("Expected 3, got %v", result[2].Value())
	}
	result = common.NSmallest(items, f, 10)
	if len(result) != 5 {
		t.Errorf("Expected 3 items, got %v", len(result))
	}
	if result[0].Value() != 1 {
		t.Errorf("Expected 1, got %v", result[0].Value())
	}
	if result[1].Value() != 2 {
		t.Errorf("Expected 2, got %v", result[1].Value())
	}
	if result[2].Value() != 3 {
		t.Errorf("Expected 3, got %v", result[2].Value())
	}
	if result[3].Value() != 4 {
		t.Errorf("Expected 4, got %v", result[3].Value())
	}
	if result[4].Value() != 5 {
		t.Errorf("Expected 5, got %v", result[4].Value())
	}

	result = common.NSmallest(items, f, 0)
	if len(result) != 0 {
		t.Errorf("Expected 0 items, got %v", len(result))
	}

	result = common.NSmallest([]TestItem{}, f, 3)
	if len(result) != 0 {
		t.Errorf("Expected 0 items, got %v", len(result))
	}
}
