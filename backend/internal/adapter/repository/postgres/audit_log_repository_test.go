package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
)

// Every audited action goes through Create, and its caller logs any error it
// returns, so a successful write must come back as nil — with or without the
// visitor's address and user agent.
func TestAuditLogRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 5)
	users := postgresadapter.NewUserRepository(pool)
	logs := postgresadapter.NewAuditLogRepository(pool)

	actor, err := users.Create(ctx, port.CreateUserParams{
		Email:        "auditor@example.com",
		PasswordHash: testPasswordHashFixture,
		Role:         domainuser.RoleCustomer,
	})
	require.NoError(t, err)

	t.Run("records_an_entry_with_the_request_details", func(t *testing.T) {
		ip := "203.0.113.7"
		userAgent := "Mozilla/5.0"
		require.NoError(t, logs.Create(ctx, port.CreateAuditLogParams{
			ActorID:    actor.ID,
			ActorRole:  domainuser.RoleCustomer,
			Action:     domainuser.AuditActionMeUpdatedProfile,
			TargetID:   &actor.ID,
			TargetType: "user",
			BeforeJSON: []byte(`{"phone":null}`),
			AfterJSON:  []byte(`{"phone":"0901234567"}`),
			IP:         &ip,
			UserAgent:  &userAgent,
		}))

		entries, err := logs.ListByTargetID(ctx, port.ListAuditLogsByTargetParams{TargetID: actor.ID, Limit: 10})
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, actor.Email, entries[0].ActorEmail)
		assert.Equal(t, domainuser.AuditActionMeUpdatedProfile, entries[0].Action)
		assert.JSONEq(t, `{"phone":"0901234567"}`, string(entries[0].AfterJSON))
	})

	t.Run("records_an_entry_without_an_address_or_user_agent", func(t *testing.T) {
		require.NoError(t, logs.Create(ctx, port.CreateAuditLogParams{
			ActorID:    actor.ID,
			ActorRole:  domainuser.RoleCustomer,
			Action:     domainuser.AuditActionMeChangedPassword,
			TargetID:   &actor.ID,
			TargetType: "user",
			BeforeJSON: []byte(`{}`),
			AfterJSON:  []byte(`{}`),
		}))

		count, err := logs.CountByTargetID(ctx, actor.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}
