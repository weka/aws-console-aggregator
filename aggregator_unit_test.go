package main

import (
	"testing"
)

func checkOverlapping(prev, latest string, expected int, t *testing.T) {
	if got := findOverlappingEndingIndex(prev, latest); got != expected {
		t.Errorf("findOverlappingEndingIndex(\"%s\",\"%s\") = %d, expected %d", prev, latest, got, expected)
	}
}

func TestFindOverlappingEndingIndex(t *testing.T) {
	prev := "abcd"
	latest := "cdef" //latest and prev are the same size
	expected := 2
	checkOverlapping(prev, latest, expected, t)
	latest = "cde" //latest is shorter than prev
	expected = 2
	checkOverlapping(prev, latest, expected, t)
	latest = "abcde" //latest is longer than prev
	expected = 4
	checkOverlapping(prev, latest, expected, t)
	latest = "cdcde" //2 overlaps
	expected = 2
	checkOverlapping(prev, latest, expected, t)
	latest = "efgh" //no overlap
	expected = 0
	checkOverlapping(prev, latest, expected, t)
}
