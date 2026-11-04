package mappers

import (
	"sbminimalistics.com/cctv-visionfusion/downloader/internal/downloader"
	"sbminimalistics.com/cctv-visionfusion/downloader/proto/downloads"
)

func DownloadProgressToProto(src downloader.DownloadProgress) *downloads.DownloadCompleteEvent {
	return &downloads.DownloadCompleteEvent{
		EventUuid:   src.EventUUID.String(),
		Prefix:      src.SavePrefix,
		ObjectName:  src.ObjectName,
		ContentType: src.ContentType,
		Extension:   src.Extension,
	}
}
