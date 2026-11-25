package email

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/port"
)

func TestReplyComposer(t *testing.T) {
	t.Parallel()
	composer, err := NewReplyComposer("https://shop.example")
	require.NoError(t, err)

	t.Run("messages_about_an_order_link_to_its_page", func(t *testing.T) {
		t.Parallel()
		orderID := uuid.New()
		email, err := composer.ComposeReplyEmail(port.ReplyEmail{
			To: "mai@example.com", ToName: "Mai", OrderID: &orderID, OrderCode: "CH-261125-004", Count: 2,
		})

		require.NoError(t, err)
		assert.Equal(t, "A message from our staff about your order CH-261125-004", email.Subject,
			"staff may have written first: no reply is claimed")
		assert.Equal(t, "Mai", email.ToName)
		for _, part := range []string{email.Text, email.HTML} {
			assert.Contains(t, part, "Hi Mai,")
			assert.Contains(t, part, "About your order CH-261125-004, our staff wrote you 2 messages")
			assert.Contains(t, part, "our staff at Choux wrote to you")
			assert.Contains(t, part, "https://shop.example/orders/"+orderID.String())
		}
	})

	t.Run("other_replies_point_to_the_assistant_on_the_shop", func(t *testing.T) {
		t.Parallel()
		email, err := composer.ComposeReplyEmail(port.ReplyEmail{To: "mai@example.com", Count: 1})

		require.NoError(t, err)
		assert.Equal(t, "Our staff replied to your message", email.Subject)
		for _, part := range []string{email.Text, email.HTML} {
			assert.Contains(t, part, "Hello,")
			assert.Contains(t, part, "a reply that you have not read yet")
			assert.Contains(t, part, "Ask Choux")
			assert.Contains(t, part, "https://shop.example/")
		}
		assert.NotContains(t, email.Text+email.HTML, "1 replies")
	})

	t.Run("needs_an_address_and_a_reply", func(t *testing.T) {
		t.Parallel()
		_, err := composer.ComposeReplyEmail(port.ReplyEmail{Count: 1})
		require.Error(t, err)
		_, err = composer.ComposeReplyEmail(port.ReplyEmail{To: "mai@example.com"})
		require.Error(t, err)
	})
}
