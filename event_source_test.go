package main

import (
	"github.com/waterlink/goactor"
	"io"
	"strings"
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

func TestReceive(t *testing.T) {
	rawEvents := []string{
		"54|F|37|99\r\n",
		"49|P|124|19\r\n",
		"67|U|47|932\r\n",
		"63|B\r\n",
		"54|F|37|99\r\n",
		"abnormalstring\r\n",
	}
	stream := strings.Join(rawEvents, "")
	userRelationships := &UserRelationships{
		Actor: goactor.NewActor(),
	}
	connection := &FakeConnection{
		Reader: strings.NewReader(stream),
		closed: false,
	}

	eventSource := &EventSource{
		Actor:             goactor.NewActor(),
		connection:        connection,
		userRelationships: userRelationships,
	}

	eventSource.Act(true)
	expectEventReceived(t, userRelationships, (&Event{
		SequenceId: 54,
		Type:       "F",
		FromUserId: 37,
		ToUserId:   99,
	}).Lift())
	<-eventSource.Inbox()

	eventSource.Act(true)
	expectEventReceived(t, userRelationships, (&Event{
		SequenceId: 49,
		Type:       "P",
		FromUserId: 124,
		ToUserId:   19,
	}).Lift())
	<-eventSource.Inbox()

	eventSource.Act(true)
	expectEventReceived(t, userRelationships, (&Event{
		SequenceId: 67,
		Type:       "U",
		FromUserId: 47,
		ToUserId:   932,
	}).Lift())
	<-eventSource.Inbox()

	eventSource.Act(true)
	expectEventReceived(t, userRelationships, (&Event{
		SequenceId: 63,
		Type:       "B",
	}).Lift())
	<-eventSource.Inbox()

	eventSource.Act(true)
	expectEventReceived(t, userRelationships, (&Event{
		SequenceId: 54,
		Type:       "F",
		FromUserId: 37,
		ToUserId:   99,
	}).Lift())
	<-eventSource.Inbox()

	eventSource.Act(true)
	expectBoolToBeEqual(t, connection.closed, true)
	expectInboxToBeDead(t, eventSource)
}
