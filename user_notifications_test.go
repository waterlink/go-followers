package main

import (
	"github.com/waterlink/goactor"
	"strings"
	"testing"
)

func TestReceiveNotification(t *testing.T) {
	userClients := &UserClients{
		Actor:   goactor.NewActor(),
		Clients: NewClients(),
	}
	userNotifications := &UserNotifications{
		Actor:   goactor.NewActor(),
		clients: userClients,
	}
	clients := &userClients.Clients

	userA := FakeClientConnection{Reader: strings.NewReader("")}
	userB := FakeClientConnection{Reader: strings.NewReader("")}

	userAConnection := ClientConnection(&userA)
	userBConnection := ClientConnection(&userB)

	(*clients)[79] = &userAConnection
	(*clients)[324] = &userBConnection

	event, _ := scanEvent(strings.NewReader("3497|F|55|79\r\n"))
	userNotifications.Act(&Notification{
		Event:  event,
		UserId: 79,
	})
	expectStringToBeEqual(t, userA.written, "3497|F|55|79\r\n")
	expectStringToBeEqual(t, userB.written, "")

	broadcastEvent, _ := scanEvent(strings.NewReader("3577|B\r\n"))
	userNotifications.Act(&Notification{
		Event:     broadcastEvent,
		Broadcast: true,
	})
	expectStringToBeEqual(t, userA.written, "3497|F|55|79\r\n3577|B\r\n")
	expectStringToBeEqual(t, userB.written, "3577|B\r\n")
}
