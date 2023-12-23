package main

import (
	service "choreography/internal/services/warehouse"
	"context"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	srv, stopService := service.NewKafkaWarehouseService()
	slog.Info("Run service", slog.String("service", srv.GetServiceName()))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// srv.AddProduct(ctx, service.ProductRequest{ItemId: 1, Name: "sad", Qty: 10, Price: 15})
	<-ctx.Done()
	stopService()
}
