package usecase

import (
	"strconv"
	"time"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/dto"
	cloudinarysvc "github.com/boms/backend/internal/service/cloudinary"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

type ManagerMediaUsecase struct {
	cloudinary config.CloudinaryConfig
}

func NewManagerMediaUsecase(cloudinary config.CloudinaryConfig) *ManagerMediaUsecase {
	return &ManagerMediaUsecase{cloudinary: cloudinary}
}

func (u *ManagerMediaUsecase) CloudinaryUploadSignature() (*dto.CloudinaryUploadSignatureResponse, error) {
	if !u.cloudinary.Enabled() {
		return nil, apperrors.ErrServiceUnavailable.WithDetail("cloudinary", "not configured")
	}

	timestamp := time.Now().Unix()
	folder := u.cloudinary.ResolvedUploadFolder()
	maxFileSize := strconv.FormatInt(cloudinarysvc.MaxProductImageBytes, 10)
	params := map[string]string{
		"allowed_formats": cloudinarysvc.AllowedImageFormats,
		"folder":          folder,
		cloudinarysvc.MaxFileSizeParam: maxFileSize,
		"timestamp":       strconv.FormatInt(timestamp, 10),
		"unique_filename": cloudinarysvc.UniqueFilenameTrue,
	}

	return &dto.CloudinaryUploadSignatureResponse{
		CloudName:      u.cloudinary.CloudName,
		APIKey:         u.cloudinary.APIKey,
		Timestamp:      timestamp,
		Signature:      cloudinarysvc.SignUpload(params, u.cloudinary.APISecret),
		Folder:         folder,
		UploadURL:      cloudinarysvc.UploadEndpoint(u.cloudinary.CloudName),
		AllowedFormats: cloudinarysvc.AllowedImageFormats,
		UniqueFilename: cloudinarysvc.UniqueFilenameTrue,
		MaxBytes:       cloudinarysvc.MaxProductImageBytes,
	}, nil
}
