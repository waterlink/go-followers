package main

import (
	"fmt"
	"github.com/waterlink/goactor"
	"log"
	"net"
)

type Client struct {
	goactor.Actor
	userId     int64
	connection *net.Conn
}

type ClientMap map[int64]*Client

func (this *Client) Act(message goactor.Any) {
	if event, ok := message.(*EventInterface); ok {
		log.Printf("Send notification %s to user %d\n", *event, this.userId)
		fmt.Fprintf(*this.connection, "%s\r\n", *event)
	}
}
