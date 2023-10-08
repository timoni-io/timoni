package set_test

import (
	"lib/utils/set"
	"lib/utils/slice"
	"sort"
	"testing"
)

func TestRigid(t *testing.T) {
	rigid := set.NewRigid[int, uint8](10)
	rigid.Add(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	if rigid.Len() != 10 {
		t.Errorf("Expected 10, got %d", rigid.Len())
	}

	if !rigid.Exists(1) {
		t.Errorf("Expected 1 to be in the set")
	}

	rigid.Add(11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	if rigid.Len() > 10 {
		t.Errorf("Expected 10, got %d", rigid.Len())
	}

	if rigid.Exists(5) {
		t.Errorf("Expected 5 not to be in the set")
	}

	list := rigid.List()
	sort.Ints(list)
	if !slice.Equal([]int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, list) {
		t.Errorf("Expected []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, got %v", list)
	}
}
