package maps_test

import (
	"lib/utils/maps"
	"testing"
)

func TestReadOnly(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	rMap := m.ReadOnly()

	rMap.Set("x", "x")

	if rMap.Len() != 0 {
		t.Error("map not read only")
	}

	m.Set("x", "x")

	rMap.Delete("x")

	if rMap.Len() != 1 {
		t.Error("map not read only")
	}
}
