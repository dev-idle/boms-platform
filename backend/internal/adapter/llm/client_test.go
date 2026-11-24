package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
)

// server answers every chat completion with the chunks given, recording the
// request it was sent.
func server(t *testing.T, status int, chunks ...string) (*Client, *chatRequest) {
	t.Helper()
	var sent chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &sent))
		w.WriteHeader(status)
		for _, chunk := range chunks {
			_, _ = io.WriteString(w, chunk)
		}
	}))
	t.Cleanup(srv.Close)
	client := New(config.AssistantConfig{BaseURL: srv.URL + "/v1/", Model: "qwen3.5:9b", ReasoningEffort: "none"})
	return client, &sent
}

func data(payload string) string {
	return "data: " + payload + "\n\n"
}

func TestClient_Chat(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	question := []port.LLMMessage{{Role: port.LLMRoleUser, Content: "Croissant price?"}}

	t.Run("streams_the_text_as_it_arrives", func(t *testing.T) {
		t.Parallel()
		client, sent := server(t, http.StatusOK,
			": keep-alive\n\n",
			data(`{"choices":[{"delta":{"content":"A croissant "}}]}`),
			data(`{"choices":[{"delta":{"content":"is $3.80."}}]}`),
			data("[DONE]"),
		)
		var streamed []string

		reply, err := client.Chat(ctx, question, nil, func(text string) { streamed = append(streamed, text) })

		require.NoError(t, err)
		assert.Equal(t, "A croissant is $3.80.", reply.Text)
		assert.Empty(t, reply.ToolCalls)
		assert.Equal(t, []string{"A croissant ", "is $3.80."}, streamed)
		assert.Equal(t, "qwen3.5:9b", sent.Model)
		assert.True(t, sent.Stream)
		assert.Equal(t, "none", sent.ReasoningEffort, "thinking is turned off")
	})

	t.Run("puts_tool_calls_together_from_their_fragments", func(t *testing.T) {
		t.Parallel()
		client, sent := server(t, http.StatusOK,
			data(`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"search_products","arguments":"{\"query\":"}}]}}]}`),
			data(`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"croissant\"}"}},{"index":1,"id":"call_2","function":{"name":"get_store_hours","arguments":"{}"}}]}}]}`),
			data("[DONE]"),
		)
		tools := []port.LLMTool{{Name: "search_products", Description: "Search", Parameters: map[string]any{"type": "object"}}}

		reply, err := client.Chat(ctx, question, tools, func(string) {})

		require.NoError(t, err)
		assert.Equal(t, []port.LLMToolCall{
			{ID: "call_1", Name: "search_products", Arguments: `{"query":"croissant"}`},
			{ID: "call_2", Name: "get_store_hours", Arguments: "{}"},
		}, reply.ToolCalls)
		require.Len(t, sent.Tools, 1)
		assert.Equal(t, "function", sent.Tools[0].Type)
		assert.Equal(t, "search_products", sent.Tools[0].Function.Name)
	})

	t.Run("sends_a_tool_result_back_to_the_call_it_answers", func(t *testing.T) {
		t.Parallel()
		client, sent := server(t, http.StatusOK, data("[DONE]"))
		conversation := append(question,
			port.LLMMessage{Role: port.LLMRoleAssistant, ToolCalls: []port.LLMToolCall{{ID: "call_1", Name: "get_store_hours", Arguments: "{}"}}},
			port.LLMMessage{Role: port.LLMRoleTool, Content: `{"open_every_day_from":"07:00"}`, ToolCallID: "call_1"},
		)

		_, err := client.Chat(ctx, conversation, nil, func(string) {})

		require.NoError(t, err)
		require.Len(t, sent.Messages, 3)
		assert.Equal(t, "function", sent.Messages[1].ToolCalls[0].Type)
		assert.Equal(t, "call_1", sent.Messages[2].ToolCallID)
	})

	t.Run("a_refusal_is_an_error", func(t *testing.T) {
		t.Parallel()
		client, _ := server(t, http.StatusNotFound, `{"error":"model not found"}`)

		_, err := client.Chat(ctx, question, nil, func(string) {})

		require.ErrorContains(t, err, "status 404")
		assert.ErrorContains(t, err, "model not found")
	})

	t.Run("a_stream_cut_short_is_an_error", func(t *testing.T) {
		t.Parallel()
		client, _ := server(t, http.StatusOK, data(`{"choices":[{"delta":{"content":"A cro"}}]}`))

		_, err := client.Chat(ctx, question, nil, func(string) {})

		require.ErrorContains(t, err, "before [DONE]")
	})

	t.Run("refuses_a_tool_call_index_past_the_bound", func(t *testing.T) {
		t.Parallel()
		client, _ := server(t, http.StatusOK,
			data(`{"choices":[{"delta":{"tool_calls":[{"index":1000000,"function":{"name":"x"}}]}}]}`),
			data("[DONE]"),
		)

		_, err := client.Chat(ctx, question, nil, func(string) {})

		require.ErrorContains(t, err, "out of range")
	})

	t.Run("sends_no_effort_it_was_not_given", func(t *testing.T) {
		t.Parallel()
		var raw string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			raw = string(body)
			_, _ = io.WriteString(w, data("[DONE]"))
		}))
		t.Cleanup(srv.Close)
		client := New(config.AssistantConfig{BaseURL: srv.URL, Model: "m"})

		_, err := client.Chat(ctx, question, nil, func(string) {})

		require.NoError(t, err)
		assert.False(t, strings.Contains(raw, "reasoning_effort"))
	})
}
