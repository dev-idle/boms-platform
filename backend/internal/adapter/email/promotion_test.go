package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/port"
)

func TestPromotionComposer(t *testing.T) {
	t.Parallel()
	composer, err := NewPromotionComposer("https://shop.example")
	require.NoError(t, err)

	t.Run("says_the_promotion_and_how_to_stop_them", func(t *testing.T) {
		t.Parallel()
		email, err := composer.ComposePromotionEmail(port.PromotionEmail{
			To:               "mai@example.com",
			ToName:           "Mai",
			Subject:          "Matcha week",
			Body:             "Ten percent off every matcha cake.\n\nThis week only.",
			UnsubscribeToken: "tok-123",
		})
		require.NoError(t, err)

		assert.Equal(t, "Matcha week", email.Subject)
		assert.Equal(t, "Mai", email.ToName)
		assert.Equal(t, "https://shop.example/unsubscribe#token=tok-123", email.ListUnsubscribe, "mail clients offer it as their own button")
		for _, part := range []string{email.HTML, email.Text} {
			assert.Contains(t, part, "Ten percent off every matcha cake.")
			assert.Contains(t, part, "This week only.")
			assert.Contains(t, part, "https://shop.example/products")
			assert.Contains(t, part, "https://shop.example/unsubscribe#token=tok-123", "the token rides in the fragment")
		}
		assert.Contains(t, email.HTML, `<p style="margin:0 0 16px;">This week only.</p>`, "each line a paragraph, blank ones dropped")
	})

	t.Run("escapes_what_the_manager_wrote", func(t *testing.T) {
		t.Parallel()
		email, err := composer.ComposePromotionEmail(port.PromotionEmail{
			To: "mai@example.com", Subject: "<b>Sale</b>", Body: "<script>alert(1)</script>", UnsubscribeToken: "tok",
		})
		require.NoError(t, err)
		assert.NotContains(t, email.HTML, "<script>")
		assert.Contains(t, email.HTML, "&lt;script&gt;")
	})
}
