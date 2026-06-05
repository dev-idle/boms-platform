package usecase

import (
	"context"
	"errors"
	"time"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/google/uuid"
)

type CartUsecase struct {
	carts    port.CartRepository
	products port.ProductRepository
	combos   port.ComboRepository
	discount port.DiscountCodeRepository
	pricer   *cartPricer
}

func NewCartUsecase(
	carts port.CartRepository,
	products port.ProductRepository,
	combos port.ComboRepository,
	discount port.DiscountCodeRepository,
) *CartUsecase {
	return &CartUsecase{
		carts:    carts,
		products: products,
		combos:   combos,
		discount: discount,
		pricer:   newCartPricer(products, combos),
	}
}

func (u *CartUsecase) Get(ctx context.Context, userID uuid.UUID) (*dto.CartResponse, error) {
	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.buildCartResponse(ctx, cart)
}

func (u *CartUsecase) AddItem(ctx context.Context, userID uuid.UUID, req dto.AddCartItemRequest) (*dto.CartResponse, error) {
	productID, comboID, err := parseCartLineTarget(req.ProductID, req.ComboID)
	if err != nil {
		return nil, err
	}
	if err := u.ensureLinePurchasable(ctx, productID, comboID); err != nil {
		return nil, err
	}

	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	if productID != nil {
		if resp, ok, err := u.mergeCartLine(ctx, cart, func() (*domaincart.Item, error) {
			return u.carts.GetItemByProduct(ctx, cart.ID, *productID)
		}, req.Quantity); err != nil {
			return nil, err
		} else if ok {
			return resp, nil
		}
	}
	if comboID != nil {
		if resp, ok, err := u.mergeCartLine(ctx, cart, func() (*domaincart.Item, error) {
			return u.carts.GetItemByCombo(ctx, cart.ID, *comboID)
		}, req.Quantity); err != nil {
			return nil, err
		} else if ok {
			return resp, nil
		}
	}

	count, err := u.carts.CountItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	if count >= cartMaxItems {
		return nil, domaincart.ErrMaxItemsReached
	}

	params := port.CreateCartItemParams{
		CartID:   cart.ID,
		Quantity: req.Quantity,
	}
	if productID != nil {
		params.LineType = domaincart.LineTypeProduct
		params.ProductID = productID
	} else {
		params.LineType = domaincart.LineTypeCombo
		params.ComboID = comboID
	}
	if _, err := u.carts.CreateItem(ctx, params); err != nil {
		return nil, err
	}
	return u.buildCartResponse(ctx, cart)
}

func (u *CartUsecase) UpdateItem(
	ctx context.Context,
	userID, itemID uuid.UUID,
	req dto.UpdateCartItemRequest,
) (*dto.CartResponse, error) {
	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	if _, err := u.carts.GetItemByID(ctx, cart.ID, itemID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaincart.ErrItemNotFound
		}
		return nil, err
	}
	if _, err := u.carts.UpdateItemQuantity(ctx, port.UpdateCartItemQuantityParams{
		ID:       itemID,
		CartID:   cart.ID,
		Quantity: req.Quantity,
	}); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaincart.ErrItemNotFound
		}
		return nil, err
	}
	return u.buildCartResponse(ctx, cart)
}

func (u *CartUsecase) RemoveItem(ctx context.Context, userID, itemID uuid.UUID) (*dto.CartResponse, error) {
	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := u.carts.DeleteItem(ctx, cart.ID, itemID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaincart.ErrItemNotFound
		}
		return nil, err
	}
	return u.buildCartResponse(ctx, cart)
}

func (u *CartUsecase) ApplyDiscount(ctx context.Context, userID uuid.UUID, req dto.ApplyCartDiscountRequest) (*dto.CartResponse, error) {
	code, err := domaindiscount.NormalizeCode(req.Code)
	if err != nil {
		return nil, apperrors.ErrValidation.WithDetail("code", "must be 3-64 uppercase letters, numbers, hyphens, or underscores")
	}
	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	loaded, err := u.loadPricedCart(ctx, cart)
	if err != nil {
		return nil, err
	}
	if !loaded.totals.CheckoutReady {
		return nil, domaincart.ErrEmpty
	}

	discountCode, err := u.discount.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaindiscount.ErrNotFound
		}
		return nil, err
	}
	now := time.Now().UTC()
	if err := domaindiscount.ValidateRedeemable(discountCode, now, loaded.totals.SubtotalCents); err != nil {
		return nil, err
	}
	if err := u.carts.SetDiscountCodeID(ctx, cart.ID, &discountCode.ID); err != nil {
		return nil, err
	}
	cart.DiscountCodeID = &discountCode.ID
	totals := summarizeCart(loaded.lines, discountCode, now)
	return u.buildCartResponseFromLines(cart, loaded.lines, discountCode, totals)
}

