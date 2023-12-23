package local_test

import (
	orders "choreography/internal/services/order"
	payments "choreography/internal/services/payment"
	warehouses "choreography/internal/services/warehouse"
	"choreography/tools"
	"context"
	"os"
	"os/signal"
	"testing"
	"time"
)

func TestLocal(t *testing.T) {
	broker := tools.NewLocalBroker()
	// t.Parallel()
	var payment *payments.PaymentServiceImpl
	var stopPayment context.CancelFunc
	t.Run("start payment service", func(t *testing.T) { payment, stopPayment = payments.NewLocalPaymentService(broker) })

	var order *orders.OrderServiceImpl
	var stopOrder context.CancelFunc
	t.Run("start order service", func(t *testing.T) { order, stopOrder = orders.NewLocalOrderService(broker) })

	var warehouse *warehouses.WarehouseServiceImpl
	var stopWarehouse context.CancelFunc
	t.Run("start order service", func(t *testing.T) { warehouse, stopWarehouse = warehouses.NewLocalWarehouseService(broker) })

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	t.Run("start broker", func(t *testing.T) { broker.Run(ctx) })
	// defer stop()

	t.Run("create user balance", func(t *testing.T) { payment.CreateUserBalance(context.Background(), "gaa") })
	t.Run("deposit money to user balance", func(t *testing.T) {
		payment.DepositToBalance(context.Background(), payments.DepositRequest{CustomerId: "gaa", Value: 500})
	})
	t.Run("add item to warehouse", func(t *testing.T) {
		warehouse.AddProduct(ctx, warehouses.ProductRequest{ItemId: 1, Name: "some item", Qty: 10, Price: 15})
	})

	timer := time.NewTicker(time.Duration(time.Second * 10))
	for {
		select {
		case <-timer.C:
			stop()
		case <-ctx.Done():
			stopPayment()
			stopWarehouse()
			stopOrder()
			return
		default:
			line := map[int]int{1: 3}
			r := orders.OrderRequest{Line: line, UserId: "gaa"}
			order.CreateOrder(ctx, r)
			time.Sleep(time.Second * 3)
		}
	}
}
