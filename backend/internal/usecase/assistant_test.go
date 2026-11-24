package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainassistant "github.com/boms/backend/internal/domain/assistant"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// scriptedLLM answers each call with the next reply, and keeps what it was sent.
type scriptedLLM struct {
	replies []port.LLMReply
	err     error
	calls   [][]port.LLMMessage
	tools   [][]port.LLMTool
}

func (s *scriptedLLM) Chat(_ context.Context, messages []port.LLMMessage, tools []port.LLMTool, onText func(string)) (*port.LLMReply, error) {
	s.calls = append(s.calls, messages)
	s.tools = append(s.tools, tools)
	if s.err != nil {
		return nil, s.err
	}
	reply := s.replies[0]
	if len(s.replies) > 1 {
		s.replies = s.replies[1:]
	}
	if reply.Text != "" {
		onText(reply.Text)
	}
	return &reply, nil
}

func (s *scriptedLLM) toolNames(call int) []string {
	names := make([]string, 0, len(s.tools[call]))
	for _, tool := range s.tools[call] {
		names = append(names, tool.Name)
	}
	return names
}

// toolResult is what the model was sent back for the call id, in its last request.
func (s *scriptedLLM) toolResult(t *testing.T, id string) map[string]any {
	t.Helper()
	last := s.calls[len(s.calls)-1]
	for _, m := range last {
		if m.Role == port.LLMRoleTool && m.ToolCallID == id {
			var out map[string]any
			require.NoError(t, json.Unmarshal([]byte(m.Content), &out))
			return out
		}
	}
	t.Fatalf("no result for %s", id)
	return nil
}

// memorySessions keeps sessions in memory.
type memorySessions struct {
	owners map[uuid.UUID]*uuid.UUID
	turns  map[uuid.UUID][]domainassistant.Turn
}

func newMemorySessions() *memorySessions {
	return &memorySessions{owners: map[uuid.UUID]*uuid.UUID{}, turns: map[uuid.UUID][]domainassistant.Turn{}}
}

func (m *memorySessions) RecentTurns(_ context.Context, id uuid.UUID, userID *uuid.UUID, _ int32) ([]domainassistant.Turn, error) {
	owner := m.owners[id]
	if (owner == nil) != (userID == nil) || (owner != nil && *owner != *userID) {
		return nil, nil
	}
	return m.turns[id], nil
}

func (m *memorySessions) SaveTurn(_ context.Context, id uuid.UUID, userID *uuid.UUID, question, answer string) error {
	m.owners[id] = userID
	m.turns[id] = append(m.turns[id],
		domainassistant.Turn{Role: domainassistant.RoleVisitor, Body: question},
		domainassistant.Turn{Role: domainassistant.RoleAssistant, Body: answer})
	return nil
}

func (m *memorySessions) DeleteIdle(context.Context, time.Duration) (int64, error) { return 0, nil }
func (m *memorySessions) ListForUser(context.Context, uuid.UUID) ([]domainassistant.Session, error) {
	return nil, nil
}
func (m *memorySessions) DeleteForUser(context.Context, uuid.UUID) error { return nil }

// menuProducts answers a search with matches, and a search for nothing with the menu.
type menuProducts struct {
	port.ProductRepository
	matches []port.CatalogListProduct
	menu    []port.CatalogListProduct
}

func (p menuProducts) CatalogList(_ context.Context, params port.CatalogListProductsParams) ([]port.CatalogListProduct, error) {
	if params.Search != nil {
		return p.matches, nil
	}
	return p.menu, nil
}

// fixedStore is a bakery open 07:00–19:00 with one closed day.
type fixedStore struct{ port.StoreSettingsRepository }

func (fixedStore) GetSettings(context.Context) (domainstore.Settings, error) {
	return domainstore.Settings{
		OpensAt: 7 * time.Hour, ClosesAt: 19 * time.Hour, SlotLength: 30 * time.Minute,
		PreorderMinLead: 3 * time.Hour, InstantPrep: 15 * time.Minute, MaxAdvanceDays: 14,
	}, nil
}

