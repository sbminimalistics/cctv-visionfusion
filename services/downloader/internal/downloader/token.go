package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"sbminimalistics.com/cctv-visionfusion/downloader/internal/config"
)

type RequestPacket struct {
	Token string `json:"token"`
}

type ResponsePacket struct {
	PictureURL      string    `json:"picture_url"`
	DestinationPath string    `json:"destination_path"`
	EventUUID       uuid.UUID `json:"event_uuid"`
}

var tokenHTTPClient = &http.Client{
	Transport: otelhttp.NewTransport(http.DefaultTransport),
	Timeout:   15 * time.Second,
}

func ExchangeToken(msgCtx context.Context, apiEndpoint string, token string) (ResponsePacket, error) {
	url, err := url.JoinPath(
		strings.Join([]string{
			config.Config.BackendApiHost,
			strconv.Itoa(int(config.Config.BackendApiPort)),
		}, ":"),
		apiEndpoint,
	)
	if err != nil {
		slog.ErrorContext(msgCtx, "failed to build token exchange URL", "error", err)
		return ResponsePacket{}, err
	}

	outPacket := RequestPacket{
		Token: token,
	}
	contents, err := json.Marshal(&outPacket)
	if err != nil {
		slog.ErrorContext(msgCtx, "failed to marshal token request", "error", err)
		return ResponsePacket{}, err
	}

	req, err := http.NewRequestWithContext(msgCtx, "POST", url, bytes.NewReader(contents))
	if err != nil {
		slog.ErrorContext(msgCtx, "failed to create token exchange request", "error", err)
		return ResponsePacket{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := tokenHTTPClient.Do(req)
	if err != nil {
		slog.ErrorContext(msgCtx, "token exchange HTTP request failed", "error", err)
		return ResponsePacket{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.ErrorContext(msgCtx, "failed reading token exchange response", "error", err)
		return ResponsePacket{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		slog.ErrorContext(msgCtx, "token exchange rejected by backend", "status", resp.StatusCode, "error", err)
		return ResponsePacket{}, err
	}

	var inPacket ResponsePacket
	err = json.Unmarshal(body, &inPacket)
	if err != nil {
		slog.ErrorContext(msgCtx, "failed unmarshaling token exchange response", "error", err)
		return ResponsePacket{}, err
	}

	slog.DebugContext(msgCtx, "ExchangeToken token exchanged", "exchanged_to", string(body))

	return inPacket, nil
}
