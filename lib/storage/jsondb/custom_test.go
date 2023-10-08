package jsondb_test

import (
	"lib/storage/encoding"
	"lib/storage/helpers"
	"strings"
	"testing"
)

type TestStruct struct {
	Name string
}

// Change struct to []string
func (ts *TestStruct) EncodeValue(c encoding.ValueCoder) ([]byte, error) {
	return c.EncodeValue(strings.Split(ts.Name, ","))
}

// Change []string to struct
func (ts *TestStruct) DecodeValue(c encoding.ValueCoder, data []byte) error {
	list, err := helpers.Decode[[]string](c, data)
	if err != nil {
		return err
	}
	ts.Name = strings.Join(list, ",")
	return nil
}

func TestCustomEncode(t *testing.T) {
	const key = "customTestStruct"
	err := db.Set(key, &TestStruct{"TEST1,TEST2"})
	if err != nil {
		t.Error(err)
	}

	list, err := helpers.Get[[]string](db, key)
	if err != nil {
		t.Error(err)
	}

	if list[0] != "TEST1" {
		t.Error("value not TEST1")
	}

	if list[1] != "TEST2" {
		t.Error("value not TEST2")
	}
}

func TestCustomDecode(t *testing.T) {
	const key = "customTestStruct"
	err := db.Set(key, &TestStruct{"TEST1,TEST2"})
	if err != nil {
		t.Error(err)
	}

	strct, err := helpers.Get[TestStruct](db, key)
	if err != nil {
		t.Error(err)
	}

	if strct.Name != "TEST1,TEST2" {
		t.Error("struct value not 'TEST1,TEST'")
	}
}
