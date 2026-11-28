package usecase

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domaincart "github.com/boms/backend/internal/domain/cart"
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

// catalogProducts answers the cart's product lookup with fixed products.
type catalogProducts struct {
	port.ProductRepository
	products []port.CatalogListProduct
}

func (p catalogProducts) CatalogGetByIDs(context.Context, []uuid.UUID) ([]port.CatalogListProduct, error) {
	return p.products, nil
}

func (catalogProducts) ListOptions(context.Context, []uuid.UUID) ([]domainproduct.Option, error) {
	return nil, nil
}

// catalogCombos answers the cart's combo lookup with fixed combos.
type catalogCombos struct {
	port.ComboRepository
	combos []port.CatalogCombo
}

func (c catalogCombos) CatalogGetByIDs(context.Context, []uuid.UUID) ([]port.CatalogCombo, error) {
	return c.combos, nil
}

func TestCartPricer_LinesCarryTheirPhoto(t *testing.T) {
	t.Parallel()
	cakePhoto := "https://res.cloudinary.com/demo/image/upload/boms/products/matcha-cake-1.jpg"
	boxPhoto := "https://res.cloudinary.com/demo/image/upload/boms/combos/tea-box.jpg"
	cake := port.CatalogListProduct{ID: uuid.New(), Name: "Matcha cake", CategoryName: "Cakes", PriceCents: 4500, ImageURLs: []string{cakePhoto}}
	tart := port.CatalogListProduct{ID: uuid.New(), Name: "Fruit tart", CategoryName: "Tarts", PriceCents: 3200}
	box := port.CatalogCombo{ID: uuid.New(), Name: "Tea box", PriceCents: 7000, ImageURL: &boxPhoto}
	plainBox := port.CatalogCombo{ID: uuid.New(), Name: "Plain box", PriceCents: 6000}
	pricer := newCartPricer(
		catalogProducts{products: []port.CatalogListProduct{cake, tart}},
		catalogCombos{combos: []port.CatalogCombo{box, plainBox}},
	)

	lines, err := pricer.priceLines(context.Background(), []domaincart.Item{
		{LineType: domaincart.LineTypeProduct, ProductID: &cake.ID, Quantity: 1, Configuration: json.RawMessage("{}")},
		{LineType: domaincart.LineTypeProduct, ProductID: &tart.ID, Quantity: 1, Configuration: json.RawMessage("{}")},
		{LineType: domaincart.LineTypeCombo, ComboID: &box.ID, Quantity: 1},
		{LineType: domaincart.LineTypeCombo, ComboID: &plainBox.ID, Quantity: 1},
	})
	require.NoError(t, err)
	require.Len(t, lines, 4)
	assert.Equal(t, [2]string{cakePhoto, "Cakes"}, [2]string{lines[0].ImageURL, lines[0].CategoryName})
	assert.Equal(t, [2]string{"", "Tarts"}, [2]string{lines[1].ImageURL, lines[1].CategoryName})
	assert.Equal(t, [2]string{boxPhoto, ""}, [2]string{lines[2].ImageURL, lines[2].CategoryName})
	assert.Equal(t, [2]string{"", ""}, [2]string{lines[3].ImageURL, lines[3].CategoryName})
}
