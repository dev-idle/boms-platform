package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/boms/backend/internal/shared/errors"
)

func TestProductLeadTime(t *testing.T) {
	t.Parallel()

	for minutes, want := range map[int32]time.Duration{0: 0, 90: 90 * time.Minute, 10080: 7 * 24 * time.Hour} {
		got, err := productLeadTime(minutes)
		require.NoError(t, err, "%d minutes", minutes)
		assert.Equal(t, want, got)
	}
	for _, minutes := range []int32{-1, 10081} {
		_, err := productLeadTime(minutes)
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr, "%d minutes", minutes)
		assert.Contains(t, appErr.Details, "lead_time_minutes")
	}
}
