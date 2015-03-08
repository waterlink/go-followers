package main

import (
	"fmt"
	"github.com/waterlink/goactor"
	"net"
)

type UserClients struct {
	goactor.Actor
	Clients ClientMap
}

func (this *UserClients) Act(message goactor.Any) {
	if connection, ok := message.(*net.Conn); ok {
		userId := int64(0)
		_, error := fmt.Fscanf(*connection, "%d\r\n", &userId)
		if error != nil {
			return
		}

		this.Clients[userId] = connection
	}
}
