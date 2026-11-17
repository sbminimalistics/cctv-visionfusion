package bootstrap

import (
	"context"
	"log/slog"

	"sbminimalistics.com/cctv-visionfusion/internal/services"
)

func WireServices(ctx context.Context) {
	slog.Info("starting download link service ...")
	services.InitDownloadLinkService(ctx)
	slog.Info("services started")
}
