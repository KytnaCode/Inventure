package auth

import (
	"github.com/kytnacode/inventure/event"
	"github.com/kytnacode/inventure/internal/user"
)

// TopicUserCreated is the event topic for user creations with payload of type [EventUserCreated].
const TopicUserCreated event.Topic = "user.created"

// EventUserCreated is the event data for user creation events.
type EventUserCreated struct {
	User *user.User
}

// GetUserCreatedEvent gets user created event data from payload.
func GetUserCreatedEvent(payload any) (EventUserCreated, error) {
	return event.GetEventData[EventUserCreated](payload)
}
