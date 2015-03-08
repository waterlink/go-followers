package main

import (
	"github.com/waterlink/goactor"
	//"log"
	//"runtime"
)

type FollowMapValue map[int64]bool
type FollowMap map[int64]FollowMapValue

type UserRelationships struct {
	goactor.Actor
	follows            FollowMap
	lastSeenSequenceId int64
	userNotifications  *UserNotifications
}

func NewFollowMap() FollowMap {
	return make(FollowMap)
}

func NewFollowMapValue() FollowMapValue {
	return make(FollowMapValue)
}

func (this *UserRelationships) Act(message goactor.Any) {
	if event, ok := message.(*EventInterface); ok {

		if eventId := (*event).getSequenceId(); this.lastSeenSequenceId+1 == eventId {

			this.lastSeenSequenceId = eventId
			(*event).Handle(&this.follows, this.userNotifications)

		} else {

			//log.Printf("goroutines: %d\n", runtime.NumGoroutine())
			this.Send(event)

		}
	}
}
