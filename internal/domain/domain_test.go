package domain

import (
	"context"
	"errors"
	"testing"
	"time"
	"strings"
)

// ---- fakes ----

type fakeCatalog struct {
	items map[int64]VideoMeta
}

func (f fakeCatalog) Get(id int64) (VideoMeta, bool) {
	v, ok := f.items[id]
	return v, ok
}

type fakeIdentityClient struct {
	identity *Identity
	err      error
}

func (f fakeIdentityClient) Fetch(ctx context.Context, token string) (*Identity, error) {
	return f.identity, f.err
}

type fakeAvailabilityClient struct {
	availability *Availability
	err          error
}

func (f fakeAvailabilityClient) Fetch(ctx context.Context, token string, videoID int64) (*Availability, error) {
	return f.availability, f.err
}

// helper to make availability window
func availability(from, to string) *Availability {
	a := &Availability{}
	a.AvailabilityWindow.From = from
	a.AvailabilityWindow.To = to
	return a
}

// ---- tests ----

func TestServiceResolve_PremiumUserGetsPremiumFilename(t *testing.T) {
	cat := fakeCatalog{items: map[int64]VideoMeta{
		46325: {VideoID: 46325, Title: "Example Video 001", StdFilename: "example001", PremiumFilename: "example001-premium", PlaybackExt: ".mp4"},
	}}

	idc := fakeIdentityClient{identity: &Identity{Roles: []string{"premium"}}}
	avc := fakeAvailabilityClient{availability: availability("2025-11-05", "2026-05-04")} // covers Feb 2026

	svc := NewService(1*time.Second, "https://s3.example/bucket", cat, idc, avc)

	resp, status, details := svc.Resolve(context.Background(), "token", 46325)

	if status != 200 || details != "" {
		t.Fatalf("expected 200 with no details; got status=%d details=%q", status, details)
	}
	if resp.PlaybackFilename != "example001-premium" {
		t.Fatalf("expected premium filename, got %q", resp.PlaybackFilename)
	}
}

func TestServiceResolve_NonPremiumUserGetsStandardFilename(t *testing.T) {
	cat := fakeCatalog{items: map[int64]VideoMeta{
		46325: {VideoID: 46325, Title: "Example Video 001", StdFilename: "example001", PremiumFilename: "example001-premium", PlaybackExt: ".mp4"},
	}}

	idc := fakeIdentityClient{identity: &Identity{Roles: []string{}}}
	avc := fakeAvailabilityClient{availability: availability("2025-11-05", "2026-05-04")}

	svc := NewService(1*time.Second, "https://s3.example/bucket", cat, idc, avc)

	resp, status, _ := svc.Resolve(context.Background(), "token", 46325)

	if status != 200 {
		t.Fatalf("expected 200; got %d", status)
	}
	if resp.PlaybackFilename != "example001" {
		t.Fatalf("expected standard filename, got %q", resp.PlaybackFilename)
	}
}

func TestServiceResolve_NotAvailableReturns403(t *testing.T) {
	cat := fakeCatalog{items: map[int64]VideoMeta{
		46325: {VideoID: 46325, Title: "Example Video 001", StdFilename: "example001", PremiumFilename: "example001-premium", PlaybackExt: ".mp4"},
	}}

	idc := fakeIdentityClient{identity: &Identity{Roles: []string{"premium"}}}
	// window that DOES NOT include Feb 2026
	avc := fakeAvailabilityClient{availability: availability("2024-01-01", "2024-02-01")}

	svc := NewService(1*time.Second, "https://s3.example/bucket", cat, idc, avc)

	_, status, details := svc.Resolve(context.Background(), "token", 46325)

	if status != 403 {
		t.Fatalf("expected 403; got %d", status)
	}
	if details == "" {
		t.Fatalf("expected details explaining availability window; got empty")
	}
}

func TestServiceResolve_UnknownVideoReturns404(t *testing.T) {
	cat := fakeCatalog{items: map[int64]VideoMeta{}}

	idc := fakeIdentityClient{identity: &Identity{Roles: []string{"premium"}}}
	avc := fakeAvailabilityClient{availability: availability("2025-11-05", "2026-05-04")}

	svc := NewService(1*time.Second, "https://s3.example/bucket", cat, idc, avc)

	_, status, details := svc.Resolve(context.Background(), "token", 99999)

	if status != 404 {
		t.Fatalf("expected 404; got %d", status)
	}
	if details != "unknown video_id" {
		t.Fatalf("expected details=%q; got %q", "unknown video_id", details)
	}
}

func TestServiceResolve_IdentityUpstreamFailureReturns502(t *testing.T) {
	cat := fakeCatalog{items: map[int64]VideoMeta{
		46325: {
			VideoID:         46325,
			Title:           "Example Video 001",
			StdFilename:     "example001",
			PremiumFilename: "example001-premium",
			PlaybackExt:     ".mp4",
		},
	}}

	idc := fakeIdentityClient{err: errors.New("boom")}
	avc := fakeAvailabilityClient{
		availability: availability("2025-11-05", "2026-05-04"),
	}

	svc := NewService(1*time.Second, "https://s3.example/bucket", cat, idc, avc)

	_, status, details := svc.Resolve(context.Background(), "token", 46325)

	if status != 502 {
		t.Fatalf("expected 502; got %d", status)
	}

	if !strings.HasPrefix(details, "identity service:") {
		t.Fatalf("expected identity service prefix; got %q", details)
	}
}

