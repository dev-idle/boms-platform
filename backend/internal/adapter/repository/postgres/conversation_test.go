package postgres_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// A customer and the counter write to each other about an order the bakery
// took. Each side's unread messages are counted, the inbox lists the
// conversation, and the customer's data export and erasure take the messages
// with them.
func TestConversations_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	conversations := postgresadapter.NewConversationRepository(f.pool)
	customerSide := usecase.NewConversationUsecase(f.users, f.orders, conversations, f.pool, f.outbox)
	counter := usecase.NewStaffConversationUsecase(f.users, f.orders, conversations, f.pool, f.outbox)
	_, err := postgresadapter.NewStaffProfileRepository(f.pool).Create(ctx, port.UpsertStaffProfileParams{
		UserID: f.clerk, FullName: "Linh Tran", EmployeeCode: "EMP-9001",
	})
	require.NoError(t, err)

	confirmedOrder := func(t *testing.T, hour int) (uuid.UUID, uuid.UUID) {
		t.Helper()
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(hour, 0)))
		require.NoError(t, err)
		f.pay(t, customer, order.ID)
		return customer, uuid.MustParse(order.ID)
	}
	write := func(t *testing.T, customer, orderID uuid.UUID, body string) {
		t.Helper()
		_, err := customerSide.Post(ctx, customer, orderID, dto.PostMessageRequest{Body: body})
		require.NoError(t, err)
	}
	inboxRow := func(t *testing.T, orderID uuid.UUID) (dto.StaffInboxConversationResponse, bool) {
		t.Helper()
		rows, _, _, _, err := counter.List(ctx, 1, 100, "")
		require.NoError(t, err)
		for _, row := range rows {
			if row.OrderID == orderID.String() {
				return row, true
			}
		}
		return dto.StaffInboxConversationResponse{}, false
	}

	t.Run("the_customer_and_the_counter_write_to_each_other", func(t *testing.T) {
		customer, orderID := confirmedOrder(t, 9)
		before, err := counter.Counts(ctx)
		require.NoError(t, err)

		asked, err := customerSide.Post(ctx, customer, orderID, dto.PostMessageRequest{Body: "  Could the box say Happy birthday?\r\nThank you  "})
		require.NoError(t, err)
		assert.Equal(t, "Could the box say Happy birthday?\nThank you", asked.Body)
		assert.Equal(t, string(domainuser.RoleCustomer), asked.From)

		row, listed := inboxRow(t, orderID)
		require.True(t, listed)
		assert.Equal(t, string(domainconversation.StatusOpen), row.Status)
		assert.Equal(t, int32(1), row.Unread)
		assert.Equal(t, asked.Body, row.Preview)
		assert.NotEmpty(t, row.OrderCode)
		counts, err := counter.Counts(ctx)
		require.NoError(t, err)
		assert.Equal(t, before.Open+1, counts.Open)
		assert.Equal(t, before.Unread+1, counts.Unread)

		answer, err := counter.Post(ctx, f.clerk, orderID, dto.PostMessageRequest{Body: "Of course, we will write it in white chocolate."})
		require.NoError(t, err)
		assert.Equal(t, string(domainuser.RoleStaff), answer.From)
		assert.Equal(t, "Linh Tran", *answer.AuthorName)

		atCounter, err := counter.Thread(ctx, orderID, nil)
		require.NoError(t, err)
		require.NotNil(t, atCounter.Conversation)
		assert.Zero(t, atCounter.Conversation.Unread, "answering reads what came before")
		assert.Equal(t, "Linh Tran", *atCounter.Conversation.AssignedStaffName, "the writer takes it over")
		require.Len(t, atCounter.Messages, 2)
		assert.Equal(t, asked.ID, atCounter.Messages[0].ID, "oldest first")

		mine, err := customerSide.Thread(ctx, customer, orderID, nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), mine.Unread)
		require.Len(t, mine.Messages, 2)
		assert.Nil(t, mine.Messages[1].AuthorName, "the customer is not told who at the counter wrote")
		orders, _, _, _, err := f.orderUC.List(ctx, customer, 1, 20, dto.OrderHistoryQuery{})
		require.NoError(t, err)
		require.Len(t, orders, 1)
		assert.Equal(t, int32(1), orders[0].UnreadMessages)

		require.NoError(t, customerSide.MarkRead(ctx, customer, orderID))
		mine, err = customerSide.Thread(ctx, customer, orderID, nil)
		require.NoError(t, err)
		assert.Zero(t, mine.Unread)

		assert.Equal(t, []string{
			"message.created:customer+staff",
			"message.created:customer+staff",
			"conversation.changed:customer",
		}, conversationNotices(t, f, orderID, customer))
	})

	t.Run("resolving_reads_it_and_the_next_message_opens_it_again", func(t *testing.T) {
		customer, orderID := confirmedOrder(t, 10)
		write(t, customer, orderID, "Is it nut free?")

		resolved, err := counter.SetStatus(ctx, orderID, dto.PatchConversationRequest{Status: "closed"})
		require.NoError(t, err)
		assert.Equal(t, string(domainconversation.StatusClosed), resolved.Status)
		assert.Zero(t, resolved.Unread)
		reopened, err := counter.SetStatus(ctx, orderID, dto.PatchConversationRequest{Status: "open"})
		require.NoError(t, err)
		assert.Equal(t, string(domainconversation.StatusOpen), reopened.Status)
		_, err = counter.SetStatus(ctx, orderID, dto.PatchConversationRequest{Status: "closed"})
		require.NoError(t, err)

		write(t, customer, orderID, "One more question")
		row, listed := inboxRow(t, orderID)
		require.True(t, listed)
		assert.Equal(t, string(domainconversation.StatusOpen), row.Status)
		assert.Equal(t, int32(1), row.Unread)

		closed, _, _, _, err := counter.List(ctx, 1, 100, "closed")
		require.NoError(t, err)
		for _, row := range closed {
			assert.NotEqual(t, orderID.String(), row.OrderID, "open again, it leaves the resolved")
		}
	})

	t.Run("the_counter_reads_a_conversation_once", func(t *testing.T) {
		customer, orderID := confirmedOrder(t, 17)
		write(t, customer, orderID, "Hello")

		require.NoError(t, counter.MarkRead(ctx, orderID))
		require.NoError(t, counter.MarkRead(ctx, orderID))

		row, listed := inboxRow(t, orderID)
		require.True(t, listed)
		assert.Zero(t, row.Unread)
		assert.Equal(t, []string{
			"message.created:customer+staff",
			"conversation.changed:staff",
		}, conversationNotices(t, f, orderID, customer), "reading again, with nothing to read, tells no one")
	})

	t.Run("an_unknown_status_is_refused", func(t *testing.T) {
		_, orderID := confirmedOrder(t, 8)
		var appErr *apperrors.AppError

		_, _, _, _, err := counter.List(ctx, 1, 20, "archived")
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "validation_error", appErr.Code)
		_, err = counter.SetStatus(ctx, orderID, dto.PatchConversationRequest{Status: "archived"})
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "validation_error", appErr.Code)
	})

	t.Run("a_conversation_resolved_before_its_first_message_is_not_found", func(t *testing.T) {
		_, orderID := confirmedOrder(t, 11)

		_, err := counter.SetStatus(ctx, orderID, dto.PatchConversationRequest{Status: "closed"})

		require.ErrorIs(t, err, domainconversation.ErrNotFound)
	})

	t.Run("a_thread_pages_back_from_a_message", func(t *testing.T) {
		customer, orderID := confirmedOrder(t, 12)
		for i := range 51 {
			write(t, customer, orderID, fmt.Sprintf("Message %d", i))
		}

		latest, err := counter.Thread(ctx, orderID, nil)
		require.NoError(t, err)
		require.Len(t, latest.Messages, 50)
		assert.True(t, latest.HasMore)
		assert.Equal(t, "Message 50", latest.Messages[49].Body)

		older, err := counter.Thread(ctx, orderID, new(uuid.MustParse(latest.Messages[0].ID)))
		require.NoError(t, err)
		require.Len(t, older.Messages, 1)
		assert.Equal(t, "Message 0", older.Messages[0].Body)
		assert.False(t, older.HasMore)
	})

	t.Run("an_order_the_bakery_has_not_taken_cannot_be_written_about", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		unpaid, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(13, 0)))
		require.NoError(t, err)
		orderID := uuid.MustParse(unpaid.ID)

		_, err = customerSide.Post(ctx, customer, orderID, dto.PostMessageRequest{Body: "Hello"})
		require.ErrorIs(t, err, domainconversation.ErrUnavailable)
		_, err = counter.Thread(ctx, orderID, nil)
		require.ErrorIs(t, err, domainorder.ErrNotFound, "the counter does not see it")
		assert.False(t, unpaid.CanMessage)
	})

	t.Run("a_guest_has_no_account_to_write_to", func(t *testing.T) {
		pastry := f.pastry.String()
		guestOrder, err := f.staff.CreateOrder(ctx, f.clerk, domainuser.RoleStaff, uuid.New(), dto.CreateStaffOrderRequest{
			Channel:  "counter",
			Guest:    &dto.StaffOrderGuestRequest{Name: "Lan", Phone: "0901234567"},
			PickupAt: tomorrowAt(14, 0),
			Items:    []dto.StaffOrderItemRequest{{ProductID: &pastry, Quantity: 1}},
		})
		require.NoError(t, err)

		_, err = counter.Post(ctx, f.clerk, uuid.MustParse(guestOrder.ID), dto.PostMessageRequest{Body: "Your order is ready"})

		require.ErrorIs(t, err, domainconversation.ErrUnavailable)
	})

	t.Run("another_customer_s_order_is_not_found", func(t *testing.T) {
		_, orderID := confirmedOrder(t, 15)
		stranger := f.newCustomer(t, nil, nil)

		_, err := customerSide.Post(ctx, stranger, orderID, dto.PostMessageRequest{Body: "Hello"})
		require.ErrorIs(t, err, domainorder.ErrNotFound)
		_, err = customerSide.Thread(ctx, stranger, orderID, nil)
		require.ErrorIs(t, err, domainorder.ErrNotFound)
	})

	t.Run("the_export_carries_the_messages_and_erasure_takes_them", func(t *testing.T) {
		customer, orderID := confirmedOrder(t, 16)
		write(t, customer, orderID, "Please cancel, I am travelling")
		_, err := counter.Post(ctx, f.clerk, orderID, dto.PostMessageRequest{Body: "Done, the refund is on its way"})
		require.NoError(t, err)
		_, err = f.orderUC.Cancel(ctx, customer, orderID)
		require.NoError(t, err)
		write(t, customer, orderID, "Thank you")

		customerProfiles := postgresadapter.NewCustomerProfileRepository(f.pool)
		audit := postgresadapter.NewAuditLogRepository(f.pool)
		export, err := usecase.NewDataExportUsecase(f.users, customerProfiles, postgresadapter.NewStaffProfileRepository(f.pool),
			postgresadapter.NewAdminProfileRepository(f.pool), f.orders, conversations, postgresadapter.NewSavedProductRepository(f.pool),
			postgresadapter.NewReviewRepository(f.pool), f.carts, fixedSessions{}, audit,
		).Export(ctx, customer)
		require.NoError(t, err)
		require.Len(t, export.Orders, 1)
		messages := export.Orders[0].Messages
		require.Len(t, messages, 3)
		assert.Equal(t, string(domainuser.RoleStaff), messages[1].From)
		assert.Nil(t, messages[1].AuthorName, "the staff member's name is theirs, not the customer's")

		erasure := usecase.NewAccountErasureUsecase(f.pool, f.users, customerProfiles, f.carts, f.orders, conversations,
			postgresadapter.NewSavedProductRepository(f.pool), postgresadapter.NewReviewRepository(f.pool), audit,
			postgresadapter.NewUserTokenRepository(f.pool), &endedSessions{}, auditlogger.NewService(audit), fixtureHasher{})
		require.NoError(t, erasure.Erase(ctx, customer, fixturePassword))

		_, err = customerSide.Post(ctx, customer, orderID, dto.PostMessageRequest{Body: "Still there?"})
		require.ErrorIs(t, err, domainorder.ErrNotFound, "a message waits for the erasure, then finds no account")
		left, err := conversations.ListMessagesByOrderIDs(ctx, []uuid.UUID{orderID})
		require.NoError(t, err)
		assert.Empty(t, left, "every message on the order is erased")
		_, listed := inboxRow(t, orderID)
		assert.False(t, listed, "nor is the conversation in the inbox")
		raw, err := pgxpool.New(ctx, f.connStr)
		require.NoError(t, err)
		t.Cleanup(raw.Close)
		var text string
		require.NoError(t, raw.QueryRow(ctx, `
			SELECT string_agg(m.body, '') FROM messages m
			JOIN conversations c ON c.id = m.conversation_id WHERE c.order_id = $1`, orderID).Scan(&text))
		assert.Empty(t, text, "their text is gone, not only hidden")
	})
}

// conversationNotices lists the outbox's conversation events about an order,
// oldest first, as topic:audience — "customer" when it reaches customer.
func conversationNotices(t *testing.T, f *ticketFixture, orderID, customer uuid.UUID) []string {
	t.Helper()
	var events []domainevent.Event
	require.NoError(t, f.pool.WithTx(context.Background(), func(txCtx context.Context) error {
		var err error
		events, err = f.outbox.ClaimUnpublished(txCtx, 0, 1000)
		return err
	}))
	var notices []string
	for _, e := range events {
		if e.Data["order_id"] != orderID.String() ||
			(e.Topic != domainconversation.TopicMessageCreated && e.Topic != domainconversation.TopicConversationChanged) {
			continue
		}
		var audience []string
		for _, id := range e.Audience.UserIDs {
			if id == customer {
				audience = append(audience, "customer")
			}
		}
		for _, role := range e.Audience.Roles {
			audience = append(audience, string(role))
		}
		notices = append(notices, fmt.Sprintf("%s:%s", e.Topic, strings.Join(audience, "+")))
	}
	return notices
}
