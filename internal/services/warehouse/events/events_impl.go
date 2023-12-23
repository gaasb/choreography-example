package events

import "choreography/tools"

type InvoiceCreated struct{}

func (o InvoiceCreated) ServiceTarget() string {
	return tools.PaymentService
}

type InvoiceNotCreated struct{}
