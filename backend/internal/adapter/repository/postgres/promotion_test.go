package postgres_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/adapter/email"
	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainevent "github.com/boms/backend/internal/domain/event"
	domainpromotion "github.com/boms/backend/internal/domain/promotion"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/boms/backend/internal/usecase"
)

// recordingPromotionQueue keeps the promotion emails it is asked to queue.
type recordingPromotionQueue struct {
	emails []port.PromotionEmailTask
}

func (q *recordingPromotionQueue) EnqueuePromotion(context.Context, port.PromotionTask) error {
	return nil
}

func (q *recordingPromotionQueue) EnqueuePromotionEmail(_ context.Context, task port.PromotionEmailTask) error {
	q.emails = append(q.emails, task)
	return nil
}

// A manager's promotion goes to every customer who agreed to promotions and
// confirmed their address, an email each; a customer stops them from the link
// in any of them, and the moment they agreed is kept until they withdraw.
func TestPromotions_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 8)
	profiles := postgresadapter.NewCustomerProfileRepository(f.pool)
	promotionRepo := postgresadapter.NewPromotionRepository(f.pool)
	audit := postgresadapter.NewAuditLogRepository(f.pool)
	queue := &recordingPromotionQueue{}
	mailer := &capturingMailer{}
	composer, err := email.NewPromotionComposer("https://shop.example")
	require.NoError(t, err)
	tokens := domainpromotion.NewUnsubscribeTokens(strings.Repeat("u", 32))
	managers := usecase.NewManagerPromotionUsecase(promotionRepo, f.pool, f.outbox, nil, zap.NewNop())
	emails := usecase.NewPromotionEmailUsecase(promotionRepo, queue, f.pool, f.outbox, composer, mailer, tokens, zap.NewNop())
	unsubscribe := usecase.NewUnsubscribeUsecase(f.users, profiles, f.pool, auditlogger.NewService(audit), tokens)
	manager := f.newWorker(t, domainuser.RoleManager)
	_, err = postgresadapter.NewStaffProfileRepository(f.pool).Create(ctx, port.UpsertStaffProfileParams{
		UserID: manager, FullName: "Hoa Pham", EmployeeCode: "EMP-9201",
	})
	require.NoError(t, err)

	customer := func(t *testing.T, optIn bool) uuid.UUID {
		t.Helper()
		id := f.newCustomer(t, nil, nil)
		_, err := profiles.Create(ctx, port.UpsertCustomerProfileParams{UserID: id, MarketingOptIn: &optIn})
		require.NoError(t, err)
		return id
	}
	send := func(t *testing.T, subject, body string) *dto.PromotionResponse {
		t.Helper()
		out, err := managers.Send(ctx, manager, domainuser.RoleManager, dto.SendPromotionRequest{Subject: subject, Body: body})
		require.NoError(t, err)
		return out
	}
	queued := func(promotionID string) []uuid.UUID {
		var out []uuid.UUID
		for _, task := range queue.emails {
			if task.PromotionID.String() == promotionID {
				out = append(out, task.UserID)
			}
		}
		return out
	}

	t.Run("a_promotion_goes_to_every_customer_who_agreed", func(t *testing.T) {
		before, err := managers.Audience(ctx)
		require.NoError(t, err)
		agreed, alsoAgreed := customer(t, true), customer(t, true)
		declined := customer(t, false)
		unconfirmed, err := f.users.Create(ctx, port.CreateUserParams{
			Email: "unconfirmed-" + uuid.NewString()[:8] + "@example.com", PasswordHash: testPasswordHashFixture, Role: domainuser.RoleCustomer,
		})
		require.NoError(t, err)
		_, err = profiles.Create(ctx, port.UpsertCustomerProfileParams{UserID: unconfirmed.ID, MarketingOptIn: new(true)})
		require.NoError(t, err)

		audience, err := managers.Audience(ctx)
		require.NoError(t, err)
		assert.Equal(t, before.Recipients+2, audience.Recipients, "only customers who agreed and confirmed their address")

		promotion := send(t, "  Matcha week  ", "Ten percent off every matcha cake.\r\nThis week only.")
		assert.Equal(t, "Matcha week", promotion.Subject)
		assert.Equal(t, string(domainpromotion.StatusSending), promotion.Status)
		assert.Nil(t, promotion.RecipientCount)
		assert.Equal(t, "Hoa Pham", *promotion.SenderName)
		promotionID := uuid.MustParse(promotion.ID)
		assert.Equal(t, []domainevent.Topic{domainpromotion.TopicCreated}, promotionTopics(t, f, promotionID),
			"the worker hears of it from the same transaction")

		require.NoError(t, emails.Queue(ctx, port.PromotionTask{EventID: uuid.New(), PromotionID: promotionID}))
		recipients := queued(promotion.ID)
		assert.Len(t, recipients, int(audience.Recipients))
		assert.Contains(t, recipients, agreed)
		assert.Contains(t, recipients, alsoAgreed)
		assert.NotContains(t, recipients, declined)
		assert.NotContains(t, recipients, unconfirmed.ID)

		listed, _, _, _, err := managers.List(ctx, 1, 20)
		require.NoError(t, err)
		require.NotEmpty(t, listed)
		assert.Equal(t, promotion.ID, listed[0].ID, "latest first")
		assert.Equal(t, string(domainpromotion.StatusSent), listed[0].Status)
		assert.Equal(t, audience.Recipients, int64(*listed[0].RecipientCount))
		assert.Equal(t, []domainevent.Topic{domainpromotion.TopicCreated, domainpromotion.TopicSent}, promotionTopics(t, f, promotionID))

		require.NoError(t, emails.Queue(ctx, port.PromotionTask{EventID: uuid.New(), PromotionID: promotionID}))
		assert.Len(t, queued(promotion.ID), len(recipients), "a promotion sent is not queued again")

		require.NoError(t, emails.Send(ctx, port.PromotionEmailTask{PromotionID: promotionID, UserID: agreed}))
		require.Len(t, mailer.sent, 1)
		account, err := f.users.GetByID(ctx, agreed)
		require.NoError(t, err)
		assert.Equal(t, account.Email, mailer.sent[0].To)
		assert.Equal(t, "Matcha week", mailer.sent[0].Subject)
		assert.Equal(t, "https://shop.example/unsubscribe#token="+tokens.Of(agreed), mailer.sent[0].ListUnsubscribe)
	})

	t.Run("the_link_in_any_promotion_stops_the_next", func(t *testing.T) {
		subscriber := customer(t, true)
		promotion := send(t, "Croissant Sunday", "Two for one.")
		promotionID := uuid.MustParse(promotion.ID)

		require.NoError(t, unsubscribe.Unsubscribe(ctx, tokens.Of(subscriber)))
		require.NoError(t, unsubscribe.Unsubscribe(ctx, tokens.Of(subscriber)), "stopping them again is no error")
		profile, err := profiles.GetByUserID(ctx, subscriber)
		require.NoError(t, err)
		assert.Nil(t, profile.MarketingConsentAt)
		trail, err := audit.ListForSubject(ctx, subscriber, nil, 10)
		require.NoError(t, err)
		require.Len(t, trail, 1, "recorded once")
		assert.Equal(t, domainpromotion.AuditActionMeUnsubscribed, trail[0].Action)

		sent := len(mailer.sent)
		require.NoError(t, emails.Send(ctx, port.PromotionEmailTask{PromotionID: promotionID, UserID: subscriber}))
		assert.Len(t, mailer.sent, sent, "an email queued before they stopped them is not sent")

		disabled := customer(t, true)
		require.NoError(t, f.users.SoftDelete(ctx, disabled))
		require.NoError(t, unsubscribe.Unsubscribe(ctx, tokens.Of(disabled)))
		disabledProfile, err := profiles.GetByUserID(ctx, disabled)
		require.NoError(t, err)
		assert.Nil(t, disabledProfile.MarketingConsentAt, "it holds if an administrator enables the account again")

		require.ErrorIs(t, unsubscribe.Unsubscribe(ctx, "forged"), domainpromotion.ErrInvalidUnsubscribeToken)
		require.NoError(t, unsubscribe.Unsubscribe(ctx, tokens.Of(uuid.New())), "an account that is gone has nothing to stop")
	})

	t.Run("the_moment_of_agreement_is_kept_until_withdrawn", func(t *testing.T) {
		subscriber := customer(t, true)
		agreed, err := profiles.GetByUserID(ctx, subscriber)
		require.NoError(t, err)
		require.NotNil(t, agreed.MarketingConsentAt)

		kept, err := profiles.UpdateByUserID(ctx, port.UpsertCustomerProfileParams{UserID: subscriber, MarketingOptIn: new(true)})
		require.NoError(t, err)
		assert.True(t, agreed.MarketingConsentAt.Equal(*kept.MarketingConsentAt), "agreeing again keeps when they first agreed")

		withdrawn, err := profiles.UpdateByUserID(ctx, port.UpsertCustomerProfileParams{UserID: subscriber, MarketingOptIn: new(false)})
		require.NoError(t, err)
		assert.Nil(t, withdrawn.MarketingConsentAt)
		left, err := profiles.UpdateByUserID(ctx, port.UpsertCustomerProfileParams{UserID: subscriber})
		require.NoError(t, err)
		assert.Nil(t, left.MarketingConsentAt, "a save that does not ask leaves it as the row has it")

		_, err = profiles.UpdateByUserID(ctx, port.UpsertCustomerProfileParams{UserID: subscriber, MarketingOptIn: new(true)})
		require.NoError(t, err)
		require.NoError(t, profiles.Erase(ctx, subscriber))
		erased, err := profiles.GetByUserID(ctx, subscriber)
		require.NoError(t, err)
		assert.Nil(t, erased.MarketingConsentAt, "an erased account gets no promotions")
	})

	t.Run("a_promotion_must_say_something", func(t *testing.T) {
		_, err := managers.Send(ctx, manager, domainuser.RoleManager, dto.SendPromotionRequest{Subject: "Two\nlines", Body: "Hi"})
		require.ErrorIs(t, err, domainpromotion.ErrInvalidSubject)
		_, err = managers.Send(ctx, manager, domainuser.RoleManager, dto.SendPromotionRequest{Subject: "Hi", Body: "  "})
		require.ErrorIs(t, err, domainpromotion.ErrInvalidBody)
	})
}

// promotionTopics lists the outbox's events about a promotion, oldest first.
func promotionTopics(t *testing.T, f *checkoutFixture, promotionID uuid.UUID) []domainevent.Topic {
	t.Helper()
	var events []domainevent.Event
	require.NoError(t, f.pool.WithTx(context.Background(), func(txCtx context.Context) error {
		var err error
		events, err = f.outbox.ClaimUnpublished(txCtx, 0, 1000)
		return err
	}))
	var topics []domainevent.Topic
	for _, e := range events {
		if e.Data["promotion_id"] == promotionID.String() {
			topics = append(topics, e.Topic)
		}
	}
	return topics
}
