package main

import (
	"io"
	"testing"
)

type FakeEvent struct {
	sequenceId int64
	handler    chan *FakeEvent
}

func (FakeEvent) String() string                                   { return "fake event" }
func (this *FakeEvent) Handle(*FollowMap, *UserNotifications)      { this.handler <- this }
func (this *FakeEvent) Lift() EventInterface                       { return this }
func (this *FakeEvent) scanRest(io.Reader) (EventInterface, error) { return this, nil }
func (this FakeEvent) getSequenceId() int64                        { return this.sequenceId }
func (FakeEvent) getType() string                                  { return "FAKE" }
func (FakeEvent) getFromUserId() int64                             { return 1 }
func (FakeEvent) getToUserId() int64                               { return 2 }

func expectToReceiveEventsInOrder(t *testing.T, handler chan *FakeEvent, events []*FakeEvent) {
	for _, expectedEvent := range events {
		actualEvent, _ := <-handler
		if actualEvent != expectedEvent {
			t.Errorf("Expected to receive %s, but received %s", expectedEvent, actualEvent)
		}
	}
}
