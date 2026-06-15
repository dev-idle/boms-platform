package media

import (
	"net/url"
	"strings"
	"unicode"
)

const (
	cloudinaryDeliveryHost     = "res.cloudinary.com"
	DefaultProductUploadFolder = "boms/products"
)

// IsCloudinaryDeliveryURL reports whether raw is an HTTPS image delivery URL for cloudName.
func IsCloudinaryDeliveryURL(cloudName, raw string) bool {
	cloudName = strings.TrimSpace(cloudName)
	trimmed := strings.TrimSpace(raw)
	if cloudName == "" || trimmed == "" {
		return false
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return false
	}
	if parsed.Scheme != "https" || strings.ToLower(parsed.Host) != cloudinaryDeliveryHost {
		return false
	}

	path := strings.Trim(parsed.Path, "/")
	prefix := cloudName + "/image/upload/"
	return strings.HasPrefix(path, prefix) && len(path) > len(prefix)
}

// IsCloudinaryDeliveryURLInFolder reports whether raw is a Cloudinary delivery URL whose
// public ID path is under folder (e.g. boms/products). Empty folder uses DefaultProductUploadFolder.
func IsCloudinaryDeliveryURLInFolder(cloudName, folder, raw string) bool {
	if !IsCloudinaryDeliveryURL(cloudName, raw) {
		return false
	}

	folder = strings.Trim(strings.TrimSpace(folder), "/")
	if folder == "" {
		folder = DefaultProductUploadFolder
	}

	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}

	path := strings.Trim(parsed.Path, "/")
	prefix := strings.TrimSpace(cloudName) + "/image/upload/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}

	publicPath := cloudinaryPublicIDPath(strings.TrimPrefix(path, prefix))
	if publicPath == "" {
		return false
	}

	return strings.HasPrefix(publicPath, folder+"/") || publicPath == folder
}

func cloudinaryPublicIDPath(afterUpload string) string {
	segments := strings.Split(strings.Trim(afterUpload, "/"), "/")
	start := 0
	for start < len(segments) {
		segment := segments[start]
		if segment == "" {
			start++
			continue
		}
		if isCloudinaryTransformSegment(segment) || isCloudinaryVersionSegment(segment) {
			start++
			continue
		}
		break
	}
	return strings.Join(segments[start:], "/")
}

func isCloudinaryVersionSegment(segment string) bool {
	if len(segment) < 2 || segment[0] != 'v' {
		return false
	}
	for _, r := range segment[1:] {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isCloudinaryTransformSegment(segment string) bool {
	if isCloudinaryVersionSegment(segment) {
		return false
	}
	return strings.Contains(segment, ",") || strings.Contains(segment, "_")
}
