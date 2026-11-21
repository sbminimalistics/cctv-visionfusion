package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"sbminimalistics.com/cctv-visionfusion/internal/db"
	"sbminimalistics.com/cctv-visionfusion/internal/db/model"
	"sbminimalistics.com/cctv-visionfusion/internal/edge_device"
)

var (
	eventServiceInstance   *EventService
	createEventServiceOnce sync.Once
)

type EventService struct {
	store *db.Store
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

func InitEventService(store *db.Store) *EventService {
	createEventServiceOnce.Do(func() {
		eventServiceInstance = &EventService{
			store: store,
		}

	})
	return eventServiceInstance
}
