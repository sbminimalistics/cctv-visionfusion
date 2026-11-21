package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	DefaultLinkTTL                = time.Duration(time.Second * 30)
	downloadLinkServiceInstance   *DownloadLinkService
	createDownloadLinkServiceOnce sync.Once
)

type LinkWithExpiry struct {
	token           uuid.UUID
	createdAt       time.Time
	expiresAt       time.Time
	url             string
	destinationPath string
	eventUUID       uuid.UUID
}

func (r *LinkWithExpiry) GetToken() uuid.UUID {
	return r.token
}

func (r *LinkWithExpiry) GetCreatedAt() time.Time {
	return r.createdAt
}

func (r *LinkWithExpiry) GetExpiresAt() time.Time {
	return r.expiresAt
}

func (r *LinkWithExpiry) GetURL() string {
	return r.url
}

func (r *LinkWithExpiry) GetDestinationPath() string {
	return r.destinationPath
}

func (r *LinkWithExpiry) GetEventUUID() uuid.UUID {
	return r.eventUUID
}

type DownloadLinkService struct {
	ctx   context.Context
	mu    sync.Mutex
	links []LinkWithExpiry
}

func (r *DownloadLinkService) GetLink(ctx context.Context, token uuid.UUID) (LinkWithExpiry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for idx, link := range r.links {
		if link.token == token {
			copy(r.links[idx:], r.links[idx+1:])
			r.links[len(r.links)-1] = LinkWithExpiry{}
			r.links = r.links[:len(r.links)-1]
			return link, nil
		}
	}
	return LinkWithExpiry{}, errors.New("link not found")
}

func (r *DownloadLinkService) GetNewToken(url, destPath string, eventUUID uuid.UUID) (uuid.UUID, time.Time, error) {
	now := time.Now().UTC()
	id, err := uuid.NewV7()

	if err != nil {
		return uuid.UUID{}, time.Time{}, err
	}

	link := LinkWithExpiry{
		token:           id,
		createdAt:       now,
		expiresAt:       now.Add(DefaultLinkTTL),
		url:             url,
		destinationPath: destPath,
		eventUUID:       eventUUID,
	}

	r.links = append(r.links, link)

	return id, link.GetExpiresAt(), nil
}

func GetDownloadLinkService() *DownloadLinkService {
	if downloadLinkServiceInstance == nil {
		panic("DownloadLinkService requested before initialization")
	}
	return downloadLinkServiceInstance
}

func InitDownloadLinkService(ctx context.Context) *DownloadLinkService {
	createDownloadLinkServiceOnce.Do(func() {
		downloadLinkServiceInstance = &DownloadLinkService{
			ctx:   ctx,
			links: make([]LinkWithExpiry, 0),
		}

		go downloadLinkServiceInstance.startCleanupWorker()
	})
	return downloadLinkServiceInstance
}

func (r *DownloadLinkService) startCleanupWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.cleanExpiredLinks()
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *DownloadLinkService) cleanExpiredLinks() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	keepIdx := 0

	for i := 0; i < len(r.links); i++ {
		if !r.links[i].expiresAt.Before(now) {
			r.links[keepIdx] = r.links[i]
			keepIdx++
		}
	}

	// zero out the remaining stale slots so the gc can free them
	for i := keepIdx; i < len(r.links); i++ {
		r.links[i] = LinkWithExpiry{}
	}

	// shrink the slice to only contain the kept items
	r.links = r.links[:keepIdx]
}
