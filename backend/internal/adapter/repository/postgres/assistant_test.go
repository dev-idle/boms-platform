package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainassistant "github.com/boms/backend/internal/domain/assistant"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/boms/backend/internal/usecase"
)

// scriptedModel answers with its replies in turn and keeps the tool results
// it was sent.
type scriptedModel struct {
	replies []port.LLMReply
	results map[string]string
}

func (m *scriptedModel) Chat(_ context.Context, messages []port.LLMMessage, _ []port.LLMTool, _ func(string)) (*port.LLMReply, error) {
	for _, msg := range messages {
		if msg.Role == port.LLMRoleTool {
			m.results[msg.ToolCallID] = msg.Content
		}
	}
	reply := m.replies[0]
	m.replies = m.replies[1:]
	return &reply, nil
}

func model(replies ...port.LLMReply) *scriptedModel {
	return &scriptedModel{replies: replies, results: map[string]string{}}
}

func callTool(name, arguments string) port.LLMReply {
	return port.LLMReply{ToolCalls: []port.LLMToolCall{{ID: "call", Name: name, Arguments: arguments}}}
}

// The assistant on the bakery's data: a customer's order read by its code and
// nobody else's, a question passed to the counter on its order or in the
// customer's general conversation, and sessions continued, exported, erased
// with the account and deleted once idle past their retention.
func TestAssistant_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	sessions := postgresadapter.NewAssistantRepository(f.pool)
	conversations := postgresadapter.NewConversationRepository(f.pool)
	messages := usecase.NewConversationUsecase(f.users, f.orders, conversations, f.pool, f.outbox)
	counter := usecase.NewStaffConversationUsecase(f.users, f.orders, conversations, f.pool, f.outbox)
	products := postgresadapter.NewProductRepository(f.pool)

	ask := func(t *testing.T, m *scriptedModel, userID *uuid.UUID, session *string, question string) (*dto.AssistantAnswerResponse, string) {
		t.Helper()
		assistant := usecase.NewAssistantUsecase(m, sessions, f.users, products, f.store, f.orders, messages, f.pool)
		q, err := assistant.Ask(ctx, userID, dto.AskAssistantRequest{SessionID: session, Question: question})
		require.NoError(t, err)
		answer, err := q.Answer(ctx, func(string) {})
		require.NoError(t, err)
		return answer, q.SessionID()
	}
	paidOrder := func(t *testing.T, hour int) (uuid.UUID, *dto.OrderResponse) {
		t.Helper()
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(hour, 0)))
		require.NoError(t, err)
		f.pay(t, customer, order.ID)
		return customer, order
	}

	t.Run("an_order_is_read_by_its_code_for_its_customer_only", func(t *testing.T) {
		customer, order := paidOrder(t, 9)
		m := model(callTool(domainassistant.ToolOrderStatus, `{"order_code":"`+order.Code+`"}`), port.LLMReply{Text: "It is confirmed."})

		ask(t, m, &customer, nil, "Where is my order?")

		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(m.results["call"]), &result))
		assert.Equal(t, true, result["found"])
		assert.Equal(t, order.Code, result["code"])

		stranger := f.newCustomer(t, nil, nil)
		m = model(callTool(domainassistant.ToolOrderStatus, `{"order_code":"`+order.Code+`"}`), port.LLMReply{Text: "Not found."})
		ask(t, m, &stranger, nil, "Where is this order?")
		assert.JSONEq(t, `{"found":false}`, m.results["call"])
	})

	t.Run("a_question_about_an_order_goes_to_its_conversation", func(t *testing.T) {
		customer, order := paidOrder(t, 10)
		m := model(callTool(domainassistant.ToolHandOff, `{"order_code":"`+order.Code+`"}`))

		answer, _ := ask(t, m, &customer, nil, "My cake came without its candles.")

		require.NotNil(t, answer.HandOff)
		require.NotNil(t, answer.HandOff.OrderID)
		assert.Equal(t, order.ID, *answer.HandOff.OrderID)
		assert.Contains(t, answer.Reply, order.Code)
		thread, err := counter.Thread(ctx, usecase.OrderThread(uuid.MustParse(order.ID)), nil)
		require.NoError(t, err)
		require.Len(t, thread.Messages, 1)
		assert.Equal(t, "My cake came without its candles.", thread.Messages[0].Body)
		assert.Equal(t, "customer", thread.Messages[0].From, "the customer's own words, not the assistant's")
		var viaAssistant bool
		require.NoError(t, raw.QueryRow(ctx, "SELECT via_assistant FROM messages WHERE body = $1",
			"My cake came without its candles.").Scan(&viaAssistant))
		assert.True(t, viaAssistant, "engagement counts customers writing, not questions the assistant passed on")
	})

	t.Run("a_question_about_no_order_goes_to_the_general_conversation", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		_, session := ask(t, model(port.LLMReply{Text: "We do not list allergens here."}), &customer, nil, "Is the matcha cake nut free?")

		answer, _ := ask(t, model(callTool(domainassistant.ToolHandOff, "{}")), &customer, &session, "Can someone check for me?")

		require.NotNil(t, answer.HandOff)
		assert.Nil(t, answer.HandOff.OrderID)
		thread, err := counter.Thread(ctx, usecase.GeneralThread(customer), nil)
		require.NoError(t, err)
		require.Len(t, thread.Messages, 1)
		assert.Equal(t, "Is the matcha cake nut free?\nCan someone check for me?", thread.Messages[0].Body,
			"the counter reads what the customer asked in this session")
	})

	t.Run("an_order_the_bakery_has_not_taken_goes_to_the_general_conversation", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		unpaid, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(11, 0)))
		require.NoError(t, err)

		answer, _ := ask(t, model(callTool(domainassistant.ToolHandOff, `{"order_code":"`+unpaid.Code+`"}`)), &customer, nil, "My payment failed")

		require.NotNil(t, answer.HandOff)
		assert.Nil(t, answer.HandOff.OrderID)
		_, err = counter.Thread(ctx, usecase.GeneralThread(customer), nil)
		require.NoError(t, err)
	})

	t.Run("an_order_of_another_customer_goes_to_the_general_conversation", func(t *testing.T) {
		_, order := paidOrder(t, 12)
		stranger := f.newCustomer(t, nil, nil)

		answer, _ := ask(t, model(callTool(domainassistant.ToolHandOff, `{"order_code":"`+order.Code+`"}`)), &stranger, nil, "Is this order ready?")

		require.NotNil(t, answer.HandOff)
		assert.Nil(t, answer.HandOff.OrderID, "nothing is written on someone else's order")
		thread, err := counter.Thread(ctx, usecase.OrderThread(uuid.MustParse(order.ID)), nil)
		require.NoError(t, err)
		assert.Empty(t, thread.Messages)
	})

	t.Run("one_answer_passes_the_customer_on_once", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		twice := port.LLMReply{ToolCalls: []port.LLMToolCall{
			{ID: "first", Name: domainassistant.ToolHandOff, Arguments: "{}"},
			{ID: "second", Name: domainassistant.ToolHandOff, Arguments: "{}"},
		}}

		ask(t, model(twice), &customer, nil, "Can someone call me?")

		thread, err := counter.Thread(ctx, usecase.GeneralThread(customer), nil)
		require.NoError(t, err)
		assert.Len(t, thread.Messages, 1)
	})

	t.Run("a_second_hand_off_passes_on_only_the_new_questions", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		_, session := ask(t, model(callTool(domainassistant.ToolHandOff, "{}")), &customer, nil, "Do you bake gluten-free bread?")
		ask(t, model(port.LLMReply{Text: "We open at 07:00."}), &customer, &session, "When do you open?")

		ask(t, model(callTool(domainassistant.ToolHandOff, "{}")), &customer, &session, "Can I order a vegan cake?")

		thread, err := counter.Thread(ctx, usecase.GeneralThread(customer), nil)
		require.NoError(t, err)
		require.Len(t, thread.Messages, 2)
		assert.Equal(t, "When do you open?\nCan I order a vegan cake?", thread.Messages[1].Body)
	})

	t.Run("a_session_is_kept_exported_and_erased_with_the_account", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		_, session := ask(t, model(port.LLMReply{Text: "From 07:00."}), &customer, nil, "When do you open?")
		_, again := ask(t, model(port.LLMReply{Text: "Until 19:00."}), &customer, &session, "And close?")
		require.Equal(t, session, again, "the session continues")

		turns, err := sessions.RecentTurns(ctx, uuid.MustParse(session), &customer, 10)
		require.NoError(t, err)
		require.Len(t, turns, 4)
		assert.Equal(t, []string{"When do you open?", "From 07:00.", "And close?", "Until 19:00."},
			[]string{turns[0].Body, turns[1].Body, turns[2].Body, turns[3].Body}, "oldest first, each question before its answer")
		stranger := f.newCustomer(t, nil, nil)
		require.Error(t, sessions.SaveTurn(ctx, uuid.MustParse(session), &stranger, "Mine now?", "No."),
			"a session never changes hands")

		customerProfiles := postgresadapter.NewCustomerProfileRepository(f.pool)
		audit := postgresadapter.NewAuditLogRepository(f.pool)
		export, err := usecase.NewDataExportUsecase(f.users, customerProfiles, postgresadapter.NewStaffProfileRepository(f.pool),
			postgresadapter.NewAdminProfileRepository(f.pool), f.orders, postgresadapter.NewPaymentRepository(f.pool), conversations, sessions,
			postgresadapter.NewSavedProductRepository(f.pool), postgresadapter.NewReviewRepository(f.pool), f.carts, fixedSessions{}, audit,
		).Export(ctx, customer)
		require.NoError(t, err)
		require.Len(t, export.Assistant, 1)
		require.Len(t, export.Assistant[0].Turns, 4)
		assert.Equal(t, "you", export.Assistant[0].Turns[0].From)
		assert.Equal(t, "assistant", export.Assistant[0].Turns[1].From)

		erasure := usecase.NewAccountErasureUsecase(f.pool, f.users, customerProfiles, f.carts, f.orders, conversations, sessions,
			postgresadapter.NewSavedProductRepository(f.pool), postgresadapter.NewReviewRepository(f.pool), audit,
			postgresadapter.NewUserTokenRepository(f.pool), &endedSessions{}, auditlogger.NewService(audit), fixtureHasher{})
		require.NoError(t, erasure.Erase(ctx, customer, fixturePassword))
		left, err := sessions.ListForUser(ctx, customer)
		require.NoError(t, err)
		assert.Empty(t, left)
	})

	t.Run("a_session_idle_past_its_retention_is_deleted", func(t *testing.T) {
		idle, fresh := uuid.New(), uuid.New()
		require.NoError(t, sessions.SaveTurn(ctx, idle, nil, "Do you sell bread?", "Yes."))
		require.NoError(t, sessions.SaveTurn(ctx, fresh, nil, "Do you sell tea?", "Yes."))
		_, err := raw.Exec(ctx, "UPDATE assistant_sessions SET last_message_at = now() - interval '61 days' WHERE id = $1", idle)
		require.NoError(t, err)

		deleted, err := sessions.DeleteIdle(ctx, 60*24*time.Hour)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))
		var turns int
		require.NoError(t, raw.QueryRow(ctx, "SELECT count(*) FROM assistant_messages WHERE session_id = $1", idle).Scan(&turns))
		assert.Zero(t, turns, "its turns go with it")
		kept, err := sessions.RecentTurns(ctx, fresh, nil, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, kept, "a session still in use stays")
	})

	t.Run("a_session_is_read_only_by_whoever_started_it", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		guestSession, customerSession := uuid.New(), uuid.New()
		require.NoError(t, sessions.SaveTurn(ctx, guestSession, nil, "Do you sell bread?", "Yes."))
		require.NoError(t, sessions.SaveTurn(ctx, customerSession, &customer, "Where is my cake?", "It is ready."))

		for name, read := range map[string]struct {
			session uuid.UUID
			userID  *uuid.UUID
			want    int
		}{
			"the_guest_s_by_a_guest":         {guestSession, nil, 2},
			"the_guest_s_by_a_customer":      {guestSession, &customer, 0},
			"the_customer_s_by_the_customer": {customerSession, &customer, 2},
			"the_customer_s_by_a_guest":      {customerSession, nil, 0},
			"the_customer_s_by_someone_else": {customerSession, new(uuid.New()), 0},
		} {
			turns, err := sessions.RecentTurns(ctx, read.session, read.userID, 10)
			require.NoError(t, err, name)
			assert.Len(t, turns, read.want, name)
		}
	})
}
