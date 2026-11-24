package assistant

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQuestion(t *testing.T) {
	t.Parallel()

	t.Run("trims_and_keeps_line_breaks", func(t *testing.T) {
		t.Parallel()
		got, err := NewQuestion("  Do you have croissants?\r\nAnd matcha?  ")
		require.NoError(t, err)
		assert.Equal(t, "Do you have croissants?\nAnd matcha?", got)
	})

	t.Run("refuses_empty_long_and_invisible_text", func(t *testing.T) {
		t.Parallel()
		for name, text := range map[string]string{
			"empty":          "   ",
			"too_long":       strings.Repeat("a", MaxQuestionLength+1),
			"direction_flip": "Hello" + string(rune(0x202e)) + "olleh",
			"control_char":   "Hello\x07",
		} {
			_, err := NewQuestion(text)
			assert.ErrorIs(t, err, ErrInvalidQuestion, name)
		}
	})

	t.Run("counts_characters_not_bytes", func(t *testing.T) {
		t.Parallel()
		_, err := NewQuestion(strings.Repeat("é", MaxQuestionLength))
		assert.NoError(t, err)
	})
}

func TestCleanReply(t *testing.T) {
	t.Parallel()

	t.Run("drops_markdown_a_visitor_would_read_as_symbols", func(t *testing.T) {
		t.Parallel()
		got, ok := CleanReply("## Croissants\nOur **Matcha Croissant** is `$4.20`.")
		require.True(t, ok)
		assert.Equal(t, "Croissants\nOur Matcha Croissant is $4.20.", got)
	})

	t.Run("withholds_a_reply_showing_the_assistant_s_workings", func(t *testing.T) {
		t.Parallel()
		for _, reply := range []string{
			"Please call hand_off_to_staff.",
			`<tool_call>{"name":"search_products"}</tool_call>`,
			"<think>The user wants bread</think>We have bread.",
			"",
			"Hi" + string(rune(0x200b)),
		} {
			_, ok := CleanReply(reply)
			assert.False(t, ok, reply)
		}
	})

	t.Run("cuts_a_long_reply_at_its_last_sentence_that_fits", func(t *testing.T) {
		t.Parallel()
		sentence := strings.Repeat("a", 99) + "."
		got, ok := CleanReply(strings.Repeat(sentence, 11))
		require.True(t, ok)
		assert.Equal(t, strings.Repeat(sentence, 10), got)
	})

	t.Run("cuts_a_reply_with_no_sentence_end_mid_way", func(t *testing.T) {
		t.Parallel()
		got, ok := CleanReply(strings.Repeat("a", MaxReplyLength+50))
		require.True(t, ok)
		assert.Equal(t, MaxReplyLength, len([]rune(got)))
		assert.True(t, strings.HasSuffix(got, "…"))
	})
}

func TestReplyStream(t *testing.T) {
	t.Parallel()

	// stream writes parts and returns what was shown.
	stream := func(parts ...string) string {
		var shown strings.Builder
		s := NewReplyStream(func(text string) { shown.WriteString(text) })
		for _, part := range parts {
			s.Write(part)
		}
		return shown.String()
	}

	t.Run("shows_the_reply_as_written_less_its_last_few_bytes", func(t *testing.T) {
		t.Parallel()
		reply := "Our matcha croissant is $4.50 and it is baked every morning."
		assert.Equal(t, reply[:len(reply)-heldBack], stream(reply[:20], reply[20:41], reply[41:]))
	})

	t.Run("never_shows_part_of_a_tool_name", func(t *testing.T) {
		t.Parallel()
		shown := stream("Let me check that for you now. I will call get_order", "_status for you.")
		assert.Equal(t, "Let me check that for you now. I wil", shown)
		assert.NotContains(t, shown, "get_")
	})

	t.Run("stops_for_good_at_the_workings", func(t *testing.T) {
		t.Parallel()
		shown := stream("Checking. <tool_call>{}</tool_call>", strings.Repeat("More words that follow. ", 5))
		assert.Empty(t, shown)
	})

	t.Run("stops_past_the_longest_answer", func(t *testing.T) {
		t.Parallel()
		shown := stream(strings.Repeat("a", MaxReplyLength), "b")
		assert.Equal(t, strings.Repeat("a", MaxReplyLength-heldBack), shown)
	})

	t.Run("cuts_between_characters", func(t *testing.T) {
		t.Parallel()
		shown := stream("Bánh mì và cà phê sữa đá — có sẵn mỗi sáng ở cửa hàng.")
		assert.True(t, utf8.ValidString(shown))
	})
}
