package domain

import "choreography/internal/services/payment/events"

type Status string

func NewStatus() Status {
	return Status(events.PaymentStatusPending)
}

func (s *Status) Accept() {
	*s = Status(events.PaymentStatusAccepted)
}
func (s *Status) Reject() {
	*s = Status(events.PaymentStatusRejected)
}
func (s Status) IsPending() bool {
	return s == Status(events.PaymentStatusPending)
}
func (s Status) IsAccepterd() bool {
	return s == Status(events.PaymentStatusAccepted)
}
func (s Status) IsRejected() bool {
	return s == Status(events.PaymentStatusRejected)
}
func (s Status) String() string {
	return string(s)
}
