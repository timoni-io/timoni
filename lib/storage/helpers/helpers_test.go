package helpers_test

import (
	"context"
	"lib/storage/etcd"
	"lib/storage/helpers"
	"lib/utils"
	"testing"
)

var (
	db     = utils.Must(etcd.New(etcd.Embed("", "", true)))
	bucket = db.Bucket("env", "123", "element")
)

func TestIterHelper(t *testing.T) {
	err := bucket.Set("one", 1)
	if err != nil {
		t.Error(err)
	}

	err = bucket.Set("two", 2)
	if err != nil {
		t.Error(err)
	}

	i := 0
	for keyval := range helpers.Iter[int](context.Background(), bucket) {
		i++
		switch keyval.Key {
		case "one":
			if keyval.Value != 1 {
				t.Error("Value not 1")
			}
		case "two":
			if keyval.Value != 2 {
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

func TestWatcherHelper(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for keyval := range helpers.Watch[int](ctx, bucket, "two") {
			if keyval.Key != "two" {
				t.Error("Key not two")
			}
			if keyval.Value != 2 {
				t.Error("Value not 2")
			}
		}
	}()

	err := bucket.Set("one", 1)
	if err != nil {
		t.Error(err)
	}

	err = bucket.Set("two", 2)
	if err != nil {
		t.Error(err)
	}
}

func TestValuesHelper(t *testing.T) {
	err := bucket.Set("one", 1)
	if err != nil {
		t.Error(err)
	}

	err = bucket.Set("two", 2)
	if err != nil {
		t.Error(err)
	}

	values := helpers.Values[int](bucket)

	if len(values) < 2 {
		t.Error("Len < 2")
	}
}

func TestGetHelper(t *testing.T) {
	err := bucket.Set("one", 1)
	if err != nil {
		t.Error(err)
	}

	err = bucket.Set("two", 2)
	if err != nil {
		t.Error(err)
	}

	value, err := helpers.Get[int](bucket, "two")
	if err != nil {
		t.Error(err)
	}

	if value != 2 {
		t.Error("value != 2")
	}
}
