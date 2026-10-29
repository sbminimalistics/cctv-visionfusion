package main

import (
	"context"
	"fmt"
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
	fmt.Println("starting CCTV VisionFusion main backend ...")

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
		fmt.Println("db connection error:,", err)
	}

	// initialize RabbitMQ client
	rmqClient := messaging.NewRMQClient(
		config.Config.RabbitMQ.User,
		config.Config.RabbitMQ.Password,
		config.Config.RabbitMQ.Host,
		int(config.Config.RabbitMQ.Port),
		config.Config.RabbitMQ.VHost,
	)

	fmt.Println("connecting to RabbitMQ broker instance...")
	if err := rmqClient.Connect(); err != nil {
		fmt.Printf("RabbitMQ connection failed to initialize: %v", err)
	}
	defer rmqClient.Close()
	fmt.Println("RabbitMQ transport connection pipeline live.")

	var wg sync.WaitGroup

	// wire controllers
	_ = bootstrap.WireEdgeDeviceEventController(ctx, &wg, rmqClient)

	// wire services
	bootstrap.WireServices(ctx)

	// wire api enpoints
	fmt.Printf("starting API on port %d ...\n", config.Config.API.Port)
	wg.Go(func() { bootstrap.WireAPIServer(ctx) })

	<-ctx.Done()
	fmt.Println("Shutdown initiated. Waiting for edge streams to exit...")

	wg.Wait()
	fmt.Println("All edge devices disconnected. Closing infrastructure...")
}
