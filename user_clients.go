package main

import (
	"fmt"
	"github.com/waterlink/goactor"
	"log"
	"net"
)

type UserClients struct {
	goactor.Actor
	Clients ClientMap
}

func NewClients() ClientMap {
	return make(ClientMap)
}

func (this *UserClients) Act(message goactor.Any) {
	if connection, ok := message.(*net.Conn); ok {
		userId := int64(0)
		_, error := fmt.Fscanf(*connection, "%d\r\n", &userId)
		if error != nil {
			return
		}

		log.Printf("user connected: %d\n", userId)

		client := &Client{
			Actor:      goactor.NewActor(),
			userId:     userId,
			connection: connection,
		}
		this.Clients[userId] = client
		goactor.Go(client, fmt.Sprintf("client %s", userId))
	}
}
