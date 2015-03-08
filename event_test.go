package main

import (
	"github.com/waterlink/goactor"
	"strings"
	"testing"
)

func TestGetters(t *testing.T) {
	event := Event{
		SequenceId: 35,
		Type:       "U",
		FromUserId: 731,
		ToUserId:   54,
	}

	expectIntToBeEqual(t, event.getSequenceId(), 35)
	expectStringToBeEqual(t, event.getType(), "U")
	expectIntToBeEqual(t, event.getFromUserId(), 731)
	expectIntToBeEqual(t, event.getToUserId(), 54)
}

func TestLift(t *testing.T) {
	followEvent := Event{
		SequenceId: 77,
		Type:       "F",
		FromUserId: 498,
		ToUserId:   99,
	}
	expectToBeA(t, followEvent, "Event")
	expectToBeA(t, followEvent.Lift(), "Follow")

	unfollowEvent := Event{
		SequenceId: 77,
		Type:       "U",
		FromUserId: 498,
		ToUserId:   99,
	}
	expectToBeA(t, unfollowEvent, "Event")
	expectToBeA(t, unfollowEvent.Lift(), "Unfollow")

	broadcastEvent := Event{
		SequenceId: 126,
		Type:       "B",
	}
	expectToBeA(t, broadcastEvent, "Event")
	expectToBeA(t, broadcastEvent.Lift(), "Broadcast")

	privateMsgEvent := Event{
		SequenceId: 77,
		Type:       "P",
		FromUserId: 498,
		ToUserId:   99,
	}
	expectToBeA(t, privateMsgEvent, "Event")
	expectToBeA(t, privateMsgEvent.Lift(), "PrivateMsg")

	statusUpdateEvent := Event{
		SequenceId: 99,
		Type:       "S",
		FromUserId: 42,
	}
	expectToBeA(t, statusUpdateEvent, "Event")
	expectToBeA(t, statusUpdateEvent.Lift(), "StatusUpdate")
}

func TestString(t *testing.T) {
	followEvent := Event{
		SequenceId: 77,
		Type:       "F",
		FromUserId: 498,
		ToUserId:   99,
	}
	expectStringToBeEqual(t, followEvent.Lift().String(), "77|F|498|99")

	unfollowEvent := Event{
		SequenceId: 77,
		Type:       "U",
		FromUserId: 498,
		ToUserId:   99,
	}
	expectStringToBeEqual(t, unfollowEvent.Lift().String(), "77|U|498|99")

	broadcastEvent := Event{
		SequenceId: 126,
		Type:       "B",
	}
	expectStringToBeEqual(t, broadcastEvent.Lift().String(), "126|B")

	privateMsgEvent := Event{
		SequenceId: 77,
		Type:       "P",
		FromUserId: 498,
		ToUserId:   99,
	}
	expectStringToBeEqual(t, privateMsgEvent.Lift().String(), "77|P|498|99")

	statusUpdateEvent := Event{
		SequenceId: 99,
		Type:       "S",
		FromUserId: 42,
	}
	expectStringToBeEqual(t, statusUpdateEvent.Lift().String(), "99|S|42")
}

func TestScanEvent(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{"77|F|498|99\r\n", "77|F|498|99"},
		{"77|U|498|99\r\n", "77|U|498|99"},
		{"126|B\r\n", "126|B"},
		{"77|P|498|99\r\n", "77|P|498|99"},
		{"99|S|42\r\n", "99|S|42"},
	}

	for _, test := range tests {
		event, err := scanEvent(strings.NewReader(test.input))
		if err != nil {
			t.Error(err)
		} else {
			expectStringToBeEqual(t, event.String(), test.expected)
		}
	}
}

