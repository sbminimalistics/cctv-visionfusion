package mappers

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/db/model"
	"sbminimalistics.com/cctv-visionfusion/internal/edge_device"
	"sbminimalistics.com/cctv-visionfusion/internal/event"
	"sbminimalistics.com/cctv-visionfusion/proto/detections"
	"sbminimalistics.com/cctv-visionfusion/proto/downloads"
	"sbminimalistics.com/cctv-visionfusion/proto/events"
)

type DownloadInstructionBase struct {
	Token       uuid.UUID
	TokenExpiry time.Time
	RedeemPath  string
}

// maps a domain EventPacket into its corresponding Protobuf structured message payload.
func DeviceEventToProto(packet edge_device.EventPacket) *events.DeviceEvent {
	return &events.DeviceEvent{
		Uuid:         packet.GetUUID().String(),
		DeviceId:     packet.DeviceID,
		Manufacturer: string(packet.Manufacturer),
		EventType:    edge_device.BasicParsedMessage(packet.Parsed).GetEventType(),
		EventTime:    timestamppb.Now(), // Utilizes the required Google Protobuf timestamp type
	}
}

// maps a domain CompositionTriggerErvent into its corresponding Protobuf structured message payload.
func CompositionTriggerToProto(packet event.CompositionTriggerEvent) *events.CompositeEvent {
	sources := toProtoSources(packet.Sources)
	return &events.CompositeEvent{
		Uuid:      packet.GetUUID(),
		Id:        packet.CompositionID,
		Sources:   sources,
		EventTime: timestamppb.Now(),
	}
}

// maps a domain DownloadInstructionBase struct into its corresponding Protobuf structured message payload.
func DeviceInstructionToProto(src DownloadInstructionBase) *downloads.DownloadInstruction {
	return &downloads.DownloadInstruction{
		EventTime:   timestamppb.Now(),
		Token:       src.Token.String(),
		TokenExpiry: timestamppb.New(src.TokenExpiry),
		RedeemPath:  src.RedeemPath,
	}
}

func toProtoSources(configs []config.CompositionSourceConfig) []*events.CompositeEventSource {
	protos := make([]*events.CompositeEventSource, len(configs))

	for i := range configs {
		protos[i] = &events.CompositeEventSource{
			DeviceId:  configs[i].DeviceID,
			EventType: configs[i].EventType,
		}
	}
	return protos
}

func StoredEventToProto(evCaptured *model.CapturedEvent, detections *detections.DetectionCompletedEvent) *events.DeviceEvent {
	var event = events.DeviceEvent{
		Uuid:         evCaptured.ID,
		DeviceId:     *evCaptured.Src,
		Manufacturer: *evCaptured.SrcType,
		EventType:    *evCaptured.Type,
		EventTime:    timestamppb.New(evCaptured.CreatedAt),
	}

	event.Detection = &events.Detection{
		Frame:      detections.Frame,
		Detections: detections.Detections,
	}
	return &event
}
