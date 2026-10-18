package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// Status represents the state of a single download task.
type Status string

const (
	StatusStarted   Status = "STARTED"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

// Progress report sent back to the user via the communication channel.
type DownloadProgress struct {
	URL      string
	SavePath string
	Status   Status
	Err      error
}

// DownloadJob defines an individual file download task.
type DownloadJob struct {
	URL      string
	SavePath string
}

// FileDownloader coordinates parallel file downloads.
type FileDownloader struct {
	workerCount int
	httpClient  *http.Client
}

// NewFileDownloader initializes the downloader with a specific concurrency limit.
func NewFileDownloader(workerCount int) *FileDownloader {
	if workerCount <= 0 {
		workerCount = 1
	}
	return &FileDownloader{
		workerCount: workerCount,
		httpClient:  &http.Client{},
	}
}

// DownloadBatch accepts a slice of jobs, downloads them concurrently up to the worker count,
// and streams live status updates to the returned channel. It closes the channel when finished.
func (fd *FileDownloader) DownloadBatch(ctx context.Context, jobs []DownloadJob) <-chan DownloadProgress {
	progressChan := make(chan DownloadProgress, len(jobs))
	jobChan := make(chan DownloadJob, len(jobs))

	// Queue up all the jobs
	for _, job := range jobs {
		jobChan <- job
	}
	close(jobChan)

	// Use a WaitGroup to track when all background workers finish
	var wg sync.WaitGroup

	// Spawn the bounded pool of workers
	for i := 0; i < fd.workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				// Check for context cancellation before starting a new job
				if err := ctx.Err(); err != nil {
					progressChan <- DownloadProgress{URL: job.URL, SavePath: job.SavePath, Status: StatusFailed, Err: err}
					continue
				}

				fd.downloadFile(ctx, job, progressChan)
			}
		}()
	}

	// Monitor workers in a separate goroutine to close the progress channel cleanly
	go func() {
		wg.Wait()
		close(progressChan)
	}()

	return progressChan
}

// downloadFile performs the actual HTTP GET request and file persistence.
func (fd *FileDownloader) downloadFile(ctx context.Context, job DownloadJob, progress chan<- DownloadProgress) {
	progress <- DownloadProgress{URL: job.URL, SavePath: job.SavePath, Status: StatusStarted}

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(job.SavePath), os.ModePerm); err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SavePath, Status: StatusFailed, Err: err}
		return
	}

	// Build the HTTP request to support context cancellation midway
	req, err := http.NewRequestWithContext(ctx, "GET", job.URL, nil)
	if err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SavePath, Status: StatusFailed, Err: err}
		return
	}

	resp, err := fd.httpClient.Do(req)
	if err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SavePath, Status: StatusFailed, Err: err}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		progress <- DownloadProgress{
			URL:      job.URL,
			SavePath: job.SavePath,
			Status:   StatusFailed,
			Err:      fmt.Errorf("bad status code: %d", resp.StatusCode),
		}
		return
	}

	// Create the destination file
	out, err := os.Create(job.SavePath)
	if err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SavePath, Status: StatusFailed, Err: err}
		return
	}
	defer out.Close()

	// Stream directly from the response body to the file on disk (low memory footprint)
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SavePath, Status: StatusFailed, Err: err}
		return
	}

	progress <- DownloadProgress{URL: job.URL, SavePath: job.SavePath, Status: StatusCompleted}
}
