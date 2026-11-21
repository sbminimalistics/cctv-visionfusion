package bootstrap

import (
	"context"
	"log/slog"

	"sbminimalistics.com/cctv-visionfusion/internal/db"
	"sbminimalistics.com/cctv-visionfusion/internal/services"
)

type Services struct {
	DownloadLinkService *services.DownloadLinkService
	EventService        *services.EventService
}

func WireServices(ctx context.Context, store *db.Store) Services {
	slog.Info("starting download link service ...")
	dlService := services.InitDownloadLinkService(ctx)
	eventService := services.InitEventService(store)
	slog.Info("services started")

	return Services{
		DownloadLinkService: dlService,
		EventService:        eventService,
	}
}
