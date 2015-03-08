package main

import (
	"github.com/waterlink/goactor"
)

type FollowMapValue map[int64]bool
type FollowMap map[int64]FollowMapValue
type EventMap map[int64]*EventInterface

type UserRelationships struct {
	goactor.Actor
	follows            FollowMap
	lastSeenSequenceId int64
	userNotifications  *UserNotifications
	bufferedEvents     EventMap
}

func NewFollowMap() FollowMap {
	return make(FollowMap)
}

func NewFollowMapValue() FollowMapValue {
	return make(FollowMapValue)
}

func NewEventMap() EventMap {
	return make(EventMap)
}

func (this *UserRelationships) Act(message goactor.Any) {
	if event, ok := message.(*EventInterface); ok {

		this.bufferedEvents[(*event).getSequenceId()] = event

		for eventFromBuffer, ok := this.bufferedEvents[this.lastSeenSequenceId+1]; ok; eventFromBuffer, ok = this.bufferedEvents[this.lastSeenSequenceId+1] {

			this.lastSeenSequenceId += 1
			delete(this.bufferedEvents, this.lastSeenSequenceId)
			(*eventFromBuffer).Handle(&this.follows, this.userNotifications)

		}
	}
}
