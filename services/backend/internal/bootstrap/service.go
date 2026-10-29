package bootstrap

import (
	"context"
	"fmt"

	"sbminimalistics.com/cctv-visionfusion/internal/services"
)

func WireServices(ctx context.Context) {
	fmt.Println("starting services ...")
	fmt.Println("starting download link service ...")
	services.InitDownloadLinkService(ctx)
	fmt.Println("services started")
}
