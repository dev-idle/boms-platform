package usecase

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainproduct "github.com/boms/backend/internal/domain/product"
	"github.com/boms/backend/internal/port"
	"github.com/google/uuid"
)

const (
	cartMaxItems        int64 = 50
	cartMaxItemQuantity int32 = 99
)

type pricedCartLine struct {
	Item           domaincart.Item
	Name           string
	Slug           string
	UnitPriceCents int64
	LineTotalCents int64
	IsAvailable    bool
	// Customization is the line as configured, priced now; nil for a plain line.
	Customization *domainorder.Customization
}

type cartTotals struct {
	SubtotalCents int64
	DiscountCents int64
	TotalCents    int64
	CheckoutReady bool
}

type cartPricer struct {
	products port.ProductRepository
	combos   port.ComboRepository
}

func newCartPricer(products port.ProductRepository, combos port.ComboRepository) *cartPricer {
	return &cartPricer{products: products, combos: combos}
}

func (p *cartPricer) priceLines(ctx context.Context, items []domaincart.Item) ([]pricedCartLine, error) {
	productIDs := make([]uuid.UUID, 0)
	comboIDs := make([]uuid.UUID, 0)
	for _, item := range items {
		switch item.LineType {
		case domaincart.LineTypeProduct:
			if item.ProductID != nil {
				productIDs = append(productIDs, *item.ProductID)
			}
		case domaincart.LineTypeCombo:
			if item.ComboID != nil {
				comboIDs = append(comboIDs, *item.ComboID)
			}
		}
	}

	products, err := p.products.CatalogGetByIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}
	combos, err := p.combos.CatalogGetByIDs(ctx, comboIDs)
	if err != nil {
		return nil, err
	}

	productByID := make(map[uuid.UUID]port.CatalogListProduct, len(products))
	customizable := make([]uuid.UUID, 0)
	for _, product := range products {
		productByID[product.ID] = product
		if product.IsCustomizable {
			customizable = append(customizable, product.ID)
		}
	}
	options, err := p.products.ListOptions(ctx, customizable)
	if err != nil {
		return nil, err
	}
	optionsByProduct := make(map[uuid.UUID][]domainproduct.Option, len(customizable))
	for _, option := range options {
		optionsByProduct[option.ProductID] = append(optionsByProduct[option.ProductID], option)
	}
	comboByID := make(map[uuid.UUID]port.CatalogCombo, len(combos))
	for _, combo := range combos {
		comboByID[combo.ID] = combo
	}

	out := make([]pricedCartLine, 0, len(items))
	for _, item := range items {
		line := pricedCartLine{Item: item, IsAvailable: false}
		switch item.LineType {
		case domaincart.LineTypeProduct:
			if item.ProductID == nil {
				continue
			}
			product, ok := productByID[*item.ProductID]
			if !ok {
				out = append(out, line)
				continue
			}
			line.Name = product.Name
			line.Slug = product.Slug
			unitPrice, customization, available, err := customizedPrice(product, optionsByProduct[product.ID], item.Configuration)
			if err != nil {
				return nil, err
			}
			line.Customization = customization
			if !available {
				out = append(out, line)
				continue
			}
			line.UnitPriceCents = unitPrice
			line.LineTotalCents = unitPrice * int64(item.Quantity)
			line.IsAvailable = true
		case domaincart.LineTypeCombo:
			if item.ComboID == nil {
				continue
			}
			combo, ok := comboByID[*item.ComboID]
			if !ok {
				out = append(out, line)
				continue
			}
			line.Name = combo.Name
			line.Slug = combo.Slug
			line.UnitPriceCents = combo.PriceCents
			line.LineTotalCents = combo.PriceCents * int64(item.Quantity)
			line.IsAvailable = true
		}
		out = append(out, line)
	}
	return out, nil
}

