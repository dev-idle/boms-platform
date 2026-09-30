package account

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainevent "github.com/boms/backend/internal/domain/event"
)

func TestEmailFor(t *testing.T) {
	t.Parallel()
	userID := uuid.New()

	for _, tc := range []struct {
		event domainevent.Event
		want  Purpose
	}{
		{VerificationRequestedEvent(userID), PurposeVerifyEmail},
		{PasswordResetRequestedEvent(userID), PurposeResetPassword},
	} {
		gotUser, purpose, ok := EmailFor(tc.event)
		require.True(t, ok)
		assert.Equal(t, tc.want, purpose)
		assert.Equal(t, userID, gotUser)
		assert.Empty(t, tc.event.Audience.UserIDs, "a request never reaches an open page")
		assert.Empty(t, tc.event.Audience.Roles)
		assert.False(t, tc.event.Audience.Public)
	}

	_, _, ok := EmailFor(domainevent.New("order.created", domainevent.Audience{}, map[string]string{"user_id": userID.String()}))
	assert.False(t, ok, "other topics call for no account email")

	bad := VerificationRequestedEvent(userID)
	bad.Data["user_id"] = "x"
	_, _, ok = EmailFor(bad)
	assert.False(t, ok)
}

func TestPurpose(t *testing.T) {
	t.Parallel()
	assert.True(t, PurposeVerifyEmail.Valid())
	assert.True(t, PurposeResetPassword.Valid())
	assert.False(t, Purpose("login").Valid())
	assert.Equal(t, 30*time.Minute, PurposeResetPassword.TTL())
	assert.Equal(t, 48*time.Hour, PurposeVerifyEmail.TTL())
}
