package domain

import (
	"choreography/internal/services/payment/events"
	"choreography/tools"
	"context"

	"github.com/google/uuid"
)

type Payment struct {
	tools.AggregateRoot
	aggregateId string
	orderId     string
	customerId  string
	invoiceId   string
	totalAmount Money
	status      Status
}

func NewPayment(orderId string, customerId string, invoiceId string, totalSum int) (payment Payment) {
	payment.aggregateId = uuid.NewString()
	payment.orderId = orderId
	payment.customerId = customerId
	payment.status = NewStatus()
	payment.invoiceId = invoiceId
	payment.totalAmount = NewMoney(totalSum)

	payment.WithEvent(events.PaymentCreated{})
	return
}
func (p *Payment) ProcessPay(ctx context.Context, repository PaymentRepository) error {
	var err error
	var userBalance *Money
	if userBalance, err = repository.GetBalance(ctx, p.customerId); err != nil {
		return err
	}
	if err = userBalance.DecreaseBy(p.totalAmount); err != nil {
		return err
	}
	repository.UpdateBalance(ctx, p.customerId, *userBalance)
	return nil
}

func (p *Payment) RefundMoney(ctx context.Context, repository PaymentRepository) error {
	return repository.AddMoneyToBalance(ctx, p.customerId, p.totalAmount)
}

func (p *Payment) GetId() string {
	return p.aggregateId
}
func (p *Payment) GetStatus() string {
	return p.status.String()
}
func (p *Payment) Accept() {
	p.status.Accept()
}
func (p *Payment) Reject() {
	p.status.Reject()
}
func (p *Payment) Update(ctx context.Context, repository PaymentRepository) error {
	if err := repository.UpdatePayment(ctx, *p); err != nil {
		return err
	}
	return nil
}

func (p *Payment) Process(ctx context.Context, dispatcher tools.EventDispatcher, storage tools.Storager) {

	serviceName := ctx.Value("current_service").(string)
	eventId := ctx.Value("event_id").(string)

	for _, event := range p.GetEvents() {
		message := tools.EventMessage{
			Header:      tools.NewMessageHeader(event, event.ServiceTarget(), serviceName),
			AggregateID: p.GetId(),
			State:       p.GetStatus(),
			Data: map[string]any{
				"payment_id":  p.GetId(),
				"customer_id": p.customerId,
				"order_id":    p.orderId,
				"invoice_id":  p.invoiceId,

				"linked_service": tools.OrderService,
			},
		}
		message.SetEventID(eventId)

		message.Save(storage)
		dispatcher.Produce(ctx, message)
	}
}
