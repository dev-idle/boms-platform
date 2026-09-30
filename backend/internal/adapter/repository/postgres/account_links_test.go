package postgres_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/adapter/email"
	"github.com/boms/backend/internal/adapter/queue"
	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	redisadapter "github.com/boms/backend/internal/adapter/repository/redis"
	"github.com/boms/backend/internal/bootstrap"
	"github.com/boms/backend/internal/config"
	domainaccount "github.com/boms/backend/internal/domain/account"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/boms/backend/internal/service/eventdispatch"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// A link token works once, for its purpose, until it expires, and only for an
// open account; the table keeps its hash, never the token.
func TestUserTokens_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool, connStr := newIntegrationDB(t, 4)
	users := postgresadapter.NewUserRepository(pool)
	tokens := postgresadapter.NewUserTokenRepository(pool)
	n := 0
	newUser := func(t *testing.T) uuid.UUID {
		t.Helper()
		n++
		user, err := users.Create(ctx, port.CreateUserParams{
			Email: fmt.Sprintf("link-%d@example.com", n), PasswordHash: testPasswordHashFixture, Role: domainuser.RoleCustomer,
		})
		require.NoError(t, err)
		return user.ID
	}

	t.Run("works_once", func(t *testing.T) {
		user := newUser(t)
		token, err := tokens.Issue(ctx, user, domainaccount.PurposeVerifyEmail)
		require.NoError(t, err)

		owner, err := tokens.Redeem(ctx, token, domainaccount.PurposeVerifyEmail)
		require.NoError(t, err)
		assert.Equal(t, user, owner)
		_, err = tokens.Redeem(ctx, token, domainaccount.PurposeVerifyEmail)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("only_the_newest_link_works", func(t *testing.T) {
		user := newUser(t)
		first, err := tokens.Issue(ctx, user, domainaccount.PurposeResetPassword)
		require.NoError(t, err)
		second, err := tokens.Issue(ctx, user, domainaccount.PurposeResetPassword)
		require.NoError(t, err)

		_, err = tokens.Redeem(ctx, first, domainaccount.PurposeResetPassword)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		_, err = tokens.Redeem(ctx, second, domainaccount.PurposeResetPassword)
		assert.NoError(t, err)
	})

	t.Run("a_link_does_one_thing", func(t *testing.T) {
		user := newUser(t)
		token, err := tokens.Issue(ctx, user, domainaccount.PurposeVerifyEmail)
		require.NoError(t, err)

		_, err = tokens.Redeem(ctx, token, domainaccount.PurposeResetPassword)
		assert.ErrorIs(t, err, apperrors.ErrNotFound, "a confirmation link cannot set a password")
	})

	t.Run("an_expired_link_opens_nothing", func(t *testing.T) {
		user := newUser(t)
		token, err := tokens.Issue(ctx, user, domainaccount.PurposeResetPassword)
		require.NoError(t, err)
		raw, err := pgxpool.New(ctx, connStr)
		require.NoError(t, err)
		t.Cleanup(raw.Close)
		_, err = raw.Exec(ctx, "UPDATE user_tokens SET expires_at = now() - interval '1 second' WHERE user_id = $1", user)
		require.NoError(t, err)

		_, err = tokens.Redeem(ctx, token, domainaccount.PurposeResetPassword)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("a_closed_account's_link_opens_nothing", func(t *testing.T) {
		user := newUser(t)
		token, err := tokens.Issue(ctx, user, domainaccount.PurposeResetPassword)
		require.NoError(t, err)
		require.NoError(t, users.SoftDelete(ctx, user))

		_, err = tokens.Redeem(ctx, token, domainaccount.PurposeResetPassword)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("the_table_keeps_hashes_only", func(t *testing.T) {
		user := newUser(t)
		token, err := tokens.Issue(ctx, user, domainaccount.PurposeVerifyEmail)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(token), 43, "256 random bits")
		raw, err := pgxpool.New(ctx, connStr)
		require.NoError(t, err)
		t.Cleanup(raw.Close)
		var stored []byte
		require.NoError(t, raw.QueryRow(ctx, "SELECT token_hash FROM user_tokens WHERE user_id = $1", user).Scan(&stored))
		sum := sha256.Sum256([]byte(token))
		assert.Equal(t, sum[:], stored)

		require.NoError(t, tokens.DeleteForUser(ctx, user))
		_, err = tokens.Redeem(ctx, token, domainaccount.PurposeVerifyEmail)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}

// plainHasher stands in for Argon2: fast, and a reset's new password can be read back.
type plainHasher struct{}

func (plainHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (plainHasher) Verify(encoded, password string) error {
	if encoded != "hashed:"+password {
		return errors.New("password mismatch")
	}
	return nil
}
func (plainHasher) NeedsRehash(string) bool { return false }

var linkPattern = regexp.MustCompile(`#token=([A-Za-z0-9_-]+)`)

// A customer confirms their address from the emailed link before they can
// order, and a forgotten password is replaced from a reset link that ends every
// session: request → event → queued task → email with a fresh link → redeem.
func TestAccountEmails_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 6)
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	redisClient, err := redisadapter.NewClient(ctx, config.RedisConfig{
		Addr: mr.Addr(), PoolSize: 4,
		DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second, HealthCheckTimeout: time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = redisClient.Close() })
	dispatcher := eventdispatch.New(f.outbox, bootstrap.EventPublisher(rdb), f.pool, zap.NewNop(), time.Second)
	f.pool.OnCommit(dispatcher.AfterCommit)

	auditRepo := postgresadapter.NewAuditLogRepository(f.pool)
	audit := auditlogger.NewService(auditRepo)
	tokens := postgresadapter.NewUserTokenRepository(f.pool)
	sessions := &endedSessions{}
	verification := usecase.NewEmailVerificationUsecase(f.pool, f.users, tokens, f.outbox, audit)
	reset := usecase.NewPasswordResetUsecase(f.pool, f.users, tokens, f.outbox, sessions, plainHasher{},
		redisadapter.NewQuota(redisClient), port.QuotaLimit{Max: 1, Window: time.Hour}, audit, zap.NewNop())
	composer, err := email.NewAccountComposer("https://shop.example")
	require.NoError(t, err)
	mailer := &capturingMailer{}
	handler := queue.AccountEmailHandler(usecase.NewAccountEmailUsecase(f.users, tokens, composer, mailer, zap.NewNop()))

	inspector := asynq.NewInspectorFromRedisClient(rdb)
	sent := map[string]bool{}
	// deliver runs the account emails queued since the last call and returns the
	// token of the last link sent.
	deliver := func(t *testing.T) string {
		t.Helper()
		waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		dispatcher.Wait(waitCtx)
		pending, err := inspector.ListPendingTasks(queue.QueueEmail)
		require.NoError(t, err)
		for _, task := range pending {
			if task.Type != queue.TypeAccountEmail || sent[task.ID] {
				continue
			}
			sent[task.ID] = true
			require.NoError(t, handler(ctx, asynq.NewTask(task.Type, task.Payload)))
		}
		if len(mailer.sent) == 0 {
			return ""
		}
		match := linkPattern.FindStringSubmatch(mailer.sent[len(mailer.sent)-1].Text)
		require.Len(t, match, 2)
		return match[1]
	}

	customer, err := f.users.Create(ctx, port.CreateUserParams{
		Email: "new-customer@example.com", PasswordHash: testPasswordHashFixture, Role: domainuser.RoleCustomer,
	})
	require.NoError(t, err)
	require.NoError(t, f.fillCart(customer.ID, []uuid.UUID{f.pastry}, nil))

	_, err = f.orderUC.Checkout(ctx, customer.ID, acceptingTerms(tomorrowAt(12, 0)))
	require.ErrorIs(t, err, domainuser.ErrEmailNotVerified, "no order before the address is confirmed")

	require.NoError(t, verification.Resend(ctx, customer.ID))
	token := deliver(t)
	require.NotEmpty(t, token)
	assert.Equal(t, "new-customer@example.com", mailer.sent[0].To)
	assert.Equal(t, "Confirm your email address for Choux", mailer.sent[0].Subject)

	require.NoError(t, verification.Verify(ctx, token))
	confirmed, err := f.users.GetByID(ctx, customer.ID)
	require.NoError(t, err)
	assert.True(t, confirmed.EmailVerified)
	require.ErrorIs(t, verification.Verify(ctx, token), domainaccount.ErrLinkInvalid, "a link works once")

	_, err = f.orderUC.Checkout(ctx, customer.ID, acceptingTerms(tomorrowAt(12, 0)))
	require.NoError(t, err, "a confirmed customer orders")

	require.NoError(t, reset.Request(ctx, "New-Customer@example.com"))
	resetToken := deliver(t)
	assert.Equal(t, "Reset your Choux password", mailer.sent[len(mailer.sent)-1].Subject)
	require.NoError(t, reset.Request(ctx, "new-customer@example.com"))
	deliver(t)
	assert.Len(t, mailer.sent, 2, "an account that used up its links is sent nothing")

	require.NoError(t, reset.Confirm(ctx, resetToken, "New-Password-9"))
	changed, err := f.users.GetByID(ctx, customer.ID)
	require.NoError(t, err)
	assert.Equal(t, "hashed:New-Password-9", changed.PasswordHash)
	assert.Contains(t, sessions.ended, customer.ID.String(), "every session ends")
	require.ErrorIs(t, reset.Confirm(ctx, resetToken, "Other-Password-9"), domainaccount.ErrLinkInvalid)

	trail, err := auditRepo.ListForSubject(ctx, customer.ID, nil, 10)
	require.NoError(t, err)
	var actions []domainuser.AuditAction
	for _, entry := range trail {
		actions = append(actions, entry.Action)
	}
	assert.Contains(t, actions, domainuser.AuditActionMeVerifiedEmail)
	assert.Contains(t, actions, domainuser.AuditActionMeResetPassword)
}
