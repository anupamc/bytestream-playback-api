package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port              string
	IdentityBaseURL   string
	AvailabilityBase  string
	S3PlaybackBaseURL string
	HTTPTimeout       time.Duration
}

func Load() (Config, error) {
	port := getenvDefault("PORT", "8080")

	identity := os.Getenv("IDENTITY_BASE_URL")
	avail := os.Getenv("AVAILABILITY_BASE_URL")
	s3 := os.Getenv("S3_PLAYBACK_BASEURL")
	if identity == "" || avail == "" || s3 == "" {
		return Config{}, fmt.Errorf("missing required env vars: IDENTITY_BASE_URL, AVAILABILITY_BASE_URL, S3_PLAYBACK_BASEURL")
	}

	timeoutStr := getenvDefault("HTTP_TIMEOUT", "3s")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid HTTP_TIMEOUT: %w", err)
	}

	return Config{
		Port:              port,
		IdentityBaseURL:   strings.TrimRight(identity, "/"),
		AvailabilityBase:  strings.TrimRight(avail, "/"),
		S3PlaybackBaseURL: strings.TrimRight(s3, "/"),
		HTTPTimeout:       timeout,
	}, nil
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
