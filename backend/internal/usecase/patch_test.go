package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvePatchString(t *testing.T) {
	t.Parallel()

	fallback := strPtr("existing")

	assert.Nil(t, resolvePatchString(nil, nil))
	assert.Equal(t, fallback, resolvePatchString(nil, fallback))
	assert.Nil(t, resolvePatchString(strPtr(""), fallback))
	assert.Nil(t, resolvePatchString(strPtr("   "), fallback))
	assert.Equal(t, strPtr("555-0100"), resolvePatchString(strPtr(" 555-0100 "), fallback))
}

func strPtr(value string) *string {
	return &value
}