func (u *CartUsecase) RemoveDiscount(ctx context.Context, userID uuid.UUID) (*dto.CartResponse, error) {
	cart, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := u.carts.ClearDiscountCode(ctx, cart.ID); err != nil {
		return nil, err
	}
	cart.DiscountCodeID = nil
	return u.buildCartResponse(ctx, cart)
}

func (u *CartUsecase) getOrCreateCart(ctx context.Context, userID uuid.UUID) (*domaincart.Cart, error) {
	cart, err := u.carts.GetByUserID(ctx, userID)
	if err == nil {
		return cart, nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	created, err := u.carts.CreateForUser(ctx, userID)
	if err == nil {
		return created, nil
	}
	if errors.Is(err, apperrors.ErrConflict) {
		return u.carts.GetByUserID(ctx, userID)
	}
	return nil, err
}

func (u *CartUsecase) mergeCartLine(
	ctx context.Context,
	cart *domaincart.Cart,
	lookup func() (*domaincart.Item, error),
	addQty int32,
) (*dto.CartResponse, bool, error) {
	existing, err := lookup()
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	newQty := existing.Quantity + addQty
	if newQty > cartMaxItemQuantity {
		return nil, false, domaincart.ErrQuantityOutOfRange
	}
	if _, err := u.carts.UpdateItemQuantity(ctx, port.UpdateCartItemQuantityParams{
		ID:       existing.ID,
		CartID:   cart.ID,
		Quantity: newQty,
	}); err != nil {
		return nil, false, err
	}
	resp, err := u.buildCartResponse(ctx, cart)
	return resp, true, err
}

func (u *CartUsecase) ensureLinePurchasable(ctx context.Context, productID, comboID *uuid.UUID) error {
	if productID != nil {
		if _, err := u.products.CatalogGetByID(ctx, *productID); err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return domaincart.ErrProductUnavailable
			}
			return err
		}
		return nil
	}
	if comboID != nil {
		if _, err := u.combos.CatalogGetByID(ctx, *comboID); err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return domaincart.ErrComboUnavailable
			}
			return err
		}
		return nil
	}
	return domaincart.ErrInvalidLine
}

func (u *CartUsecase) buildCartResponse(ctx context.Context, cart *domaincart.Cart) (*dto.CartResponse, error) {
	loaded, err := u.loadPricedCart(ctx, cart)
	if err != nil {
		return nil, err
	}
	var discountCode *domaindiscount.Code
	if cart.DiscountCodeID != nil {
		discountCode, err = u.discount.GetByID(ctx, *cart.DiscountCodeID)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				_ = u.carts.ClearDiscountCode(ctx, cart.ID)
				discountCode = nil
			} else {
				return nil, err
			}
		} else if domaindiscount.ValidateRedeemable(discountCode, time.Now().UTC(), loaded.totals.SubtotalCents) != nil {
			_ = u.carts.ClearDiscountCode(ctx, cart.ID)
			discountCode = nil
			loaded.totals = summarizeCart(loaded.lines, nil, time.Now().UTC())
		} else {
			loaded.totals = summarizeCart(loaded.lines, discountCode, time.Now().UTC())
		}
	}
	return u.buildCartResponseFromLines(cart, loaded.lines, discountCode, loaded.totals)
}

type loadedPricedCart struct {
	lines  []pricedCartLine
	totals cartTotals
}

func (u *CartUsecase) loadPricedCart(ctx context.Context, cart *domaincart.Cart) (loadedPricedCart, error) {
	items, err := u.carts.ListItemsByCartID(ctx, cart.ID)
	if err != nil {
		return loadedPricedCart{}, err
	}
	lines, err := u.pricer.priceLines(ctx, items)
	if err != nil {
		return loadedPricedCart{}, err
	}
	now := time.Now().UTC()
	totals := summarizeCart(lines, nil, now)
	return loadedPricedCart{lines: lines, totals: totals}, nil
}

