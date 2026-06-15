package usecase

import (
	"testing"

	"github.com/boms/backend/internal/config"
)

func TestManagerMediaUsecase_CloudinaryUploadSignature(t *testing.T) {
	t.Parallel()

	t.Run("disabled when not configured", func(t *testing.T) {
		t.Parallel()
		uc := NewManagerMediaUsecase(config.CloudinaryConfig{})
		if _, err := uc.CloudinaryUploadSignature(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("returns signature when configured", func(t *testing.T) {
		t.Parallel()
		uc := NewManagerMediaUsecase(config.CloudinaryConfig{
			CloudName:    "demo",
			APIKey:       "key",
			APISecret:    "secret",
			UploadFolder: "boms/products",
		})
		out, err := uc.CloudinaryUploadSignature()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.CloudName != "demo" || out.APIKey != "key" || out.Folder != "boms/products" {
			t.Fatalf("unexpected payload: %+v", out)
		}
		if out.AllowedFormats == "" || out.UniqueFilename == "" {
			t.Fatalf("expected signed upload params in response: %+v", out)
		}
		if out.MaxBytes != 5*1024*1024 {
			t.Fatalf("expected max_bytes 5 MiB, got %d", out.MaxBytes)
		}
		if len(out.Signature) != 40 {
			t.Fatalf("expected sha1 hex signature, got %q", out.Signature)
		}
	})
}
