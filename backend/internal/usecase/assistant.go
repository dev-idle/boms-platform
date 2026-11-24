package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	domainassistant "github.com/boms/backend/internal/domain/assistant"
	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

const (
	// assistantHistoryTurns is how many earlier turns of a session the model
	// reads with a new question.
	assistantHistoryTurns int32 = 10
	// assistantToolRounds bounds how often the model may call tools for one
	// answer before it must answer.
	assistantToolRounds = 3
)

// Answers the model's own words cannot give. Each hand-off reply opens with
// assistantHandOffLead, so a later hand-off in the session knows which
// questions the counter already has.
const (
	assistantFallbackReply   = "Sorry, I can't answer that. Our staff can help you instead."
	assistantHandOffLead     = "I've passed your question to our staff."
	assistantHandOffReply    = assistantHandOffLead + " They'll reply here."
	assistantOrderHandOffFmt = assistantHandOffLead + " They'll reply on the page of order %s."
)

// AssistantUsecase answers visitors' questions with a language model that
// calls tools for the bakery's facts, and passes what it cannot answer to the
// counter.
type AssistantUsecase struct {
	llm      port.LLMClient
	sessions port.AssistantRepository
	users    port.UserRepository
	products port.ProductRepository
	store    port.StoreSettingsRepository
	orders   port.OrderRepository
	messages *ConversationUsecase
	tx       port.TxManager
}

func NewAssistantUsecase(
	llm port.LLMClient,
	sessions port.AssistantRepository,
	users port.UserRepository,
	products port.ProductRepository,
	store port.StoreSettingsRepository,
	orders port.OrderRepository,
	messages *ConversationUsecase,
	tx port.TxManager,
) *AssistantUsecase {
	return &AssistantUsecase{
		llm:      llm,
		sessions: sessions,
		users:    users,
		products: products,
		store:    store,
		orders:   orders,
		messages: messages,
		tx:       tx,
	}
}

// AssistantQuestion is a question checked and ready to answer, with the
// session it continues.
type AssistantQuestion struct {
	u         *AssistantUsecase
	userID    *uuid.UUID
	sessionID uuid.UUID
	question  string
	history   []domainassistant.Turn
	// handedOff is where the answer passed the customer on: once per answer.
	handedOff *dto.AssistantHandOffResponse
}

// Ask checks a question from userID — nil for a visitor who is not signed
// in — and finds the session it continues. A session that is gone, or that
// is not theirs, is not continued: a new one starts.
func (u *AssistantUsecase) Ask(ctx context.Context, userID *uuid.UUID, req dto.AskAssistantRequest) (*AssistantQuestion, error) {
	question, err := domainassistant.NewQuestion(req.Question)
	if err != nil {
		return nil, err
	}
	out := &AssistantQuestion{u: u, userID: userID, sessionID: uuid.New(), question: question}
	if req.SessionID == nil {
		return out, nil
	}
	sessionID, err := uuid.Parse(*req.SessionID)
	if err != nil {
		return nil, apperrors.ErrValidation.WithDetail("session_id", "invalid session id")
	}
	// A session is written with its first turn, so one with none is not theirs.
	history, err := u.sessions.RecentTurns(ctx, sessionID, userID, assistantHistoryTurns)
	if err != nil {
		return nil, err
	}
	if len(history) > 0 {
		out.sessionID, out.history = sessionID, history
	}
	return out, nil
}

// SessionID is the session the answer belongs to.
func (q *AssistantQuestion) SessionID() string {
	return q.sessionID.String()
}

// Answer asks the model, runs the tools it calls, streams its text through
// onText as far as a visitor may read it, and keeps the question and the
// answer in the session. The answer returned replaces what was streamed.
func (q *AssistantQuestion) Answer(ctx context.Context, onText func(string)) (*dto.AssistantAnswerResponse, error) {
	messages := q.conversation(time.Now())
	tools := assistantTools(q.userID != nil)
	stream := domainassistant.NewReplyStream(onText)
	var text string
	for round := 0; round < assistantToolRounds && q.handedOff == nil; round++ {
		reply, err := q.u.llm.Chat(ctx, messages, tools, stream.Write)
		if err != nil {
			return nil, errors.Join(domainassistant.ErrUnavailable, err)
		}
		if len(reply.ToolCalls) == 0 {
			text = reply.Text
			break
		}
		lastRound := round == assistantToolRounds-1
		messages = append(messages, port.LLMMessage{Role: port.LLMRoleAssistant, ToolCalls: reply.ToolCalls})
		for _, call := range reply.ToolCalls {
			// No round follows the last to read a lookup: only a hand-off still runs.
			if lastRound && call.Name != domainassistant.ToolHandOff {
				continue
			}
			result, err := q.runTool(ctx, call)
			if err != nil {
				return nil, err
			}
			messages = append(messages, port.LLMMessage{Role: port.LLMRoleTool, Content: result, ToolCallID: call.ID})
		}
	}
	handOff := q.handedOff
	answer := &dto.AssistantAnswerResponse{HandOff: handOff}
	switch {
	case handOff != nil && handOff.OrderCode != nil:
		answer.Reply = fmt.Sprintf(assistantOrderHandOffFmt, *handOff.OrderCode)
	case handOff != nil:
		answer.Reply = assistantHandOffReply
	default:
		reply, ok := domainassistant.CleanReply(text)
		answer.Reply, answer.OfferStaff = reply, !ok
		if !ok {
			answer.Reply = assistantFallbackReply
		}
	}
	if err := q.save(ctx, answer.Reply); err != nil {
		return nil, err
	}
	return answer, nil
}

