package jsondb_test

import (
	"context"
	"lib/storage/helpers"
	"testing"
	"time"
)

var (
	bucket = db.Bucket("env", "123", "element")
)

func TestBucketGetSetDelete(t *testing.T) {
	err := bucket.Set("one", 1)
	if err != nil {
		t.Error(err)
	}

	err = bucket.Set("two", 2)
	if err != nil {
		t.Error(err)
	}

	val, err := helpers.Get[uint64](bucket, "one")
	if err != nil {
		t.Error(err)
	}
	if val != 1 {
		t.Error("Value not 1")
	}

	var two int
	err = bucket.Get("two", &two)
	if err != nil {
		t.Error(err)
	}
	if two != 2 {
		t.Error("Value not 2")
	}
}

func TestBucketDeleteExists(t *testing.T) {
	const key = "desd1"
	err := bucket.Set(key, 100)
	if err != nil {
		t.Error(err)
	}

	if !bucket.Exists(key) {
		t.Log("Value not created")
	}

	err = bucket.Delete(key)
	if err != nil {
		t.Error(err)
	}

	if bucket.Exists(key) {
		t.Log("Value not deleted")
	}

	val, err := helpers.Get[uint64](bucket, key)
	if err == nil {
		t.Error("Got value, expected error")
	}
	if val != 0 {
		t.Error("Expected zero value")
	}
}

func TestBucketIter(t *testing.T) {
	err := bucket.Set("one", '1')
	if err != nil {
		t.Error(err)
	}

	err = bucket.Set("two", '2')
	if err != nil {
		t.Error(err)
	}

	i := 0
	for item := range helpers.Iter[rune](context.Background(), bucket) {
		i++
		switch item.Key {
		case "one":
			if item.Value != '1' {
				t.Error("Value not 1")
			}
		case "two":
			if item.Value != '2' {
				t.Error("Value not 2")
			}
		default:
			t.Log("Unknown item:", item)
			i--
		}
	}

	if i < 2 {
		t.Error("Not full iteration")
	}
}

func TestBucketWatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hit := make(chan struct{})

	go func() {
		for keyval := range helpers.Watch[rune](ctx, bucket, "two") {
			if keyval.Key != "two" {
				t.Error("Key not two")
			}
			if keyval.Value != '2' {
				t.Error("Value not 2")
			}
			hit <- struct{}{}
		}
	}()

	time.Sleep(time.Second)

	err := bucket.Set("one", '1')
	if err != nil {
		t.Error(err)
	}

	err = bucket.Set("two", '2')
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

func TestBucketLenKeyVals(t *testing.T) {
	err := bucket.Set("one", 1)
	if err != nil {
		t.Error(err)
	}

	err = bucket.Set("two", 2)
	if err != nil {
		t.Error(err)
	}

	length, err := bucket.Len()
	if err != nil {
		t.Error(err)
	}

	if length < 2 {
		t.Error("Len < 2")
	}

	keys, err := bucket.Keys()
	if err != nil {
		t.Error(err)
	}

	if len(keys) < 2 {
		t.Error("Keys Len < 2")
	}

	vals, err := bucket.Values()
	if err != nil {
		t.Error(err)
	}

	if len(vals) < 2 {
		t.Error("Vals Len < 2")
	}
}

func TestBucketUnmarshal(t *testing.T) {
	data, err := db.Encoding().EncodeValue(bucket)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(data))

	newBucket := db.Bucket()
	err = newBucket.Encoding().DecodeValue(data, newBucket)
	if err != nil {
		t.Error(err)
	}

	if newBucket.Prefix() != bucket.Prefix() {
		t.Error("unmarshal newBucket failed, wrong prefix", newBucket.Prefix(), "!=", bucket.Prefix())
	}

	err = newBucket.Set("one", 1)
	if err != nil {
		t.Error(err)
	}

	val, err := helpers.Get[uint64](newBucket, "one")
	if err != nil {
		t.Error(err)
	}
	if val != 1 {
		t.Log(`Get "one" != 1`)
	}
}
