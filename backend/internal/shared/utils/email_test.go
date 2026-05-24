package utils_test

import (
	"testing"

	"github.com/boms/backend/internal/shared/utils"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "user@example.com", utils.NormalizeEmail("  User@Example.COM  "))
	assert.Equal(t, "a@b.co", utils.NormalizeEmail("a@b.co"))
}
