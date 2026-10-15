package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"sbminimalistics.com/vms/internal/cam"
	"sbminimalistics.com/vms/internal/config"
	"sbminimalistics.com/vms/internal/db"
	"sbminimalistics.com/vms/internal/event"
	"sbminimalistics.com/vms/internal/messaging"
)

func main() {
	fmt.Println("starting vms...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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
	broker := event.NewBroker(100)

	broker.StartConsumer(ctx, &wg, func(ev event.CommsEvent) {
		fmt.Printf(" [ALERT TRIGGERED] Camera %s reported event: %s\n", ev.CameraID, ev.EventType)
		routingKey := strings.Join([]string{"camera.", ev.CameraID}, "")
		rmqClient.Publish(ctx, "vms.events", routingKey, []byte("camera.event"))
	})

	for _, camConfig := range config.Config.Cameras {

		cameraStream := cam.NewStreamInstance(camConfig)

		wg.Add(1)
		go func(c *cam.EventNotificationAlert, cfg config.CameraConfig) {
			defer wg.Done()

			fmt.Printf(" [INIT] Starting connection routine for %s...\n", cfg.Name)

			broker.OrchestrateStream(ctx, &wg, c)

			c.ConnectAndRetryForever(ctx, camConfig)

		}(cameraStream, camConfig)
	}

	// alertChan := make(chan event.CommsEvent, 100)

	// var cam08 cam.EventNotificationAlert
	// go cam08.Connect(config.Cfg.Cameras[1], alertChan)

	select {}
}
