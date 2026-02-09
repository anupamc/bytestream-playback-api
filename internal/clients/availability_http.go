package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anupamc/bytestream-playback-api/internal/domain"
)

type AvailabilityHTTP struct {
	baseURL string
	client  *http.Client
}

func NewAvailabilityHTTP(baseURL string, c *http.Client) *AvailabilityHTTP {
	return &AvailabilityHTTP{baseURL: baseURL, client: c}
}

func (c *AvailabilityHTTP) Fetch(ctx context.Context, token string, videoID int64) (*domain.Availability, error) {
	url := fmt.Sprintf("%s/availability/availabilityinfo/%d", c.baseURL, videoID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("unauthorized (%d)", res.StatusCode)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("bad status (%d)", res.StatusCode)
	}

	var out domain.Availability
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &out, nil
}
