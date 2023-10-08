package maps_test

import (
	"encoding/json"
	"lib/utils/maps"
	"lib/utils/set"
	"testing"
)

func TestExists(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	if m.Exists("x") {
		t.Fail()
	}

	m.Set("x", "x")
	if !m.Exists("x") {
		t.Fail()
	}
}

func TestGet(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	if m.Get("x") != "" {
		t.Fail()
	}

	m.Set("x", "x")
	if m.Get("x") != "x" {
		t.Fail()
	}
}

func TestGetFull(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	if v, ok := m.GetFull("x"); v != "" || ok {
		t.Fail()
	}

	m.Set("x", "x")
	if v, ok := m.GetFull("x"); v != "x" || !ok {
		t.Fail()
	}
}

func TestSet(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	m.Set("x", "x")
	if m.Len() == 0 {
		t.Fail()
	}
}

func TestDelete(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	m.Set("x", "x")
	if m.Len() == 0 {
		t.Fail()
	}

	m.Delete("x")
	if m.Len() != 0 {
		t.Fail()
	}
}

func TestCommit(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	m.Commit(func(data map[string]string) {
		data["x"] = "x"
	})

	if m.Len() == 0 {
		t.Fail()
	}
}

func TestIter(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	m.Set("x", "x")

	for i := range m.Iter() {
		if i.Key != "x" {
			t.Fail()
		}
		if i.Value != "x" {
			t.Fail()
		}
	}
}

func TestForEach(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	m.Set("x", "x")

	m.ForEach(func(k, v string) error {
		if k != "x" {
			t.Fail()
		}
		if v != "x" {
			t.Fail()
		}
		return nil
	})
}

func TestKeys(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	m.Set("1", "a")
	m.Set("2", "b")
	m.Set("3", "c")

	keys := m.Keys()

	if len(keys) != 3 {
		t.Fail()
	}

	expected := set.New("1", "2", "3")

	for _, k := range keys {
		if !expected.Exists(k) {
			t.Errorf("key %s not found", k)
		}
	}
}

func TestValues(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	m.Set("1", "a")
	m.Set("2", "b")
	m.Set("3", "c")

	values := m.Values()

	if len(values) != 3 {
		t.Fail()
	}

	expected := set.New("a", "b", "c")

	for _, v := range values {
		if !expected.Exists(v) {
			t.Errorf("value %s not found", v)
		}
	}
}

func TestLen(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	m.Set("1", "a")
	m.Set("2", "b")
	m.Set("3", "c")

	if m.Len() != 3 {
		t.Fail()
	}
}

func TestCopy(t *testing.T) {
	m := maps.New(map[string]string{"1": "a"})
	if m == nil {
		t.Fail()
	}

	cp := m.Copy()
	if cp.Len() != 1 {
		t.Errorf("invalid len %d", cp.Len())
	}

	if cp.Get("1") != "a" {
		t.Error("invalid cp value")
	}

	cp.Set("1", "b")

	if m.Get("1") == "b" {
		t.Error("invalid m value")
	}

	cp.Set("2", "b")
	if m.Len() != 1 {
		t.Error("invalid m len")
	}
}

func TestMarshalJSON(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}
	m.Set("x", "x")
	b, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}

	if string(b) != `{"x":"x"}` {
		t.Error(string(b))
	}
}

func TestUnmarshalJSON(t *testing.T) {
	m := maps.New[string, string](nil)
	if m == nil {
		t.Fail()
	}

	err := json.Unmarshal([]byte(`{"x":"x"}`), m)
	if err != nil {
		t.Error(err)
	}

	if m.Len() != 1 {
		t.Error("invalid len")
	}

	if !m.Exists("x") || m.Get("x") != "x" {
		t.Fail()
	}
}
