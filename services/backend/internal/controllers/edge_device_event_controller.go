package controllers

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/edge_device"
	"sbminimalistics.com/cctv-visionfusion/internal/event"
	"sbminimalistics.com/cctv-visionfusion/internal/mappers"
	"sbminimalistics.com/cctv-visionfusion/internal/messaging"
	"sbminimalistics.com/cctv-visionfusion/internal/services"
)

type TrafficSource struct {
	SourceID        string
	EventTypeRegexp regexp.Regexp
}

type TrafficDestination struct {
	Exchange string
}

type EdgeDeviceEventController struct {
	rmqClient *messaging.RMQClient
	wg        *sync.WaitGroup

	// channel for separate device events (edge_devices in config)
	incomingEvents chan edge_device.EventPacket
	// channel for composite events (event_composition in config)
	incomingCompositionEvents chan event.CompositionTriggerEvent
	compositions              map[string]*event.EventComposition
}

func NewDeviceEventController(rmq *messaging.RMQClient, parentWG *sync.WaitGroup) *EdgeDeviceEventController {
	var controller = EdgeDeviceEventController{
		rmqClient:                 rmq,
		wg:                        parentWG,
		incomingEvents:            make(chan edge_device.EventPacket, 100),
		incomingCompositionEvents: make(chan event.CompositionTriggerEvent, 10),
		compositions:              make(map[string]*event.EventComposition),
	}

	return &controller
}

func (c *EdgeDeviceEventController) RegisterSSEListener(listener *edge_device.SSEListener) {
	listener.AttachChannel(c.incomingEvents)
}

func (c *EdgeDeviceEventController) RegisterEventComposition(composition *event.EventComposition) {
	id := composition.GetId()
	c.compositions[id] = composition
	composition.AttachChannel(c.incomingCompositionEvents)
}

func (c *EdgeDeviceEventController) StartMessaging(ctx context.Context, routing []config.Route) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case cev := <-c.incomingCompositionEvents:
				fmt.Println("\033[31mcomposition event received:\033[0m", cev)
				for _, route := range routing {
					re := route.Source.EventTypeRegexp.Regexp
					if route.Source.SourceID == cev.CompositionID && re.MatchString(string(cev.EventType)) {
						routingKey := strings.Join([]string{"composition.", cev.CompositionID, ".trigger"}, "")
						triggerEvent := mappers.CompositionTriggerToProto(cev)
						bytes, err := proto.Marshal(triggerEvent)
						if err != nil {
							fmt.Println("failed marshaling proto.Message, error:", err)
						}
						c.rmqClient.Publish(ctx, route.DestinationExchange, routingKey, bytes)
					}
				}
			case ev := <-c.incomingEvents:
				for _, comp := range c.compositions {
					comp.TrackEvent(ev)
				}
				routingKey := strings.Join([]string{"device.", ev.DeviceID, ".event"}, "")
				for _, route := range routing {
					re := route.Source.EventTypeRegexp.Regexp
					if route.Source.SourceID == ev.DeviceID && re.MatchString(ev.Parsed.GetEventType()) {
						deviceEvent := mappers.DeviceEventToProto(ev)
						bytes, err := proto.Marshal(deviceEvent)
						if err != nil {
							fmt.Println("failed marshaling proto.Message, error:", err)
						}
						c.rmqClient.Publish(ctx, route.DestinationExchange, routingKey, bytes)
					}
				}

				now := time.Now().UTC()
				token, expiry, err := services.GetDownloadLinkService().GetNewToken(
					ev.GetAbsolutePicturePath(),
					fmt.Sprintf("%s/%d/%d", ev.DeviceID, now.Year(), now.Month()),
					ev.GetUUID().String(),
				)

				if err != nil {
					fmt.Printf("token generation failed: %v", err)
				} else {
					source := mappers.DownloadInstructionBase{
						Token:       token,
						TokenExpiry: expiry,
						RedeemPath:  "v1/links/redeem",
					}
					instruction := mappers.DeviceInstructionToProto(source)

					bytes, _ := proto.Marshal(instruction)
					c.rmqClient.Publish(
						ctx,
						"requests.download.screenshot",
						routingKey,
						bytes,
					)
				}
			}
		}
	}()
}

type UUIDHolder interface {
	GetUUID() string
	SetUUID(string)
}
