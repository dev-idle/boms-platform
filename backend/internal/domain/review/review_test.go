package review

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainuser "github.com/boms/backend/internal/domain/user"
)

func TestNewRating(t *testing.T) {
	t.Parallel()
	for rating := MinRating; rating <= MaxRating; rating++ {
		got, err := NewRating(rating)
		require.NoError(t, err)
		assert.Equal(t, int16(rating), got)
	}
	for _, bad := range []int{0, -1, 6, 100} {
		_, err := NewRating(bad)
		assert.ErrorIs(t, err, ErrInvalidRating, "%d", bad)
	}
}

func TestNewComment(t *testing.T) {
	t.Parallel()
	comment, err := NewComment(new("  Light and crisp,\r\nnot too sweet  "))
	require.NoError(t, err)
	assert.Equal(t, "Light and crisp,\nnot too sweet", *comment, "trimmed, one kind of line break")

	for _, nothing := range []*string{nil, new(""), new(" \n ")} {
		comment, err := NewComment(nothing)
		require.NoError(t, err)
		assert.Nil(t, comment, "nothing written is no comment")
	}

	longest := strings.Repeat("é", MaxCommentLength)
	comment, err = NewComment(&longest)
	require.NoError(t, err)
	assert.Equal(t, longest, *comment, "the limit counts characters, not bytes")

	for _, bad := range []string{longest + "a", "Good" + string(rune(0x202e)) + "bad", "tab\there"} {
		_, err := NewComment(&bad)
		assert.ErrorIs(t, err, ErrInvalidComment, "%q", bad)
	}
}

func TestNewModeration(t *testing.T) {
	t.Parallel()
	for _, status := range []Status{StatusPublished, StatusHidden} {
		got, err := NewModeration(string(status))
		require.NoError(t, err)
		assert.Equal(t, status, got)
	}
	for _, bad := range []string{"pending", "", "deleted"} {
		_, err := NewModeration(bad)
		assert.ErrorIs(t, err, ErrInvalidStatus, "%q", bad)
	}
}

func TestStatusValid(t *testing.T) {
	t.Parallel()
	assert.True(t, StatusPending.Valid())
	assert.True(t, StatusPublished.Valid())
	assert.True(t, StatusHidden.Valid())
	assert.False(t, Status("deleted").Valid())
	assert.False(t, Status("").Valid())
}

func TestEventAudiences(t *testing.T) {
	t.Parallel()
	productID, authorID := uuid.New(), uuid.New()

	changed := ChangedEvent(productID, authorID)
	assert.Equal(t, TopicChanged, changed.Topic)
	assert.Equal(t, []uuid.UUID{authorID}, changed.Audience.UserIDs)
	assert.Equal(t, []domainuser.Role{domainuser.RoleManager}, changed.Audience.Roles)
	assert.False(t, changed.Audience.Public, "nobody else is told")
	assert.Equal(t, productID.String(), changed.Data["product_id"])
}
