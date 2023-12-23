package domain

import "choreography/internal/services/warehouse/events"

type Status string

func NewStatus() Status {
	return Status(events.InvoiceStatusPending)
}

func (s *Status) Accept() {
	*s = Status(events.InvoiceStatusAccepted)
}
func (s *Status) Reject() {
	*s = Status(events.InvoiceStatusRejected)
}
func (s Status) IsPending() bool {
	return s == Status(events.InvoiceStatusPending)
}
func (s Status) IsAccepterd() bool {
	return s == Status(events.InvoiceStatusAccepted)
}
func (s Status) IsRejected() bool {
	return s == Status(events.InvoiceStatusRejected)
}
func (s Status) String() string {
	return string(s)
}
