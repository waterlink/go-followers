package main

import (
	"github.com/waterlink/goactor"
	"net"
)

type EventSource struct {
	goactor.Actor
	connection        net.Conn
	userRelationships *UserRelationships
}

func (this *EventSource) Act(message goactor.Any) {
	if event, error := scanEvent(this.connection); error == nil {

		this.userRelationships.Send(&event)
		this.Send(message)

	} else {

		this.connection.Close()
		this.Die()

	}
}
