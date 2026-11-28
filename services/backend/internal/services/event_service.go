package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/db"
	"sbminimalistics.com/cctv-visionfusion/internal/db/model"
	"sbminimalistics.com/cctv-visionfusion/internal/edge_device"
	"sbminimalistics.com/cctv-visionfusion/internal/mappers"
	"sbminimalistics.com/cctv-visionfusion/internal/messaging"
	"sbminimalistics.com/cctv-visionfusion/proto/detections"
	"sbminimalistics.com/cctv-visionfusion/proto/events"
)

var (
	eventServiceInstance   *EventService
	createEventServiceOnce sync.Once
)

type EventService struct {
	rmqClient *messaging.RMQClient
	store     *db.Store
	linkSvc   *DownloadLinkService
	routing   []config.Route
}

func (r *EventService) SaveEvent(ctx context.Context, data edge_device.EventPacket) error {
	var traceID *string
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.HasTraceID() {
		id := spanCtx.TraceID().String()
		traceID = &id
	}

	manufacturer := string(data.Manufacturer)
	eventType := data.Parsed.GetEventType()

	ev := &model.CapturedEvent{
		ID:      data.GetUUID().String(),
		Src:     &data.DeviceID,
		SrcType: &manufacturer,
		Type:    &eventType,
		TraceID: traceID,
	}

	if err := r.store.Events.Create(ctx, ev); err != nil {
		slog.ErrorContext(
			ctx,
			"failed to persist edge event",
			"event_id", data.GetUUID().String(),
			"device_id", data.DeviceID,
			"error", err,
		)
		return fmt.Errorf("save event %s: %w", data.GetUUID(), err)
	}

	slog.InfoContext(
		ctx,
		"saved edge event to database",
		"event_id", data.GetUUID().String(),
		"device_id", data.DeviceID,
		"event_type", ev.Type,
	)

	return nil
}

func (r *EventService) ProcessEvent(ctx context.Context, ev edge_device.EventPacket) {
	routingKey := strings.Join([]string{"device.", ev.DeviceID, ".event"}, "")
	deviceEvent := mappers.DeviceEventToProto(ev)
	r.publishMatchingRoutes(ctx, ev.DeviceID, ev.Parsed.GetEventType(), false, deviceEvent)

	// if ev.Parsed.GetEventType() == "fielddetection" {
	if ev.Parsed.GetEventType() == "VMD" {
		r.SaveEvent(ctx, ev)
		now := time.Now().UTC()
		token, expiry, err := r.linkSvc.GetNewToken(
			ev.GetAbsolutePicturePath(),
			fmt.Sprintf("%s/%d/%d", ev.DeviceID, now.Year(), now.Month()),
			ev.GetUUID(),
		)

		if err != nil {
			slog.ErrorContext(ctx, "token generation failed", "error", err)
		} else {
			source := mappers.DownloadInstructionBase{
				Token:       token,
				TokenExpiry: expiry,
				RedeemPath:  "v1/links/redeem",
			}
			instruction := mappers.DeviceInstructionToProto(source)

			bytes, _ := proto.Marshal(instruction)
			r.rmqClient.Publish(
				ctx,
				"requests.download.screenshot",
				routingKey,
				bytes,
			)
		}
	}
}

func (r *EventService) RouteCompletedEvent(ctx context.Context, eventUUID uuid.UUID, evWithDetections *detections.DetectionCompletedEvent) error {
	if len(evWithDetections.GetDetections()) == 0 {
		return nil
	}

	evRecord, err := r.store.Events.GetByUUID(ctx, eventUUID)
	if err != nil {
		return fmt.Errorf("fetch event %s for routing: %w", eventUUID, err)
	}

	deviceEvent := mappers.StoredEventToProto(evRecord, evWithDetections) // attaches optional detections onto DeviceEvent
	fmt.Println("RouteCompletedEvent", deviceEvent)
	return r.publishMatchingRoutes(ctx, *evRecord.Src, *evRecord.Type, true, deviceEvent)
}

func (r *EventService) publishMatchingRoutes(
	ctx context.Context,
	deviceID string,
	eventType string,
	requireDetections bool,
	payload *events.DeviceEvent,
) error {
	routingKey := fmt.Sprintf("device.%s.event", deviceID)

	for _, route := range r.routing {
		if route.OnlyWithDetections != requireDetections {
			continue
		}
		if route.Source.SourceID == deviceID && route.Source.EventTypeRegexp.Regexp.MatchString(eventType) {
			bytes, err := proto.Marshal(payload)
			if err != nil {
				return fmt.Errorf("marshal DeviceEvent: %w", err)
			}
			if err := r.rmqClient.Publish(ctx, route.DestinationExchange, routingKey, bytes); err != nil {
				return fmt.Errorf("publish to %s: %w", route.DestinationExchange, err)
			}
		}
	}
	return nil
}

func InitEventService(
	store *db.Store,
	linkSvc *DownloadLinkService,
	rmqClient *messaging.RMQClient,
	routing []config.Route,
) *EventService {
	createEventServiceOnce.Do(func() {
		eventServiceInstance = &EventService{
			store:     store,
			linkSvc:   linkSvc,
			rmqClient: rmqClient,
			routing:   routing,
		}

	})
	return eventServiceInstance
}
