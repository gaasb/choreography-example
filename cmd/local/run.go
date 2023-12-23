package main

import (
	orders "choreography/internal/services/order"
	payments "choreography/internal/services/payment"
	warehouses "choreography/internal/services/warehouse"
	"choreography/tools"
	"context"
	"os"
	"os/signal"
	"time"
)

func main() {
	runLocal()
}

func runKafka() {
	_, stopPayment := payments.NewKafkaPaymentService()
	_, stopOrder := orders.NewKafkaOrderService()
	_, stopWarehouse := warehouses.NewKafkaWarehouseService()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	<-ctx.Done()
	stopPayment()
	stopWarehouse()
	stopOrder()
	time.Sleep(time.Second * 1)
}

func runLocal() {
	broker := tools.NewLocalBroker()

	_, stopPayment := payments.NewLocalPaymentService(broker)
	_, stopOrder := orders.NewLocalOrderService(broker)
	_, stopWarehouse := warehouses.NewLocalWarehouseService(broker)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	broker.Run(ctx)
	defer stop()

	<-ctx.Done()
	stopPayment()
	stopWarehouse()
	stopOrder()
	time.Sleep(time.Second * 1)
}
