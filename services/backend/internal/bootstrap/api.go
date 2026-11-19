package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/services"
)

type RedeemVoucher struct {
	Token uuid.UUID `json:"token"`
}

type RedeemVoucherResponse struct {
	PictureURL      string    `json:"picture_url"`
	DestinationPath string    `json:"destination_path"`
	EventUUID       uuid.UUID `json:"event_uuid"`
}

func redeemHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.ErrorContext(ctx, "redeemHandler error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var voucher RedeemVoucher
	err = json.Unmarshal(body, &voucher)

	if err != nil {
		slog.ErrorContext(ctx, "redeemHandler unmarshal error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	link, err := services.GetDownloadLinkService().GetLink(ctx, voucher.Token)

	if err != nil {
		slog.ErrorContext(ctx, "redeemHandler GetLink failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := RedeemVoucherResponse{
		PictureURL:      link.GetURL(),
		DestinationPath: link.GetDestinationPath(),
		EventUUID:       link.GetEventUUID(),
	}

	contents, err := json.Marshal(&resp)

	if err != nil {
		slog.ErrorContext(ctx, "redeemHandler marshal to final json error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "download token redeemed", "event_id", link.GetEventUUID())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(contents)
}

func linksRouter() http.Handler {
	linksMux := http.NewServeMux()

	linksMux.HandleFunc("POST /redeem", redeemHandler)

	return linksMux
}

func v1Router() http.Handler {
	v1Mux := http.NewServeMux()

	v1Mux.Handle("/links/", http.StripPrefix("/links", linksRouter()))
	v1Mux.HandleFunc("GET /health", redeemHandler)

	return v1Mux
}

func WireAPIServer(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/v1/", http.StripPrefix("/v1", v1Router()))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	otelHandler := otelhttp.NewHandler(
		mux,
		"http-server",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Config.API.Port),
		Handler:      otelHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		slog.Info("shutdown signal received, closing API server...")

		// 3-second deadline for in-flight requests to finish
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		srv.Shutdown(shutdownCtx)
	}()

	// blocks the current worker goroutine until the server stops or crashes
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("API server failed to listen: %v", err)
	}

	slog.Info("API server worker clean exit")
}