func TestHandleFollow(t *testing.T) {
	userNotifications := &UserNotifications{
		Actor: goactor.NewActor(),
	}
	follows := NewFollowMap()
	event := Event{
		SequenceId: 77,
		Type:       "F",
		FromUserId: 498,
		ToUserId:   99,
	}

	event.Handle(&follows, userNotifications)
	expectToFollow(t, &follows, 99, 498)
	expectReceivedNotification(t, userNotifications, Notification{
		Event:  &event,
		UserId: 99,
	})
}

func TestHandleUnfollowWhenNotFollowing(t *testing.T) {
	userNotifications := &UserNotifications{
		Actor: goactor.NewActor(),
	}
	follows := NewFollowMap()
	event := Event{
		SequenceId: 77,
		Type:       "U",
		FromUserId: 498,
		ToUserId:   99,
	}

	event.Handle(&follows, userNotifications)
	expectNotToFollow(t, &follows, 99, 498)
	expectNotToReceiveNotification(t, userNotifications)
}

func TestHandleUnfollowWhenAlreadyFollowing(t *testing.T) {
	userNotifications := &UserNotifications{
		Actor: goactor.NewActor(),
	}
	follows := NewFollowMap()
	event := Event{
		SequenceId: 77,
		Type:       "U",
		FromUserId: 498,
		ToUserId:   99,
	}

	makeTestFollow(&follows, 67, 99, 498)

	event.Handle(&follows, userNotifications)
	expectNotToFollow(t, &follows, 99, 498)
	expectNotToReceiveNotification(t, userNotifications)
}

func TestHandleBroadcast(t *testing.T) {
	userNotifications := &UserNotifications{
		Actor: goactor.NewActor(),
	}
	follows := NewFollowMap()
	event := Event{
		SequenceId: 123,
		Type:       "B",
	}

	event.Handle(&follows, userNotifications)
	expectIntToBeEqual(t, int64(len(follows)), 0)

	liftedEvent := event.Lift()
	expectReceivedNotification(t, userNotifications, Notification{
		Event:     liftedEvent,
		Broadcast: true,
	})
}

func TestHandlePrivateMsg(t *testing.T) {
	userNotifications := &UserNotifications{
		Actor: goactor.NewActor(),
	}
	follows := NewFollowMap()
	event := Event{
		SequenceId: 77,
		Type:       "P",
		FromUserId: 498,
		ToUserId:   99,
	}

	event.Handle(&follows, userNotifications)
	expectIntToBeEqual(t, int64(len(follows)), 0)
	expectReceivedNotification(t, userNotifications, Notification{
		Event:  &event,
		UserId: 99,
	})
}

func TestHandleStatusUpdateWhenNoFollowers(t *testing.T) {
	userNotifications := &UserNotifications{
		Actor: goactor.NewActor(),
	}
	follows := NewFollowMap()
	event := Event{
		SequenceId: 59,
		Type:       "S",
		FromUserId: 71,
	}

	event.Handle(&follows, userNotifications)
	expectIntToBeEqual(t, int64(len(follows)), 0)
	expectNotToReceiveNotification(t, userNotifications)
}

func TestHandleStatusUpdateWhenThereAreFollowers(t *testing.T) {
	userNotifications := &UserNotifications{
		Actor: goactor.NewActor(),
	}
	follows := NewFollowMap()
	event := Event{
		SequenceId: 59,
		Type:       "S",
		FromUserId: 71,
	}

	makeTestFollow(&follows, 14, 59, 67)
	makeTestFollow(&follows, 15, 71, 94)
	makeTestFollow(&follows, 16, 71, 125)
	makeTestFollow(&follows, 17, 59, 494)
	makeTestFollow(&follows, 18, 71, 129)
	makeTestFollow(&follows, 19, 71, 54)
	makeTestFollow(&follows, 20, 11, 356)

	followers := map[int64]bool{
		94:  true,
		125: true,
		129: true,
		54:  true,
	}

	event.Handle(&follows, userNotifications)
	expectAllToReceiveEventNotification(t, userNotifications, event.Lift(), followers)
}
