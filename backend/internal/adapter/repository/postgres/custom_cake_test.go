package postgres_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainproduct "github.com/boms/backend/internal/domain/product"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// A customer configures a custom cake from the options a manager offers: it is
// priced by what they chose, paid into staff review, and accepted only while
// the bakery still has the notice it needs; the kitchen sees how to make it.
func TestCustomCakes_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	products := postgresadapter.NewProductRepository(f.pool)
	manager := usecase.NewManagerProductUsecase(products, postgresadapter.NewCategoryRepository(f.pool), f.pool, nil, testCloudinary, zap.NewNop())
	managerID := f.newWorker(t, domainuser.RoleManager)
	kitchenCake, err := products.GetByID(ctx, f.cake)
	require.NoError(t, err)
	cakeRequest := dto.CreateProductRequest{
		CategoryID: kitchenCake.CategoryID.String(), Name: "Celebration cake", PriceCents: 3000, IsActive: true,
		LeadTimeMinutes: 180, IsCustomizable: true,
		Options: []dto.ProductOptionInput{
			{Group: "size", Label: "16 cm", IsActive: true},
			{Group: "size", Label: "20 cm", PriceDeltaCents: 1000, IsActive: true},
			{Group: "flavor", Label: "Matcha", PriceDeltaCents: 200, IsActive: true},
			{Group: "flavor", Label: "Durian", IsActive: false},
		},
	}
	cake, err := manager.Create(ctx, managerID, domainuser.RoleManager, cakeRequest)
	require.NoError(t, err)
	require.Len(t, cake.Options, 4)
	option := func(label string) string {
		for _, o := range cake.Options {
			if o.Label == label {
				return o.ID
			}
		}
		t.Fatalf("no option %q", label)
		return ""
	}
	add := func(customer uuid.UUID, customization *dto.CartCustomizationRequest) (*dto.CartResponse, error) {
		return f.cartUC.AddItem(ctx, customer, dto.AddCartItemRequest{ProductID: &cake.ID, Quantity: 1, Customization: customization})
	}
	choose := func(labels ...string) *dto.CartCustomizationRequest {
		ids := make([]string, 0, len(labels))
		for _, label := range labels {
			ids = append(ids, option(label))
		}
		return &dto.CartCustomizationRequest{OptionIDs: ids, Message: "Happy birthday Mai"}
	}
	ownPhoto := func(customer uuid.UUID) string {
		return "https://res.cloudinary.com/demo/image/upload/v1/boms/references/" + customer.String() + "/cake.jpg"
	}
	validationField := func(t *testing.T, err error) string {
		t.Helper()
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		for field := range appErr.Details {
			return field
		}
		return ""
	}

	t.Run("a_custom_cake_is_priced_by_the_options_chosen", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)

		_, err := add(customer, choose("20 cm", "Matcha"))
		require.NoError(t, err)
		cart, err := add(customer, choose("20 cm", "Matcha"))

		require.NoError(t, err)
		require.Len(t, cart.Items, 2, "each configured cake is a line of its own")
		line := cart.Items[0]
		assert.Equal(t, int64(4200), line.UnitPriceCents)
		require.NotNil(t, line.Customization)
		assert.Equal(t, "Happy birthday Mai", line.Customization.Message)
		assert.Equal(t, "20 cm", line.Customization.Options[0].Label)
		assert.Equal(t, "Matcha", line.Customization.Options[1].Label)
	})

	t.Run("a_choice_must_fit_what_the_cake_offers", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		pastry := f.pastry.String()

		_, err := add(customer, choose("20 cm"))
		require.ErrorIs(t, err, domainproduct.ErrInvalidChoice, "a flavor is missing")
		_, err = add(customer, choose("16 cm", "20 cm", "Matcha"))
		require.ErrorIs(t, err, domainproduct.ErrInvalidChoice, "two sizes")
		_, err = add(customer, choose("16 cm", "Durian"))
		require.ErrorIs(t, err, domainproduct.ErrInvalidChoice, "a flavor no longer offered")
		_, err = add(customer, nil)
		require.ErrorIs(t, err, domainproduct.ErrCustomization, "a custom cake is configured")
		_, err = f.cartUC.AddItem(ctx, customer, dto.AddCartItemRequest{ProductID: &pastry, Quantity: 1, Customization: choose("16 cm", "Matcha")})
		require.ErrorIs(t, err, domainproduct.ErrCustomization, "a plain pastry is not")
		long := choose("16 cm", "Matcha")
		long.Message = strings.Repeat("a", domainproduct.MaxMessageLength+1)
		_, err = add(customer, long)
		require.ErrorIs(t, err, domainproduct.ErrInvalidMessage)
	})

	t.Run("a_reference_photo_is_the_customers_own_upload", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		withPhoto := choose("16 cm", "Matcha")
		withPhoto.ReferenceImageURL = ownPhoto(customer)

		_, err := add(customer, withPhoto)
		assert.Equal(t, "reference_rights_confirmed", validationField(t, err))
		someoneElses := *withPhoto
		someoneElses.ReferenceImageURL = ownPhoto(uuid.New())
		someoneElses.ReferenceRightsConfirmed = true
		_, err = add(customer, &someoneElses)
		assert.Equal(t, "reference_image_url", validationField(t, err))

		withPhoto.ReferenceRightsConfirmed = true
		cart, err := add(customer, withPhoto)
		require.NoError(t, err)
		assert.Equal(t, ownPhoto(customer), cart.Items[0].Customization.ReferenceImageURL)
	})

	t.Run("a_paid_custom_order_waits_for_staff_who_accept_it_in_time", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		_, err := add(customer, choose("20 cm", "Matcha"))
		require.NoError(t, err)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(15, 0)))
		require.NoError(t, err)
		require.NotNil(t, order.Items[0].Customization, "the order keeps the cake as it was bought")

		f.pay(t, customer, order.ID)

		paid, err := f.orderUC.Get(ctx, customer, uuid.MustParse(order.ID))
		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusPending), paid.Status, "staff review it before the kitchen starts")
		accepted, err := f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.PatchStaffOrderStatusRequest{Status: "confirmed"})
		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusConfirmed), accepted.Status)
		require.Len(t, accepted.Tickets, 1)
		ticket, err := f.kitchen.Get(ctx, uuid.MustParse(accepted.Tickets[0].ID))
		require.NoError(t, err)
		require.NotNil(t, ticket.Items[0].Customization)
		assert.Equal(t, "Happy birthday Mai", ticket.Items[0].Customization.Message, "the kitchen sees how to make it")
	})

	t.Run("a_custom_order_with_nothing_to_pay_still_waits_for_staff", func(t *testing.T) {
		_, err := postgresadapter.NewDiscountCodeRepository(f.pool).Create(ctx, port.CreateDiscountCodeParams{
			Code: "CAKEONUS", DiscountType: domaindiscount.TypePercent, Value: 100, IsActive: true,
			StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(24 * time.Hour),
		})
		require.NoError(t, err)
		customer := f.newCustomer(t, nil, nil)
		_, err = add(customer, choose("16 cm", "Matcha"))
		require.NoError(t, err)
		_, err = f.cartUC.ApplyDiscount(ctx, customer, dto.ApplyCartDiscountRequest{Code: "CAKEONUS"})
		require.NoError(t, err)

		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(14, 0)))

		require.NoError(t, err)
		assert.Zero(t, order.TotalCents)
		assert.Equal(t, string(domainorder.StatusPending), order.Status, "the discount pays it; staff still review it")
	})

	t.Run("a_request_left_waiting_too_long_is_not_accepted", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		_, err := add(customer, choose("16 cm", "Matcha"))
		require.NoError(t, err)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(16, 0)))
		require.NoError(t, err)
		f.pay(t, customer, order.ID)
		longer := dto.UpdateProductRequest{
			CategoryID: cakeRequest.CategoryID, Name: cake.Name, Slug: cake.Slug, PriceCents: cake.PriceCents,
			IsActive: true, LeadTimeMinutes: 7 * 24 * 60, IsCustomizable: true,
		}
		_, err = manager.Update(ctx, managerID, domainuser.RoleManager, uuid.MustParse(cake.ID), longer)
		require.NoError(t, err)

		_, err = f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.PatchStaffOrderStatusRequest{Status: "confirmed"})

		require.ErrorIs(t, err, domainorder.ErrPickupTooSoon, "the bakery no longer has the notice the cake needs")
	})

	t.Run("retiring_an_option_takes_the_cakes_that_chose_it_off_sale", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		_, err := add(customer, choose("16 cm", "Matcha"))
		require.NoError(t, err)
		keep := func(labels ...string) *[]dto.ProductOptionInput {
			kept := make([]dto.ProductOptionInput, 0, len(labels))
			for _, label := range labels {
				id := option(label)
				kept = append(kept, dto.ProductOptionInput{ID: &id, Group: "size", Label: label, IsActive: true})
			}
			return &kept
		}
		update := dto.UpdateProductRequest{
			CategoryID: cakeRequest.CategoryID, Name: cake.Name, Slug: cake.Slug, PriceCents: cake.PriceCents,
			IsActive: true, LeadTimeMinutes: 180, IsCustomizable: true, Options: keep("16 cm", "20 cm"),
		}
		unknown := uuid.NewString()
		withUnknown := append(*keep("16 cm"), dto.ProductOptionInput{ID: &unknown, Group: "size", Label: "30 cm", IsActive: true})
		update.Options = &withUnknown
		_, err = manager.Update(ctx, managerID, domainuser.RoleManager, uuid.MustParse(cake.ID), update)
		require.ErrorIs(t, err, domainproduct.ErrInvalidOption, "an option the cake does not have")

		update.Options = keep("16 cm", "20 cm")
		updated, err := manager.Update(ctx, managerID, domainuser.RoleManager, uuid.MustParse(cake.ID), update)
		require.NoError(t, err)

		assert.Len(t, updated.Options, 2, "the flavors are retired")
		cart, err := f.cartUC.Get(ctx, customer)
		require.NoError(t, err)
		assert.False(t, cart.Items[0].IsAvailable, "its flavor is no longer offered")
		assert.False(t, cart.CheckoutReady)
		assert.Equal(t, "Happy birthday Mai", cart.Items[0].Customization.Message, "the customer still sees what it was")
	})
}