// customizedPrice prices a product line as configured: the product's price
// plus what each chosen option adds now. The line stops being available when
// it no longer matches what the product offers — configured when the product
// is not customizable, or the other way round, or its options retired or
// changed — and keeps its message and photo so the customer sees what it was.
func customizedPrice(
	product port.CatalogListProduct,
	offered []domainproduct.Option,
	configuration json.RawMessage,
) (int64, *domainorder.Customization, bool, error) {
	c, err := domaincart.ParseCustomization(configuration)
	if err != nil {
		return 0, nil, false, err
	}
	if c == nil {
		return product.PriceCents, nil, !product.IsCustomizable, nil
	}
	snapshot := &domainorder.Customization{Message: c.Message, ReferenceImageURL: c.ReferenceImageURL}
	if !product.IsCustomizable {
		return 0, snapshot, false, nil
	}
	chosen, err := domainproduct.Choose(offered, c.OptionIDs)
	if err != nil {
		// The choice no longer fits the product: the line waits to be removed.
		return 0, snapshot, false, nil //nolint:nilerr // an outdated choice is a state of the line, not a failure
	}
	price := product.PriceCents
	snapshot.Options = make([]domainorder.ChosenOption, 0, len(chosen))
	for _, option := range chosen {
		price += option.PriceDeltaCents
		snapshot.Options = append(snapshot.Options, domainorder.ChosenOption{
			Group: option.Group, Label: option.Label, PriceDeltaCents: option.PriceDeltaCents,
		})
	}
	return price, snapshot, true, nil
}

// hasCustomLine reports whether any line is configured, which sends the order
// to staff review once it is paid.
func hasCustomLine(lines []pricedCartLine) bool {
	return slices.ContainsFunc(lines, func(line pricedCartLine) bool { return line.Customization != nil })
}

// fulfillmentOf is what the priced lines ask of the bakery. Checkout reads it
// from the lines it turns into the order, so an edit to the cart racing the
// checkout cannot change the order's type without changing its items.
func (u *CartUsecase) fulfillmentOf(ctx context.Context, lines []pricedCartLine) (domainorder.Fulfillment, error) {
	productIDs := make([]uuid.UUID, 0, len(lines))
	comboIDs := make([]uuid.UUID, 0, len(lines))
	for _, line := range lines {
		if line.Item.ProductID != nil {
			productIDs = append(productIDs, *line.Item.ProductID)
		}
		if line.Item.ComboID != nil {
			comboIDs = append(comboIDs, *line.Item.ComboID)
		}
	}
	return u.products.FulfillmentOf(ctx, productIDs, comboIDs)
}

// orderFulfillment is what an order's items ask of the bakery now, read the
// way a cart's are: moving its pickup follows the checkout rules.
func (u *CartUsecase) orderFulfillment(ctx context.Context, items []domainorder.Item) (domainorder.Fulfillment, error) {
	productIDs := make([]uuid.UUID, 0, len(items))
	comboIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if item.ProductID != nil {
			productIDs = append(productIDs, *item.ProductID)
		}
		if item.ComboID != nil {
			comboIDs = append(comboIDs, *item.ComboID)
		}
	}
	return u.products.FulfillmentOf(ctx, productIDs, comboIDs)
}

func summarizeCart(lines []pricedCartLine, discount *domaindiscount.Code, now time.Time) cartTotals {
	var subtotal int64
	checkoutReady := false
	for _, line := range lines {
		if !line.IsAvailable {
			continue
		}
		subtotal += line.LineTotalCents
		checkoutReady = true
	}
	if subtotal == 0 {
		return cartTotals{}
	}

	discountCents := int64(0)
	if discount != nil {
		if domaindiscount.ValidateRedeemable(discount, now, subtotal) == nil {
			discountCents = domaindiscount.ComputeDiscountCents(discount, subtotal)
		}
	}
	total := subtotal - discountCents
	if total < 0 {
		total = 0
	}
	return cartTotals{
		SubtotalCents: subtotal,
		DiscountCents: discountCents,
		TotalCents:    total,
		CheckoutReady: checkoutReady,
	}
}
