package service

import (
	"choreography/internal/services/order/domain"
	"choreography/tools"
	"context"
	"log/slog"
	"os"
	"os/signal"
)

var currentServiceName = tools.OrderService

func (*OrderServiceImpl) GetServiceName() string {
	return currentServiceName
}

type Service interface {
	CreateOrder(ctx context.Context, req OrderRequest) (*OrderResponse, error)
	AcceptOrder(ctx context.Context, orderId string) error
	RejectOrder(ctx context.Context, orderId string) error
	StartDispatch(ctx context.Context, withEvents ...tools.EventOp)
}

func NewKafkaOrderService() (*OrderServiceImpl, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	service := newOrderServiceWithKafka()
	ctx = context.WithValue(ctx, "current_service", service.GetServiceName())
	service.StartDispatch(ctx, WithWarehouseService(&service))
	return &service, stop
}

func NewLocalOrderService(broker *tools.LocalBroker) (*OrderServiceImpl, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	service := newOrderServiceWithLocalBroker(broker)
	ctx = context.WithValue(ctx, "current_service", service.GetServiceName())
	service.StartDispatch(ctx, WithWarehouseService(&service))
	return &service, stop
}

func newOrderServiceWithLocalBroker(broker *tools.LocalBroker) OrderServiceImpl {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := tools.NewEventDispatcherLocal(broker, currentServiceName)
	db := domain.NewInMemoryDB()
	storage := tools.NewEventStorage()
	slog.SetDefault(logger)
	return NewOrderService(db, dispatcher, storage, logger)
}

func newOrderServiceWithKafka() OrderServiceImpl {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := tools.NewEventDispatcherKafka(currentServiceName)
	db := domain.NewInMemoryDB()
	storage := tools.NewEventStorage()
	slog.SetDefault(logger)
	return NewOrderService(db, dispatcher, storage, logger)
}
