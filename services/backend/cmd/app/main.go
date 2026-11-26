package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"sbminimalistics.com/cctv-visionfusion/internal/bootstrap"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/db"
	"sbminimalistics.com/cctv-visionfusion/internal/messaging"
)

func main() {
	defer bootstrap.InitTelemetry("cctv-BE")()

	slog.Info("starting CCTV VisionFusion backend")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// initialize db connection
	orm, err := db.Connect(
		config.Config.DB.Host,
		int(config.Config.DB.Port),
		config.Config.DB.User,
		config.Config.DB.Password,
		config.Config.DB.Name,
		config.Config.DB.Schema,
		config.Config.DB.LogLevel,
	)
	if err != nil {
		slog.Error("db connection error", "error", err)
		os.Exit(1)
	}

	// initialize db store
	store := db.NewStore(orm)

	var wg sync.WaitGroup

	// initialize RabbitMQ client
	rmqClient := messaging.NewRMQClient(
		config.Config.RabbitMQ.User,
		config.Config.RabbitMQ.Password,
		config.Config.RabbitMQ.Host,
		int(config.Config.RabbitMQ.Port),
		config.Config.RabbitMQ.VHost,
	)

	slog.Info("connecting to RabbitMQ broker instance...")
	if err := rmqClient.Connect(); err != nil {
		slog.Error("RabbitMQ connection failed to initialize", "error", err)
	} else {
		slog.Info("RabbitMQ transport connection pipeline live.")
		defer rmqClient.Close()
	}

	// wire services
	services := bootstrap.WireServices(ctx, store)

	// wire controllers
	_ = bootstrap.WireEdgeDeviceEventController(ctx, &wg, rmqClient, services.EventService)

	_ = bootstrap.WireDetectionEventController(ctx, &wg, rmqClient, services.DetectionWFService)

	// wire api enpoints
	slog.Info("starting API", "port", config.Config.API.Port)
	wg.Go(func() { bootstrap.WireAPIServer(ctx) })

	<-ctx.Done()
	slog.Info("shutdown initiated, waiting for edge streams to exit...")

	wg.Wait()
	slog.Info("all edge devices disconnected, closing infrastructure...")
}
