package main

import (
	"github.com/waterlink/goactor"
)

type UserNotifications struct {
	goactor.Actor
	clients *UserClients
}

type Notification struct {
	Event     EventInterface
	UserId    int64
	Broadcast bool
}

func sendTo(client *Client, notification *Notification) {
	if client != nil {
		client.Send(&notification.Event)
	}
}

func (this *UserNotifications) Act(message goactor.Any) {
	if notification, ok := message.(*Notification); ok {
		if notification.Broadcast {

			for _, client := range this.clients.Clients {
				sendTo(client, notification)
			}

		} else {

			sendTo(this.clients.Clients[notification.UserId], notification)

		}
	}
}
