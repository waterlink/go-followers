package main

import (
	"github.com/waterlink/goactor"
	"testing"
)

func TestReceiveEvent(t *testing.T) {
	userNotifications := &UserNotifications{}

	userRelationships := &UserRelationships{
		Actor:              goactor.NewActor(),
		follows:            NewFollowMap(),
		lastSeenSequenceId: 0,
		userNotifications:  userNotifications,
		bufferedEvents:     NewEventMap(),
	}

	handler := make(chan *FakeEvent)

	firstEvent := &FakeEvent{sequenceId: 1, handler: handler}
	secondEvent := &FakeEvent{sequenceId: 2, handler: handler}
	thirdEvent := &FakeEvent{sequenceId: 3, handler: handler}

	firstMessage := EventInterface(firstEvent)
	secondMessage := EventInterface(secondEvent)
	thirdMessage := EventInterface(thirdEvent)

	go userRelationships.Act(&secondMessage)
	go userRelationships.Act(&thirdMessage)
	go userRelationships.Act(&firstMessage)

	expectToReceiveEventsInOrder(t, handler, []*FakeEvent{firstEvent, secondEvent, thirdEvent})
}
