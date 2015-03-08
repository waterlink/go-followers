package main

import (
	"fmt"
	"github.com/waterlink/goactor"
	"log"
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

func sendTo(userId int64, client *ClientConnection, notification *Notification) {
	if client != nil {
		log.Printf("send %s to %d\n", notification.Event, userId)
		fmt.Fprintf(*client, "%s\r\n", notification.Event)
	}
}

func (this *UserNotifications) Act(message goactor.Any) {
	if notification, ok := message.(*Notification); ok {
		if notification.Broadcast {

			for userId, client := range this.clients.Clients {
				sendTo(userId, client, notification)
			}

		} else {

			sendTo(notification.UserId, this.clients.Clients[notification.UserId], notification)

		}
	}
}
