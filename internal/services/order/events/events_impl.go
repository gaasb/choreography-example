package events

import "choreography/tools"

type OrderCreated struct{}

func (o OrderCreated) ServiceTarget() string {
	return tools.WarehouseService
}

// type OrderNotCreated struct{}
