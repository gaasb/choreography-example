package service

import (
	"choreography/internal/services/warehouse/domain"
	"choreography/tools"
	"context"
	"log/slog"
	"os"
	"os/signal"
)

var currentServiceName = tools.WarehouseService

func (*WarehouseServiceImpl) GetServiceName() string {
	return currentServiceName
}

type Service interface {
	CreateInvoiceFromOrder(ctx context.Context, items ...*WarehouseRequest) (*WarehouseResponse, error)
	AddProduct(ctx context.Context, req ProductRequest) (bool, error)
	AcceptInvoice(ctx context.Context, invoiceId string) error
	RejectInvoice(ctx context.Context, invoiceId string) error

	StartDispatch(ctx context.Context, withEvents ...tools.EventOp)
}

func NewKafkaWarehouseService() (*WarehouseServiceImpl, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	service := newWarehouseServiceWithKafka()
	ctx = context.WithValue(ctx, "current_service", service.GetServiceName())
	service.StartDispatch(ctx, WithOrderService(&service), WithPaymentService(&service))
	return &service, stop
}

func NewLocalWarehouseService(broker *tools.LocalBroker) (*WarehouseServiceImpl, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	service := newWarehouseServiceWithLocalBroker(broker)
	ctx = context.WithValue(ctx, "current_service", service.GetServiceName())
	service.StartDispatch(ctx, WithOrderService(&service), WithPaymentService(&service))
	return &service, stop
}

func newWarehouseServiceWithLocalBroker(broker *tools.LocalBroker) WarehouseServiceImpl {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := tools.NewEventDispatcherLocal(broker, currentServiceName)
	db := domain.NewInMemoryDB()
	storage := tools.NewEventStorage()
	slog.SetDefault(logger)
	return NewWarehouseService(db, dispatcher, storage, logger)
}

func newWarehouseServiceWithKafka() WarehouseServiceImpl {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := tools.NewEventDispatcherKafka(currentServiceName)
	db := domain.NewInMemoryDB()
	storage := tools.NewEventStorage()
	slog.SetDefault(logger)
	return NewWarehouseService(db, dispatcher, storage, logger)
}
