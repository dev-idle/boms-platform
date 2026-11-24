// Package llm asks a language model through the OpenAI-compatible chat
// completions API, which a model served locally by Ollama and the hosted
// providers both speak.
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
)

const (
	// maxReplyBytes bounds how much of one streamed reply is read.
	maxReplyBytes = 1 << 20
	// maxErrorBytes bounds how much of a refusal is kept for the error.
	maxErrorBytes = 2 << 10
	// maxToolCalls bounds how many tools one reply may ask for.
	maxToolCalls = 8
	// temperature keeps answers close to the tool results they repeat.
	temperature = 0.2
)

// Client sends chat completions to one model.
type Client struct {
	endpoint        string
	model           string
	reasoningEffort string
	http            *http.Client
}

// New returns a client for the model cfg names. It sets no timeout of its
// own: a reply streams for as long as the caller's context allows.
func New(cfg config.AssistantConfig) *Client {
	return &Client{
		endpoint:        strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions",
		model:           cfg.Model,
		reasoningEffort: cfg.ReasoningEffort,
		http:            &http.Client{},
	}
}

type chatRequest struct {
	Model           string        `json:"model"`
	Messages        []chatMessage `json:"messages"`
	Tools           []chatTool    `json:"tools,omitempty"`
	Stream          bool          `json:"stream"`
	Temperature     float64       `json:"temperature"`
	ReasoningEffort string        `json:"reasoning_effort,omitempty"`
}

type chatMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []chatToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

type chatToolCall struct {
	// Index places a streamed fragment within the reply's calls.
	Index    *int         `json:"index,omitempty"`
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function chatFunction `json:"function"`
}

type chatFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type chatTool struct {
	Type     string         `json:"type"`
	Function chatToolSchema `json:"function"`
}

type chatToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type chatChunk struct {
	Choices []struct {
		Delta struct {
			Content   string         `json:"content"`
			ToolCalls []chatToolCall `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
}

// Chat implements port.LLMClient.
func (c *Client) Chat(
	ctx context.Context,
	messages []port.LLMMessage,
	tools []port.LLMTool,
	onText func(string),
) (*port.LLMReply, error) {
	body, err := json.Marshal(c.request(messages, tools))
	if err != nil {
		return nil, fmt.Errorf("llm: encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("llm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBytes))
		return nil, fmt.Errorf("llm: status %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}
	return readStream(io.LimitReader(resp.Body, maxReplyBytes), onText)
}

func (c *Client) request(messages []port.LLMMessage, tools []port.LLMTool) chatRequest {
	out := chatRequest{
		Model:           c.model,
		Messages:        make([]chatMessage, 0, len(messages)),
		Stream:          true,
		Temperature:     temperature,
		ReasoningEffort: c.reasoningEffort,
	}
	for _, m := range messages {
		msg := chatMessage{Role: string(m.Role), Content: m.Content, ToolCallID: m.ToolCallID}
		for _, call := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, chatToolCall{
				ID:       call.ID,
				Type:     "function",
				Function: chatFunction{Name: call.Name, Arguments: call.Arguments},
			})
		}
		out.Messages = append(out.Messages, msg)
	}
	for _, t := range tools {
		out.Tools = append(out.Tools, chatTool{
			Type:     "function",
			Function: chatToolSchema{Name: t.Name, Description: t.Description, Parameters: t.Parameters},
		})
	}
	return out
}

// readStream reads server-sent chunks until [DONE]: text as it arrives, and
// tool calls put together from their fragments.
func readStream(r io.Reader, onText func(string)) (*port.LLMReply, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64<<10), maxReplyBytes)
	var text strings.Builder
	var calls []port.LLMToolCall
	for scanner.Scan() {
		payload, ok := strings.CutPrefix(scanner.Text(), "data:")
		if !ok {
			continue
		}
		payload = strings.TrimSpace(payload)
		if payload == "[DONE]" {
			return &port.LLMReply{Text: text.String(), ToolCalls: calls}, nil
		}
		var chunk chatChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return nil, fmt.Errorf("llm: decode chunk: %w", err)
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				text.WriteString(choice.Delta.Content)
				onText(choice.Delta.Content)
			}
			for _, fragment := range choice.Delta.ToolCalls {
				var err error
				if calls, err = mergeToolCall(calls, fragment); err != nil {
					return nil, err
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("llm: read stream: %w", err)
	}
	return nil, errors.New("llm: stream ended before [DONE]")
}

// mergeToolCall adds a streamed fragment to the call its index names, or
// starts a new call.
func mergeToolCall(calls []port.LLMToolCall, fragment chatToolCall) ([]port.LLMToolCall, error) {
	i := len(calls)
	if fragment.Index != nil {
		i = *fragment.Index
	}
	if i < 0 || i >= maxToolCalls {
		return nil, fmt.Errorf("llm: tool call index %d out of range", i)
	}
	for len(calls) <= i {
		calls = append(calls, port.LLMToolCall{})
	}
	calls[i].ID += fragment.ID
	calls[i].Name += fragment.Function.Name
	calls[i].Arguments += fragment.Function.Arguments
	return calls, nil
}

var _ port.LLMClient = (*Client)(nil)
