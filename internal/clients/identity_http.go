package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	
	"github.com/anupamc/bytestream-playback-api/internal/domain"
)

type IdentityHTTP struct {
	baseURL string
	client  *http.Client
}

func NewIdentityHTTP(baseURL string, c *http.Client) *IdentityHTTP {
	return &IdentityHTTP{baseURL: baseURL, client: c}
}

func (c *IdentityHTTP) Fetch(ctx context.Context, token string) (*domain.Identity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/identity/userinfo", nil)
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

	var out domain.Identity
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &out, nil
}
