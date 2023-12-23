package service

import (
	"choreography/internal/services/payment/domain"
	"choreography/tools"
	"context"
	"log/slog"
	"os"
	"os/signal"
)

var currentServiceName = tools.PaymentService

func (*PaymentServiceImpl) GetServiceName() string {
	return currentServiceName
}

type Service interface {
	CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	CreateUserBalance(ctx context.Context, userId string) (bool, error)
	DepositToBalance(ctx context.Context, req DepositRequest) (bool, error)
	AcceptPayment(ctx context.Context, paymentId string) error
	RejectPayment(ctx context.Context, paymentId string) error

	StartDispatch(ctx context.Context, withEvents ...tools.EventOp)
}

func NewKafkaPaymentService() (*PaymentServiceImpl, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	service := newPaymentServiceWithKafka()
	ctx = context.WithValue(ctx, "current_service", service.GetServiceName())
	service.StartDispatch(ctx, WithWarehouseService(&service))
	return &service, stop
}

func NewLocalPaymentService(broker *tools.LocalBroker) (*PaymentServiceImpl, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	service := newPaymentServiceWithLocalBroker(broker)
	ctx = context.WithValue(ctx, "current_service", service.GetServiceName())
	service.StartDispatch(ctx, WithWarehouseService(&service))
	return &service, stop
}

func newPaymentServiceWithLocalBroker(broker *tools.LocalBroker) PaymentServiceImpl {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := tools.NewEventDispatcherLocal(broker, currentServiceName)
	db := domain.NewInMemoryDB()
	storage := tools.NewEventStorage()
	slog.SetDefault(logger)
	return NewPaymentService(db, dispatcher, storage, logger)
}

func newPaymentServiceWithKafka() PaymentServiceImpl {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := tools.NewEventDispatcherKafka(currentServiceName)
	db := domain.NewInMemoryDB()
	storage := tools.NewEventStorage()
	slog.SetDefault(logger)
	return NewPaymentService(db, dispatcher, storage, logger)
}
