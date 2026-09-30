package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainaccount "github.com/boms/backend/internal/domain/account"
	"github.com/boms/backend/internal/port"
)

func TestAccountComposer(t *testing.T) {
	t.Parallel()
	composer, err := NewAccountComposer("https://shop.example")
	require.NoError(t, err)

	t.Run("a_confirmation_links_to_the_verify_page", func(t *testing.T) {
		t.Parallel()
		email, err := composer.ComposeAccountEmail(port.AccountEmail{
			Purpose: domainaccount.PurposeVerifyEmail, To: "mai@example.com", Token: "tok-123",
		})
		require.NoError(t, err)

		assert.Equal(t, "Confirm your email address for Choux", email.Subject)
		assert.Equal(t, "mai@example.com", email.To)
		assert.Empty(t, email.ToName, "a name the customer typed never reaches an account email")
		for _, part := range []string{email.HTML, email.Text} {
			assert.Contains(t, part, "https://shop.example/verify-email#token=tok-123", "the token rides in the fragment")
			assert.Contains(t, part, "48 hours")
			assert.Contains(t, part, "123 Greige Lane")
		}
	})

	t.Run("a_reset_links_to_the_reset_page", func(t *testing.T) {
		t.Parallel()
		email, err := composer.ComposeAccountEmail(port.AccountEmail{
			Purpose: domainaccount.PurposeResetPassword, To: "mai@example.com", Token: "tok-456",
		})
		require.NoError(t, err)

		assert.Equal(t, "Reset your Choux password", email.Subject)
		assert.Contains(t, email.Text, "https://shop.example/reset-password#token=tok-456")
		assert.Contains(t, email.Text, "30 minutes, and only once")
		assert.Contains(t, email.HTML, "signs you out on every device")
	})

	t.Run("refuses_an_email_it_cannot_write", func(t *testing.T) {
		t.Parallel()
		_, err := composer.ComposeAccountEmail(port.AccountEmail{Purpose: "login", To: "mai@example.com", Token: "t"})
		require.Error(t, err)
		_, err = composer.ComposeAccountEmail(port.AccountEmail{Purpose: domainaccount.PurposeVerifyEmail, To: "mai@example.com"})
		require.Error(t, err, "an email without its link is useless")
	})
}

func TestLifetime(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "48 hours", lifetime(48*time.Hour))
	assert.Equal(t, "30 minutes", lifetime(30*time.Minute))
	assert.Equal(t, "90 minutes", lifetime(90*time.Minute))
	assert.Equal(t, "1 hour", lifetime(time.Hour))
	assert.Equal(t, "1 minute", lifetime(time.Minute))
}
