package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/config"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	cloudinarysvc "github.com/boms/backend/internal/service/cloudinary"
)

var demoCloudinary = config.CloudinaryConfig{
	CloudName: "demo", APIKey: "key", APISecret: "secret",
	UploadFolder: "boms/products", ReferenceFolder: "boms/references",
}

// customers stands in for the account a reference upload is signed for.
type customers struct {
	port.UserRepository
	verified bool
}

func (c customers) GetByID(_ context.Context, id uuid.UUID) (*domainuser.User, error) {
	return &domainuser.User{ID: id, EmailVerified: c.verified}, nil
}

func TestMediaUsecase_ProductImageSignature(t *testing.T) {
	t.Parallel()

	t.Run("disabled_when_not_configured", func(t *testing.T) {
		t.Parallel()
		_, err := NewMediaUsecase(config.CloudinaryConfig{}, nil).ProductImageSignature()
		require.Error(t, err)
	})

	t.Run("signs_uploads_into_the_catalog_folder", func(t *testing.T) {
		t.Parallel()
		out, err := NewMediaUsecase(demoCloudinary, nil).ProductImageSignature()
		require.NoError(t, err)
		assert.Equal(t, "boms/products", out.Folder)
		assert.Equal(t, "boms/products", out.Params["folder"])
		assert.Equal(t, cloudinarysvc.UniqueFilenameTrue, out.Params["unique_filename"])
		assert.Equal(t, cloudinarysvc.MaxImageBytes, out.MaxBytes)
		assert.Equal(t, cloudinarysvc.SignUpload(out.Params, "secret"), out.Signature, "every field sent is signed")
	})
}

func TestMediaUsecase_ReferenceImageSignature(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	customer := uuid.New()

	t.Run("one_image_in_the_customers_own_folder", func(t *testing.T) {
		t.Parallel()
		out, err := NewMediaUsecase(demoCloudinary, customers{verified: true}).ReferenceImageSignature(ctx, customer)

		require.NoError(t, err)
		assert.Equal(t, "boms/references/"+customer.String(), out.Folder)
		assert.Contains(t, out.Params["public_id"], out.Folder+"/", "named here, inside their folder")
		assert.Equal(t, "false", out.Params["overwrite"], "a reused signature replaces nothing")
		assert.Equal(t, cloudinarysvc.ReferenceImageTransformation, out.Params["transformation"])
		assert.NotContains(t, out.Params, "unique_filename")
		assert.Equal(t, cloudinarysvc.SignUpload(out.Params, "secret"), out.Signature, "every field sent is signed")
	})

	t.Run("refused_until_the_address_is_confirmed", func(t *testing.T) {
		t.Parallel()
		_, err := NewMediaUsecase(demoCloudinary, customers{}).ReferenceImageSignature(ctx, customer)
		require.ErrorIs(t, err, domainuser.ErrEmailNotVerified)
	})
}
