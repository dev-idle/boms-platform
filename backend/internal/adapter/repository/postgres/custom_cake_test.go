package postgres_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainproduct "github.com/boms/backend/internal/domain/product"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
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
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
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
			{Group: "filling", Label: "Salted caramel", PriceDeltaCents: 300, IsActive: true},
		},
	}
	cake, err := manager.Create(ctx, managerID, domainuser.RoleManager, cakeRequest)
	require.NoError(t, err)
	require.Len(t, cake.Options, 5)
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
	withPhoto := func(customer uuid.UUID) *dto.CartCustomizationRequest {
		customization := choose("16 cm", "Matcha", "Salted caramel")
		customization.ReferenceImageURL = ownPhoto(customer)
		customization.ReferenceRightsConfirmed = true
		return customization
	}
	paidWithPhoto := func(t *testing.T, pickupAt time.Time) *dto.OrderResponse {
		t.Helper()
		customer := f.newCustomer(t, nil, nil)
		_, err := add(customer, withPhoto(customer))
		require.NoError(t, err)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(pickupAt))
		require.NoError(t, err)
		f.pay(t, customer, order.ID)
		return order
	}
	accept := func(orderID string) (*dto.StaffOrderResponse, error) {
		return f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(orderID), dto.PatchStaffOrderStatusRequest{Status: "confirmed"})
	}
	setLeadTime := func(t *testing.T, minutes int32) {
		t.Helper()
		_, err := manager.Update(ctx, managerID, domainuser.RoleManager, uuid.MustParse(cake.ID), dto.UpdateProductRequest{
			CategoryID: cakeRequest.CategoryID, Name: cake.Name, Slug: cake.Slug, PriceCents: cake.PriceCents,
			IsActive: true, LeadTimeMinutes: minutes, IsCustomizable: true,
		})
		require.NoError(t, err)
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

		_, err := add(customer, choose("20 cm", "Matcha", "Salted caramel"))
		require.NoError(t, err)
		cart, err := add(customer, choose("20 cm", "Matcha", "Salted caramel"))

		require.NoError(t, err)
		require.Len(t, cart.Items, 2, "each configured cake is a line of its own")
		line := cart.Items[0]
		assert.Equal(t, int64(4500), line.UnitPriceCents)
		require.NotNil(t, line.Customization)
		assert.Equal(t, "Happy birthday Mai", line.Customization.Message)
		assert.Equal(t, "20 cm", line.Customization.Options[0].Label)
		assert.Equal(t, "Matcha", line.Customization.Options[1].Label)
		assert.Equal(t, "filling", line.Customization.Options[2].Group)
		assert.Equal(t, "Salted caramel", line.Customization.Options[2].Label)
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
		_, err = f.cartUC.AddItem(ctx, customer, dto.AddCartItemRequest{ProductID: &pastry, Quantity: 1, Customization: choose("16 cm", "Matcha", "Salted caramel")})
		require.ErrorIs(t, err, domainproduct.ErrCustomization, "a plain pastry is not")
		long := choose("16 cm", "Matcha", "Salted caramel")
		long.Message = strings.Repeat("a", domainproduct.MaxMessageLength+1)
		_, err = add(customer, long)
		require.ErrorIs(t, err, domainproduct.ErrInvalidMessage)
	})

	t.Run("a_reference_photo_is_the_customers_own_upload", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		photo := withPhoto(customer)
		photo.ReferenceRightsConfirmed = false

		_, err := add(customer, photo)
		assert.Equal(t, "reference_rights_confirmed", validationField(t, err))
		someoneElses := *photo
		someoneElses.ReferenceImageURL = ownPhoto(uuid.New())
		someoneElses.ReferenceRightsConfirmed = true
		_, err = add(customer, &someoneElses)
		assert.Equal(t, "reference_image_url", validationField(t, err))

		photo.ReferenceRightsConfirmed = true
		cart, err := add(customer, photo)
		require.NoError(t, err)
		assert.Equal(t, ownPhoto(customer), cart.Items[0].Customization.ReferenceImageURL)
	})

	t.Run("a_cake_built_from_the_options_goes_straight_to_the_kitchen", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		_, err := add(customer, choose("20 cm", "Matcha", "Salted caramel"))
		require.NoError(t, err)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(15, 0)))
		require.NoError(t, err)
		require.NotNil(t, order.Items[0].Customization, "the order keeps the cake as it was bought")

		f.pay(t, customer, order.ID)

		paid, err := f.staff.Get(ctx, uuid.MustParse(order.ID))
		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusConfirmed), paid.Status, "nothing for staff to look at first")
		require.Len(t, paid.Tickets, 1)
		ticket, err := f.kitchen.Get(ctx, uuid.MustParse(paid.Tickets[0].ID))
		require.NoError(t, err)
		require.NotNil(t, ticket.Items[0].Customization)
		assert.Equal(t, "Happy birthday Mai", ticket.Items[0].Customization.Message, "the kitchen sees how to make it")
		assert.Equal(t, "Salted caramel", ticket.Items[0].Customization.Options[2].Label)
	})

	t.Run("a_cake_with_a_reference_photo_waits_for_staff_who_accept_it", func(t *testing.T) {
		order := paidWithPhoto(t, tomorrowAt(15, 30))

		staffView, err := f.staff.Get(ctx, uuid.MustParse(order.ID))
		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusPending), staffView.Status, "staff look at the photo before the kitchen starts")
		accepted, err := accept(order.ID)

		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusConfirmed), accepted.Status)
	})

	t.Run("a_custom_order_with_nothing_to_pay_still_waits_for_staff", func(t *testing.T) {
		_, err := postgresadapter.NewDiscountCodeRepository(f.pool).Create(ctx, port.CreateDiscountCodeParams{
			Code: "CAKEONUS", DiscountType: domaindiscount.TypePercent, Value: 100, IsActive: true,
			StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(24 * time.Hour),
		})
		require.NoError(t, err)
		customer := f.newCustomer(t, nil, nil)
		_, err = add(customer, withPhoto(customer))
		require.NoError(t, err)
		_, err = f.cartUC.ApplyDiscount(ctx, customer, dto.ApplyCartDiscountRequest{Code: "CAKEONUS"})
		require.NoError(t, err)

		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(14, 0)))

		require.NoError(t, err)
		assert.Zero(t, order.TotalCents)
		assert.Equal(t, string(domainorder.StatusPending), order.Status, "the discount pays it; staff still review it")
	})

	t.Run("the_notice_counts_from_when_the_order_was_placed", func(t *testing.T) {
		pickupAt := tomorrowAt(16, 30)
		order := paidWithPhoto(t, pickupAt)
		_, err := raw.Exec(ctx, "UPDATE orders SET created_at = now() - interval '2 days' WHERE id = $1", order.ID)
		require.NoError(t, err)
		// From now the cake would need an hour more than its pickup gives.
		setLeadTime(t, utils.Int32FromInt64(int64(time.Until(pickupAt)/time.Minute)+60))
		t.Cleanup(func() { setLeadTime(t, 180) })

		accepted, err := accept(order.ID)

		require.NoError(t, err, "the time the order waited for staff is the bakery's to make up")
		assert.Equal(t, string(domainorder.StatusConfirmed), accepted.Status)
	})

	t.Run("a_pickup_already_passed_is_not_accepted", func(t *testing.T) {
		order := paidWithPhoto(t, tomorrowAt(17, 0))
		// Placed long enough ago that the notice alone would let it through.
		_, err := raw.Exec(ctx, "UPDATE orders SET pickup_at = now() - interval '1 hour', created_at = now() - interval '2 days' WHERE id = $1", order.ID)
		require.NoError(t, err)

		_, err = accept(order.ID)

		require.ErrorIs(t, err, domainorder.ErrPickupTooSoon)
	})

	t.Run("an_order_left_unreviewed_for_a_day_is_cancelled_and_refunded", func(t *testing.T) {
		waited := paidWithPhoto(t, tomorrowAt(17, 30))
		fresh := paidWithPhoto(t, tomorrowAt(13, 30))
		_, err := raw.Exec(ctx, "UPDATE order_status_events SET created_at = now() - interval '25 hours' WHERE order_id = $1 AND to_status = 'pending'", waited.ID)
		require.NoError(t, err)

		cancelled, err := f.orderUC.CancelUnreviewed(ctx)

		require.NoError(t, err)
		assert.Equal(t, 1, cancelled)
		staffView, err := f.staff.Get(ctx, uuid.MustParse(waited.ID))
		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusCancelled), staffView.Status)
		require.NotNil(t, staffView.Payment.RefundRequestedAt, "the customer gets the money back")
		last := staffView.Timeline[len(staffView.Timeline)-1]
		assert.Nil(t, last.ActorRole, "the system cancelled it")
		require.NotNil(t, last.Reason)
		assert.Equal(t, domainorder.UnreviewedReason, *last.Reason)
		other, err := f.staff.Get(ctx, uuid.MustParse(fresh.ID))
		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusPending), other.Status, "staff still have time for this one")
		again, err := f.orderUC.CancelUnreviewed(ctx)
		require.NoError(t, err)
		assert.Zero(t, again)
	})

	t.Run("a_cake_now_needing_more_notice_than_its_pickup_gives_is_not_accepted", func(t *testing.T) {
		order := paidWithPhoto(t, tomorrowAt(16, 0))
		setLeadTime(t, 7*24*60)
		t.Cleanup(func() { setLeadTime(t, 180) })

		_, err := accept(order.ID)

		require.ErrorIs(t, err, domainorder.ErrPickupTooSoon, "the bakery no longer has the notice the cake needs")
	})

	t.Run("retiring_an_option_takes_the_cakes_that_chose_it_off_sale", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		_, err := add(customer, choose("16 cm", "Matcha", "Salted caramel"))
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

	t.Run("a_product_in_a_hidden_category_stays_editable_and_none_moves_in", func(t *testing.T) {
		categories := postgresadapter.NewCategoryRepository(f.pool)
		newCategory := func(slug string) uuid.UUID {
			category, err := categories.Create(ctx, port.CreateCategoryParams{
				Name: slug, Slug: slug, IsActive: true, Station: domaincategory.StationKitchen,
			})
			require.NoError(t, err)
			return category.ID
		}
		winter, spring := newCategory("winter"), newCategory("spring")
		chestnut, err := manager.Create(ctx, managerID, domainuser.RoleManager, dto.CreateProductRequest{
			CategoryID: winter.String(), Name: "Chestnut log", PriceCents: 3200, IsActive: true,
		})
		require.NoError(t, err)
		_, err = raw.Exec(ctx, "UPDATE categories SET is_active = false WHERE id = ANY($1)", []uuid.UUID{winter, spring})
		require.NoError(t, err)
		update := dto.UpdateProductRequest{
			CategoryID: winter.String(), Name: chestnut.Name, Slug: chestnut.Slug, PriceCents: 3500, IsActive: true,
		}

		_, err = manager.Update(ctx, managerID, domainuser.RoleManager, uuid.MustParse(chestnut.ID), update)
		require.NoError(t, err, "a season being prepared is edited where it is")

		update.CategoryID = spring.String()
		_, err = manager.Update(ctx, managerID, domainuser.RoleManager, uuid.MustParse(chestnut.ID), update)
		require.ErrorIs(t, err, domaincategory.ErrInactive, "nothing moves into a hidden category")
	})
}
