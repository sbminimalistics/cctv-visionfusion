package downloader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/minio/minio-go/v7"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/config"
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/services"
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
	EventUUID   uuid.UUID
	SavePrefix  string
	ObjectName  string
	ContentType string
	Extension   string
	Status      Status
	Err         error
}

// DownloadJob defines an individual file download task.
type DownloadJob struct {
	EventUUID        uuid.UUID
	URL              string
	SavePrefix       string
	SaveBaseFilename string
}

// FileDownloader coordinates parallel file downloads.
type FileDownloader struct {
	ctx         context.Context
	workerCount int
	httpClient  *http.Client
	progress    chan DownloadProgress
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
		progress:    make(chan DownloadProgress),
	}
}

// DownloadBatch accepts a slice of jobs, downloads them concurrently up to the worker count,
// and streams live status updates to the returned channel. It closes the channel when finished.
func (r *FileDownloader) DownloadBatch(jobs []DownloadJob) {
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
					r.progress <- DownloadProgress{EventUUID: job.EventUUID, SavePrefix: job.SavePrefix, Status: StatusFailed, Err: err}
					continue
				}

				r.downloadFile(r.ctx, job)
			}
		}()
	}

	// Monitor workers in a separate goroutine to close the progress channel cleanly
	go func() {
		wg.Wait()
		// close(r.progress)
	}()
}

// downloadFile performs the actual HTTP GET request and file persistence.
func (fd *FileDownloader) downloadFile(ctx context.Context, job DownloadJob) {
	// fmt.Println("downloadFile", job)
	fd.progress <- DownloadProgress{EventUUID: job.EventUUID, SavePrefix: job.SavePrefix, Status: StatusStarted}

	// build the HTTP request to support context cancellation midway
	req, err := http.NewRequestWithContext(ctx, "GET", job.URL, nil)
	if err != nil {
		fd.progress <- DownloadProgress{EventUUID: job.EventUUID, SavePrefix: job.SavePrefix, Status: StatusFailed, Err: err}
		return
	}

	resp, err := fd.httpClient.Do(req)
	if err != nil {
		fd.progress <- DownloadProgress{EventUUID: job.EventUUID, SavePrefix: job.SavePrefix, Status: StatusFailed, Err: err}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fd.progress <- DownloadProgress{
			EventUUID:  job.EventUUID,
			SavePrefix: job.SavePrefix,
			Status:     StatusFailed,
			Err:        fmt.Errorf("bad status code: %d", resp.StatusCode),
		}
		return
	}

	br := bufio.NewReader(resp.Body)
	if err != nil {
		fd.progress <- DownloadProgress{EventUUID: job.EventUUID, SavePrefix: job.SavePrefix, Status: StatusFailed, Err: err}
		return
	}

	headerBytes, err := br.Peek(512)
	if err != nil && err != io.EOF {
		fd.progress <- DownloadProgress{EventUUID: job.EventUUID, SavePrefix: job.SavePrefix, Status: StatusFailed, Err: err}
		return
	}

	contentType, ext, err := getTypeAndExtension(&headerBytes)
	if err != nil {
		fd.progress <- DownloadProgress{EventUUID: job.EventUUID, SavePrefix: job.SavePrefix, Status: StatusFailed, Err: err}
		return
	}

	objectName := fmt.Sprintf("%s%s%s", job.SavePrefix, job.SaveBaseFilename, ext)
	// fmt.Printf("Target Object Key: '%s'\n", objectName)

	objectTags := map[string]string{
		"processed": "false",
	}
	_, err = services.GetS3Service().PutObject(ctx, config.Config.Seaweed.Bucket, objectName, br, -1, minio.PutObjectOptions{
		ContentType: contentType,
		PartSize:    5 * 1024 * 1024, // 5MB internal buffer chunks for dynamic streaming
		UserTags:    objectTags,
	})

	if err != nil {
		fmt.Println("s3 save failed, error:", err, config.Config.Seaweed.Bucket)
		fd.progress <- DownloadProgress{EventUUID: job.EventUUID, SavePrefix: job.SavePrefix, Status: StatusFailed, Err: err}
		return
	}

	fmt.Printf("s3 save %s success\n", objectName)
	fd.progress <- DownloadProgress{
		EventUUID:   job.EventUUID,
		SavePrefix:  job.SavePrefix,
		ObjectName:  job.EventUUID.String(),
		Status:      StatusCompleted,
		ContentType: contentType,
		Extension:   ext,
	}
}

func getTypeAndExtension(header *[]byte) (string, string, error) {
	contentType := http.DetectContentType(*header)

	// return contentType, nil

	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		// Fallback for custom overrides or unrecognized binaries
		return contentType, "", err
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
	return contentType, pick, nil
}
