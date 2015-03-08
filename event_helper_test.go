package main

import (
	"fmt"
	"github.com/waterlink/goactor"
	"strings"
	"testing"
	"time"
)

func expectEventToBeA(t *testing.T, actual EventInterface, expectedType string) {
	switch actual.(type) {

	case Follow:
		if expectedType != "Follow" {
			t.Errorf("Expected to be a %s, but was Follow", expectedType)
		}

	case Unfollow:
		if expectedType != "Unfollow" {
			t.Errorf("Expected to be a %s, but was Unfollow", expectedType)
		}

	case Broadcast:
		if expectedType != "Broadcast" {
			t.Errorf("Expected to be a %s, but was Broadcast", expectedType)
		}

	case PrivateMsg:
		if expectedType != "PrivateMsg" {
			t.Errorf("Expected to be a %s, but was PrivateMsg", expectedType)
		}

	case StatusUpdate:
		if expectedType != "StatusUpdate" {
			t.Errorf("Expected to be a %s, but was StatusUpdate", expectedType)
		}

	case Event:
		if expectedType != "Event" {
			t.Errorf("Expected to be a %s, but was Event", expectedType)
		}

	default:
		t.Errorf("Expected to be a %s, but was an unknown type")
	}
}

func expectToFollow(t *testing.T, follows *FollowMap, user int64, follower int64) {
	ok2 := false
	followsOfUser, ok := (*follows)[user]
	if ok {
		_, ok2 = followsOfUser[follower]
	}
	if !ok2 {
		t.Errorf("Expected user %d to follow user %d", follower, user)
	}
}

func expectNotToFollow(t *testing.T, follows *FollowMap, user int64, follower int64) {
	ok2 := false
	followsOfUser, ok := (*follows)[user]
	if ok {
		_, ok2 = followsOfUser[follower]
	}
	if ok2 {
		t.Errorf("Expected user %d to follow user %d", follower, user)
	}
}

func makeTestFollow(follows *FollowMap, seqId int64, user int64, follower int64) {
	userNotifications := &UserNotifications{
		Actor: goactor.NewActor(),
	}

	followEvent, _ := scanEvent(
		strings.NewReader(
			fmt.Sprintf("%d|F|%d|%d\r\n", seqId, follower, user),
		),
	)

	followEvent.Handle(follows, userNotifications)
	<-userNotifications.Inbox()
}

func expectReceivedNotification(t *testing.T, userNotifications *UserNotifications, expectedNotification Notification) {
	received := <-userNotifications.Inbox()
	receivedNotification, _ := received.(*Notification)

	expectIntToBeEqual(t, receivedNotification.UserId, expectedNotification.UserId)
	expectToBeTheSame(t, receivedNotification.Event, expectedNotification.Event)
	expectBoolToBeEqual(t, receivedNotification.Broadcast, expectedNotification.Broadcast)
}

func expectNotToReceiveNotification(t *testing.T, userNotifications *UserNotifications) {
	time.Sleep(1 * time.Millisecond)

	select {
	case received := <-userNotifications.Inbox():
		t.Errorf("Expected not to receive any notification, but received: %s", received)
	default:
	}
}

func expectAllToReceiveEventNotification(t *testing.T, userNotifications *UserNotifications, event EventInterface, users map[int64]bool) {
	for len(users) > 0 {
		received := <-userNotifications.Inbox()
		notification, _ := received.(*Notification)

		userId := notification.UserId

		if _, ok := users[userId]; ok {
			delete(users, userId)
			expectToBeTheSame(t, notification.Event, event)
		} else {
			t.Errorf("Received unexpected notification for user %d", userId)
		}
	}
}
