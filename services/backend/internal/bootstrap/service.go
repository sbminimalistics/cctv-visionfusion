package bootstrap

import (
	"context"
	"log/slog"

	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/db"
	"sbminimalistics.com/cctv-visionfusion/internal/messaging"
	"sbminimalistics.com/cctv-visionfusion/internal/services"
)

type Services struct {
	DownloadLinkService *services.DownloadLinkService
	EventService        *services.EventService
	DetectionService    *services.DetectionService
	DetectionWFService  *services.DetectionWorkflowService
}

func WireServices(ctx context.Context, store *db.Store, rmqClient *messaging.RMQClient) Services {
	slog.Info("starting download link service ...")

	dlService := services.InitDownloadLinkService(ctx)
	eventService := services.InitEventService(store, dlService, rmqClient, config.Config.Routing)
	detectionService := services.InitDetectionService(store)
	detectionWFService := services.InitDetectionWorkflowService(store, detectionService, eventService)
	slog.Info("services started")

	return Services{
		DownloadLinkService: dlService,
		EventService:        eventService,
		DetectionService:    detectionService,
		DetectionWFService:  detectionWFService,
	}
}
