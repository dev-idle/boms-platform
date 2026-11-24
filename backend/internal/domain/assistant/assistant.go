// Package assistant is the storefront's automated helper: it answers questions
// about products, opening hours and a customer's own orders from what the
// bakery holds, and passes anything else to the counter.
package assistant

import (
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	domainstore "github.com/boms/backend/internal/domain/store"
)

const (
	// MaxQuestionLength is the most characters one question holds.
	MaxQuestionLength = 500
	// MaxReplyLength is the most characters one answer shows; the model is
	// asked for three short sentences, so a longer one is cut.
	MaxReplyLength = 1000
)

// The tools the model may call. A visitor who is not signed in is offered
// only the catalogue and the opening hours.
const (
	ToolSearchProducts = "search_products"
	ToolStoreHours     = "get_store_hours"
	ToolOrderStatus    = "get_order_status"
	ToolHandOff        = "hand_off_to_staff"
)

// Role is who wrote one turn of a session.
type Role string

const (
	RoleVisitor   Role = "visitor"
	RoleAssistant Role = "assistant"
)

// Turn is one question or answer of a session.
type Turn struct {
	Role Role
	Body string
	At   time.Time
}

// Session is one conversation with the assistant, its turns oldest first.
type Session struct {
	ID        uuid.UUID
	StartedAt time.Time
	Turns     []Turn
}

// workings are what a reply holds when the model shows its own machinery
// instead of answering: a tool's name, or the markup it calls tools and
// thinks in.
var workings = []string{
	ToolSearchProducts, ToolStoreHours, ToolOrderStatus, ToolHandOff,
	"<tool_call", "</tool_call", "<think", "</think", "<|",
}

// markdown is the emphasis a model writes that plain text shows as symbols.
var markdown = strings.NewReplacer("**", "", "__", "", "`", "")

// NewQuestion checks a question as written: trimmed, 1 to MaxQuestionLength
// characters, plain text that may break across lines.
func NewQuestion(text string) (string, error) {
	text = domainstore.TrimLines(text)
	if text == "" || utf8.RuneCountInString(text) > MaxQuestionLength || !domainstore.PlainLines(text) {
		return "", ErrInvalidQuestion
	}
	return text, nil
}

// CleanReply turns what the model wrote into what a visitor reads: plain
// text without markdown emphasis or headings, no longer than MaxReplyLength.
// It reports false for a reply to withhold — empty, holding invisible
// characters, or showing the assistant's workings.
func CleanReply(text string) (string, bool) {
	lines := strings.Split(markdown.Replace(domainstore.TrimLines(text)), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(strings.TrimLeft(line, "#"))
	}
	text = strings.TrimSpace(strings.Join(lines, "\n"))
	if text == "" || withheld(text) {
		return "", false
	}
	return cut(text), true
}

// withheld reports whether text holds characters beyond plain text or shows
// the assistant's workings.
func withheld(text string) bool {
	lower := strings.ToLower(text)
	return !domainstore.PlainLines(text) ||
		slices.ContainsFunc(workings, func(marker string) bool { return strings.Contains(lower, marker) })
}

// heldBack is how many bytes at the end of a reply still being written wait
// for the next part: one fewer than the longest marker of the workings, so
// none shows in part before it is whole.
var heldBack = len(slices.MaxFunc(workings, func(a, b string) int { return len(a) - len(b) })) - 1

// ReplyStream shows a reply while the model writes it, as far as a visitor
// may read it: it keeps the last few characters back until they can no
// longer turn into the assistant's workings, and stops for good once the
// text would be withheld or runs past MaxReplyLength. The answer CleanReply
// makes replaces what was shown.
type ReplyStream struct {
	show    func(string)
	text    strings.Builder
	shown   int
	stopped bool
}

// NewReplyStream returns a stream that passes what may be shown to show.
func NewReplyStream(show func(string)) *ReplyStream {
	return &ReplyStream{show: show}
}

// Write adds the next part of the reply.
func (s *ReplyStream) Write(part string) {
	if s.stopped {
		return
	}
	s.text.WriteString(part)
	text := s.text.String()
	if utf8.RuneCountInString(text) > MaxReplyLength || withheld(text) {
		s.stopped = true
		return
	}
	end := len(text) - heldBack
	for end > s.shown && !utf8.RuneStart(text[end]) {
		end--
	}
	if end > s.shown {
		s.show(text[s.shown:end])
		s.shown = end
	}
}

// cut shortens text past MaxReplyLength at its last sentence that fits, or
// mid-sentence with an ellipsis when not even one does.
func cut(text string) string {
	runes := []rune(text)
	if len(runes) <= MaxReplyLength {
		return text
	}
	head := string(runes[:MaxReplyLength])
	if end := strings.LastIndexAny(head, ".!?"); end > 0 {
		return head[:end+1]
	}
	return strings.TrimSpace(string(runes[:MaxReplyLength-1])) + "…"
}
