package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

func TestUserRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 5)
	repo := postgresadapter.NewUserRepository(pool)

	t.Run("CRUD round trip", func(t *testing.T) {
		created, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "alice@example.com",
			PasswordHash: testPasswordHashFixture,
			Role:         domainuser.RoleCustomer,
		})
		require.NoError(t, err)
		assert.False(t, created.EmailVerified)

		byEmail, err := repo.GetByEmail(ctx, "alice@example.com")
		require.NoError(t, err)
		assert.Equal(t, created.ID, byEmail.ID)

		byID, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.Email, byID.Email)
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "alice@example.com",
			PasswordHash: "hash",
			Role:         domainuser.RoleCustomer,
		})
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrConflict.Code, ae.Code)
	})

	t.Run("soft deleted invisible", func(t *testing.T) {
		u, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "ghost@example.com",
			PasswordHash: "hash",
			Role:         domainuser.RoleCustomer,
		})
		require.NoError(t, err)
		require.NoError(t, repo.SoftDelete(ctx, u.ID))
		_, err = repo.GetByID(ctx, u.ID)
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrNotFound.Code, ae.Code)
	})

	t.Run("GetByEmail not found", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nobody@example.com")
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrNotFound.Code, ae.Code)
	})

	t.Run("AdminList filters by role", func(t *testing.T) {
		_, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "role-filter-customer@example.com",
			PasswordHash: testPasswordHashFixture,
			Role:         domainuser.RoleCustomer,
		})
		require.NoError(t, err)

		_, err = repo.AdminCreate(ctx, port.CreateUserParams{
			Email:        "role-filter-manager@example.com",
			PasswordHash: testPasswordHashFixture,
			Role:         domainuser.RoleManager,
		})
		require.NoError(t, err)

		customerRole := domainuser.RoleCustomer
		rows, err := repo.AdminList(ctx, port.AdminListUsersParams{
			Search: "",
			Role:   &customerRole,
			Limit:  20,
			Offset: 0,
		})
		require.NoError(t, err)

		total, err := repo.AdminListCount(ctx, "", &customerRole)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))

		for _, row := range rows {
			assert.Equal(t, domainuser.RoleCustomer, row.Role)
		}

		var foundCustomer bool
		for _, row := range rows {
			if row.Email == "role-filter-customer@example.com" {
				foundCustomer = true
			}
			assert.NotEqual(t, "role-filter-manager@example.com", row.Email)
		}
		assert.True(t, foundCustomer)
	})

	t.Run("AdminList includes customer without staff or admin full_name", func(t *testing.T) {
		_, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "customer-list@example.com",
			PasswordHash: testPasswordHashFixture,
			Role:         domainuser.RoleCustomer,
		})
		require.NoError(t, err)

		rows, err := repo.AdminList(ctx, port.AdminListUsersParams{
			Search: "",
			Limit:  20,
			Offset: 0,
		})
		require.NoError(t, err)

		total, err := repo.AdminListCount(ctx, "", nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		assert.NotEmpty(t, rows)

		var found bool
		for _, row := range rows {
			if row.Email == "customer-list@example.com" {
				found = true
				assert.Nil(t, row.FullName)
				break
			}
		}
		assert.True(t, found, "customer row should be listable")
	})
	t.Run("ClaimPhone counts other active holders only", func(t *testing.T) {
		customers := postgresadapter.NewCustomerProfileRepository(pool)
		phone := "+84912345678"
		holder, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "phone-holder@example.com",
			PasswordHash: testPasswordHashFixture,
			Role:         domainuser.RoleCustomer,
		})
		require.NoError(t, err)
		_, err = customers.Create(ctx, port.UpsertCustomerProfileParams{UserID: holder.ID, Phone: &phone})
		require.NoError(t, err)
		other, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "phone-other@example.com",
			PasswordHash: testPasswordHashFixture,
			Role:         domainuser.RoleCustomer,
		})
		require.NoError(t, err)

		claim := func(userID uuid.UUID) bool {
			var held bool
			require.NoError(t, pool.WithTx(ctx, func(txCtx context.Context) error {
				var claimErr error
				held, claimErr = repo.ClaimPhone(txCtx, phone, userID)
				return claimErr
			}))
			return held
		}
		assert.True(t, claim(other.ID), "another active account holds the number")
		assert.False(t, claim(holder.ID), "the holder does not block itself")

		require.NoError(t, repo.SoftDelete(ctx, holder.ID))
		assert.False(t, claim(other.ID), "a disabled holder releases the number")

		_, err = repo.ClaimPhone(ctx, phone, other.ID)
		assert.Error(t, err, "outside a transaction the lock would guard nothing")
	})

	t.Run("ReleasePhone clears the phone on the user's profile", func(t *testing.T) {
		customers := postgresadapter.NewCustomerProfileRepository(pool)
		phone := "+84987654321"
		user, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "phone-release@example.com",
			PasswordHash: testPasswordHashFixture,
			Role:         domainuser.RoleCustomer,
		})
		require.NoError(t, err)
		_, err = customers.Create(ctx, port.UpsertCustomerProfileParams{UserID: user.ID, Phone: &phone})
		require.NoError(t, err)

		require.NoError(t, repo.ReleasePhone(ctx, user.ID))

		profile, err := customers.GetByUserID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, profile.Phone)
	})

	t.Run("account locks see only open accounts and need a transaction", func(t *testing.T) {
		user, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "account-lock@example.com",
			PasswordHash: testPasswordHashFixture,
			Role:         domainuser.RoleCustomer,
		})
		require.NoError(t, err)

		require.NoError(t, pool.WithTx(ctx, func(txCtx context.Context) error {
			if _, err := repo.GetByIDForShare(txCtx, user.ID); err != nil {
				return err
			}
			_, err := repo.GetByIDForUpdate(txCtx, user.ID)
			return err
		}))
		_, err = repo.GetByIDForShare(ctx, user.ID)
		assert.Error(t, err, "outside a transaction the lock would guard nothing")
		_, err = repo.GetByIDForUpdate(ctx, user.ID)
		assert.Error(t, err, "outside a transaction the lock would guard nothing")

		require.NoError(t, repo.SoftDelete(ctx, user.ID))
		require.NoError(t, pool.WithTx(ctx, func(txCtx context.Context) error {
			_, err := repo.GetByIDForShare(txCtx, user.ID)
			assert.ErrorIs(t, err, apperrors.ErrNotFound, "a closed account is not held open")
			return nil
		}))
	})
}
