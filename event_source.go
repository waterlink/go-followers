package main

import (
	"github.com/waterlink/goactor"
	"net"
)

type EventSource struct {
	goactor.Actor
	connection        net.Conn
	userRelationships UserRelationships
}

func (this EventSource) Act(message goactor.Any) {
	if event, error := scanEvent(this.connection); error == nil {
		goactor.Send(this.userRelationships, event)
		goactor.Send(this, message)
	} else {
		this.connection.Close()
		close(this.Inbox())
	}
}
