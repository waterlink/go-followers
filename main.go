package main

import (
	"fmt"
	"log"
	"net"
)

const bufferSize = 128

type Notification struct {
	Event     EventInterface
	UserId    int64
	Broadcast bool
}

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

	eventSourceDone := make(chan bool)
	userClientsDone := make(chan bool)

	userClientsInbox := make(chan Notification)
	relationshipsInbox := make(chan EventInterface)

	go handleSourceEvents(eventSourceListener, eventSourceDone, relationshipsInbox)
	go handleUserClients(userClientsListener, userClientsDone, userClientsInbox)
	go handleRelationships(relationshipsInbox, userClientsInbox)

	<-eventSourceDone
	<-userClientsDone
}

func handleSourceEvents(listener net.Listener, done chan bool, relationshipsInbox chan EventInterface) {
	for {
		connection, error := listener.Accept()
		if error != nil {
			log.Print(error)
			break
		}

		go handleEventSourceConnection(connection, relationshipsInbox)
	}
	done <- true
}

func handleUserClients(listener net.Listener, done chan bool, userClientsInbox chan Notification) {
	clients := make(map[int64]net.Conn)

	go func() {
		for {
			connection, error := listener.Accept()
			if error != nil {
				log.Print(error)
				break
			}
			defer connection.Close()

			go func() {
				userId := int64(0)
				_, error := fmt.Fscanf(connection, "%d\r\n", &userId)
				if error != nil {
					log.Print(error)
				}

				clients[userId] = connection
			}()
		}
	}()

	for {
		notification, ok := <-userClientsInbox
		if !ok {
			break
		}

		if notification.Broadcast {

			for userId, client := range clients {
				if client != nil {
					log.Printf("Send notification %s to user %d\n\n", notification.Event, userId)
					fmt.Fprintf(client, "%s\r\n", notification.Event)
				}
			}

		} else {

			client := clients[notification.UserId]
			if client != nil {
				log.Printf("Send notification %s to user %d\n\n", notification.Event, notification.UserId)
				fmt.Fprintf(client, "%s\r\n", notification.Event)
			}

		}

	}

	done <- true
}

func handleRelationships(relationshipsInbox chan EventInterface, userClientsInbox chan Notification) {
	follows := make(map[int64]map[int64]bool)
	lastSeenSequenceId := int64(0)

	for {
		event, ok := <-relationshipsInbox
		if !ok {
			break
		}

		if lastSeenSequenceId+1 == event.getSequenceId() {

			lastSeenSequenceId = event.getSequenceId()

			event.Handle(follows, userClientsInbox)

		} else {

			go func() {
				relationshipsInbox <- event
			}()

		}
	}
}

func handleEventSourceConnection(connection net.Conn, relationshipsInbox chan EventInterface) {
	defer connection.Close()

	for {
		event, error := scanEvent(connection)
		if error == nil {
			relationshipsInbox <- event
		}
	}
}
