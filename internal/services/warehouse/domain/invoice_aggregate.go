package domain

import (
	"choreography/internal/services/warehouse/events"
	"choreography/tools"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Invoice struct {
	tools.AggregateRoot
	aggregateId string
	orderId     string
	customerId  string
	items       Line
	totalSum    Money
	status      Status
}

func NewInvoice(orderId string, customerId string, items ...*Item) (invoice Invoice) {
	invoice.orderId = orderId
	invoice.customerId = customerId
	invoice.aggregateId = uuid.NewString()
	invoice.status = NewStatus()
	invoice.items = NewItemLine()

	for _, item := range items {
		invoice.items[item.itemId] = item.qty.ToInt()
		invoice.totalSum.CalculateSum(item.price.Multiply(item.qty.ToInt()))
	}

	invoice.WithEvent(events.InvoiceCreated{})
	return
}

func (o *Invoice) GetId() string {
	return o.aggregateId
}
func (o *Invoice) GetStatus() string {
	return o.status.String()
}
func (o *Invoice) Accept() {
	o.status.Accept()
}
func (o *Invoice) Reject() {
	o.status.Reject()
}
func (o *Invoice) Update(ctx context.Context, repository WarehouseRepository) error {
	if err := repository.UpdateInvoice(ctx, *o); err != nil {
		return err
	}
	return nil
}

func (o *Invoice) RestoreItems(ctx context.Context, repository WarehouseRepository) error {
	var errOut error
	if err := o.Update(ctx, repository); err != nil {
		return err
	}
	for itemId, qty := range o.items {
		product, err := repository.GetProductBy(ctx, itemId)
		if err != nil {
			errOut = fmt.Errorf("error on get product: %w", err)
			continue
		}
		product.IncreaseBy(Quantity(qty))
		if err := repository.UpdateProduct(ctx, *product); err != nil {
			errOut = fmt.Errorf("error on update product: %w", err)
		}
	}
	return errOut
}

func (i *Invoice) Process(ctx context.Context, dispatcher tools.EventDispatcher, storage tools.Storager) {

	serviceName := ctx.Value("current_service").(string)
	eventId := ctx.Value("event_id").(string)

	for _, event := range i.GetEvents() {
		message := tools.EventMessage{
			Header:      tools.NewMessageHeader(event, event.ServiceTarget(), serviceName),
			AggregateID: i.GetId(),
			State:       i.GetStatus(),
			Data: map[string]any{
				"invoice_id":  i.GetId(),
				"customer_id": i.customerId,
				"order_id":    i.orderId,
				"total_sum":   i.totalSum.ToInt(),
			},
		}
		message.SetEventID(eventId)

		message.Save(storage)
		dispatcher.Produce(ctx, message)
	}
}
