package main

import (
	"github.com/waterlink/goactor"
	"log"
)

type FollowMapValue map[int64]bool
type FollowMap map[int64]FollowMapValue

type UserRelationships struct {
	goactor.Actor
	follows            FollowMap
	lastSeenSequenceId int64
	userNotifications  UserNotifications
}

func NewFollowMap() FollowMap {
	return make(FollowMap)
}

func NewFollowMapValue() FollowMapValue {
	return make(FollowMapValue)
}

func (this UserRelationships) Act(message goactor.Any) {
	if event, ok := message.(EventInterface); ok {
		log.Print("got event")
		log.Print(event)

		if this.lastSeenSequenceId+1 == event.getSequenceId() {

			this.lastSeenSequenceId = event.getSequenceId()
			event.Handle(this.follows, this.userNotifications)

		} else {

			goactor.Send(this, event)

		}
	}
}
