package main

import (
	"fmt"
	"io"
)

type EventInterface interface {
	String() string
	Handle(*FollowMap, *UserNotifications)
	Lift() EventInterface

	scanRest(io.Reader) (EventInterface, error)

	getSequenceId() int64
	getType() string
	getFromUserId() int64
	getToUserId() int64
}

type Event struct {
	SequenceId int64
	Type       string
	FromUserId int64
	ToUserId   int64
}

type Follow struct {
	Event
}

type Unfollow struct {
	Event
}

type Broadcast struct {
	Event
}

type PrivateMsg struct {
	Event
}

type StatusUpdate struct {
	Event
}

func (event Event) getSequenceId() int64 { return event.SequenceId }
func (event Event) getType() string      { return event.Type }
func (event Event) getFromUserId() int64 { return event.FromUserId }
func (event Event) getToUserId() int64   { return event.ToUserId }

func (event Event) Lift() EventInterface {
	switch event.Type {
	case "F":
		return Follow{event}
	case "U":
		return Unfollow{event}
	case "B":
		return Broadcast{event}
	case "P":
		return PrivateMsg{event}
	case "S":
		return StatusUpdate{event}
	}
	return event
}

func (event Event) String() string {
	return fmt.Sprintf(
		"%d|%s|%d|%d",
		event.SequenceId,
		event.Type,
		event.FromUserId,
		event.ToUserId,
	)
}

func (event Broadcast) String() string {
	return fmt.Sprintf(
		"%d|%s",
		event.SequenceId,
		event.Type,
	)
}

func (event StatusUpdate) String() string {
	return fmt.Sprintf(
		"%d|%s|%d",
		event.SequenceId,
		event.Type,
		event.FromUserId,
	)
}

func scanEvent(reader io.Reader) (EventInterface, error) {
	var eventSequenceId int64
	var eventType string

	_, error := fmt.Fscanf(reader, "%d|%1s", &eventSequenceId, &eventType)
	if error != nil {
		return Event{}, error
	}

	event := Event{
		SequenceId: eventSequenceId,
		Type:       eventType,
	}

	return event.Lift().scanRest(reader)
}

func (event Event) scanRest(reader io.Reader) (EventInterface, error) {
	_, error := fmt.Fscanf(reader, "|%d|%d\r\n", &event.FromUserId, &event.ToUserId)
	if error != nil {
		return Event{}, error
	}

	return event, nil
}

func (event Broadcast) scanRest(reader io.Reader) (EventInterface, error) {
	_, error := fmt.Fscanf(reader, "\r\n")
	if error != nil {
		return Event{}, error
	}

	return event, nil
}

func (event StatusUpdate) scanRest(reader io.Reader) (EventInterface, error) {
	_, error := fmt.Fscanf(reader, "|%d\r\n", &event.FromUserId)
	if error != nil {
		return Event{}, error
	}

	return event, nil
}

func (event Event) Handle(follows *FollowMap, userNotifications *UserNotifications) {
	event.Lift().Handle(follows, userNotifications)
}

func (event Follow) Handle(follows *FollowMap, userNotifications *UserNotifications) {
	if (*follows)[event.getToUserId()] == nil {
		(*follows)[event.getToUserId()] = NewFollowMapValue()
	}
	(*follows)[event.getToUserId()][event.getFromUserId()] = true

	userNotifications.Send(&Notification{
		Event:  &event,
		UserId: event.getToUserId(),
	})
}

func (event Unfollow) Handle(follows *FollowMap, userNotifications *UserNotifications) {
	if (*follows)[event.getToUserId()] == nil {
		(*follows)[event.getToUserId()] = NewFollowMapValue()
	}
	delete((*follows)[event.getToUserId()], event.getFromUserId())
}

func (event Broadcast) Handle(follows *FollowMap, userNotifications *UserNotifications) {
	userNotifications.Send(&Notification{
		Event:     &event,
		Broadcast: true,
	})
}

func (event PrivateMsg) Handle(follows *FollowMap, userNotifications *UserNotifications) {
	userNotifications.Send(&Notification{
		Event:  &event,
		UserId: event.getToUserId(),
	})
}

func (event StatusUpdate) Handle(follows *FollowMap, userNotifications *UserNotifications) {
	for followerId := range (*follows)[event.getFromUserId()] {
		userNotifications.Send(&Notification{
			Event:  &event,
			UserId: followerId,
		})
	}
}
