// Package event contains an in-memory event broker implementation.
package event

import (
	"context"
	"fmt"
)

// Topic is the an event topic that publishers can send and subscribers expect.
type Topic string

// Event contains event's data.
type Event struct {
	// Topic is event's topic.
	Topic Topic

	// Payload contains event specific payload.
	Payload any
}

type subscriber struct {
	topic Topic
	recv  chan Event
}

type unsubscribe struct {
	topic Topic
	recv  <-chan Event
}

// Broker is an in-memory event broker.
type Broker struct {
	events      <-chan Event
	sub         chan subscriber
	unsub       chan unsubscribe
	subscribers map[Topic][]chan Event
}

// NewBroker creates a new in-memory event broker.
func NewBroker(events <-chan Event) *Broker {
	return &Broker{
		subscribers: make(map[Topic][]chan Event, 10),
		sub:         make(chan subscriber, 2),
		unsub:       make(chan unsubscribe, 2),
		events:      events,
	}
}

// Subscribe return a channel that receives events from specified topic.
func (b *Broker) Subscribe(topic Topic, chanLen int) <-chan Event {
	if chanLen == 0 {
		chanLen = 3
	}

	recv := make(chan Event, chanLen)

	b.sub <- subscriber{
		topic: topic,
		recv:  recv,
	}

	return recv
}

// Unsubscribe unsubscribes the given channel from a topic if it is subscribed. Channel will be
// closed.
func (b *Broker) Unsubscribe(topic Topic, recv <-chan Event) {
	b.unsub <- unsubscribe{
		topic: topic,
		recv:  recv,
	}
}

// Run handles event broker loop.
func (b *Broker) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer b.close()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-b.events:
			if !ok {
				return
			}

			b.handleEvent(ev)
		case sub := <-b.sub:
			_, ok := b.subscribers[sub.topic]
			if !ok {
				b.subscribers[sub.topic] = make([]chan Event, 0, 8)
			}

			b.subscribers[sub.topic] = append(b.subscribers[sub.topic], sub.recv)
		case unsub := <-b.unsub:
			fmt.Println("abc")
			b.handleUnsub(unsub.topic, unsub.recv)
		}
	}
}

func (b *Broker) handleEvent(ev Event) {
	for _, sub := range b.subscribers[ev.Topic] {
		sub <- ev
	}
}

func (b *Broker) handleUnsub(topic Topic, recv <-chan Event) {
	subscribers := b.subscribers[topic]

	for i, sub := range subscribers {
		if sub == recv {
			close(sub)

			subscribers[i] = subscribers[len(subscribers)-1]
			subscribers = subscribers[:len(subscribers)-1]

			b.subscribers[topic] = subscribers

			break
		}
	}
}

func (b *Broker) close() {
	for _, topic := range b.subscribers {
		for _, sub := range topic {
			close(sub)
		}
	}
}
