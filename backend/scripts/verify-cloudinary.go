//go:build ignore

// One-off Cloudinary wiring check — run from backend/: go run scripts/verify-cloudinary.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/usecase"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		fail("config load: %v", err)
	}

	c := cfg.Cloudinary
	fmt.Println("=== BOMS Cloudinary config ===")
	fmt.Printf("enabled: %v\n", c.Enabled())
	fmt.Printf("cloud_name: %q\n", c.CloudName)
	fmt.Printf("upload_folder: %q\n", c.ResolvedUploadFolder())
	fmt.Printf("api_key_set: %v (len=%d)\n", strings.TrimSpace(c.APIKey) != "", len(strings.TrimSpace(c.APIKey)))
	fmt.Printf("api_secret_set: %v (len=%d)\n", strings.TrimSpace(c.APISecret) != "", len(strings.TrimSpace(c.APISecret)))

	if !c.Enabled() {
		fail("Cloudinary not fully configured — set CLOUDINARY_CLOUD_NAME, API_KEY, API_SECRET")
	}

	uc := usecase.NewManagerMediaUsecase(c)
	sig, err := uc.CloudinaryUploadSignature()
	if err != nil {
		fail("signature issue: %v", err)
	}
	fmt.Printf("signature_endpoint: ok (upload_url=%s)\n", sig.UploadURL)

	// 1x1 PNG
	png := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	_ = w.WriteField("api_key", sig.APIKey)
	_ = w.WriteField("timestamp", fmt.Sprintf("%d", sig.Timestamp))
	_ = w.WriteField("signature", sig.Signature)
	_ = w.WriteField("folder", sig.Folder)
	_ = w.WriteField("allowed_formats", sig.AllowedFormats)
	_ = w.WriteField("unique_filename", sig.UniqueFilename)
	_ = w.WriteField("max_file_size", fmt.Sprintf("%d", sig.MaxBytes))
	part, _ := w.CreateFormFile("file", "probe.png")
	_, _ = part.Write(png)
	_ = w.Close()

	req, err := http.NewRequest(http.MethodPost, sig.UploadURL, body)
	if err != nil {
		fail("build upload request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fail("upload request failed: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fail("Cloudinary upload HTTP %d: %s", resp.StatusCode, truncate(string(raw), 240))
	}

	var out struct {
		SecureURL string `json:"secure_url"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.SecureURL == "" {
		fail("unexpected upload response: %s", truncate(string(raw), 240))
	}

	fmt.Printf("cloudinary_upload: ok\n")
	fmt.Printf("sample_secure_url: %s\n", out.SecureURL)
	fmt.Println("=== All checks passed ===")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "FAIL: "+format+"\n", args...)
	os.Exit(1)
}
