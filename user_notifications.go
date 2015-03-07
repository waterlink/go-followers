package main

import (
	"fmt"
	"github.com/waterlink/goactor"
	"log"
)

type UserNotifications struct {
	goactor.Actor
	clients UserClients
}

type Notification struct {
	Event     EventInterface
	UserId    int64
	Broadcast bool
}

func (this UserNotifications) Act(message goactor.Any) {
	if notification, ok := message.(Notification); ok {
		if notification.Broadcast {

			for userId, client := range this.clients.Clients {
				if client != nil {
					log.Printf("Send notification %s to user %d\n\n", notification.Event, userId)
					fmt.Fprintf(client, "%s\r\n", notification.Event)
				}
			}

		} else {

			client := this.clients.Clients[notification.UserId]
			if client != nil {
				log.Printf("Send notification %s to user %d\n\n", notification.Event, notification.UserId)
				fmt.Fprintf(client, "%s\r\n", notification.Event)
			}

		}
	}
}
