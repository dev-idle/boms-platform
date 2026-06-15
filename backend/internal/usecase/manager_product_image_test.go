package usecase

import (
	"testing"

	"github.com/boms/backend/internal/config"
)

func TestSanitizeManagerProductImageURLs(t *testing.T) {
	t.Parallel()

	external := "https://images.example.test/loaf.jpg"
	cfg := config.CloudinaryConfig{}

	urls, err := sanitizeManagerProductImageURLs(cfg, []string{external})
	if err != nil || len(urls) != 1 || urls[0] != external {
		t.Fatalf("expected external URL when cloudinary disabled, got %v err %v", urls, err)
	}

	cfg = config.CloudinaryConfig{
		CloudName:    "demo",
		APIKey:       "key",
		APISecret:    "secret",
		UploadFolder: "boms/products",
	}
	valid := "https://res.cloudinary.com/demo/image/upload/v1/boms/products/loaf.jpg"
	urls, err = sanitizeManagerProductImageURLs(cfg, []string{valid})
	if err != nil || len(urls) != 1 || urls[0] != valid {
		t.Fatalf("expected cloudinary URL, got %v err %v", urls, err)
	}

	wrongFolder := "https://res.cloudinary.com/demo/image/upload/v1/boms/other/loaf.jpg"
	if _, err = sanitizeManagerProductImageURLs(cfg, []string{wrongFolder}); err == nil {
		t.Fatal("expected validation error for URL outside upload folder")
	}

	if _, err = sanitizeManagerProductImageURLs(cfg, []string{external}); err == nil {
		t.Fatal("expected validation error for non-cloudinary URL")
	}

	if _, err = sanitizeManagerProductImageURLs(cfg, []string{valid, valid}); err == nil {
		t.Fatal("expected validation error for duplicate URLs")
	}

	tooMany := make([]string, 6)
	for i := range tooMany {
		tooMany[i] = valid
	}
	if _, err = sanitizeManagerProductImageURLs(cfg, tooMany); err == nil {
		t.Fatal("expected validation error for more than 5 images")
	}
}
