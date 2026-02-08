package domain

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type VideoMeta struct {
	VideoID         int64
	Title           string
	StdFilename     string
	PremiumFilename string
	PlaybackExt     string
}

type Identity struct {
	ID    int64    `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

type Availability struct {
	VideoID            int64 `json:"video_id"`
	AvailabilityWindow struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"availability_window"`
}

type PlaybackResponse struct {
	VideoID          int64  `json:"video_id"`
	Title            string `json:"title"`
	PlaybackBaseURL  string `json:"playback_baseurl"`
	PlaybackFilename string `json:"playback_filename"`
	PlaybackExt      string `json:"playback_extension"`
}

type Catalog interface {
	Get(id int64) (VideoMeta, bool)
}

type IdentityClient interface {
	Fetch(ctx context.Context, token string) (*Identity, error)
}

type AvailabilityClient interface {
	Fetch(ctx context.Context, token string, videoID int64) (*Availability, error)
}

type Service struct {
	timeout time.Duration
	baseURL  string
	cat      Catalog
	idc      IdentityClient
	avc      AvailabilityClient
}

func NewService(timeout time.Duration, baseURL string, cat Catalog, idc IdentityClient, avc AvailabilityClient) *Service {
	return &Service{timeout: timeout, baseURL: baseURL, cat: cat, idc: idc, avc: avc}
}

func (s *Service) Resolve(ctx context.Context, token string, videoID int64) (PlaybackResponse, int, string) {
	// returns: response, httpStatus, errorDetails
	if strings.TrimSpace(token) == "" {
		return PlaybackResponse{}, 401, "missing Authorization header"
	}

	meta, ok := s.cat.Get(videoID)
	if !ok {
		return PlaybackResponse{}, 404, "unknown video_id"
	}

	cctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	identity, err := s.idc.Fetch(cctx, token)
	if err != nil {
		return PlaybackResponse{}, 502, "identity service: " + err.Error()
	}

	avail, err := s.avc.Fetch(cctx, token, videoID)
	if err != nil {
		return PlaybackResponse{}, 502, "availability service: " + err.Error()
	}

	from := avail.AvailabilityWindow.From
	to := avail.AvailabilityWindow.To
	if !withinWindowUTC(from, to, time.Now().UTC()) {
		return PlaybackResponse{}, 403, fmt.Sprintf("outside availability window %s..%s", from, to)
	}

	filename := meta.StdFilename
	if hasRole(identity.Roles, "premium") && meta.PremiumFilename != "" {
		filename = meta.PremiumFilename
	}

	return PlaybackResponse{
		VideoID:          meta.VideoID,
		Title:            meta.Title,
		PlaybackBaseURL:  s.baseURL,
		PlaybackFilename: filename,
		PlaybackExt:      meta.PlaybackExt,
	}, 200, ""
}

func hasRole(roles []string, want string) bool {
	for _, r := range roles {
		if strings.EqualFold(r, want) {
			return true
		}
	}
	return false
}

func withinWindowUTC(fromYYYYMMDD, toYYYYMMDD string, now time.Time) bool {
	from, err1 := time.ParseInLocation("2006-01-02", fromYYYYMMDD, time.UTC)
	to, err2 := time.ParseInLocation("2006-01-02", toYYYYMMDD, time.UTC)
	if err1 != nil || err2 != nil {
		return false
	}
	toEnd := to.Add(24*time.Hour - time.Nanosecond)
	return !now.Before(from) && !now.After(toEnd)
}
