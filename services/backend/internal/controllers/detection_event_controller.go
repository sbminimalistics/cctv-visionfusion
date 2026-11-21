package controllers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"google.golang.org/protobuf/proto"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/messaging"
	"sbminimalistics.com/cctv-visionfusion/internal/services"
	"sbminimalistics.com/cctv-visionfusion/proto/detections"
)

type DetectionEventController struct {
	rmqClient *messaging.RMQClient
	wg        *sync.WaitGroup
	service   *services.EventService
}

func NewDetectionEventController(
	ctx context.Context,
	rmq *messaging.RMQClient,
	parentWG *sync.WaitGroup,
	service *services.EventService,
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
	var msg detections.DetectionCompletedEvent
	if err := proto.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("proto unmarshal failed: %w", err)
	}
	slog.InfoContext(msgCtx, "detection result received", "msq_raw", &msg)

	// r.service.SaveEvent(msgCtx, msg)

	return nil
}
