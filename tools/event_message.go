package tools

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/google/uuid"
)

type EventOp func(EventMessage)

type EventMessageHeader struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Producer  string `json:"event_producer"`
	Receiver  string `json:"event_receiver"`
}

type EventMessage struct {
	Header      EventMessageHeader `json:"message_header"`
	AggregateID string             `json:"aggregate_id"`
	Data        map[string]any     `json:"data"`
	State       string             `json:"state"`
}

type MessageProcessor interface {
	Process(ctx context.Context, dispatcher EventDispatcher, storage Storager)
}

func NewEventMessage() EventMessage { return EventMessage{} }

func NewMessageHeader(event RootHelper, serviceTarget, currentService string) EventMessageHeader {
	header := EventMessageHeader{
		Producer: currentService,
		Receiver: serviceTarget,
	}
	header.SetEvent(event)
	return header
}

func (s *EventMessageHeader) SetEvent(event any) {
	eventName := reflect.TypeOf(event).Name()
	s.EventType = eventName
}

func (m EventMessage) Save(s Storager) {
	s.saveEvent(m)
}
func (m *EventMessage) SetState(state string) {
	m.State = state
}

func (m *EventMessage) Restore(s Storager) error {
	if message, err := s.getSavedEventBy(m.Header.EventID); err != nil {
		return fmt.Errorf("fail to restore message: %w", err)
	} else {
		*m = *message
		slog.Info("order restored", slog.String("order_id", m.AggregateID))
		return nil
	}
}

func (s *EventMessageHeader) Swap() {
	s.Producer, s.Receiver = s.Receiver, s.Producer
}
func (s *EventMessage) NewEventID() {
	s.Header.EventID = uuid.NewString()
}
func (s *EventMessage) SetEventID(id string) {
	s.Header.EventID = id
}
func (s *EventMessage) GetEventID() string {
	return s.Header.EventID
}
func (e *EventMessage) GetAggregateId() string {
	return e.AggregateID
}
func (e *EventMessage) GetProducer() string {
	return e.Header.Producer
}
func (e *EventMessage) GetEventName() string {
	return e.Header.EventType
}

func (e *EventMessage) EventIsEqual(to any) bool {
	eventName := reflect.TypeOf(to).Name()
	return e.Header.EventID == eventName
}
func (e *EventMessage) ReceiverIsEqual(to string) bool {
	return e.Header.Receiver == to
}
func (e *EventMessage) ProducerIsEqual(to string) bool {
	return e.Header.Producer == to
}
func (e *EventMessage) StatusIsEqual(to string) bool {
	return e.State == to
}
func (e *EventMessage) GetReceiver() string {
	return e.Header.Receiver
}
func (e *EventMessage) SetReceiver(newReceiver string) {
	e.Header.Receiver = newReceiver
}
