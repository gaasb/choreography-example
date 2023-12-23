package tools

import (
	"context"
	"log/slog"
)

type EventDispatcherLocal struct {
	localBroker *LocalBroker
	messages    chan EventMessage
}

func NewEventDispatcherLocal(broker *LocalBroker, serviceName string) *EventDispatcherLocal {
	dispatcher := &EventDispatcherLocal{
		messages:    make(chan EventMessage),
		localBroker: broker,
	}

	broker.RegistreService(serviceName, dispatcher)
	return dispatcher
}

func (d *EventDispatcherLocal) Handle(ctx context.Context, operation ...EventOp) {
	currentService := ctx.Value("current_service").(string)
	select {
	case <-ctx.Done():
		return
	case message := <-d.messages:
		slog.Info(
			"event handled",
			slog.String("event_id", message.GetEventID()),
			slog.String("event_type", message.GetEventName()),
			slog.String("from", message.GetProducer()),
			slog.String("to", message.GetReceiver()),
			slog.String("service", currentService),
		)
		for _, run := range operation {
			run(message)
		}
		// return
	}
}

func (d *EventDispatcherLocal) Produce(ctx context.Context, message EventMessage) {
	currentService := ctx.Value("current_service").(string)
	d.localBroker.SendEvent(message)
	slog.Info(
		"event produced",
		slog.String("event_id", message.GetEventID()),
		slog.String("event_type", message.GetEventName()),
		slog.String("from", message.GetProducer()),
		slog.String("to", message.GetReceiver()),
		slog.String("service", currentService),
	)
}

func (d *EventDispatcherLocal) MessagePipe(msg EventMessage) {
	d.messages <- msg
}
