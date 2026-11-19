package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"google.golang.org/protobuf/proto"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/bootstrap"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/config"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/downloader"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/mappers"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/messaging"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/services"
	"sbminimalistics.com/cctv-visionfusion/downloader/proto/downloads"
)

func main() {
	defer bootstrap.InitTelemetry("cctv-DWNLDR")()

	slog.Info("starting CCTV VisionFusion downloader", "backend_host", config.Config.BackendApiHost, "port", config.Config.BackendApiPort)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	services.InitS3Service(ctx)

	var wg sync.WaitGroup

	// initialize RabbitMQ client
	rmqClient := messaging.NewRMQClient(
		config.Config.RabbitMQ.User,
		config.Config.RabbitMQ.Password,
		config.Config.RabbitMQ.Host,
		int(config.Config.RabbitMQ.Port),
		config.Config.RabbitMQ.VHost,
	)

	// initialize downloader allowing a maximum of 2 downloads in parallel
	dl := downloader.NewFileDownloader(ctx, 2)
	progress := dl.GetProgress()

	slog.Info("connecting to RabbitMQ broker instance...")
	if err := rmqClient.Connect(); err != nil {
		slog.Error("RabbitMQ connection failed to initialize", "error", err)
	} else {
		slog.Info("RabbitMQ transport connection pipeline live.")
		defer rmqClient.Close()

		err := rmqClient.StartConsumer(
			ctx,
			&wg,
			config.Config.RabbitMQQueue,
			"cctv-DWNLDR",
			func(msgCtx context.Context, body []byte) error {
				var msg downloads.DownloadInstruction
				if err := proto.Unmarshal(body, &msg); err != nil {
					return fmt.Errorf("proto unmarshal failed: %w", err)
				}
				slog.InfoContext(msgCtx, "download request received")

				data, err := downloader.ExchangeToken(msgCtx, msg.RedeemPath, msg.Token)
				if err != nil {
				} else {
					savePrefix := path.Join(
						config.Config.PathPrefix,
						data.DestinationPath,
					) + "/"
					dl.DownloadBatch([]downloader.DownloadJob{
						{
							JobCtx:           msgCtx,
							EventUUID:        data.EventUUID,
							URL:              data.PictureURL,
							SavePrefix:       savePrefix,
							SaveBaseFilename: data.EventUUID.String(),
						},
					})
				}

				return nil
			},
		)
		if err != nil {
			slog.Error("failed to start consumer", "error", err)
		}

		/* ch, err := rmqClient.Consume(ctx, config.Config.RabbitMQQueue, "downloader")

		if err != nil {
			slog.Error("call for consume failed", "error", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
		messageReadLoop:
			for {
				select {
				case <-ctx.Done():
					break messageReadLoop
				case src := <-ch:
					err := rmqClient.Ack(src.DeliveryTag, false)
					if err != nil {
						continue
					}
					var msg downloads.DownloadInstruction
					proto.Unmarshal(src.Body, &msg)
					if src.Body == nil {
						continue
					}

					data, err := downloader.ExchangeToken(msg.RedeemPath, msg.Token)
					if err != nil {
					} else {
						savePrefix := path.Join(
							config.Config.PathPrefix,
							data.DestinationPath,
						) + "/"
						dl.DownloadBatch([]downloader.DownloadJob{
							{
								EventUUID:        data.EventUUID,
								URL:              data.PictureURL,
								SavePrefix:       savePrefix,
								SaveBaseFilename: data.EventUUID.String(),
							},
						})
					}
				}
			}
		}() */
	}

	// listen to the progress reports over the channel
downloaderProgressLoop:
	for {
		select {
		case <-ctx.Done():
			break downloaderProgressLoop

		case update := <-progress:
			switch update.Status {
			case downloader.StatusStarted:
				slog.InfoContext(update.JobCtx, "started downloading", "event_id", update.EventUUID.String())

			case downloader.StatusCompleted:
				slog.InfoContext(update.JobCtx, "finished downloading", "event_id", update.EventUUID, "saved_to", update.SavePrefix)

				msg := mappers.DownloadProgressToProto(update)
				bytes, _ := proto.Marshal(msg)

				err := rmqClient.Publish(update.JobCtx,
					"downloads.requests.complete",
					"downloads.complete",
					bytes,
				)
				if err != nil {
					slog.ErrorContext(update.JobCtx, "message publication failed", "error", err)
				}

			case downloader.StatusFailed:
				slog.ErrorContext(update.JobCtx, "failed downloading", "event_id", update.EventUUID, "error", update.Err)
			}
		}
	}

	<-ctx.Done()
	slog.Info("shutdown initiated, waiting for edge streams to exit...")

	wg.Wait()
	slog.Info("closing infrastructure...")
}
