package events

import "choreography/tools"

type PaymentCreated struct{}

func (o PaymentCreated) ServiceTarget() string {
	return tools.WarehouseService
}

type PaymentNotCreated struct{}

// type PaymentAccepted struct{}
