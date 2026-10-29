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
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/messaging"
	"sbminimalistics.com/cctv-visionfusion/downloader/proto/downloads"
)

func main() {
	fmt.Println("starting CCTV VisionFusion downloader ...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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
						destDirPath := path.Join(
							"./downloads",
							data.DestinationPath,
							"/",
						)
						dl.DownloadBatch([]downloader.DownloadJob{
							{
								URL:              data.PictureURL,
								SaveDir:          destDirPath,
								SaveBaseFilename: data.DestinationBaseFilename,
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
				fmt.Printf("[→] Started downloading: %s\n", update.URL)
			case downloader.StatusCompleted:
				fmt.Printf("[✓] Finished downloading: %s saved to %s\n", update.URL, update.SavePath)
			case downloader.StatusFailed:
				fmt.Printf("[X] Failed downloading %s: %v\n", update.URL, update.Err)
			}
		}
	}

	<-ctx.Done()
	fmt.Println("Shutdown initiated. Waiting for edge streams to exit...")

	wg.Wait()
	fmt.Println("All edge devices disconnected. Closing infrastructure...")
}
