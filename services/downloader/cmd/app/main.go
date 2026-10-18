package main

import (
	"context"
	"fmt"
	"time"

	"sbminimalistics.com/downloader/internal/downloader"
)

func main() {
	// Create a context that we can use to cancel downloads if needed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Define tasks to process
	jobs := []downloader.DownloadJob{
		{URL: "https://example.com", SavePath: "./downloads/file1.zip"},
		{URL: "https://example.com", SavePath: "./downloads/file2.png"},
		{URL: "https://example.com", SavePath: "./downloads/file3.pdf"},
	}

	// Initialize downloader allowing a maximum of 2 downloads in parallel
	dl := downloader.NewFileDownloader(2)

	// Start downloading!
	progressUpdates := dl.DownloadBatch(ctx, jobs)

	// Actively listen to the progress reports over the channel
	for update := range progressUpdates {
		switch update.Status {
		case downloader.StatusStarted:
			fmt.Printf("[→] Started downloading: %s\n", update.URL)
		case downloader.StatusCompleted:
			fmt.Printf("[✓] Finished downloading: %s saved to %s\n", update.URL, update.SavePath)
		case downloader.StatusFailed:
			fmt.Printf("[X] Failed downloading %s: %v\n", update.URL, update.Err)
		}
	}

	fmt.Println("All batch download operations completed.")
}
