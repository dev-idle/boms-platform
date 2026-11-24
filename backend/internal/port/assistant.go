package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainassistant "github.com/boms/backend/internal/domain/assistant"
)

// LLMRole is who wrote one message a language model reads.
type LLMRole string

const (
	LLMRoleSystem    LLMRole = "system"
	LLMRoleUser      LLMRole = "user"
	LLMRoleAssistant LLMRole = "assistant"
	LLMRoleTool      LLMRole = "tool"
)

// LLMMessage is one message of the conversation sent to the model. An
// assistant message may ask for tools instead of answering; a tool message
// returns one tool's result to the call ToolCallID names.
type LLMMessage struct {
	Role       LLMRole
	Content    string
	ToolCalls  []LLMToolCall
	ToolCallID string
}

// LLMTool is a function the model may ask the caller to run. Parameters is
// its arguments' JSON Schema.
type LLMTool struct {
	Name        string
	Description string
	Parameters  map[string]any
}

// LLMToolCall is the model asking for one tool, with its arguments as JSON.
type LLMToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// LLMReply is the model's turn: text, or the tools it wants run first.
type LLMReply struct {
	Text      string
	ToolCalls []LLMToolCall
}

// LLMClient asks a language model.
type LLMClient interface {
	// Chat sends the conversation and the tools on offer, streams the reply's
	// text through onText as it arrives, and returns the whole reply.
	Chat(ctx context.Context, messages []LLMMessage, tools []LLMTool, onText func(string)) (*LLMReply, error)
}

// AssistantRepository keeps the assistant's sessions.
type AssistantRepository interface {
	// RecentTurns returns the last turns of a session of userID (nil for a
	// visitor not signed in), oldest first; none when the session is gone or
	// belongs to someone else.
	RecentTurns(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID, limit int32) ([]domainassistant.Turn, error)
	// SaveTurn records a question and its answer, starting the session for
	// userID (nil when not signed in) on its first turn.
	SaveTurn(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID, question, answer string) error
	// DeleteIdle deletes the sessions with no message for longer than
	// retention, measured on the database clock, and reports how many.
	DeleteIdle(ctx context.Context, retention time.Duration) (int64, error)
	// ListForUser returns every session of the account, newest first, for a
	// personal data export.
	ListForUser(ctx context.Context, userID uuid.UUID) ([]domainassistant.Session, error)
	// DeleteForUser deletes every session of the account.
	DeleteForUser(ctx context.Context, userID uuid.UUID) error
}
