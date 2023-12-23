package service

import (
	"choreography/internal/services/order/domain"
	"choreography/tools"
	"context"
	"log/slog"
)

func WithWarehouseService(service *OrderServiceImpl) tools.EventOp {
	return func(m tools.EventMessage) {

		// && m.EventIsEqual((*events.OrderCreated)(nil))
		if m.ReceiverIsEqual(service.GetServiceName()) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "event_id", m.GetEventID())
			ctx = context.WithValue(ctx, "current_service", service.GetServiceName())

			orderId := m.Data["order_id"].(string)

			if domain.Status(m.State).IsAccepted() {
				if err := service.AcceptOrder(ctx, orderId); err != nil {
					slog.Warn("operation WithWarehouseService called AcceptOrder", slog.String("err", err.Error()))
				}
			} else {
				if err := service.RejectOrder(ctx, orderId); err != nil {
					slog.Warn("operation WithWarehouseService called RejectOrder", slog.String("err", err.Error()))
				}
			}
		}
	}
}
