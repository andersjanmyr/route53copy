package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	actual := main()
	expected := "main"
	if actual != expected {
		t.Errorf("main(): %v, expected %v", actual, expected)
	}
}
