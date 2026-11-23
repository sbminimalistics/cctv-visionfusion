package controllers

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/edge_device"
	"sbminimalistics.com/cctv-visionfusion/internal/event"
	"sbminimalistics.com/cctv-visionfusion/internal/mappers"
	"sbminimalistics.com/cctv-visionfusion/internal/messaging"
	"sbminimalistics.com/cctv-visionfusion/internal/services"
)

type pkgAnchor struct{}

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
	service                   *services.EventService
	throttler                 *services.EventThrottle
}

var pkgName = reflect.TypeOf(pkgAnchor{}).PkgPath()

func NewDeviceEventController(rmq *messaging.RMQClient, parentWG *sync.WaitGroup, service *services.EventService) *EdgeDeviceEventController {
	var controller = EdgeDeviceEventController{
		rmqClient:                 rmq,
		wg:                        parentWG,
		incomingEvents:            make(chan edge_device.EventPacket, 100),
		incomingCompositionEvents: make(chan event.CompositionTriggerEvent, 10),
		compositions:              make(map[string]*event.EventComposition),
		service:                   service,
		throttler:                 services.NewEventThrottle(config.Config.EventCooldown),
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
	var tracer = otel.Tracer(pkgName)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case cev := <-c.incomingCompositionEvents:
				slog.Debug("composition event received", "raw_event", cev)
				for _, route := range routing {
					re := route.Source.EventTypeRegexp.Regexp
					if route.Source.SourceID == cev.CompositionID && re.MatchString(string(cev.EventType)) {
						routingKey := strings.Join([]string{"composition.", cev.CompositionID, ".trigger"}, "")
						triggerEvent := mappers.CompositionTriggerToProto(cev)
						bytes, err := proto.Marshal(triggerEvent)
						if err != nil {
							slog.Error("failed marshaling proto.Message", "error", err)
						}
						c.rmqClient.Publish(ctx, route.DestinationExchange, routingKey, bytes)
					}
				}
			case ev, ok := <-c.incomingEvents:
				if !ok {
					return
				}

				eventCtx, span := tracer.Start(
					ev.Ctx,
					"controller.event.handling",
					trace.WithAttributes(
						attribute.String("device_id", ev.DeviceID),
						attribute.String("event_id", ev.GetUUID().String()),
					),
				)

				now := time.Now().UTC()
				allow := c.throttler.Allow(ev.DeviceID, ev.Parsed.GetEventType(), now)
				if !allow {
					slog.DebugContext(eventCtx, "edge device event suppressed (cooldown period active)")
					continue
				} else {
					slog.DebugContext(eventCtx, "edge device event allowed (cooldown period inactive)")
				}

				mDevice, _ := baggage.NewMember("device.id", ev.DeviceID)
				bag, _ := baggage.New(mDevice)
				eventCtx = baggage.ContextWithBaggage(eventCtx, bag)

				for _, comp := range c.compositions {
					comp.TrackEvent(ev)
				}
				routingKey := strings.Join([]string{"device.", ev.DeviceID, ".event"}, "")
				// slog.Debug("<- event type:", "type", ev.Parsed.GetEventType())
				for _, route := range routing {
					re := route.Source.EventTypeRegexp.Regexp
					match := re.MatchString(ev.Parsed.GetEventType())
					if route.Source.SourceID == ev.DeviceID && match {
						deviceEvent := mappers.DeviceEventToProto(ev)
						bytes, err := proto.Marshal(deviceEvent)
						if err != nil {
							slog.ErrorContext(eventCtx, "failed marshaling proto.Message", "error", err)
						}
						c.rmqClient.Publish(eventCtx, route.DestinationExchange, routingKey, bytes)
					}
				}

				// if ev.Parsed.GetEventType() == "fielddetection" {
				if ev.Parsed.GetEventType() == "VMD" {
					c.service.SaveEvent(eventCtx, ev)
					now := time.Now().UTC()
					token, expiry, err := services.GetDownloadLinkService().GetNewToken(
						ev.GetAbsolutePicturePath(),
						fmt.Sprintf("%s/%d/%d", ev.DeviceID, now.Year(), now.Month()),
						ev.GetUUID(),
					)

					if err != nil {
						slog.ErrorContext(eventCtx, "token generation failed", "error", err)
					} else {
						source := mappers.DownloadInstructionBase{
							Token:       token,
							TokenExpiry: expiry,
							RedeemPath:  "v1/links/redeem",
						}
						instruction := mappers.DeviceInstructionToProto(source)

						bytes, _ := proto.Marshal(instruction)
						c.rmqClient.Publish(
							eventCtx,
							"requests.download.screenshot",
							routingKey,
							bytes,
						)
					}
				}

				slog.InfoContext(eventCtx, "handled edge device event",
					"device_id", ev.DeviceID,
					"event_id", ev.GetUUID().String(),
					"event_type", ev.Parsed.GetEventType(),
				)

				span.End()
			}
		}
	}()
}

type UUIDHolder interface {
	GetUUID() string
	SetUUID(string)
}
