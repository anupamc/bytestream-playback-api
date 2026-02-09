package main

import (
	"log"
	"net/http"

	"github.com/anupamc/bytestream-playback-api/internal/api"
	"github.com/anupamc/bytestream-playback-api/internal/catalog"
	"github.com/anupamc/bytestream-playback-api/internal/clients"
	"github.com/anupamc/bytestream-playback-api/internal/config"
	"github.com/anupamc/bytestream-playback-api/internal/domain"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	httpClient := &http.Client{Timeout: cfg.HTTPTimeout}

	cat := catalog.NewHardcoded(cfg.S3PlaybackBaseURL)
	idc := clients.NewIdentityHTTP(cfg.IdentityBaseURL, httpClient)
	avc := clients.NewAvailabilityHTTP(cfg.AvailabilityBase, httpClient)

	svc := domain.NewService(cfg.HTTPTimeout, cfg.S3PlaybackBaseURL, cat, idc, avc)
	h := api.NewHandler(svc, cat)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.Healthz)
	mux.HandleFunc("/readyz", h.Readyz)
	mux.HandleFunc("/playback/", h.Playback)

	addr := ":" + cfg.Port
	log.Printf("bytestream-playback listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, api.WithRequestLogging(mux)))
}
