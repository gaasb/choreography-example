package domain

import "choreography/internal/services/order/events"

type Status string

func NewStatus() Status {
	return Status(events.OrderStatusPending)
}

func (s *Status) Accept() {
	*s = Status(events.OrderStatusAccepted)
}
func (s *Status) Reject() {
	*s = Status(events.OrderStatusRejected)
}
func (s Status) IsPending() bool {
	return s == Status(events.OrderStatusPending)
}
func (s Status) IsAccepted() bool {
	return s == Status(events.OrderStatusAccepted)
}
func (s Status) IsRejected() bool {
	return s == Status(events.OrderStatusRejected)
}
func (s Status) String() string {
	return string(s)
}
