package main

import (
	"io"
	"testing"
)

type FakeClientConnection struct {
	io.Reader
}

func (this FakeClientConnection) Write(buffer []byte) (int, error) {
	return len(buffer), nil
}

func (this FakeClientConnection) Close() error {
	return nil
}

func expectToBeFakeClientConnection(t *testing.T, actual *ClientConnection, expected FakeClientConnection) {
	if actualFake, ok := (*actual).(FakeClientConnection); ok {

		if actualFake != expected {
			t.Errorf("Expected connection %s to be %s", actualFake, expected)
		}

	} else {

		t.Errorf("Unable to convert %s to FakeClientConnection", actual)

	}
}
