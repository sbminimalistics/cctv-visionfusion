package bootstrap

import (
	"context"
	"fmt"
	"sync"

	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/controllers"
	"sbminimalistics.com/cctv-visionfusion/internal/edge_device"
	"sbminimalistics.com/cctv-visionfusion/internal/event"
	"sbminimalistics.com/cctv-visionfusion/internal/messaging"
)

// WireEdgeDeviceController reads config data and hooks up all inputs to the orchestrator.
func WireEdgeDeviceEventController(
	ctx context.Context,
	wg *sync.WaitGroup,
	rmq *messaging.RMQClient,
) *controllers.EdgeDeviceEventController {
	controller := controllers.NewDeviceEventController(rmq, wg)
	for _, compositionConfig := range config.Config.EventCompositions {
		c := event.NewComposition(compositionConfig)
		controller.RegisterEventComposition(c)
	}

	// setup edge device inputs based on configuration data
	for _, device := range config.Config.EdgeDevices {
		stream := createListener(device)
		cfg := device

		wg.Add(1)
		go func(stream *edge_device.SSEListener, item config.EdgeDeviceConfig) {
			defer wg.Done()
			controller.RegisterSSEListener(stream)
			stream.ConnectAndRetryForever(ctx, item)
		}(stream, cfg)
	}

	// start message routing based on the provided config
	controller.StartMessaging(ctx, config.Config.Routing)

	fmt.Printf("successfully wired %d edge devices\n", len(config.Config.EdgeDevices))

	return controller
}

func createListener(device config.EdgeDeviceConfig) *edge_device.SSEListener {
	switch device.Manufacturer {
	case config.Hikvision:
		return edge_device.NewSSEListener(
			device,
			&edge_device.HikvisionPacketParser{},
		)
	default:
		panic(fmt.Sprintf("unknown edge device manufacturer '%s'", device.Manufacturer))
	}
}
