package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetDownloadLinkService(t *testing.T) {
	ctx := t.Context()

	InitDownloadLinkService(ctx)

	instance0 := GetDownloadLinkService()
	instance1 := GetDownloadLinkService()

	if instance0 != instance1 {
		t.Errorf("expected both instances to be identical pointers, but got %p and %p", instance0, instance1)
	}
}

func Test_DownloadLinkService_Lifecycle(t *testing.T) {
	testURL := "https://example.com"
	testDestPath := "/directory"
	testDestBaseFileName := "file0"

	ctx := t.Context()

	InitDownloadLinkService(ctx)

	service := GetDownloadLinkService()

	var token uuid.UUID

	t.Run("create new token", func(t *testing.T) {
		token_, expiresAt, err := service.GetNewToken(testURL, testDestPath, testDestBaseFileName)

		token = token_

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if expiresAt.Before(time.Now()) {
			t.Error("extected expiry time to by after current time")
		}
	})

	t.Run("redeem link based on token", func(t *testing.T) {
		link, ok := service.GetLink(token)

		if !ok {
			t.Error("expected to redeem the link by using earlier created token")
		}

		if link.GetURL() != testURL {
			t.Error("expected to get the same url provided during token creation")
		}

		if link.GetDestinationPath() != testDestPath {
			t.Error("expected to get the same destination path provided during token creation")
		}

		if link.GetDestinationBaseFileName() != testDestBaseFileName {
			t.Error("expected to get the same destination base file name provided during token creation")
		}
	})

	t.Run("check if link was removed upon link redeem", func(t *testing.T) {
		linkCount := len(service.links)
		if linkCount > 0 {
			t.Errorf("expected to find an empty links slice in service, found %d items", linkCount)
		}
	})
}
