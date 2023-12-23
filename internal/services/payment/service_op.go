package service

import (
	"choreography/internal/services/payment/events"
	"choreography/tools"
	"context"
	"log/slog"
)

func WithWarehouseService(service *PaymentServiceImpl) tools.EventOp {
	return func(m tools.EventMessage) {
		if m.ReceiverIsEqual(service.GetServiceName()) && !m.StatusIsEqual(events.PaymentStatusRejected) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "event_id", m.GetEventID())
			ctx = context.WithValue(ctx, "current_service", service.GetServiceName())

			paymentRequest, err := NewPaymentRequest(m)
			if err != nil {
				slog.Warn("operation WithWarehouseService called NewPaymentRequest", slog.String("err", err.Error()))
			} else {
				_, err = service.CreatePayment(ctx, paymentRequest)
			}
			if err != nil {
				m.Header.Swap()
				m.Header.SetEvent(events.PaymentNotCreated{})
				m.SetState(events.PaymentStatusRejected)
				m.Data["linked_service"] = tools.OrderService
				slog.Warn("operation WithWarehouseService called CreatePayment", slog.String("err", err.Error()))
				service.dispatcher.Produce(ctx, m)
			}

		}
	}
}
