package maps_test

import (
	"fmt"
	"lib/utils/maps"
	"testing"
)

func TestWeightedForEach(t *testing.T) {
	m := maps.NewWeighted(map[string]maps.Weighted[string]{
		"1": {Value: "x", Weight: 3},
		"2": {Value: "x", Weight: 2},
		"3": {Value: "x", Weight: 1},
	})

	if m == nil {
		t.Fail()
		return
	}

	i := 0
	m.ForEach(func(k string, v maps.Weighted[string]) error {
		i++

		if k != fmt.Sprint(i) {
			t.Errorf("%s != %d", k, i)
		}
		return nil
	})
}

func TestWeightedKeys(t *testing.T) {
	m := maps.NewWeighted(map[string]maps.Weighted[string]{
		"1": {Value: "x", Weight: 3},
		"2": {Value: "x", Weight: 2},
		"3": {Value: "x", Weight: 1},
	})
	if m == nil {
		t.Fail()
		return
	}

	keys := m.Keys()

	if len(keys) != 3 {
		t.Fail()
	}

	expected := []string{
		"1",
		"2",
		"3",
	}

	for i, k := range keys {
		if expected[i] != k {
			t.Errorf("%s != %s", expected[i], k)
		}
	}
}

func TestWeightedValues(t *testing.T) {
	m := maps.NewWeighted(map[string]maps.Weighted[string]{
		"1": {Value: "a", Weight: 3},
		"2": {Value: "b", Weight: 2},
		"3": {Value: "c", Weight: 1},
	})
	if m == nil {
		t.Fail()
		return
	}

	values := m.Values()

	if len(values) != 3 {
		t.Fail()
	}

	expected := []string{
		"a",
		"b",
		"c",
	}

	for i, v := range values {
		if expected[i] != v.Value {
			t.Errorf("%s != %s", expected[i], v.Value)
		}
	}
}
