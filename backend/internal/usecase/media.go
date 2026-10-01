package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/config"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	cloudinarysvc "github.com/boms/backend/internal/service/cloudinary"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// MediaUsecase signs uploads the browser sends straight to Cloudinary, each
// into the folder its use allows, so the API secret never leaves the server.
type MediaUsecase struct {
	cloudinary config.CloudinaryConfig
	users      port.UserRepository
}

func NewMediaUsecase(cloudinary config.CloudinaryConfig, users port.UserRepository) *MediaUsecase {
	return &MediaUsecase{cloudinary: cloudinary, users: users}
}

// ProductImageSignature signs a manager's catalog image uploads into the
// upload folder, each under a name Cloudinary makes unique.
func (u *MediaUsecase) ProductImageSignature() (*dto.CloudinaryUploadSignatureResponse, error) {
	folder := u.cloudinary.ResolvedUploadFolder()
	return u.sign(folder, map[string]string{
		"folder":          folder,
		"unique_filename": cloudinarysvc.UniqueFilenameTrue,
	})
}

// ReferenceImageSignature signs one reference photo upload for a customer who
// confirmed their address: into a folder of their own, so a cart line can only
// name a photo its customer uploaded, under a name chosen here and never
// overwritten, so the signature makes one image and the rate limit caps the
// uploads. The stored image is scaled down to fit 2000 px, whatever was sent.
func (u *MediaUsecase) ReferenceImageSignature(ctx context.Context, userID uuid.UUID) (*dto.CloudinaryUploadSignatureResponse, error) {
	customer, err := u.users.GetByID(ctx, userID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, ErrMeNotFound
	}
	if err != nil {
		return nil, err
	}
	if !customer.EmailVerified {
		return nil, domainuser.ErrEmailNotVerified
	}
	folder := u.cloudinary.CustomerReferenceFolder(userID)
	return u.sign(folder, map[string]string{
		"public_id":      folder + "/" + uuid.NewString(),
		"overwrite":      "false",
		"transformation": cloudinarysvc.ReferenceImageTransformation,
	})
}

// sign signs an upload into folder with params, plus the formats every
// upload is held to and the time it is signed at.
func (u *MediaUsecase) sign(folder string, params map[string]string) (*dto.CloudinaryUploadSignatureResponse, error) {
	if !u.cloudinary.Enabled() {
		return nil, apperrors.ErrServiceUnavailable.WithDetail("cloudinary", "not configured")
	}
	params["allowed_formats"] = cloudinarysvc.AllowedImageFormats
	params["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	return &dto.CloudinaryUploadSignatureResponse{
		CloudName: u.cloudinary.CloudName,
		APIKey:    u.cloudinary.APIKey,
		Signature: cloudinarysvc.SignUpload(params, u.cloudinary.APISecret),
		UploadURL: cloudinarysvc.UploadEndpoint(u.cloudinary.CloudName),
		Folder:    folder,
		Params:    params,
		MaxBytes:  cloudinarysvc.MaxImageBytes,
	}, nil
}
