package main

import (
	"github.com/waterlink/goactor"
	"log"
	"net"
)

const bufferSize = 128

func main() {
	log.Print("Starting the server")

	eventSourceListener, error := net.Listen("tcp", ":9090")
	if error != nil {
		log.Fatal(error)
		return
	}
	defer eventSourceListener.Close()

	userClientsListener, error := net.Listen("tcp", ":9099")
	if error != nil {
		log.Fatal(error)
		return
	}
	defer userClientsListener.Close()

	userClients := &UserClients{
		Actor:   goactor.NewActor(),
		Clients: NewClients(),
	}
	goactor.Go(userClients, "user clients")

	userNotifications := &UserNotifications{
		Actor:   goactor.NewActor(),
		clients: userClients,
	}
	goactor.Go(userNotifications, "user notifications")

	userRelationships := &UserRelationships{
		Actor:              goactor.NewActor(),
		follows:            NewFollowMap(),
		lastSeenSequenceId: int64(0),
		userNotifications:  userNotifications,
		bufferedEvents:     NewEventMap(),
	}
	goactor.Go(userRelationships, "user relationships")

	eventSourceDone := make(chan bool)

	go handleSourceEvents(eventSourceListener, eventSourceDone, userRelationships)
	go handleUserClients(userClientsListener, userClients)

	<-eventSourceDone
}

func handleSourceEvents(listener net.Listener, done chan bool, userRelationships *UserRelationships) {
	for {
		connection, error := listener.Accept()
		if error != nil {
			log.Print(error)
			break
		}

		eventSource := &EventSource{
			Actor:             goactor.NewActor(),
			connection:        connection,
			userRelationships: userRelationships,
		}
		goactor.Go(eventSource, "event source")
		eventSource.Send(true)
	}
	done <- true
}

func handleUserClients(listener net.Listener, userClients *UserClients) {
	for {
		connection, error := listener.Accept()
		if error != nil {
			log.Print(error)
			break
		}
		defer connection.Close()

		clientConnection, _ := connection.(ClientConnection)
		userClients.Send(&clientConnection)
	}
}
