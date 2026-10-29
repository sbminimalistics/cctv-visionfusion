package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
	"sbminimalistics.com/cctv-visionfusion/internal/services"
)

type RedeemVoucher struct {
	Token uuid.UUID `json:"token"`
}

type RedeemVoucherResponse struct {
	PictureURL              string `json:"picture_url"`
	DestinationPath         string `json:"destination_path"`
	DestinationBaseFilename string `json:"destination_base_filename"`
}

func redeemHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("api downloadPathHandler", r.URL.Path)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("redeemHandler error: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var voucher RedeemVoucher
	err = json.Unmarshal(body, &voucher)

	if err != nil {
		log.Printf("redeemHandler unmarshal error: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	link, ok := services.GetDownloadLinkService().GetLink(voucher.Token)

	if !ok {
		log.Println("redeemHandler GetLink failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := RedeemVoucherResponse{
		PictureURL:              link.GetURL(),
		DestinationPath:         link.GetDestinationPath(),
		DestinationBaseFilename: link.GetDestinationBaseFileName(),
	}

	contents, err := json.Marshal(&resp)

	if err != nil {
		log.Printf("redeemHandler marshal to final json error: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Config.API.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		fmt.Println("shutdown signal received, closing API server...")

		// 3-second deadline for in-flight requests to finish
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		srv.Shutdown(shutdownCtx)
	}()

	// blocks the current worker goroutine until the server stops or crashes
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("API server failed to listen: %v", err)
	}

	fmt.Println("API server worker clean exit")
}
