package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/protobuf/proto"
	"sbminimalistics.com/cctv-visionfusion/internal/bootstrap"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/db"
	"sbminimalistics.com/cctv-visionfusion/internal/messaging"
	"sbminimalistics.com/cctv-visionfusion/proto/downloads"
)

func main() {
	defer bootstrap.InitTelemetry("cctv-BE")()

	slog.Info("starting CCTV VisionFusion backend")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// initialize db
	_, err := db.Connect(
		config.Config.DB.Host,
		int(config.Config.DB.Port),
		config.Config.DB.User,
		config.Config.DB.Password,
		config.Config.DB.Name,
	)

	if err != nil {
		slog.Error("db connection error", "error", err)
	}

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

		err := rmqClient.StartConsumer(
			ctx,
			&wg,
			config.Config.RabbitMQ.QueueDetectionsCompleted,
			"cctv-BE",
			func(msgCtx context.Context, body []byte) error {
				var msg downloads.DownloadInstruction
				if err := proto.Unmarshal(body, &msg); err != nil {
					return fmt.Errorf("proto unmarshal failed: %w", err)
				}
				slog.InfoContext(msgCtx, "detection result received")
				// pass msgCtx & msg to your service layer
				return nil
			},
		)
		if err != nil {
			slog.Error("failed to start consumer", "error", err)
		}
	}

	// wire controllers
	_ = bootstrap.WireEdgeDeviceEventController(ctx, &wg, rmqClient)

	// wire services
	bootstrap.WireServices(ctx)

	// wire api enpoints
	slog.Info("starting API", "port", config.Config.API.Port)
	wg.Go(func() { bootstrap.WireAPIServer(ctx) })

	<-ctx.Done()
	slog.Info("shutdown initiated, waiting for edge streams to exit...")

	wg.Wait()
	slog.Info("all edge devices disconnected, closing infrastructure...")
}
