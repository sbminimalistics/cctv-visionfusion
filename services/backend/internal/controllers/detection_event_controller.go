package controllers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/messaging"
	"sbminimalistics.com/cctv-visionfusion/internal/services"
	"sbminimalistics.com/cctv-visionfusion/proto/detections"
)

type DetectionEventController struct {
	rmqClient *messaging.RMQClient
	wg        *sync.WaitGroup
	service   *services.DetectionWorkflowService
}

func NewDetectionEventController(
	ctx context.Context,
	rmq *messaging.RMQClient,
	parentWG *sync.WaitGroup,
	service *services.DetectionWorkflowService,
) *DetectionEventController {
	var controller = DetectionEventController{
		rmqClient: rmq,
		wg:        parentWG,
		service:   service,
	}

	err := rmq.StartConsumer(
		ctx,
		parentWG,
		config.Config.RabbitMQ.QueueDetectionsCompleted,
		"cctv-BE",
		controller.detectionCompletedHandler,
	)
	if err != nil {
		slog.Error("failed to start consumer", "error", err)
	}

	return &controller
}

func (r *DetectionEventController) detectionCompletedHandler(msgCtx context.Context, body []byte) error {
	var tracer = otel.Tracer(pkgName)

	var msg detections.DetectionCompletedEvent
	if err := proto.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("proto unmarshal failed: %w", err)
	}

	detectionCtx, span := tracer.Start(
		msgCtx,
		"controller.detection.handling",
		trace.WithAttributes(
			attribute.String("event_id", msg.GetEventUuid()),
		),
	)
	defer span.End()
	slog.InfoContext(detectionCtx, "detection result received", "msq_raw", &msg)

	err := r.service.CompleteDetection(detectionCtx, &msg)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("save detections: %w", err)
	}

	return nil
}
