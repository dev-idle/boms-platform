package postgres_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domaincart "github.com/boms/backend/internal/domain/cart"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainproduct "github.com/boms/backend/internal/domain/product"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
)

func TestOrderCheckoutRepositories_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 5)
	users := postgresadapter.NewUserRepository(pool)
	categories := postgresadapter.NewCategoryRepository(pool)
	products := postgresadapter.NewProductRepository(pool)
	carts := postgresadapter.NewCartRepository(pool)
	orders := postgresadapter.NewOrderRepository(pool)

	customer, err := users.Create(ctx, port.CreateUserParams{
		Email:        "checkout@example.com",
		PasswordHash: testPasswordHashFixture,
		Role:         domainuser.RoleCustomer,
	})
	require.NoError(t, err)
	category, err := categories.Create(ctx, port.CreateCategoryParams{Name: "Cakes", Slug: "cakes", IsActive: true, Station: domaincategory.StationKitchen})
	require.NoError(t, err)
	cake, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: category.ID, Name: "Matcha cake", Slug: "matcha-cake", PriceCents: 4500, IsActive: true,
	})
	require.NoError(t, err)
	tart, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: category.ID, Name: "Fruit tart", Slug: "fruit-tart", PriceCents: 3200, IsActive: true,
	})
	require.NoError(t, err)
	_, err = carts.CreateForUser(ctx, customer.ID)
	require.NoError(t, err)

	t.Run("create_items_inserts_every_line_in_one_statement", func(t *testing.T) {
		order, err := orders.Create(ctx, port.CreateOrderParams{
			Code: "CH-260101-001", Type: domainorder.TypePreOrder, UserID: &customer.ID, Channel: domainorder.ChannelOnline, Status: domainorder.StatusPending, SubtotalCents: 12200, TotalCents: 12200,
		})
		require.NoError(t, err)

		require.NoError(t, orders.CreateItems(ctx, []port.CreateOrderItemParams{
			{
				OrderID: order.ID, LineType: domaincart.LineTypeProduct, ProductID: &cake.ID,
				Name: cake.Name, Slug: cake.Slug, Quantity: 2, UnitPriceCents: 4500, LineTotalCents: 9000,
			},
			{
				OrderID: order.ID, LineType: domaincart.LineTypeProduct, ProductID: &tart.ID,
				Customization: &domainorder.Customization{
					Options: []domainorder.ChosenOption{{Group: domainproduct.OptionSize, Label: "20 cm", PriceDeltaCents: 800}},
					Message: "Happy birthday",
				},
				Name: tart.Name, Slug: tart.Slug, Quantity: 1, UnitPriceCents: 3200, LineTotalCents: 3200,
			},
		}))

		items, err := orders.ListItemsByOrderID(ctx, order.ID)
		require.NoError(t, err)
		require.Len(t, items, 2)
		byName := map[string]domainorder.Item{items[0].Name: items[0], items[1].Name: items[1]}
		assert.Equal(t, int32(2), byName[cake.Name].Quantity)
		assert.Equal(t, int64(9000), byName[cake.Name].LineTotalCents)
		assert.Nil(t, byName[cake.Name].ComboID)
		assert.Nil(t, byName[cake.Name].Customization)
		require.NotNil(t, byName[tart.Name].Customization)
		assert.Equal(t, "Happy birthday", byName[tart.Name].Customization.Message)
		assert.Equal(t, "20 cm", byName[tart.Name].Customization.Options[0].Label)
	})

	t.Run("create_items_is_atomic_when_a_line_violates_a_constraint", func(t *testing.T) {
		order, err := orders.Create(ctx, port.CreateOrderParams{
			Code: "CH-260101-002", Type: domainorder.TypePreOrder, UserID: &customer.ID, Channel: domainorder.ChannelOnline, Status: domainorder.StatusPending, SubtotalCents: 4500, TotalCents: 4500,
		})
		require.NoError(t, err)

		err = orders.CreateItems(ctx, []port.CreateOrderItemParams{
			{
				OrderID: order.ID, LineType: domaincart.LineTypeProduct, ProductID: &cake.ID,
				Name: cake.Name, Slug: cake.Slug, Quantity: 1, UnitPriceCents: 4500, LineTotalCents: 4500,
			},
			{
				// A product line without a product violates order_items_line_target_check.
				OrderID: order.ID, LineType: domaincart.LineTypeProduct,
				Name: "broken", Slug: "broken", Quantity: 1, UnitPriceCents: 0, LineTotalCents: 0,
			},
		})
		require.Error(t, err)

		items, err := orders.ListItemsByOrderID(ctx, order.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("get_for_update_requires_a_transaction", func(t *testing.T) {
		_, err := carts.GetByUserIDForUpdate(ctx, customer.ID)
		require.Error(t, err)
	})

	t.Run("get_for_update_serializes_concurrent_checkouts", func(t *testing.T) {
		const holdFor = 300 * time.Millisecond
		var firstReleased atomic.Bool
		locked := make(chan struct{})
		secondAcquired := make(chan bool, 1)

		go func() {
			<-locked
			err := pool.WithTx(ctx, func(txCtx context.Context) error {
				_, err := carts.GetByUserIDForUpdate(txCtx, customer.ID)
				secondAcquired <- firstReleased.Load()
				return err
			})
			if err != nil {
				secondAcquired <- false
			}
		}()

		require.NoError(t, pool.WithTx(ctx, func(txCtx context.Context) error {
			if _, err := carts.GetByUserIDForUpdate(txCtx, customer.ID); err != nil {
				return err
			}
			close(locked)
			time.Sleep(holdFor)
			firstReleased.Store(true)
			return nil
		}))

		select {
		case afterRelease := <-secondAcquired:
			assert.True(t, afterRelease, "second transaction acquired the cart lock while the first still held it")
		case <-time.After(10 * time.Second):
			t.Fatal("second transaction never acquired the cart lock")
		}
	})

	t.Run("the_staff_list_filters_by_an_optional_status", func(t *testing.T) {
		confirmed, err := orders.Create(ctx, port.CreateOrderParams{
			Code: "CH-260101-003", Type: domainorder.TypePreOrder, UserID: &customer.ID, Channel: domainorder.ChannelOnline, Status: domainorder.StatusConfirmed, SubtotalCents: 4500, TotalCents: 4500,
		})
		require.NoError(t, err)
		status := domainorder.StatusConfirmed

		all, err := orders.StaffList(ctx, port.StaffListOrdersParams{Limit: 100})
		require.NoError(t, err)
		allCount, err := orders.StaffListCount(ctx, port.StaffListOrdersParams{})
		require.NoError(t, err)
		assert.Equal(t, int64(len(all)), allCount)
		assert.Greater(t, allCount, int64(1), "an unset status lists every order")

		filtered, err := orders.StaffList(ctx, port.StaffListOrdersParams{Status: &status, Limit: 100})
		require.NoError(t, err)
		filteredCount, err := orders.StaffListCount(ctx, port.StaffListOrdersParams{Status: &status})
		require.NoError(t, err)
		require.Len(t, filtered, 1)
		assert.Equal(t, int64(1), filteredCount)
		assert.Equal(t, confirmed.ID, filtered[0].Order.ID)
		assert.Equal(t, &customer.Email, filtered[0].CustomerEmail)

		for _, term := range []string{"260101-003", "ch-260101-003"} {
			byCode := port.StaffListOrdersParams{Search: term, Limit: 100}
			found, err := orders.StaffList(ctx, byCode)
			require.NoError(t, err)
			foundCount, err := orders.StaffListCount(ctx, byCode)
			require.NoError(t, err)
			require.Len(t, found, 1, "a code is found whole or in part, in any case: %q", term)
			assert.Equal(t, int64(1), foundCount)
			assert.Equal(t, confirmed.ID, found[0].Order.ID)
		}
		byEmail := port.StaffListOrdersParams{Search: customer.Email, Limit: 100}
		found, err := orders.StaffList(ctx, byEmail)
		require.NoError(t, err)
		foundCount, err := orders.StaffListCount(ctx, byEmail)
		require.NoError(t, err)
		assert.Equal(t, int64(len(found)), foundCount)
		require.NotEmpty(t, found)
		for _, row := range found {
			assert.Equal(t, &customer.Email, row.CustomerEmail, "only that customer's orders")
		}
		none, err := orders.StaffListCount(ctx, port.StaffListOrdersParams{Search: "no such order"})
		require.NoError(t, err)
		assert.Zero(t, none)
	})

	t.Run("staff_see_the_phone_on_an_order_but_not_across_the_list", func(t *testing.T) {
		caller, err := users.Create(ctx, port.CreateUserParams{
			Email: "phone-on-file@example.com", PasswordHash: testPasswordHashFixture, Role: domainuser.RoleCustomer,
		})
		require.NoError(t, err)
		phone := "+84901234567"
		_, err = postgresadapter.NewCustomerProfileRepository(pool).Create(ctx, port.UpsertCustomerProfileParams{
			UserID: caller.ID, Phone: &phone,
		})
		require.NoError(t, err)
		order, err := orders.Create(ctx, port.CreateOrderParams{
			Code: "CH-260101-004", Type: domainorder.TypePreOrder, UserID: &caller.ID, Channel: domainorder.ChannelOnline, Status: domainorder.StatusPending, SubtotalCents: 4500, TotalCents: 4500,
		})
		require.NoError(t, err)

		detail, err := orders.StaffGetByID(ctx, order.ID)
		require.NoError(t, err)
		require.NotNil(t, detail.CustomerPhone)
		assert.Equal(t, phone, *detail.CustomerPhone)

		list, err := orders.StaffList(ctx, port.StaffListOrdersParams{Limit: 100})
		require.NoError(t, err)
		for _, row := range list {
			assert.Nil(t, row.CustomerPhone, "one list page must not hand out every customer's number")
		}
	})

	t.Run("sums_item_quantities_for_a_list_of_order_ids", func(t *testing.T) {
		order, err := orders.Create(ctx, port.CreateOrderParams{
			Code: "CH-260101-005", Type: domainorder.TypePreOrder, UserID: &customer.ID, Channel: domainorder.ChannelOnline, Status: domainorder.StatusPending, SubtotalCents: 12200, TotalCents: 12200,
		})
		require.NoError(t, err)
		require.NoError(t, orders.CreateItems(ctx, []port.CreateOrderItemParams{
			{
				OrderID: order.ID, LineType: domaincart.LineTypeProduct, ProductID: &cake.ID,
				Name: cake.Name, Slug: cake.Slug, Quantity: 2, UnitPriceCents: 4500, LineTotalCents: 9000,
			},
			{
				OrderID: order.ID, LineType: domaincart.LineTypeProduct, ProductID: &tart.ID,
				Name: tart.Name, Slug: tart.Slug, Quantity: 1, UnitPriceCents: 3200, LineTotalCents: 3200,
			},
		}))

		sums, err := orders.SumItemQuantitiesByOrderIDs(ctx, []uuid.UUID{order.ID, uuid.New()})
		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]int32{order.ID: 3}, sums)
	})

	t.Run("items_carry_their_product_or_combo_photo", func(t *testing.T) {
		cakePhoto := "https://res.cloudinary.com/demo/image/upload/boms/products/matcha-cake-1.jpg"
		require.NoError(t, products.ReplaceProductImages(ctx, cake.ID, []string{
			cakePhoto,
			"https://res.cloudinary.com/demo/image/upload/boms/products/matcha-cake-2.jpg",
		}))
		boxPhoto := "https://res.cloudinary.com/demo/image/upload/boms/combos/tea-box.jpg"
		now := time.Now().UTC()
		combos := postgresadapter.NewComboRepository(pool)
		box, err := combos.Create(ctx, port.CreateComboParams{
			Name: "Tea box", Slug: "tea-box", PriceCents: 7000, ImageURL: &boxPhoto,
			StartsAt: now.Add(-time.Hour), EndsAt: now.Add(24 * time.Hour), IsActive: true,
		})
		require.NoError(t, err)
		order, err := orders.Create(ctx, port.CreateOrderParams{
			Code: "CH-260101-006", Type: domainorder.TypePreOrder, UserID: &customer.ID, Channel: domainorder.ChannelOnline, Status: domainorder.StatusPending, SubtotalCents: 14700, TotalCents: 14700,
		})
		require.NoError(t, err)
		require.NoError(t, orders.CreateItems(ctx, []port.CreateOrderItemParams{
			{
				OrderID: order.ID, LineType: domaincart.LineTypeProduct, ProductID: &cake.ID,
				Name: cake.Name, Slug: cake.Slug, Quantity: 1, UnitPriceCents: 4500, LineTotalCents: 4500,
			},
			{
				OrderID: order.ID, LineType: domaincart.LineTypeProduct, ProductID: &tart.ID,
				Name: tart.Name, Slug: tart.Slug, Quantity: 1, UnitPriceCents: 3200, LineTotalCents: 3200,
			},
			{
				OrderID: order.ID, LineType: domaincart.LineTypeCombo, ComboID: &box.ID,
				Name: box.Name, Slug: box.Slug, Quantity: 1, UnitPriceCents: 7000, LineTotalCents: 7000,
			},
		}))

		photosOf := func(items []domainorder.Item) map[string]string {
			photos := make(map[string]string, len(items))
			for _, item := range items {
				photos[item.Name] = item.ImageURL
			}
			return photos
		}
		items, err := orders.ListItemsByOrderID(ctx, order.ID)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{cake.Name: cakePhoto, tart.Name: "", box.Name: boxPhoto}, photosOf(items))

		exported, err := orders.ListItemsByOrderIDs(ctx, []uuid.UUID{order.ID})
		require.NoError(t, err)
		assert.Equal(t, map[string]string{cake.Name: cakePhoto, tart.Name: "", box.Name: boxPhoto}, photosOf(exported[order.ID]),
			"the data export reads the same photos")

		// Last subtest: the cake and the box are deleted for good here.
		require.NoError(t, products.SoftDelete(ctx, cake.ID))
		require.NoError(t, combos.SoftDelete(ctx, box.ID))
		items, err = orders.ListItemsByOrderID(ctx, order.ID)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{cake.Name: "", tart.Name: "", box.Name: ""}, photosOf(items),
			"a deleted product or combo no longer lends its photo")
	})
}
