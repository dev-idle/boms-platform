package conversation

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainuser "github.com/boms/backend/internal/domain/user"
)

func TestNewBody(t *testing.T) {
	t.Parallel()
	body, err := NewBody("  Hello,\r\nis it nut free?  ")
	require.NoError(t, err)
	assert.Equal(t, "Hello,\nis it nut free?", body, "trimmed, one kind of line break")

	cook := "Thank you " + string([]rune{0x1f469, 0x200d, 0x1f373})
	body, err = NewBody(cook)
	require.NoError(t, err, "an emoji built with a zero-width joiner is plain text")
	assert.Equal(t, cook, body)

	longest := strings.Repeat("é", MaxBodyLength)
	body, err = NewBody(longest)
	require.NoError(t, err)
	assert.Equal(t, longest, body, "the limit counts characters, not bytes")

	for _, bad := range []string{"", " \n ", longest + "a", "Hi" + string(rune(0x202e)) + "there", "tab\there"} {
		_, err := NewBody(bad)
		assert.ErrorIs(t, err, ErrInvalidBody, "%q", bad)
	}
}

func TestStatusValid(t *testing.T) {
	t.Parallel()
	assert.True(t, StatusOpen.Valid())
	assert.True(t, StatusClosed.Valid())
	assert.False(t, Status("archived").Valid())
	assert.False(t, Status("").Valid())
}

func TestEventAudiences(t *testing.T) {
	t.Parallel()
	orderID, customerID := uuid.New(), uuid.New()

	created := MessageCreatedEvent(orderID, customerID)
	assert.Equal(t, TopicMessageCreated, created.Topic)
	assert.Equal(t, []uuid.UUID{customerID}, created.Audience.UserIDs)
	assert.Equal(t, []domainuser.Role{domainuser.RoleStaff}, created.Audience.Roles)
	assert.Equal(t, orderID.String(), created.Data["order_id"])

	read := ReadByCustomerEvent(orderID, customerID)
	assert.Equal(t, []uuid.UUID{customerID}, read.Audience.UserIDs)
	assert.Empty(t, read.Audience.Roles, "the counter sees no read receipts")

	counter := ChangedAtCounterEvent(orderID)
	assert.Empty(t, counter.Audience.UserIDs, "the customer does not see the inbox")
	assert.Equal(t, []domainuser.Role{domainuser.RoleStaff}, counter.Audience.Roles)
}
