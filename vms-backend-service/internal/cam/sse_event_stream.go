package cam

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"strings"
	"time"

	"sbminimalistics.com/vms/internal/config"
)

type EventData struct {
	XMLName xml.Name `xml:"EventNotificationAlert"`

	Name             string `xml:"id"`
	IPAddress        string `xml:"ipAddress"`
	EventType        string `xml:"eventType"`
	EventDescription string `xml:"eventDescription"`
	ActivePostCount  int    `xml:"activePostCount"`
}

type EventNotificationAlert struct {
	Name      string //value from config
	IPAddress string //value from config
	events    chan EventData
}

func NewStreamInstance(cfg config.CameraConfig) *EventNotificationAlert {
	return &EventNotificationAlert{
		Name:      cfg.Name,
		IPAddress: cfg.IP,
		events:    make(chan EventData, 10),
	}
}

func (r *EventNotificationAlert) ConnectAndRetryForever(ctx context.Context, camConfig config.CameraConfig) {
	url := fmt.Sprintf("http://%s/ISAPI/Event/notification/alertStream", camConfig.IP)
	username := camConfig.User
	password := camConfig.Password

	/* client := &http.Client{
		Timeout: 30 * time.Second,
	} */

	client := &http.Client{
		Transport: &http.Transport{
			// Time allowed to establish the initial network handshake
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,

			// Max time the TCP connection can stay idle if no data passes through
			// Hikvision devices usually send silent heartbeats every few seconds.
			// Setting this to 5 minutes gives ample breathing room.
			IdleConnTimeout: 5 * time.Minute,
		},

		// CRUCIAL: Remove the global client timeout entirely (set to 0)
		// so the stream can stay open indefinitely.
		Timeout: 0,
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Printf(" [SHUTDOWN] Exiting connection routine loop for %s\n", r.Name)
			return
		default:
		}

		fmt.Printf(" [NETWORK] Attempting to open Hikvision ISAPI alert stream for %s...\n", r.Name)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			log.Printf(" [ERROR] Failed to build request schema for %s: %v\n", r.Name, err)
			goto RETRY_DELAY
		}

		req.SetBasicAuth(username, password)
		req.Header.Set("Accept", "multipart/x-mixed-replace")

		if resp, err := client.Do(req); err == nil {
			// Pass the successful response body straight down to your clean reader method
			err = r.readAlertStream(ctx, resp)
			if err != nil {
				log.Printf(" [STREAM DISCONNECTED] Loop broken for camera %s: %v\n", r.Name, err)
			}
		} else {
			log.Printf(" [WARN] Connection to camera %s failed: %v\n", r.Name, err)
		}

	RETRY_DELAY:
		fmt.Printf(" [RETRY] Camera %s resting. Re-attempting connection link in 5 seconds...\n", r.Name)
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (r *EventNotificationAlert) readAlertStream(ctx context.Context, resp *http.Response) error {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server rejected request with status: %s", resp.Status)
	}

	_, params, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	boundary := params["boundary"]
	if boundary == "" {
		boundary = "boundary"
	}

	fmt.Printf(" [SUCCESS] Connected to %s! Tracking stream boundary: %s\n", r.Name, boundary)
	multipartReader := multipart.NewReader(resp.Body, boundary)

	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			return fmt.Errorf("multipart stream closed cleanly by camera host")
		}
		if err != nil {
			return fmt.Errorf("error navigating boundary tracking layer: %w", err)
		}

		xmlData, err := io.ReadAll(part)
		part.Close()
		if err != nil {
			log.Printf(" [WARN] Frame extraction failed on camera %s: %v\n", r.Name, err)
			continue
		}

		cleanXML := strings.TrimSpace(string(xmlData))
		if cleanXML == "" {
			continue
		}

		var ev EventData
		if err := xml.Unmarshal([]byte(cleanXML), &ev); err == nil {
			if ev.EventDescription != "Motion alarm" {
				fmt.Printf("@camera EVENT MATCHED: [%s] Description: %s ip: %s \n", ev.EventType, ev.EventDescription, ev.IPAddress)

				select {
				case <-ctx.Done():
					return ctx.Err()
				case r.events <- ev:
				}
			}
		} else {
			log.Printf(" [PARSE ERROR] Camera %s dropped unreadable XML payload: %v\n", r.Name, err)
		}
	}
}

func (e *EventNotificationAlert) Events() <-chan EventData {
	return e.events
}
