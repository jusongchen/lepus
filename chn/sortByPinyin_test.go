package chn

import (
	"sort"
	"testing"
)

func TestSortByPinyin(t *testing.T) {
	s := []string{"一", "二", "三", "ha", "AN", ","}

	// sort.Strings(s)
	sort.Sort(ByPinyin(s))
	expected := []string{",", "AN", "ha", "二", "三", "一"}

	for i, v := range expected {
		if s[i] != v {
			t.Errorf("expect %v but get %v", expected, s)
		}
	}
}