func (fixedStore) ListClosedDates(context.Context, time.Time, time.Time) ([]domainstore.ClosedDate, error) {
	return nil, nil
}

// ownedOrders finds an order only by its customer and code.
type ownedOrders struct {
	port.OrderRepository
	owner uuid.UUID
	order domainorder.Order
	asked []string
}

func (o *ownedOrders) GetByCodeForUser(_ context.Context, userID uuid.UUID, code string) (*domainorder.Order, error) {
	o.asked = append(o.asked, code)
	if userID != o.owner || code != o.order.Code {
		return nil, apperrors.ErrNotFound
	}
	return &o.order, nil
}

type assistantFixture struct {
	llm      *scriptedLLM
	sessions *memorySessions
	products menuProducts
	orders   *ownedOrders
	users    *mockUserRepo
	customer uuid.UUID
}

func newAssistantFixture(replies ...port.LLMReply) *assistantFixture {
	customer := uuid.New()
	pickup := time.Date(2026, 11, 26, 1, 0, 0, 0, time.UTC)
	f := &assistantFixture{
		llm:      &scriptedLLM{replies: replies},
		sessions: newMemorySessions(),
		orders: &ownedOrders{owner: customer, order: domainorder.Order{
			ID: uuid.New(), Code: "CH-261125-004", UserID: &customer, Status: domainorder.StatusReady,
			Type: domainorder.TypePreOrder, Channel: domainorder.ChannelOnline, PickupAt: &pickup,
		}},
		users:    new(mockUserRepo),
		customer: customer,
	}
	f.users.On("GetByIDForShare", mock.Anything, customer).Return(&domainuser.User{ID: customer, Role: domainuser.RoleCustomer}, nil).Maybe()
	return f
}

func (f *assistantFixture) usecase() *usecase.AssistantUsecase {
	return usecase.NewAssistantUsecase(f.llm, f.sessions, f.users, f.products, fixedStore{}, f.orders, nil, passthroughTxManager{})
}

func (f *assistantFixture) answer(t *testing.T, userID *uuid.UUID, req dto.AskAssistantRequest) (*dto.AssistantAnswerResponse, string, error) {
	t.Helper()
	question, err := f.usecase().Ask(context.Background(), userID, req)
	require.NoError(t, err)
	var streamed strings.Builder
	answer, err := question.Answer(context.Background(), func(text string) { streamed.WriteString(text) })
	return answer, streamed.String(), err
}

func toolCall(id, name, arguments string) port.LLMReply {
	return port.LLMReply{ToolCalls: []port.LLMToolCall{{ID: id, Name: name, Arguments: arguments}}}
}

