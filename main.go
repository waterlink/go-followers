package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

const bufferSize = 128

type EventInterface interface {
	String() string
	scanRest(io.Reader) (EventInterface, error)

	getSequenceId() int64
	getType() string
	getFromUserId() int64
	getToUserId() int64
}

type Event struct {
	SequenceId int64
	Type       string
	FromUserId int64
	ToUserId   int64
}

type Follow struct {
	Event
}

type Unfollow struct {
	Event
}

type Broadcast struct {
	Event
}

type PrivateMsg struct {
	Event
}

type StatusUpdate struct {
	Event
}

type Notification struct {
	Event     EventInterface
	UserId    int64
	Broadcast bool
}

func (event Event) getSequenceId() int64 { return event.SequenceId }
func (event Event) getType() string      { return event.Type }
func (event Event) getFromUserId() int64 { return event.FromUserId }
func (event Event) getToUserId() int64   { return event.ToUserId }

func (event Event) Lift() EventInterface {
	switch event.Type {
	case "F":
		return Follow{event}
	case "U":
		return Unfollow{event}
	case "B":
		return Broadcast{event}
	case "P":
		return PrivateMsg{event}
	case "S":
		return StatusUpdate{event}
	}
	return event
}

func (event Event) String() string {
	return fmt.Sprintf(
		"%d|%s|%d|%d",
		event.SequenceId,
		event.Type,
		event.FromUserId,
		event.ToUserId,
	)
}

func (event Broadcast) String() string {
	return fmt.Sprintf(
		"%d|%s",
		event.SequenceId,
		event.Type,
	)
}

func (event StatusUpdate) String() string {
	return fmt.Sprintf(
		"%d|%s|%d",
		event.SequenceId,
		event.Type,
		event.FromUserId,
	)
}

func scanEvent(reader io.Reader) (EventInterface, error) {
	var eventSequenceId int64
	var eventType string

	_, error := fmt.Fscanf(reader, "%d|%1s", &eventSequenceId, &eventType)
	if error != nil {
		return Event{}, nil
	}

	event := Event{
		SequenceId: eventSequenceId,
		Type:       eventType,
	}

	return event.Lift().scanRest(reader)
}

func (event Event) scanRest(reader io.Reader) (EventInterface, error) {
	_, error := fmt.Fscanf(reader, "|%d|%d\r\n", &event.FromUserId, &event.ToUserId)
	if error != nil {
		return Event{}, nil
	}

	return event, nil
}

func (event Broadcast) scanRest(reader io.Reader) (EventInterface, error) {
	return event, nil
}

func (event StatusUpdate) scanRest(reader io.Reader) (EventInterface, error) {
	_, error := fmt.Fscanf(reader, "|%d\r\n", &event.FromUserId)
	if error != nil {
		return Event{}, nil
	}

	return event, nil
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

			switch event.getType() {

			case "F":
				if follows[event.getToUserId()] == nil {
					follows[event.getToUserId()] = make(map[int64]bool)
				}
				follows[event.getToUserId()][event.getFromUserId()] = true

				userClientsInbox <- Notification{
					Event:  event,
					UserId: event.getToUserId(),
				}

			case "U":
				if follows[event.getToUserId()] == nil {
					follows[event.getToUserId()] = make(map[int64]bool)
				}
				delete(follows[event.getToUserId()], event.getFromUserId())

			case "B":
				userClientsInbox <- Notification{
					Event:     event,
					Broadcast: true,
				}

			case "P":
				userClientsInbox <- Notification{
					Event:  event,
					UserId: event.getToUserId(),
				}

			case "S":
				for followerId := range follows[event.getFromUserId()] {
					userClientsInbox <- Notification{
						Event:  event,
						UserId: followerId,
					}
				}
			}

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
