package retail

import (
	"context"
	"log/slog"

	"github.com/kytnacode/inventure/event"
	"github.com/kytnacode/inventure/internal/auth"
	"github.com/kytnacode/inventure/logging"
)

const userCreatedHandlerBuffer = 100

// EventHandler contain retail related event subscribers.
type EventHandler struct {
	service *Service
	broker  *event.Broker
}

// NewEventHandler creates a new [EventHandler].
func NewEventHandler(service *Service, broker *event.Broker) *EventHandler {
	return &EventHandler{
		broker:  broker,
		service: service,
	}
}

func (h *EventHandler) handleUserCreated(ctx context.Context) {
	logger := logging.FromCtx(ctx)

	logger = logger.With(slog.String("event.topic", string(auth.TopicUserCreated)))

	sub := h.broker.Subscribe(auth.TopicUserCreated, userCreatedHandlerBuffer)

	for ev := range sub {
		data, err := auth.GetUserCreatedEvent(ev.Payload)
		if err != nil {
			logger.Error("invalid user created event payload type", logging.Error(err))

			continue
		}

		_, err = h.service.CreateDefaultTenant(ctx, *data.User)
		if err != nil {
			logger.Error("could not create user default retail", logging.Error(err))

			continue
		}
	}
}

// Handle register retail event handlers.
func (h *EventHandler) Handle(ctx context.Context) {
	logger := logging.FromCtx(ctx)

	logger = logger.With(slog.String("event.handler", "retail.EventHandler"))

	ctx = logging.WithLogger(ctx, logger)

	h.handleUserCreated(ctx)
}
