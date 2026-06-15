package cloudinary

import (
	"crypto/sha1" //nolint:gosec // G505: Cloudinary upload API requires SHA-1
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

const (
	// AllowedImageFormats is signed on every manager product upload.
	AllowedImageFormats = "jpg,png,webp,avif"
	// UniqueFilenameTrue is the Cloudinary upload flag value for unique filenames.
	UniqueFilenameTrue = "true"
	// MaxProductImageBytes caps manager catalog image uploads (5 MiB).
	MaxProductImageBytes int64 = 5 * 1024 * 1024
)

// SignUpload builds a Cloudinary upload authentication signature.
// params must include every signed field sent with the upload except file, api_key, and resource_type.
func SignUpload(params map[string]string, apiSecret string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+params[key])
	}

	toSign := strings.Join(parts, "&") + strings.TrimSpace(apiSecret)
	sum := sha1.Sum([]byte(toSign)) //nolint:gosec // G401: Cloudinary signed upload requires SHA-1
	return hex.EncodeToString(sum[:])
}

// UploadEndpoint returns the HTTPS image upload URL for a cloud.
func UploadEndpoint(cloudName string) string {
	return fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", strings.TrimSpace(cloudName))
}
