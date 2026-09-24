package usecase

import (
	"testing"

	"github.com/boms/backend/internal/config"
)

func TestSanitizeManagerComboImageURL(t *testing.T) {
	t.Parallel()

	cfg := config.CloudinaryConfig{
		CloudName:    "demo",
		APIKey:       "key",
		APISecret:    "secret",
		UploadFolder: "boms/products",
	}
	valid := "https://res.cloudinary.com/demo/image/upload/v1/boms/products/morning-box.jpg"

	t.Run("keeps_an_uploaded_url", func(t *testing.T) {
		t.Parallel()
		url, err := sanitizeManagerComboImageURL(cfg, &valid)
		if err != nil || url == nil || *url != valid {
			t.Fatalf("expected the cloudinary URL, got %v err %v", url, err)
		}
	})

	t.Run("clears_the_image_when_absent_or_blank", func(t *testing.T) {
		t.Parallel()
		blank := "   "
		for _, raw := range []*string{nil, &blank} {
			url, err := sanitizeManagerComboImageURL(cfg, raw)
			if err != nil || url != nil {
				t.Fatalf("expected no image, got %v err %v", url, err)
			}
		}
	})

	t.Run("rejects_a_url_from_outside_the_upload_folder", func(t *testing.T) {
		t.Parallel()
		for _, raw := range []string{
			"https://res.cloudinary.com/demo/image/upload/v1/boms/other/morning-box.jpg",
			"https://images.example.test/morning-box.jpg",
		} {
			if _, err := sanitizeManagerComboImageURL(cfg, &raw); err == nil {
				t.Fatalf("expected a validation error for %q", raw)
			}
		}
	})

	t.Run("accepts_any_url_when_cloudinary_is_not_configured", func(t *testing.T) {
		t.Parallel()
		external := "https://images.example.test/morning-box.jpg"
		url, err := sanitizeManagerComboImageURL(config.CloudinaryConfig{}, &external)
		if err != nil || url == nil || *url != external {
			t.Fatalf("expected the external URL, got %v err %v", url, err)
		}
	})
}
