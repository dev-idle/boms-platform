package usecase

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainproduct "github.com/boms/backend/internal/domain/product"
	"github.com/boms/backend/internal/port"
)

func TestCustomizedPrice(t *testing.T) {
	t.Parallel()
	large := domainproduct.Option{ID: uuid.New(), Group: domainproduct.OptionSize, Label: "20 cm", PriceDeltaCents: 1000, IsActive: true}
	configured, err := json.Marshal(map[string]any{"option_ids": []uuid.UUID{large.ID}, "message": "Happy birthday"})
	require.NoError(t, err)
	custom := port.CatalogListProduct{PriceCents: 3000, IsCustomizable: true}
	plain := port.CatalogListProduct{PriceCents: 300}

	price, customization, available, err := customizedPrice(custom, []domainproduct.Option{large}, configured)
	require.NoError(t, err)
	assert.True(t, available)
	assert.Equal(t, int64(4000), price, "the product's price plus what the option adds")
	assert.Equal(t, "20 cm", customization.Options[0].Label)

	price, customization, available, err = customizedPrice(plain, nil, json.RawMessage("{}"))
	require.NoError(t, err)
	assert.True(t, available)
	assert.Equal(t, int64(300), price)
	assert.Nil(t, customization)

	for name, tc := range map[string]struct {
		product       port.CatalogListProduct
		offered       []domainproduct.Option
		configuration json.RawMessage
	}{
		"configured_when_the_product_no_longer_is_customizable": {plain, nil, configured},
		"plain_when_the_product_became_customizable":            {custom, []domainproduct.Option{large}, json.RawMessage("{}")},
		"its_option_no_longer_offered":                          {custom, nil, configured},
	} {
		_, _, available, err := customizedPrice(tc.product, tc.offered, tc.configuration)
		require.NoError(t, err, name)
		assert.False(t, available, name)
	}
}
