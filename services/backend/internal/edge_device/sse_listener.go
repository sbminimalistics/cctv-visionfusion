package edge_device

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"time"

	"sbminimalistics.com/cctv-visionfusion/internal/config"
)

type DeviceEventType string

type EventPacket struct {
	uuid         string
	DeviceID     string
	Manufacturer config.Manufacturer
	Raw          []byte
	Parsed       BasicParsedMessage
}

func (r *EventPacket) GetUUID() string {
	return r.uuid
}

func (r *EventPacket) SetUUID(uuid string) {
	r.uuid = uuid
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
	url := fmt.Sprintf("http://%s/ISAPI/Event/notification/alertStream", deviceConfig.Host)
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
			fmt.Printf("[SSE] exiting connection routine loop for %s\n", r.DeviceID)
			return
		default:
		}

		fmt.Printf("[SSE] attempting to open Hikvision ISAPI alert stream for %s...\n", r.DeviceID)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			log.Printf("[SSE] failed to build request schema for %s: %v\n", r.DeviceID, err)
			goto RETRY_DELAY
		}

		req.SetBasicAuth(username, password)
		req.Header.Set("Accept", "multipart/x-mixed-replace")

		if resp, err := client.Do(req); err == nil {
			err = r.readAlertStream(ctx, resp)
			if err != nil {
				log.Printf("[SSE] loop broken for camera %s: %v\n", r.DeviceID, err)
			}
		} else {
			log.Printf("[SSE] connection to camera %s failed: %v\n", r.DeviceID, err)
		}

	RETRY_DELAY:
		fmt.Printf("[SSE] device %s resting, re-attempting connection link in 5 seconds...\n", r.DeviceID)
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (r *SSEListener) readAlertStream(ctx context.Context, resp *http.Response) error {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server rejected request with status: %s", resp.Status)
	}

	_, params, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	boundary := params["boundary"]
	if boundary == "" {
		boundary = "boundary"
	}

	fmt.Printf("[SSE] connected to %s, tracking stream boundary: %s\n", r.DeviceID, boundary)
	multipartReader := multipart.NewReader(resp.Body, boundary)

	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			return fmt.Errorf("[SSE] multipart stream closed cleanly by device host")
		}
		if err != nil {
			return fmt.Errorf("[SSE] error navigating boundary tracking layer: %w", err)
		}

		data, err := io.ReadAll(part)
		part.Close()
		if err != nil {
			log.Printf("[SSE] frame extraction failed on camera %s: %v\n", r.DeviceID, err)
			continue
		}

		if data == nil {
			continue
		}

		if r.parser == nil {
			continue
		}

		fields, err := r.parser.Parse(data)

		if err != nil {
			fmt.Println("[SSE] error parsing packet:", err)
		}

		var packet = EventPacket{
			DeviceID:     r.DeviceID,
			Manufacturer: r.Config.Manufacturer,
			Raw:          data,
			Parsed:       fields,
		}

		if r.events != nil {
			r.events <- packet
		} else {
			fmt.Printf("[SSE] device %s: data dropped, listener not registered to controller yet\n", r.DeviceID)
		}
	}
}

func (r *SSEListener) AttachChannel(pipe chan<- EventPacket) {
	r.events = pipe
}

func (r *SSEListener) Events() chan<- EventPacket {
	return r.events
}
