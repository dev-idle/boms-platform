package usecase

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
)

// The counter sees which role moved an order; the customer sees only when it
// moved. Asserted on the JSON the API sends.
func TestOrderTimelineVisibilityByRole(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 9, 28, 3, 5, 0, 0, time.UTC)
	events := []domainorder.StatusEvent{{To: domainorder.StatusConfirmed, ActorRole: domainuser.RoleStaff, At: at}}
	entries := func(t *testing.T, v any) []map[string]any {
		t.Helper()
		raw, err := json.Marshal(v)
		require.NoError(t, err)
		var out []map[string]any
		require.NoError(t, json.Unmarshal(raw, &out))
		return out
	}

	t.Run("customers_see_the_status_and_when", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, []map[string]any{{"status": "confirmed", "at": "2026-09-28T03:05:00Z"}},
			entries(t, mapOrderTimelineToDTO(events)))
	})

	t.Run("staff_also_see_who_moved_it", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, []map[string]any{{"status": "confirmed", "actor_role": "staff", "at": "2026-09-28T03:05:00Z"}},
			entries(t, mapStaffOrderTimelineToDTO(events)))
	})
}
