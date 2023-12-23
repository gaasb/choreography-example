package tools

import (
	"errors"
	"log/slog"
)

type Storager interface {
	saveEvent(m EventMessage)
	getSavedEventBy(eventID string) (*EventMessage, error)
}

type Storage map[string][]EventMessage

type EventStorage struct {
	events Storage
}

func NewEventStorage() *EventStorage {
	return &EventStorage{
		events: make(map[string][]EventMessage),
	}
}

func (s *EventStorage) getSavedEventBy(eventID string) (*EventMessage, error) {
	if v, ok := s.events[eventID]; ok {
		return &v[0], nil
	}
	return nil, errors.New("event missing")
}

func (s *EventStorage) saveEvent(m EventMessage) {
	id := m.GetEventID()
	s.events[id] = append(s.events[id], m)
	slog.Info("event saved", slog.String("event_id", id))
}
