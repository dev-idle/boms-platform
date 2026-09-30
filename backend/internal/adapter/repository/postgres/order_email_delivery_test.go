package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/adapter/email"
	"github.com/boms/backend/internal/adapter/queue"
	"github.com/boms/backend/internal/bootstrap"
	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/eventdispatch"
	"github.com/boms/backend/internal/usecase"
)

// capturingMailer keeps the emails it is asked to send.
type capturingMailer struct {
	sent []port.Email
}

func (m *capturingMailer) Send(_ context.Context, e port.Email) error {
	m.sent = append(m.sent, e)
	return nil
}

// A placed order is emailed: delivering its event queues the email, delivering
// the event again queues nothing more, and the worker writes the email from
// the order as stored.
func TestOrderEmailDelivery_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 5)
	rdb := goredis.NewClient(&goredis.Options{Addr: miniredis.RunT(t).Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	publisher := bootstrap.EventPublisher(rdb)
	dispatcher := eventdispatch.New(f.outbox, publisher, f.pool, zap.NewNop(), time.Second)
	f.pool.OnCommit(dispatcher.AfterCommit)

	customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
	account, err := f.users.GetByID(ctx, customer)
	require.NoError(t, err)
	order, err := f.orderUC.Checkout(ctx, customer, acceptingTerms(tomorrowAt(12, 0)))
	require.NoError(t, err)
	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	dispatcher.Wait(waitCtx)

	inspector := asynq.NewInspectorFromRedisClient(rdb)
	pending, err := inspector.ListPendingTasks(queue.QueueEmail)
	require.NoError(t, err)
	require.Len(t, pending, 1, "the order was received")
	eventID := uuid.MustParse(pending[0].ID)

	// The outbox resends an event whose delivery it could not record.
	require.NoError(t, publisher.Publish(ctx, []domainevent.Event{{
		ID: eventID, Topic: domainorder.TopicOrderCreated,
		Data: map[string]string{"order_id": order.ID, "status": string(domainorder.StatusPending)},
	}}))
	pending, err = inspector.ListPendingTasks(queue.QueueEmail)
	require.NoError(t, err)
	require.Len(t, pending, 1, "the customer is not mailed twice")

	composer, err := email.NewOrderComposer("https://shop.example")
	require.NoError(t, err)
	mailer := &capturingMailer{}
	handler := queue.OrderEmailHandler(usecase.NewOrderEmailUsecase(f.orders, composer, mailer, zap.NewNop()))
	require.NoError(t, handler(ctx, asynq.NewTask(pending[0].Type, pending[0].Payload)))

	require.Len(t, mailer.sent, 1)
	sent := mailer.sent[0]
	assert.Equal(t, account.Email, sent.To)
	assert.Equal(t, "We received your order "+order.Code, sent.Subject)
	assert.Contains(t, sent.Text, "1 x Croissant")
	assert.Contains(t, sent.Text, "https://shop.example/orders/"+order.ID)

	// The customer closes their account before a later email goes out.
	require.NoError(t, f.users.SoftDelete(ctx, customer))
	require.NoError(t, handler(ctx, asynq.NewTask(pending[0].Type, pending[0].Payload)))
	assert.Len(t, mailer.sent, 1, "a closed account gets no more email")
}
