package main

import (
	"fmt"
	"github.com/waterlink/goactor"
)

type UserClients struct {
	goactor.Actor
	Clients ClientMap
}

func (this *UserClients) Act(message goactor.Any) {
	if connection, ok := message.(*ClientConnection); ok {
		userId := int64(0)
		_, error := fmt.Fscanf(*connection, "%d\r\n", &userId)
		if error != nil {
			return
		}

		this.Clients[userId] = connection
	}
}
