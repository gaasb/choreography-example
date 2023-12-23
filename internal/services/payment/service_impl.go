package service

import (
	"choreography/internal/services/payment/domain"

	"choreography/tools"
	"context"
	"fmt"
	"log/slog"
)

type PaymentServiceImpl struct {
	db           domain.PaymentRepository
	dispatcher   tools.EventDispatcher
	eventStorage *tools.EventStorage
	logger       *slog.Logger
}

type PaymentResponse struct {
	PaymentId string `json:"payment_id"`
}
type PaymentRequest struct {
	InvoiceId   string `json:"invoice_id"`
	OrderId     string `json:"order_id"`
	CustomerId  string `json:"customer_id"`
	TotalAmount int    `json:"total_amount"`
}

type DepositRequest struct {
	CustomerId string `json:"customer_id"`
	Value      int    `json:"value"`
}

func NewPaymentService(db domain.PaymentRepository, dispatcher tools.EventDispatcher, storage *tools.EventStorage, logger *slog.Logger) PaymentServiceImpl {
	return PaymentServiceImpl{
		db:           db,
		dispatcher:   dispatcher,
		eventStorage: storage,
		logger:       logger,
	}
}
func NewPaymentRequest(msg tools.EventMessage) (*PaymentRequest, error) {

	invoiceId := msg.GetAggregateId()
	orderId := msg.Data["order_id"].(string)
	customerId := msg.Data["customer_id"].(string)
	var totalAmount int //by default json converting numbers to float64
	if value, ok := msg.Data["total_sum"].(float64); ok {
		totalAmount = int(value)
	} else {
		totalAmount = msg.Data["total_sum"].(int)
	}

	return &PaymentRequest{
		InvoiceId:   invoiceId,
		OrderId:     orderId,
		CustomerId:  customerId,
		TotalAmount: totalAmount,
	}, nil
}

func (s *PaymentServiceImpl) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {

	var err error

	payment := domain.NewPayment(req.OrderId, req.CustomerId, req.InvoiceId, req.TotalAmount)
	if err = payment.ProcessPay(ctx, s.db); err != nil {
		slog.Warn("on payment processng", slog.String("payment_id", payment.GetId()), slog.String("err", err.Error()))
		return nil, err
	}
	slog.Info("payment successfully processed", slog.String("payment_id", payment.GetId()))
	payment.Accept()
	if err = s.db.SavePayment(ctx, payment); err != nil {
		err = fmt.Errorf("payment created but failed to insert into database: %w", err)
		slog.Warn("on save payment", slog.String("err", err.Error()))
		return nil, err
	}
	slog.Info("payment created and accepted", slog.String("payment_id", payment.GetId()))

	payment.Process(ctx, s.dispatcher, s.eventStorage)
	return &PaymentResponse{PaymentId: payment.GetId()}, nil
}

func (s *PaymentServiceImpl) CreateUserBalance(ctx context.Context, userId string) (bool, error) {
	err := s.db.CreateBalance(ctx, userId)
	return err == nil, err
}

func (s *PaymentServiceImpl) DepositToBalance(ctx context.Context, req DepositRequest) (bool, error) {

	isSuccessful := true
	if err := s.db.AddMoneyToBalance(ctx, req.CustomerId, domain.NewMoney(req.Value)); err != nil {
		slog.Warn("on deposit money", slog.String("err", err.Error()), slog.String("customer_id", req.CustomerId), slog.Int("amount", req.Value))
		return !isSuccessful, err
	}
	slog.Info("balance has been successfuly deposited", slog.String("customer_id", req.CustomerId), slog.Int("amount", req.Value))
	return isSuccessful, nil
}

func (s *PaymentServiceImpl) AcceptPayment(ctx context.Context, paymentId string) error {
	payment, err := s.db.GetPayment(ctx, paymentId)
	if err != nil {
		slog.Warn("fail on accept payment", slog.String("payment_id", paymentId), slog.String("err", err.Error()))
		return err
	}
	payment.Accept()
	slog.Info("payment accepted", slog.String("invoice_id", paymentId))
	return payment.Update(ctx, s.db)
}
func (s *PaymentServiceImpl) RejectPayment(ctx context.Context, paymentId string) error {

	eventId := ctx.Value("event_id").(string)

	payment, err := s.db.GetPayment(ctx, paymentId)
	if err != nil {
		slog.Warn("fail on reject payment", slog.String("payment_id", paymentId), slog.String("err", err.Error()))
		return err
	}
	payment.Reject()

	if err := payment.RefundMoney(ctx, s.db); err != nil {
		slog.Warn("cant refund money", slog.String("payment_id", paymentId), slog.String("err", err.Error()))
		return err
	}

	restoredMessage := tools.NewEventMessage()
	restoredMessage.SetEventID(eventId)
	restoredMessage.Header.Swap()
	if err := restoredMessage.Restore(s.eventStorage); err != nil {
		slog.Warn("fail to restore payment", slog.String("event_id", eventId), slog.String("payment_id", paymentId), slog.String("err", err.Error()))
		return err
	}
	restoredMessage.SetState(payment.GetStatus())
	if err := payment.Update(ctx, s.db); err != nil {
		slog.Warn("fail to update payment", slog.String("payment_id", paymentId), slog.String("err", err.Error()))
		return err
	}
	restoredMessage.Save(s.eventStorage)
	slog.Info("payment rejected and restored", slog.String("payment_id", paymentId))

	s.dispatcher.Produce(ctx, restoredMessage)

	return nil
}

func (s *PaymentServiceImpl) StartDispatch(ctx context.Context, withEvents ...tools.EventOp) {
	slog.Info("Starting message dispatching", slog.String("service", s.GetServiceName()))
	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("Dispatcher stopped", slog.String("service", s.GetServiceName()))
				return
			default:
				s.dispatcher.Handle(ctx, withEvents...)
			}
		}

	}()
}
