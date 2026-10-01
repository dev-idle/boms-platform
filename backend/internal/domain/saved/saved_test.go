package saved

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListValid(t *testing.T) {
	t.Parallel()
	assert.True(t, ListFavorite.Valid())
	assert.True(t, ListWishlist.Valid())
	assert.False(t, List("basket").Valid())
	assert.False(t, List("").Valid())
}
