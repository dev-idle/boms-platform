// Package cloudinary deletes images through Cloudinary's Admin API.
package cloudinary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
)

// adminTimeout bounds one call to the Admin API.
const adminTimeout = 15 * time.Second

// Admin calls Cloudinary's Admin API with the account's API key and secret.
type Admin struct {
	baseURL   string
	cloudName string
	apiKey    string
	apiSecret string
	http      *http.Client
}

func NewAdmin(cfg config.CloudinaryConfig) *Admin {
	return &Admin{
		baseURL:   "https://api.cloudinary.com",
		cloudName: strings.TrimSpace(cfg.CloudName),
		apiKey:    strings.TrimSpace(cfg.APIKey),
		apiSecret: strings.TrimSpace(cfg.APISecret),
		http:      &http.Client{Timeout: adminTimeout},
	}
}

// DeleteFolder implements port.ImageStore. One call deletes up to 1,000
// images and says when more are left, so it asks again until none are.
func (a *Admin) DeleteFolder(ctx context.Context, folder string) error {
	query := url.Values{"prefix": {folder + "/"}, "invalidate": {"true"}}
	for {
		answer, err := a.deleteImages(ctx, query)
		if err != nil {
			return fmt.Errorf("delete cloudinary images: %w", err)
		}
		if !answer.Partial {
			return nil
		}
		query.Set("next_cursor", answer.NextCursor)
	}
}

type deleteAnswer struct {
	Partial    bool   `json:"partial"`
	NextCursor string `json:"next_cursor"`
	Error      struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (a *Admin) deleteImages(ctx context.Context, query url.Values) (deleteAnswer, error) {
	endpoint := a.baseURL + "/v1_1/" + url.PathEscape(a.cloudName) + "/resources/image/upload?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return deleteAnswer{}, err
	}
	req.SetBasicAuth(a.apiKey, a.apiSecret)
	resp, err := a.http.Do(req)
	if err != nil {
		return deleteAnswer{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	var answer deleteAnswer
	decodeErr := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&answer)
	if resp.StatusCode != http.StatusOK {
		return deleteAnswer{}, fmt.Errorf("cloudinary answered %d %s", resp.StatusCode, answer.Error.Message)
	}
	if decodeErr != nil {
		return deleteAnswer{}, fmt.Errorf("read the answer: %w", decodeErr)
	}
	return answer, nil
}

var _ port.ImageStore = (*Admin)(nil)