func TestAssistantUsecase(t *testing.T) {
	t.Parallel()

	t.Run("a_guest_is_offered_the_catalogue_and_the_hours_only", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(port.LLMReply{Text: "We open at 07:00."})

		answer, streamed, err := f.answer(t, nil, dto.AskAssistantRequest{Question: "When do you open?"})

		require.NoError(t, err)
		assert.Equal(t, "We open at 07:00.", answer.Reply)
		assert.True(t, strings.HasPrefix(answer.Reply, streamed), "what streamed leads into the answer")
		assert.Equal(t, []string{domainassistant.ToolSearchProducts, domainassistant.ToolStoreHours}, f.llm.toolNames(0))
		assert.Contains(t, f.llm.calls[0][0].Content, "must sign in first")
	})

	t.Run("a_customer_is_also_offered_their_orders_and_the_counter", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(port.LLMReply{Text: "Hello."})

		_, _, err := f.answer(t, &f.customer, dto.AskAssistantRequest{Question: "Hi"})

		require.NoError(t, err)
		assert.Equal(t, []string{
			domainassistant.ToolSearchProducts, domainassistant.ToolStoreHours,
			domainassistant.ToolOrderStatus, domainassistant.ToolHandOff,
		}, f.llm.toolNames(0))
	})

	t.Run("a_guest_cannot_reach_orders_or_staff_whatever_the_model_asks", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(port.LLMReply{ToolCalls: []port.LLMToolCall{
			{ID: "c1", Name: domainassistant.ToolOrderStatus, Arguments: `{"order_code":"CH-261125-004"}`},
			{ID: "c2", Name: domainassistant.ToolHandOff, Arguments: `{"order_code":"CH-261125-004"}`},
		}}, port.LLMReply{Text: "Please sign in to see your order."})

		answer, _, err := f.answer(t, nil, dto.AskAssistantRequest{Question: "Status of CH-261125-004?"})

		require.NoError(t, err)
		assert.Equal(t, "no such tool", f.llm.toolResult(t, "c1")["error"])
		assert.Equal(t, "no such tool", f.llm.toolResult(t, "c2")["error"])
		assert.Empty(t, f.orders.asked, "no order is read for a guest")
		assert.Nil(t, answer.HandOff, "nothing is passed to staff")
	})

	t.Run("an_order_s_status_is_read_from_the_customer_s_own_orders", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(toolCall("c1", domainassistant.ToolOrderStatus, `{"order_code":" ch-261125-004 "}`),
			port.LLMReply{Text: "Your order is ready to collect."})

		answer, _, err := f.answer(t, &f.customer, dto.AskAssistantRequest{Question: "Is CH-261125-004 ready?"})

		require.NoError(t, err)
		assert.Equal(t, "Your order is ready to collect.", answer.Reply)
		result := f.llm.toolResult(t, "c1")
		assert.Equal(t, true, result["found"])
		assert.Equal(t, "ready to collect at the shop", result["status"])
		assert.Equal(t, "Thursday 26 November 2026, 08:00", result["pickup_at"], "in the bakery's time")
		assert.NotContains(t, result, "pickup_code", "the handover secret never reaches the model")
	})

	t.Run("another_customer_s_order_is_not_found", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(toolCall("c1", domainassistant.ToolOrderStatus, `{"order_code":"CH-261125-004"}`),
			port.LLMReply{Text: "I could not find that order."})
		stranger := uuid.New()
		f.users.On("GetByIDForShare", mock.Anything, stranger).Return(&domainuser.User{ID: stranger}, nil)

		_, _, err := f.answer(t, &stranger, dto.AskAssistantRequest{Question: "Status of CH-261125-004?"})

		require.NoError(t, err)
		assert.Equal(t, false, f.llm.toolResult(t, "c1")["found"])
	})

	t.Run("a_search_with_no_match_returns_the_menu", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(toolCall("c1", domainassistant.ToolSearchProducts, `{"query":"cakes"}`),
			port.LLMReply{Text: "We have a Hojicha Cheesecake."})
		f.products.menu = []port.CatalogListProduct{{Name: "Hojicha Cheesecake", CategoryName: "Cakes", PriceCents: 540}}

		_, _, err := f.answer(t, nil, dto.AskAssistantRequest{Question: "What cakes do you sell?"})

		require.NoError(t, err)
		result := f.llm.toolResult(t, "c1")
		assert.Equal(t, true, result["no_exact_match"])
		products, ok := result["products"].([]any)
		require.True(t, ok)
		require.Len(t, products, 1)
		assert.Equal(t, "$5.40", products[0].(map[string]any)["price"])
	})

	t.Run("the_opening_hours_come_from_the_settings", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(toolCall("c1", domainassistant.ToolStoreHours, "{}"), port.LLMReply{Text: "From 07:00 to 19:00."})

		_, _, err := f.answer(t, nil, dto.AskAssistantRequest{Question: "Opening hours?"})

		require.NoError(t, err)
		result := f.llm.toolResult(t, "c1")
		assert.Equal(t, "07:00", result["open_every_day_from"])
		assert.Equal(t, "19:00", result["open_every_day_until"])
		assert.InDelta(t, 180, result["made_to_order_notice_minutes"], 0)
	})

	t.Run("a_reply_showing_its_workings_is_withheld_and_staff_offered", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(port.LLMReply{Text: "Please call hand_off_to_staff."})

		answer, streamed, err := f.answer(t, &f.customer, dto.AskAssistantRequest{Question: "I need help"})

		require.NoError(t, err)
		assert.Empty(t, streamed, "the workings never stream")
		assert.True(t, answer.OfferStaff)
		assert.NotContains(t, answer.Reply, "hand_off_to_staff")
		assert.Equal(t, answer.Reply, f.sessions.turns[mustSession(t, f)][1].Body, "what is kept is what the visitor read")
	})

	t.Run("a_model_that_keeps_calling_tools_must_stop", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(toolCall("c1", domainassistant.ToolStoreHours, "{}"))

		answer, _, err := f.answer(t, nil, dto.AskAssistantRequest{Question: "Hours?"})

		require.NoError(t, err)
		assert.Len(t, f.llm.calls, 3, "three rounds of tools at most")
		assert.True(t, answer.OfferStaff)
		assert.Len(t, f.llm.calls[2], len(f.llm.calls[1])+2, "the last round's lookup is not run")
	})

	t.Run("a_model_out_of_reach_leaves_nothing_kept", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture()
		f.llm.err = errors.New("connection refused")

		_, _, err := f.answer(t, nil, dto.AskAssistantRequest{Question: "Hours?"})

		require.ErrorIs(t, err, domainassistant.ErrUnavailable)
		assert.Empty(t, f.sessions.turns)
	})

	t.Run("a_session_continues_with_its_earlier_turns", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(port.LLMReply{Text: "Yes."})
		session := uuid.New()
		f.sessions.owners[session] = &f.customer
		f.sessions.turns[session] = []domainassistant.Turn{
			{Role: domainassistant.RoleVisitor, Body: "Do you have croissants?"},
			{Role: domainassistant.RoleAssistant, Body: "Yes, $3.80."},
		}
		id := session.String()

		question, err := f.usecase().Ask(context.Background(), &f.customer, dto.AskAssistantRequest{SessionID: &id, Question: "Matcha too?"})
		require.NoError(t, err)
		_, err = question.Answer(context.Background(), func(string) {})

		require.NoError(t, err)
		assert.Equal(t, id, question.SessionID())
		sent := f.llm.calls[0]
		require.Len(t, sent, 4, "instructions, two earlier turns, the question")
		assert.Equal(t, port.LLMRoleAssistant, sent[2].Role)
		assert.Equal(t, "Matcha too?", sent[3].Content)
	})

	t.Run("someone_else_s_session_starts_a_new_one", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture(port.LLMReply{Text: "Hello."})
		session := uuid.New()
		f.sessions.owners[session] = nil
		f.sessions.turns[session] = []domainassistant.Turn{{Role: domainassistant.RoleVisitor, Body: "A guest's question"}}
		id := session.String()

		question, err := f.usecase().Ask(context.Background(), &f.customer, dto.AskAssistantRequest{SessionID: &id, Question: "Hi"})

		require.NoError(t, err)
		assert.NotEqual(t, id, question.SessionID())
		_, err = question.Answer(context.Background(), func(string) {})
		require.NoError(t, err)
		assert.Len(t, f.llm.calls[0], 2, "the guest's turns are not read")
	})

	t.Run("refuses_a_question_or_session_it_cannot_read", func(t *testing.T) {
		t.Parallel()
		f := newAssistantFixture()
		bad := "not-a-uuid"

		_, err := f.usecase().Ask(context.Background(), nil, dto.AskAssistantRequest{Question: "   "})
		require.ErrorIs(t, err, domainassistant.ErrInvalidQuestion)
		_, err = f.usecase().Ask(context.Background(), nil, dto.AskAssistantRequest{SessionID: &bad, Question: "Hi"})
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "validation_error", appErr.Code)
	})
}

// mustSession is the one session the fixture kept.
func mustSession(t *testing.T, f *assistantFixture) uuid.UUID {
	t.Helper()
	require.Len(t, f.sessions.turns, 1)
	for id := range f.sessions.turns {
		return id
	}
	return uuid.Nil
}
