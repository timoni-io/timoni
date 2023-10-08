package maps_test

import (
	"encoding/json"
	"lib/utils/maps"
	"testing"
)

func TestSafeCopy(t *testing.T) {
	m := maps.New(map[string]string{"1": "a"}).Safe()
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

func TestSafeUnmarshalJSON(t *testing.T) {
	m := maps.SafeMap[string, string]{}
	err := json.Unmarshal([]byte(`{"x":"x"}`), &m)
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
