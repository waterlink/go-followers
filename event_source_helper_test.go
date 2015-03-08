package main

import (
	"github.com/waterlink/goactor"
	"io"
	"testing"
)

type FakeConnection struct {
	io.Reader
	closed bool
}

func (this *FakeConnection) Close() error {
	this.closed = true
	return nil
}

func expectEventReceived(t *testing.T, userRelationships *UserRelationships, event EventInterface) {
	received := <-userRelationships.Inbox()
	if actualEvent, ok := received.(*EventInterface); ok {
		expectToBeTheSame(t, *actualEvent, event)
	} else {
		t.Errorf("Expected to receive event %s, but received nothing", event)
	}
}

func expectInboxToBeDead(t *testing.T, actor goactor.ActorInterface) {
	if _, ok := <-actor.Inbox(); ok {
		t.Errorf("Expected actor %s's inbox to be dead", actor)
	}
}
