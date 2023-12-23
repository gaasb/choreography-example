package main

import (
	service "choreography/internal/services/order"
	"context"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	srv, stopService := service.NewKafkaOrderService()
	slog.Info("Run service", slog.String("service", srv.GetServiceName()))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		stopService()
	// 		return
	// 	default:
	// 		line := map[int]int{1: 3}
	// 		r := service.OrderRequest{Line: line, UserId: "gaa"}
	// 		srv.CreateOrder(ctx, r)
	// 		time.Sleep(time.Second * 3)
	// 	}
	// }
	<-ctx.Done()
	stopService()
}
