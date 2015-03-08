package main

import (
	"testing"
)

type Stringable interface {
	String() string
}

func expectToBeTheSame(t *testing.T, actual Stringable, expected Stringable) {
	if actual.String() != expected.String() {
		t.Errorf("Expected to be equal to: %s, but got: %s", expected, actual)
	}
}

func expectBoolToBeEqual(t *testing.T, actual bool, expected bool) {
	if actual != expected {
		t.Errorf("Expected to be equal to: %s, but got: %s", expected, actual)
	}
}

func expectIntToBeEqual(t *testing.T, actual int64, expected int64) {
	if actual != expected {
		t.Errorf("Expected to be equal to: %d, but got: %d", expected, actual)
	}
}

func expectStringToBeEqual(t *testing.T, actual string, expected string) {
	if actual != expected {
		t.Errorf("Expected to be equal to: '%s', but got: '%s'", expected, actual)
	}
}