// conversation is what the model reads: its instructions, the session's
// earlier turns and the new question.
func (q *AssistantQuestion) conversation(now time.Time) []port.LLMMessage {
	messages := make([]port.LLMMessage, 0, len(q.history)+2)
	messages = append(messages, port.LLMMessage{Role: port.LLMRoleSystem, Content: assistantInstructions(q.userID != nil, now)})
	for _, turn := range q.history {
		role := port.LLMRoleUser
		if turn.Role == domainassistant.RoleAssistant {
			role = port.LLMRoleAssistant
		}
		messages = append(messages, port.LLMMessage{Role: role, Content: turn.Body})
	}
	return append(messages, port.LLMMessage{Role: port.LLMRoleUser, Content: q.question})
}

// save keeps the question and its answer. A customer's account is held
// while it writes, as a message holds it, so an erasure waits and then
// deletes the session with the rest.
func (q *AssistantQuestion) save(ctx context.Context, answer string) error {
	return q.u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if q.userID != nil {
			if err := holdCustomer(txCtx, q.u.users, *q.userID); err != nil {
				return err
			}
		}
		return q.u.sessions.SaveTurn(txCtx, q.sessionID, q.userID, q.question, answer)
	})
}

// handOffBody is what the counter reads when the assistant passes a customer
// on: the customer's own questions of this session since it last passed them
// on, oldest first, as many of the latest as fit in one message.
func (q *AssistantQuestion) handOffBody() string {
	lines := []string{q.question}
	length := utf8.RuneCountInString(q.question)
	for i := len(q.history) - 1; i >= 0; i-- {
		turn := q.history[i]
		if turn.Role == domainassistant.RoleAssistant && strings.HasPrefix(turn.Body, assistantHandOffLead) {
			break
		}
		if turn.Role != domainassistant.RoleVisitor {
			continue
		}
		length += utf8.RuneCountInString(turn.Body) + 1
		if length > domainconversation.MaxBodyLength {
			break
		}
		lines = append([]string{turn.Body}, lines...)
	}
	return strings.Join(lines, "\n")
}

// WarmUp asks the model one short question so it is loaded before the first
// visitor's: a model served locally takes seconds to load after a restart.
func (u *AssistantUsecase) WarmUp(ctx context.Context) error {
	_, err := u.llm.Chat(ctx, []port.LLMMessage{
		{Role: port.LLMRoleSystem, Content: "Reply with one word."},
		{Role: port.LLMRoleUser, Content: "Ready?"},
	}, nil, func(string) {})
	return err
}

// assistantInstructions tells the model who it is and what it may say: only
// what a tool returned, about this bakery, in a few plain sentences.
func assistantInstructions(signedIn bool, now time.Time) string {
	var b strings.Builder
	b.WriteString("You are the assistant of Choux, a bakery in Vietnam. Customers order online and pick their order up at the shop; there is no delivery.\n")
	b.WriteString("Reply in English only, in at most three short sentences of plain text.\n")
	b.WriteString("Use the tools for every fact about products, prices, opening hours, pickup rules and orders; never state such a fact without a tool result.\n")
	if signedIn {
		b.WriteString("The customer is signed in. Order status needs the order code they give (format CH-YYMMDD-NNN); you can only see their own orders.\n")
		b.WriteString("If the customer asks for a person, reports a problem with an order or a payment, or you cannot help, call hand_off_to_staff.\n")
	} else {
		b.WriteString("The visitor is not signed in: to hear about one of their orders or to talk to staff, they must sign in first.\n")
	}
	b.WriteString("Politely refuse anything unrelated to the bakery and its orders.\n")
	b.WriteString("It is now " + now.In(domainstore.Location).Format("Monday 2 January 2006, 15:04") + " at the bakery.")
	return b.String()
}
