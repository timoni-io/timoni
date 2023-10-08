package etcd_test

import (
	"context"
	"lib/storage/etcd"
	"lib/storage/helpers"
	"lib/utils"
	"lib/utils/types"
	"testing"
	"time"
)

var (
	db = utils.Must(etcd.New(etcd.Embed("", "", true)))
)

func TestGetSet(t *testing.T) {
	err := db.Set("one", 1)
	if err != nil {
		t.Error(err)
	}

	err = db.Set("two", 2)
	if err != nil {
		t.Error(err)
	}

	val, err := helpers.Get[uint64](db, "one")
	if err != nil {
		t.Error(err)
	}
	if val != 1 {
		t.Error("Value not 1")
	}

	var two int
	err = db.Get("two", &two)
	if err != nil {
		t.Error(err)
	}
	if two != 2 {
		t.Error("Value not 2")
	}
}

func TestDeleteExists(t *testing.T) {
	err := db.Set("test1", 1)
	if err != nil {
		t.Error(err)
	}

	if !db.Exists("test1") {
		t.Log("Value not created")
	}

	err = db.Delete("test1")
	if err != nil {
		t.Error(err)
	}

	if db.Exists("test1") {
		t.Log("Value not deleted")
	}

	val, err := helpers.Get[uint64](db, "test1")
	if err == nil {
		t.Error("Got value, expected error")
	}
	if val != 0 {
		t.Error("Expected zero value")
	}
}

func TestIter(t *testing.T) {
	err := db.Set("one", '1')
	if err != nil {
		t.Error(err)
	}

	err = db.Set("two", '2')
	if err != nil {
		t.Error(err)
	}

	i := 0
	for keyval := range helpers.Iter[rune](context.Background(), db) {
		i++
		switch keyval.Key {
		case "one":
			if keyval.Value != '1' {
				t.Error("Value not 1")
			}
		case "two":
			if keyval.Value != '2' {
				t.Error("Value not 2")
			}
		default:
			i--
		}
	}

	if i < 2 {
		t.Error("Not full iteration")
	}
}

func TestWatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hit := make(chan struct{})

	go func() {
		for item := range helpers.Watch[rune](ctx, db, "two") {
			if item.Event != types.PutEvent {
				t.Error("Not Put Event")
			}
			if item.Key != "two" {
				t.Error("Key not two")
			}
			if item.Value != '2' {
				t.Error("Value not 2")
			}
			hit <- struct{}{}
		}

	}()

	err := db.Set("one", '1')
	if err != nil {
		t.Error(err)
	}

	err = db.Set("two", '2')
	if err != nil {
		t.Error(err)
	}

	ts := time.NewTimer(3 * time.Second)
	defer ts.Stop()

	select {
	case <-hit:
	case <-ts.C:
		t.Error("Change not found")
	}
}

func TestLenKeyVals(t *testing.T) {
	err := db.Set("test/one", 1)
	if err != nil {
		t.Error(err)
	}

	err = db.Set("test/two", 2)
	if err != nil {
		t.Error(err)
	}

	length, err := db.Len("test/")
	if err != nil {
		t.Error(err)
	}

	if length < 2 {
		t.Error("Len < 2")
	}

	keys, err := db.Keys("test/")
	if err != nil {
		t.Error(err)
	}

	if len(keys) < 2 {
		t.Error("Keys Len < 2")
	}

	vals, err := db.Values("test/")
	if err != nil {
		t.Error(err)
	}

	if len(vals) < 2 {
		t.Error("Vals Len < 2")
	}
}
