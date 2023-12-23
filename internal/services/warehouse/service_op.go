package service

import (
	"choreography/internal/services/warehouse/events"
	"choreography/tools"
	"context"
	"log/slog"
)

func WithOrderService(service *WarehouseServiceImpl) tools.EventOp {
	return func(m tools.EventMessage) {
		if m.ReceiverIsEqual(service.GetServiceName()) && m.ProducerIsEqual(tools.OrderService) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "event_id", m.GetEventID())
			ctx = context.WithValue(ctx, "current_service", service.GetServiceName())

			orderRequest, err := NewWarehouseRequest(m)
			if err != nil {
				slog.Warn("operation WithOrderService called NewWarehouseRequest", slog.String("err", err.Error()))
				return
			}
			if _, err := service.CreateInvoiceFromOrder(ctx, orderRequest); err != nil {
				m.Header.Swap()
				m.Header.SetEvent(events.InvoiceNotCreated{})
				m.SetState(events.InvoiceStatusRejected)
				slog.Warn("operation WithOrderService called CreateInvoiceFromOrder", slog.String("err", err.Error()))
				service.dispatcher.Produce(ctx, m)
				return
			}
		}
	}
}

func WithPaymentService(service *WarehouseServiceImpl) tools.EventOp {
	return func(m tools.EventMessage) {
		if m.ReceiverIsEqual(service.GetServiceName()) && m.ProducerIsEqual(tools.PaymentService) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "event_id", m.GetEventID())
			ctx = context.WithValue(ctx, "current_service", service.GetServiceName())
			if v, ok := m.Data["linked_service"].(string); ok {
				ctx = context.WithValue(ctx, "linked_service", v)
			}

			invoiceId := m.Data["invoice_id"].(string)
			if m.StatusIsEqual(events.InvoiceStatusAccepted) {
				if err := service.AcceptInvoice(ctx, invoiceId); err != nil {
					slog.Warn("operation WithPaymentService called AcceptInvoice", slog.String("err", err.Error()))
				}
			} else {
				if err := service.RejectInvoice(ctx, invoiceId); err != nil {
					slog.Warn("operation WithPaymentService called RejectInvoice", slog.String("err", err.Error()))
				}
			}
		}
	}
}
