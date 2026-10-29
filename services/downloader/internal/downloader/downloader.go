package downloader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	URL              string
	SaveDir          string
	SaveBaseFilename string
}

// FileDownloader coordinates parallel file downloads.
type FileDownloader struct {
	ctx         context.Context
	workerCount int
	httpClient  *http.Client
	progress    <-chan DownloadProgress
}

func (r *FileDownloader) GetProgress() <-chan DownloadProgress {
	return r.progress
}

// NewFileDownloader initializes the downloader with a specific concurrency limit.
func NewFileDownloader(ctx context.Context, workerCount int) *FileDownloader {
	if workerCount <= 0 {
		workerCount = 1
	}
	return &FileDownloader{
		ctx:         ctx,
		workerCount: workerCount,
		httpClient:  &http.Client{},
		progress:    make(<-chan DownloadProgress),
	}
}

// DownloadBatch accepts a slice of jobs, downloads them concurrently up to the worker count,
// and streams live status updates to the returned channel. It closes the channel when finished.
func (r *FileDownloader) DownloadBatch(jobs []DownloadJob) <-chan DownloadProgress {
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
	for i := 0; i < r.workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				// Check for context cancellation before starting a new job
				if err := r.ctx.Err(); err != nil {
					progressChan <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusFailed, Err: err}
					continue
				}

				r.downloadFile(r.ctx, job, progressChan)
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
	progress <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusStarted}

	// Build the HTTP request to support context cancellation midway
	req, err := http.NewRequestWithContext(ctx, "GET", job.URL, nil)
	if err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusFailed, Err: err}
		return
	}

	resp, err := fd.httpClient.Do(req)
	if err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusFailed, Err: err}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		progress <- DownloadProgress{
			URL:      job.URL,
			SavePath: job.SaveDir,
			Status:   StatusFailed,
			Err:      fmt.Errorf("bad status code: %d", resp.StatusCode),
		}
		return
	}

	br := bufio.NewReader(resp.Body)
	if err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusFailed, Err: err}
		return
	}

	headerBytes, err := br.Peek(512)
	if err != nil && err != io.EOF {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusFailed, Err: err}
		return
	}

	ext, err := getFileExtension(&headerBytes)
	if err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusFailed, Err: err}
		return
	}
	finalPath := path.Join(job.SaveDir, fmt.Sprintf("%s%s", job.SaveBaseFilename, ext))

	// ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(finalPath), os.ModePerm); err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusFailed, Err: err}
		return
	}

	// create the destination file
	out, err := os.Create(
		path.Join(job.SaveDir, fmt.Sprintf("%s%s", job.SaveBaseFilename, ext)),
	)
	if err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusFailed, Err: err}
		return
	}
	defer out.Close()

	// Stream directly from the response body to the file on disk (low memory footprint)
	_, err = io.Copy(out, br)
	if err != nil {
		progress <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusFailed, Err: err}
		return
	}

	progress <- DownloadProgress{URL: job.URL, SavePath: job.SaveDir, Status: StatusCompleted}
}

func getFileExtension(header *[]byte) (string, error) {
	contentType := http.DetectContentType(*header)

	// return contentType, nil

	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		// Fallback for custom overrides or unrecognized binaries
		return "", err
	}

	ctParts := strings.Split(contentType, "/")
	var pick string
	for _, ext := range extensions {
		e, ok := strings.CutPrefix(ext, ".")
		if ok && e == ctParts[1] {
			pick = ext
			break
		}
	}

	if pick == "" {
		pick = extensions[0]
	}
	return pick, nil
}
