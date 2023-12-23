package service

import (
	"choreography/internal/services/order/domain"

	"choreography/tools"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type OrderServiceImpl struct {
	db           domain.OrderRepository
	dispatcher   tools.EventDispatcher
	eventStorage *tools.EventStorage
	logger       *slog.Logger
}

type OrderRequest struct {
	UserId string      `json:"user_id"`
	Line   domain.Line `json:"order_line"`
}

type OrderResponse struct {
	OrderId string `json:"order_id"`
}

func NewOrderService(db domain.OrderRepository, dispatcher tools.EventDispatcher, storage *tools.EventStorage, logger *slog.Logger) OrderServiceImpl {
	return OrderServiceImpl{
		db:           db,
		dispatcher:   dispatcher,
		eventStorage: storage,
		logger:       logger,
	}
}

func (s *OrderServiceImpl) CreateOrder(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	ctx = context.WithValue(ctx, "current_service", s.GetServiceName())

	if len(req.UserId) > 0 {
		if len(req.Line) == 0 {
			return nil, errors.New("line is empty")
		}

		var order *domain.Order
		var err error

		if order, err = domain.NewOrder(req.UserId, req.Line); err != nil {
			err = fmt.Errorf("failed to create order: %w", err)
			slog.Warn("create order", slog.String("err", err.Error()))
			return nil, err
		}
		if err = s.db.Add(ctx, *order); err != nil {
			err = fmt.Errorf("order created but failed to insert into database: %w", err)
			slog.Warn("create order", slog.String("err", err.Error()))
			return nil, err
		}
		slog.Info("order created", slog.String("id", order.GetId()))
		order.Process(ctx, s.dispatcher, s.eventStorage)
		return &OrderResponse{OrderId: order.GetId()}, nil
	}
	err := errors.New("invalid user")
	slog.Warn("create order", slog.String("err", err.Error()))
	return nil, err
}
func (s *OrderServiceImpl) AcceptOrder(ctx context.Context, orderId string) error {
	order, err := s.db.GetBy(ctx, orderId)
	if err != nil {
		slog.Warn("fail on accept order", slog.String("order_id", orderId), slog.String("err", err.Error()))
		return err
	}
	order.Accept()
	if err := order.Update(ctx, s.db); err != nil {
		slog.Warn("fail on update order", slog.String("order_id", orderId), slog.String("err", err.Error()))
		return err
	}
	slog.Info("order accepted and updated", slog.String("id", orderId))
	return nil
}
func (s *OrderServiceImpl) RejectOrder(ctx context.Context, orderId string) error {

	eventId := ctx.Value("event_id").(string)

	order, err := s.db.GetBy(ctx, orderId)
	if err != nil {
		slog.Warn("reject order", slog.String("order_id", orderId), slog.String("err", err.Error()))
		return err
	}
	order.Reject()

	restoredMessage := tools.NewEventMessage()
	restoredMessage.SetEventID(eventId)
	if err := restoredMessage.Restore(s.eventStorage); err != nil {
		slog.Warn("failed on restore order", slog.String("id", orderId), slog.String("err", err.Error()))
		return err
	}

	if err := order.Update(ctx, s.db); err != nil {
		slog.Warn("failed on update order", slog.String("id", orderId), slog.String("err", err.Error()))
		return err
	}

	restoredMessage.SetState(order.GetStatus())
	restoredMessage.Save(s.eventStorage)
	slog.Info("order rejected and restored", slog.String("id", orderId))
	return nil
}

func (s *OrderServiceImpl) StartDispatch(ctx context.Context, withEvents ...tools.EventOp) {
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
