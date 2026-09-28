package category

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStationValid(t *testing.T) {
	t.Parallel()
	assert.True(t, StationKitchen.Valid())
	assert.True(t, StationCounter.Valid())
	assert.False(t, Station("").Valid())
	assert.False(t, Station("bar").Valid())
}
