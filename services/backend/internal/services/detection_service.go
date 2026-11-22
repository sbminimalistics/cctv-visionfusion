package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"sbminimalistics.com/cctv-visionfusion/internal/db"
	"sbminimalistics.com/cctv-visionfusion/internal/db/model"
	"sbminimalistics.com/cctv-visionfusion/proto/detections"
)

var (
	detectionServiceInstance   *DetectionService
	createDetectionServiceOnce sync.Once
)

type DetectionService struct {
	store *db.Store
}

func (r *DetectionService) SaveDetections(ctx context.Context, data *detections.DetectionCompletedEvent) error {
	var traceID *string
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.HasTraceID() {
		id := spanCtx.TraceID().String()
		traceID = &id
	}

	eventID := data.GetEventUuid()

	if data.GetProcessedAt() == nil {
		return errors.New("processed data missing")
	}

	deviceID := baggage.FromContext(ctx).Member("device.id").Value()

	detectRec := &model.Detection{
		EventID:          &eventID,
		ModelName:        data.GetModelName(),
		SeaweedObjectKey: data.GetFrame().GetSeaweedObjectKey(),
		ImageWidth:       int16(data.GetFrame().GetImageWidth()),
		ImageHeight:      int16(data.GetFrame().GetImageHeight()),
		TraceID:          traceID,
		DetectedAt:       data.GetProcessedAt().AsTime(),
	}
	if err := r.store.WithTx(ctx, func(tx *db.Store) error {
		if err := tx.Detections.Create(ctx, detectRec); err != nil {
			return err
		}

		objects := data.Detections
		if len(objects) > 0 {
			detectionObjects := make([]*model.DetectionObject, 0, len(objects))
			for _, obj := range objects {
				box := obj.GetBox()
				detectionObjects = append(detectionObjects, &model.DetectionObject{
					DetectionID: detectRec.ID,
					ClassID:     int16(obj.GetClassId()),
					ClassName:   obj.GetClassName(),
					X:           int16(box.GetXLeft()),
					Y:           int16(box.GetYLeft()),
					Width:       int16(box.GetWidth()),
					Height:      int16(box.GetHeight()),
					Confidence:  obj.GetConfidence(),
				})
			}

			if err := tx.DetectionObjects.CreateBatch(ctx, detectionObjects); err != nil {
				return err
			}
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

func InitDetectionService(store *db.Store) *DetectionService {
	createDetectionServiceOnce.Do(func() {
		detectionServiceInstance = &DetectionService{
			store: store,
		}

	})
	return detectionServiceInstance
}
