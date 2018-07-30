package main

import (
	"testing"
	"reflect"
)

func TestHello(t *testing.T) {

	testArray := []byte{1,2,3}
	got := parseBody(testArray)
	want := []byte{1,2,3}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got '%s' want '%s'", got, want)
	}
}