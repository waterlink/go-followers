package main

import (
	"io"
)

type ClientConnection interface {
	io.Reader
	io.Writer
	io.Closer
}

type ClientMap map[int64]*ClientConnection

func NewClients() ClientMap {
	return make(ClientMap)
}
