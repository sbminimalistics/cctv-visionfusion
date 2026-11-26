package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/baggage"
	"sbminimalistics.com/cctv-visionfusion/internal/db"
	"sbminimalistics.com/cctv-visionfusion/proto/detections"
)

var (
	detectionWFSInstance   *DetectionWorkflowService
	createDetectionWFSOnce sync.Once
)

type DetectionWorkflowService struct {
	store            *db.Store
	detectionService *DetectionService
	eventService     *EventService
}

func (r *DetectionWorkflowService) CompleteDetection(ctx context.Context, data *detections.DetectionCompletedEvent) error {
	eventID := data.GetEventUuid()
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return fmt.Errorf("invalid event uuid %q: %w", eventID, err)
	}

	deviceID := baggage.FromContext(ctx).Member("device.id").Value()

	status := "empty"
	if len(data.GetDetections()) > 0 {
		status = "confirmed"
	}

	if err := r.store.WithTx(ctx, func(ctxTx context.Context, tx *db.Store) error {
		if err := r.detectionService.SaveDetections(ctxTx, data); err != nil {
			return fmt.Errorf("save detections: %w", err)
		}
		if err := tx.Events.UpdateStatus(ctxTx, eventUUID, status); err != nil {
			return fmt.Errorf("update event status: %w", err)
		}
		return nil
	}); err != nil {

		slog.ErrorContext(
			ctx,
			"failed to persist detection record",
			"event_id", eventID,
			"device_id", deviceID,
			"error", err,
		)
		return fmt.Errorf("save detecion error: %w", err)
	}

	slog.InfoContext(
		ctx,
		"saved detection to database",
		"event_id", eventID,
		"device_id", deviceID,
	)

	return nil
}

func InitDetectionWorkflowService(
	store *db.Store,
	detectionService *DetectionService,
	eventService *EventService,
) *DetectionWorkflowService {
	createDetectionWFSOnce.Do(func() {
		detectionWFSInstance = &DetectionWorkflowService{
			store:            store,
			detectionService: detectionService,
			eventService:     eventService,
		}

	})
	return detectionWFSInstance
}
