package domain

import (
	"choreography/internal/services/order/events"
	"choreography/tools"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Order struct {
	tools.AggregateRoot
	aggregateId string
	customerId  string
	line        Line
	status      Status
}

func NewOrder(userId string, items Line) (*Order, error) {
	if len(items) <= 0 {
		return nil, errors.New("line is empty")
	}

	if err := items.Verify(); err != nil {
		return nil, fmt.Errorf("failed on create new order: %w", err)
	}

	order := &Order{
		aggregateId: uuid.NewString(),
		status:      NewStatus(),
		customerId:  userId,
		line:        items,
	}
	order.WithEvent(events.OrderCreated{})
	return order, nil
}

func (o *Order) AddItem(itemId int, qty int) error {
	return o.line.AddItem(itemId, qty)
}

func (o *Order) GetStatus() string {
	return string(o.status)
}

func (o *Order) GetId() string {
	return o.aggregateId
}

func (o *Order) Accept() {
	o.status.Accept()
}
func (o *Order) Reject() {
	o.status.Reject()
}

func (o *Order) Update(ctx context.Context, repository OrderRepository) error {
	return repository.Update(ctx, o)
}

func (o *Order) Process(ctx context.Context, dispatcher tools.EventDispatcher, storage tools.Storager) {

	serviceName := ctx.Value("current_service").(string)
	// eventId := ctx.Value("event_id").(string)
	for _, event := range o.GetEvents() {
		message := tools.EventMessage{
			Header:      tools.NewMessageHeader(event, event.ServiceTarget(), serviceName),
			AggregateID: o.GetId(),
			State:       o.status.String(),
			Data: map[string]any{
				"order_id":    o.GetId(),
				"customer_id": o.customerId,
				"order_line":  o.line,
			},
		}
		message.NewEventID()
		message.Save(storage)
		dispatcher.Produce(ctx, message)
	}
}
