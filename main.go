package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

const bufferSize = 128

type Event struct {
	SequenceId string
	Type       string
	FromUserId string
	ToUserId   string
}

type Notification struct {
	Event     Event
	UserId    string
	Broadcast bool
}

func (event Event) String() string {
	result := []rune("")

	result = append(result, []rune(event.SequenceId)...)

	result = append(result, '|')
	result = append(result, []rune(event.Type)...)

	if event.FromUserId != "" {
		result = append(result, '|')
		result = append(result, []rune(event.FromUserId)...)
	}

	if event.ToUserId != "" {
		result = append(result, '|')
		result = append(result, []rune(event.ToUserId)...)
	}

	return string(result)
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
	relationshipsInbox := make(chan Event)

	go handleSourceEvents(eventSourceListener, eventSourceDone, relationshipsInbox)
	go handleUserClients(userClientsListener, userClientsDone, userClientsInbox)
	go handleRelationships(relationshipsInbox, userClientsInbox)

	<-eventSourceDone
	<-userClientsDone
}

func handleSourceEvents(listener net.Listener, done chan bool, relationshipsInbox chan Event) {
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
	clients := make(map[string]net.Conn)

	go func() {
		for {
			connection, error := listener.Accept()
			if error != nil {
				log.Print(error)
				break
			}
			defer connection.Close()

			go func() {
				userId := ""
				_, error := fmt.Fscanf(connection, "%s\r\n", &userId)
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
					log.Printf("Send notification %s to user %s\n\n", notification.Event, userId)
					fmt.Fprintf(client, "%s\r\n", notification.Event)
				}
			}

		} else {

			client := clients[notification.UserId]
			if client != nil {
				log.Printf("Send notification %s to user %s\n\n", notification.Event, notification.UserId)
				fmt.Fprintf(client, "%s\r\n", notification.Event)
			}

		}

	}

	done <- true
}

func handleRelationships(relationshipsInbox chan Event, userClientsInbox chan Notification) {
	follows := make(map[string]map[string]bool)
	lastSeenSequenceId := 0

	for {
		event, ok := <-relationshipsInbox
		if !ok {
			break
		}

		sequenceId, error := strconv.Atoi(event.SequenceId)
		if error != nil {
			log.Print(error)
			continue
		}

		if lastSeenSequenceId+1 == sequenceId {

			lastSeenSequenceId = sequenceId

			switch event.Type {

			case "F":
				if follows[event.ToUserId] == nil {
					follows[event.ToUserId] = make(map[string]bool)
				}
				follows[event.ToUserId][event.FromUserId] = true

				userClientsInbox <- Notification{
					Event:  event,
					UserId: event.ToUserId,
				}

			case "U":
				if follows[event.ToUserId] == nil {
					follows[event.ToUserId] = make(map[string]bool)
				}
				delete(follows[event.ToUserId], event.FromUserId)

			case "B":
				userClientsInbox <- Notification{
					Event:     event,
					Broadcast: true,
				}

			case "P":
				userClientsInbox <- Notification{
					Event:  event,
					UserId: event.ToUserId,
				}

			case "S":
				for followerId := range follows[event.FromUserId] {
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

func handleEventSourceConnection(connection net.Conn, relationshipsInbox chan Event) {
	defer connection.Close()

	buffer := make([]byte, bufferSize)

	eventString := []rune("")

	state := "waitSequenceId"

	eventSequenceId := []rune("")
	eventType := []rune("")
	eventFromUserId := []rune("")
	eventToUserId := []rune("")

	event := Event{}

	for {
		count, error := connection.Read(buffer)
		if error != nil {
			log.Print(error)
			break
		}

		for _, value := range string(buffer[:count]) {
			switch value {

			case '\n':
				switch state {
				case "waitType":
					event.Type = string(eventType)
					eventType = []rune("")
				case "waitFromUserId":
					event.FromUserId = string(eventFromUserId)
					eventFromUserId = []rune("")
				case "waitToUserId":
					event.ToUserId = string(eventToUserId)
					eventToUserId = []rune("")
				}

				relationshipsInbox <- event
				event = Event{}
				eventString = []rune("")
				state = "waitSequenceId"

			case '\r':

			default:

				eventString = append(eventString, value)

				switch state {

				case "waitSequenceId":
					if value == '|' {
						event.SequenceId = string(eventSequenceId)
						eventSequenceId = []rune("")
						state = "waitType"
					} else {
						eventSequenceId = append(eventSequenceId, value)
					}

				case "waitType":
					if value == '|' {
						event.Type = string(eventType)
						eventType = []rune("")
						state = "waitFromUserId"
					} else {
						eventType = append(eventType, value)
					}

				case "waitFromUserId":
					if value == '|' {
						event.FromUserId = string(eventFromUserId)
						eventFromUserId = []rune("")
						state = "waitToUserId"
					} else {
						eventFromUserId = append(eventFromUserId, value)
					}

				case "waitToUserId":
					eventToUserId = append(eventToUserId, value)

				default:
				}
			}
		}
	}
}
