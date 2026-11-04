package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"google.golang.org/protobuf/proto"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/config"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/downloader"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/mappers"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/messaging"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/services"
	"sbminimalistics.com/cctv-visionfusion/downloader/proto/downloads"
)

func main() {
	fmt.Println("starting CCTV VisionFusion downloader ...")

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

	fmt.Println("connecting to RabbitMQ broker instance...")
	if err := rmqClient.Connect(); err != nil {
		fmt.Printf("RabbitMQ connection failed to initialize: %v", err)
	} else {
		fmt.Println("RabbitMQ transport connection pipeline live.")

		ch, err := rmqClient.Consume(ctx, config.Config.RabbitMQQueue, "downloader")

		if err != nil {
			fmt.Printf("call for consume failed: %w", err)
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
		}()
	}
	defer rmqClient.Close()

	// listen to the progress reports over the channel
downloaderProgressLoop:
	for {
		select {
		case <-ctx.Done():
			break downloaderProgressLoop

		case update := <-progress:
			switch update.Status {
			case downloader.StatusStarted:
				fmt.Printf("started downloading: %s\n", update.EventUUID)

			case downloader.StatusCompleted:
				fmt.Printf("finished downloading: %s saved to %s\n", update.EventUUID, update.SavePrefix)

				msg := mappers.DownloadProgressToProto(update)
				bytes, _ := proto.Marshal(msg)

				err := rmqClient.Publish(ctx,
					"downloads.requests.complete",
					"downloads.complete",
					bytes,
				)
				if err != nil {
					fmt.Println("message publication failed:", err)
				}

			case downloader.StatusFailed:
				fmt.Printf("failed downloading %s: %v\n", update.EventUUID, update.Err)
			}
		}
	}

	<-ctx.Done()
	fmt.Println("shutdown initiated, waiting for edge streams to exit...")

	wg.Wait()
	fmt.Println("all edge devices disconnected, closing infrastructure...")
}
