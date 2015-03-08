package main

import (
	"net"
)

type ClientMap map[int64]*net.Conn

func NewClients() ClientMap {
	return make(ClientMap)
}
