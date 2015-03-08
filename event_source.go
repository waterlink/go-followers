package main

import (
	"github.com/waterlink/goactor"
	"io"
)

type Connection interface {
	io.Reader
	io.Closer
}

type EventSource struct {
	goactor.Actor
	connection        Connection
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
