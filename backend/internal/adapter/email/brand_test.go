package email

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// contracts/brand.json, which the storefront's BRAND is tested against too: a
// customer reads the same address and phone in an email and on the site.
func TestBrand_MatchesTheStorefront(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../../contracts/brand.json")
	require.NoError(t, err)
	var contract struct {
		Name         string `json:"name"`
		AddressLine  string `json:"address_line"`
		ContactEmail string `json:"contact_email"`
		ContactPhone string `json:"contact_phone"`
	}
	require.NoError(t, json.Unmarshal(raw, &contract))

	assert.Equal(t, brand{
		Name:         contract.Name,
		AddressLine:  contract.AddressLine,
		ContactEmail: contract.ContactEmail,
		ContactPhone: contract.ContactPhone,
	}, choux)
}
