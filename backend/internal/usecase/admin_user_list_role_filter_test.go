package usecase

import (
	"testing"

	domainuser "github.com/boms/backend/internal/domain/user"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAdminListRoleFilter(t *testing.T) {
	t.Run("empty means no filter", func(t *testing.T) {
		role, err := parseAdminListRoleFilter("")
		require.NoError(t, err)
		assert.Nil(t, role)
	})

	t.Run("accepts all application roles", func(t *testing.T) {
		for _, want := range []domainuser.Role{
			domainuser.RoleCustomer,
			domainuser.RoleStaff,
			domainuser.RoleBaker,
			domainuser.RoleManager,
			domainuser.RoleAdmin,
		} {
			role, err := parseAdminListRoleFilter(string(want))
			require.NoError(t, err)
			require.NotNil(t, role)
			assert.Equal(t, want, *role)
		}
	})

	t.Run("rejects unknown role", func(t *testing.T) {
		role, err := parseAdminListRoleFilter("superuser")
		assert.Nil(t, role)
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrValidation.Code, ae.Code)
	})
}