func (u *CartUsecase) buildCartResponseFromLines(
	cart *domaincart.Cart,
	lines []pricedCartLine,
	discountCode *domaindiscount.Code,
	totals cartTotals,
) (*dto.CartResponse, error) {
	resp := &dto.CartResponse{
		ID:            cart.ID.String(),
		Items:         make([]dto.CartItemResponse, 0, len(lines)),
		SubtotalCents: totals.SubtotalCents,
		DiscountCents: totals.DiscountCents,
		TotalCents:    totals.TotalCents,
		CheckoutReady: totals.CheckoutReady,
	}
	for _, line := range lines {
		item := dto.CartItemResponse{
			ID:             line.Item.ID.String(),
			LineType:       string(line.Item.LineType),
			Name:           line.Name,
			Slug:           line.Slug,
			Quantity:       line.Item.Quantity,
			UnitPriceCents: line.UnitPriceCents,
			LineTotalCents: line.LineTotalCents,
			IsAvailable:    line.IsAvailable,
		}
		if line.Item.ProductID != nil {
			id := line.Item.ProductID.String()
			item.ProductID = &id
		}
		if line.Item.ComboID != nil {
			id := line.Item.ComboID.String()
			item.ComboID = &id
		}
		resp.Items = append(resp.Items, item)
	}
	if discountCode != nil && totals.DiscountCents > 0 {
		resp.Discount = &dto.CartDiscountResponse{
			Code:          discountCode.Code,
			DiscountType:  string(discountCode.DiscountType),
			Value:         discountCode.Value,
			DiscountCents: totals.DiscountCents,
		}
	}
	return resp, nil
}

func parseCartLineTarget(productIDRaw, comboIDRaw *string) (*uuid.UUID, *uuid.UUID, error) {
	hasProduct := productIDRaw != nil && *productIDRaw != ""
	hasCombo := comboIDRaw != nil && *comboIDRaw != ""
	if hasProduct == hasCombo {
		return nil, nil, domaincart.ErrInvalidLine
	}
	if hasProduct {
		id, err := uuid.Parse(*productIDRaw)
		if err != nil {
			return nil, nil, apperrors.ErrValidation.WithDetail("product_id", "invalid uuid")
		}
		return &id, nil, nil
	}
	id, err := uuid.Parse(*comboIDRaw)
	if err != nil {
		return nil, nil, apperrors.ErrValidation.WithDetail("combo_id", "invalid uuid")
	}
	return nil, &id, nil
}

func (u *CartUsecase) pricedCartForCheckout(ctx context.Context, userID uuid.UUID) (
	*domaincart.Cart,
	[]pricedCartLine,
	*domaindiscount.Code,
	cartTotals,
	error,
) {
	cart, err := u.carts.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, nil, nil, cartTotals{}, domaincart.ErrEmpty
		}
		return nil, nil, nil, cartTotals{}, err
	}
	items, err := u.carts.ListItemsByCartID(ctx, cart.ID)
	if err != nil {
		return nil, nil, nil, cartTotals{}, err
	}
	if len(items) == 0 {
		return nil, nil, nil, cartTotals{}, domaincart.ErrEmpty
	}
	lines, err := u.pricer.priceLines(ctx, items)
	if err != nil {
		return nil, nil, nil, cartTotals{}, err
	}
	now := time.Now().UTC()
	subtotal := summarizeCart(lines, nil, now).SubtotalCents
	var discountCode *domaindiscount.Code
	if cart.DiscountCodeID != nil {
		discountCode, err = u.discount.GetByID(ctx, *cart.DiscountCodeID)
		if err != nil {
			return nil, nil, nil, cartTotals{}, err
		}
		if err := domaindiscount.ValidateRedeemable(discountCode, now, subtotal); err != nil {
			return nil, nil, nil, cartTotals{}, err
		}
	}
	totals := summarizeCart(lines, discountCode, now)
	if !totals.CheckoutReady {
		return nil, nil, nil, cartTotals{}, domaincart.ErrProductUnavailable
	}
	for _, line := range lines {
		if !line.IsAvailable {
			return nil, nil, nil, cartTotals{}, domaincart.ErrProductUnavailable
		}
	}
	return cart, lines, discountCode, totals, nil
}
