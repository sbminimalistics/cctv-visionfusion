package edge_device

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
)

type pkgAnchor struct{}

type DeviceEventType string

type EventPacket struct {
	Ctx                 context.Context
	uuid                uuid.UUID
	absolutePicturePath string
	DeviceID            string
	Manufacturer        config.Manufacturer
	Raw                 []byte
	Parsed              BasicParsedMessage
}

var pkgName = reflect.TypeOf(pkgAnchor{}).PkgPath()

func (r *EventPacket) GetUUID() uuid.UUID {
	return r.uuid
}

func (r *EventPacket) SetUUID(uuid uuid.UUID) {
	r.uuid = uuid
}

func (r *EventPacket) GetAbsolutePicturePath() string {
	return r.absolutePicturePath
}

type SSEListener struct {
	DeviceID string                  //value from config
	Config   config.EdgeDeviceConfig //raw config
	parser   PacketParser
	events   chan<- EventPacket
}

func NewSSEListener(cfg config.EdgeDeviceConfig, parser PacketParser) *SSEListener {
	return &SSEListener{
		DeviceID: cfg.Id,
		Config:   cfg,
		parser:   parser,
	}
}

func (r *SSEListener) ConnectAndRetryForever(ctx context.Context, deviceConfig config.EdgeDeviceConfig) {
	url := fmt.Sprintf("http://%s:%d/ISAPI/Event/notification/alertStream", deviceConfig.Host, deviceConfig.Port)
	username := deviceConfig.User
	password := deviceConfig.Password

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,

			// max time the TCP connection can stay idle if no data passes through
			IdleConnTimeout: 5 * time.Minute,
		},

		// remove the global client timeout entirely (set to 0)
		// so the stream can stay open indefinitely.
		Timeout: 0,
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("[SSE] exiting connection routine loop", "device_id", r.DeviceID)
			return
		default:
		}

		slog.Info("[SSE] attempting to open Hikvision ISAPI alert stream...", "device_id", r.DeviceID)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			slog.Warn("[SSE] failed to build request schema", "device_id", r.DeviceID, "error", err)
			goto RETRY_DELAY
		}

		req.SetBasicAuth(username, password)
		req.Header.Set("Accept", "multipart/x-mixed-replace")

		if resp, err := client.Do(req); err == nil {
			err = r.readAlertStream(ctx, resp)
			if err != nil {
				slog.Warn("[SSE] loop broken", "device_id", r.DeviceID, "error", err)
			}
		} else {
			slog.Warn("[SSE] connection to device failed", "device_id", r.DeviceID, "error", err)
		}

	RETRY_DELAY:
		slog.Info("[SSE] re-attempting connection link in 5 seconds", "device_id", r.DeviceID)
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (r *SSEListener) readAlertStream(ctx context.Context, resp *http.Response) error {
	var tracer = otel.Tracer(pkgName)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server rejected request with status: %s", resp.Status)
	}

	_, params, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	boundary := params["boundary"]
	if boundary == "" {
		boundary = "boundary"
	}

	slog.Info("[SSE] connected to device", "device_id", r.DeviceID, "stream_boundary", boundary)
	multipartReader := multipart.NewReader(resp.Body, boundary)

	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			return fmt.Errorf("[SSE] multipart stream closed cleanly by device host")
		}
		if err != nil {
			return fmt.Errorf("[SSE] error navigating boundary tracking layer: %w", err)
		}

		eventID, _ := uuid.NewV7()
		eventCtx, span := tracer.Start(
			ctx,
			"edge.alert.received",
			trace.WithNewRoot(),
			trace.WithAttributes(
				attribute.String("device_id", r.DeviceID),
				attribute.String("event_id", eventID.String()),
			),
		)

		data, err := io.ReadAll(part)
		part.Close()
		if err != nil {
			slog.WarnContext(eventCtx, "[SSE] frame extraction failed", "device_id", r.DeviceID, "error", err)
			span.End()
			continue
		}

		if data == nil || r.parser == nil {
			span.End()
			continue
		}

		fields, err := r.parser.Parse(eventCtx, data)

		if err != nil {
			slog.Warn("[SSE] error parsing packet", "error", err)
		}

		var packet = EventPacket{
			Ctx:  eventCtx,
			uuid: eventID,
			absolutePicturePath: fmt.Sprintf(
				"%s://%s:%s@%s:%d%s",
				r.Config.Protocol,
				r.Config.User,
				r.Config.Password,
				r.Config.Host,
				r.Config.Port,
				r.Config.PicturePath,
			),
			DeviceID:     r.DeviceID,
			Manufacturer: r.Config.Manufacturer,
			Raw:          data,
			Parsed:       fields,
		}

		if r.events != nil {
			slog.InfoContext(eventCtx, "[SSE] device event queued", "device_id", r.DeviceID, "event_id", eventID.String())
			r.events <- packet
		} else {
			slog.WarnContext(eventCtx, "[SSE] device data dropped, listener not registered to controller yet", "device_id", r.DeviceID)
		}

		span.End()
	}
}

func (r *SSEListener) AttachChannel(pipe chan<- EventPacket) {
	r.events = pipe
}

func (r *SSEListener) Events() chan<- EventPacket {
	return r.events
}
