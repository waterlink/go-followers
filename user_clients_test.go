package main

import (
	"github.com/waterlink/goactor"
	"strings"
	"testing"
)

func TestUserConnectionReceive(t *testing.T) {
	userClients := &UserClients{
		Actor:   goactor.NewActor(),
		Clients: NewClients(),
	}
	clients := &userClients.Clients

	userA := FakeClientConnection{strings.NewReader("79\r\n")}
	userB := FakeClientConnection{strings.NewReader("936\r\n")}
	broken := FakeClientConnection{strings.NewReader("noise")}

	userAConnection := ClientConnection(userA)
	userBConnection := ClientConnection(userB)
	brokenConnection := ClientConnection(broken)

	userClients.Act(&userAConnection)
	userClients.Act(&brokenConnection)
	userClients.Act(&userBConnection)

	expectToBeFakeClientConnection(t, (*clients)[79], userA)
	expectToBeFakeClientConnection(t, (*clients)[936], userB)
	expectIntToBeEqual(t, int64(len(*clients)), 2)
}
