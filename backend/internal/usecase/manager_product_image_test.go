package usecase

import (
	"testing"

	"github.com/boms/backend/internal/config"
)

func TestSanitizeManagerProductImageURL(t *testing.T) {
	t.Parallel()

	external := "https://images.example.test/loaf.jpg"
	cfg := config.CloudinaryConfig{}

	url, err := sanitizeManagerProductImageURL(cfg, &external)
	if err != nil || url == nil || *url != external {
		t.Fatalf("expected external URL when cloudinary disabled, got %v err %v", url, err)
	}

	cfg = config.CloudinaryConfig{
		CloudName:    "demo",
		APIKey:       "key",
		APISecret:    "secret",
		UploadFolder: "boms/products",
	}
	valid := "https://res.cloudinary.com/demo/image/upload/v1/boms/products/loaf.jpg"
	url, err = sanitizeManagerProductImageURL(cfg, &valid)
	if err != nil || url == nil || *url != valid {
		t.Fatalf("expected cloudinary URL, got %v err %v", url, err)
	}

	wrongFolder := "https://res.cloudinary.com/demo/image/upload/v1/boms/other/loaf.jpg"
	if _, err = sanitizeManagerProductImageURL(cfg, &wrongFolder); err == nil {
		t.Fatal("expected validation error for URL outside upload folder")
	}

	if _, err = sanitizeManagerProductImageURL(cfg, &external); err == nil {
		t.Fatal("expected validation error for non-cloudinary URL")
	}
}
