package event_test

import (
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/kytnacode/inventure/event"
)

func TestBrokerShouldCallSubscriber(t *testing.T) {
	t.Parallel()

	var testEventTopic event.Topic = "test"

	expected := event.Event{
		Topic: testEventTopic,
	}

	br := event.NewBroker()

	go br.Run(t.Context())

	sub := br.Subscribe(expected.Topic, 1)

	time.Sleep(time.Millisecond * 10)

	br.Publish(expected.Topic, expected.Payload)

	select {
	case <-t.Context().Done():
		t.Fatalf("context canceled: %v", t.Context().Err())
	case ev := <-sub:
		if ev.Topic != expected.Topic {
			t.Errorf("expected topic to be '%v': got '%v'", expected.Topic, ev.Topic)
		}

		if ev.Payload != expected.Payload {
			t.Errorf("expected payload to be '%v': got '%v'", expected.Payload, ev.Payload)
		}
	}
}

func TestBrokerShouldUnsuscribe(t *testing.T) {
	t.Parallel()

	var testEventTopic event.Topic = "test"

	expected := event.Event{
		Topic: testEventTopic,
	}

	br := event.NewBroker()

	go br.Run(t.Context())

	sub := br.Subscribe(expected.Topic, 1)

	time.Sleep(time.Millisecond * 10)

	br.Unsubscribe(expected.Topic, sub)

	br.Publish(expected.Topic, expected.Payload)

	time.Sleep(time.Millisecond * 10)

	select {
	case <-t.Context().Done():
		t.Fatalf("context canceled: %v", t.Context().Err())
	case _, ok := <-sub:
		if ok {
			t.Error("expected event channel to be closed")
		}
	}
}

func TestBrokerShouldCallMultipleSubscribersConcurrently(t *testing.T) {
	t.Parallel()

	var testEventTopic event.Topic = "test"

	expected := event.Event{
		Topic: testEventTopic,
	}

	br := event.NewBroker()

	subN := 10
	pubN := 5
	eventsPerPub := 3

	subs := make([]<-chan event.Event, 0, subN)

	go br.Run(t.Context())

	for range subN {
		subs = append(subs, br.Subscribe(testEventTopic, 10))
	}

	time.Sleep(time.Millisecond * 10)

	for range pubN {
		go func() {
			for range eventsPerPub {
				//nolint:gosec // cryptographically secure RNG is not needed.
				time.Sleep(time.Duration(rand.Float64()) * time.Second / 8)

				br.Publish(expected.Topic, expected.Payload)
			}
		}()
	}

	expectedCount := eventsPerPub * pubN * subN

	var wg sync.WaitGroup

	wg.Add(expectedCount)

	done := make(chan struct{}, 1)

	go func() {
		wg.Wait()

		close(done)
	}()

	for _, sub := range subs {
		go func() {
			for range sub {
				wg.Done()
			}
		}()
	}

	select {
	case <-t.Context().Done():
		t.Fatalf("context canceled: %v", t.Context().Err())
	case <-done:
	}
}
