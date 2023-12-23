package main

import (
	service "choreography/internal/services/payment"
	"context"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	srv, stopService := service.NewKafkaPaymentService()
	slog.Info("Run service", slog.String("service", srv.GetServiceName()))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// srv.CreateUserBalance(context.Background(), "gaa")
	// srv.DepositToBalance(context.Background(), service.DepositRequest{CustomerId: "gaa", Value: 500})
	<-ctx.Done()
	stopService()
}
