package usecase

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
)

// Each role receives only the customer data its work needs (proposal 4.17.2):
// the counter phones a customer about a pickup, the kitchen never contacts one.
// Asserted on the JSON the API sends, since that is what leaves the server.
func TestOrderCustomerVisibilityByRole(t *testing.T) {
	t.Parallel()

	name := "Mai"
	phone := "0901234567"
	row := &port.StaffOrderListRow{
		Order:               domainorder.Order{ID: uuid.New(), UserID: uuid.New()},
		CustomerEmail:       "mai@example.com",
		CustomerDisplayName: &name,
		CustomerPhone:       &phone,
	}

	customerJSON := func(t *testing.T, v any) map[string]any {
		t.Helper()
		raw, err := json.Marshal(v)
		require.NoError(t, err)
		var out map[string]any
		require.NoError(t, json.Unmarshal(raw, &out))
		return out
	}

	t.Run("staff_can_reach_the_customer", func(t *testing.T) {
		t.Parallel()
		got := customerJSON(t, toStaffOrderCustomer(row))
		assert.Equal(t, "mai@example.com", got["email"])
		assert.Equal(t, phone, got["phone"])
		assert.Equal(t, name, got["display_name"])
	})

	t.Run("stations_see_a_name_and_no_contact_details", func(t *testing.T) {
		t.Parallel()
		ticket := toStationTicketResponse(port.StationTicket{
			OrderCode: "CH-260928-001", OrderStatus: domainorder.StatusConfirmed, CustomerDisplayName: row.CustomerDisplayName,
		})
		got := customerJSON(t, ticket.Customer)
		assert.Equal(t, map[string]any{"display_name": name}, got)
	})
}
