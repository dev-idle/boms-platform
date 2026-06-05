package usecase

import (
	"context"
	"time"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
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
	for _, product := range products {
		productByID[product.ID] = product
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
			line.UnitPriceCents = product.PriceCents
			line.LineTotalCents = product.PriceCents * int64(item.Quantity)
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
