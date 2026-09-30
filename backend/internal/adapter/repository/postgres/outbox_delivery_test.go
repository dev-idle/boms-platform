package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/adapter/eventbus"
	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/eventdispatch"
	"github.com/boms/backend/internal/usecase"
)

// staffChannel is the bus channel every staff session listens on.
const staffChannel = "boms:events:role:staff"

// A change committed while the bus is down is not lost: its event waits in the
// outbox and the sweeper delivers it once the bus is back.
func TestOutboxDelivery_RecoversAfterABusOutage_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 5)
	outbox := postgresadapter.NewOutboxRepository(pool)
	bus := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: bus.Addr(), MaxRetries: -1})
	t.Cleanup(func() { _ = rdb.Close() })

	dispatcher := eventdispatch.New(outbox, eventbus.NewRedisPublisher(rdb), pool, zap.NewNop(), time.Second)
	pool.OnCommit(dispatcher.AfterCommit)

	bus.Close()
	e := newOrderEvent()
	record(t, pool, outbox, e)
	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	dispatcher.Wait(waitCtx)
	require.Contains(t, eventIDs(claim(t, pool, outbox, 0)), e.ID, "the failed delivery left the event pending")

	require.NoError(t, bus.Restart())
	sub := rdb.Subscribe(ctx, staffChannel)
	t.Cleanup(func() { _ = sub.Close() })
	_, err := sub.Receive(ctx)
	require.NoError(t, err)

	delivered, err := dispatcher.Sweep(ctx, 0, 100)
	require.NoError(t, err)
	assert.Equal(t, 1, delivered)

	receiveCtx, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	msg, err := sub.ReceiveMessage(receiveCtx)
	require.NoError(t, err)
	var got struct {
		ID   string            `json:"id"`
		Type string            `json:"type"`
		Data map[string]string `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(msg.Payload), &got))
	assert.Equal(t, e.ID.String(), got.ID)
	assert.Equal(t, string(domainorder.TopicOrderStatusChanged), got.Type)
	assert.Empty(t, claim(t, pool, outbox, 0), "nothing is left pending once the sweep delivered")
}

// Checkout numbers the new order, starts its history and announces it, all in
// the transaction that creates it.
func TestCheckoutRecordsOrderCreated_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 5)
	users := postgresadapter.NewUserRepository(pool)
	categories := postgresadapter.NewCategoryRepository(pool)
	products := postgresadapter.NewProductRepository(pool)
	combos := postgresadapter.NewComboRepository(pool)
	carts := postgresadapter.NewCartRepository(pool)
	orders := postgresadapter.NewOrderRepository(pool)
	discounts := postgresadapter.NewDiscountCodeRepository(pool)
	outbox := postgresadapter.NewOutboxRepository(pool)
	store := postgresadapter.NewStoreSettingsRepository(pool)

	cartUC := usecase.NewCartUsecase(carts, products, combos, discounts)
	orderUC := usecase.NewOrderUsecase(users, orders, carts, discounts, cartUC, pool, outbox, store, postgresadapter.NewTicketRepository(pool))

	customer, err := users.Create(ctx, port.CreateUserParams{
		Email: "outbox-checkout@example.com", PasswordHash: testPasswordHashFixture, Role: domainuser.RoleCustomer,
	})
	require.NoError(t, err)
	category, err := categories.Create(ctx, port.CreateCategoryParams{Name: "Cakes", Slug: "cakes", IsActive: true, Station: domaincategory.StationKitchen})
	require.NoError(t, err)
	cake, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: category.ID, Name: "Matcha cake", Slug: "matcha-cake", PriceCents: 4500, IsActive: true,
	})
	require.NoError(t, err)
	productID := cake.ID.String()
	_, err = cartUC.AddItem(ctx, customer.ID, dto.AddCartItemRequest{ProductID: &productID, Quantity: 1})
	require.NoError(t, err)

	order, err := orderUC.Checkout(ctx, customer.ID, acceptingTerms(nextPickupSlot(time.Now())))
	require.NoError(t, err)

	var created []string
	for _, e := range claim(t, pool, outbox, 0) {
		if e.Topic == domainorder.TopicOrderCreated {
			created = append(created, e.Data["order_id"])
			assert.Equal(t, []uuid.UUID{customer.ID}, e.Audience.UserIDs)
		}
	}
	assert.Equal(t, []string{order.ID}, created)

	placedOn := domainstore.DayOf(order.CreatedAt).Format("060102")
	assert.Equal(t, "CH-"+placedOn+"-001", order.Code, "the first order of the bakery day it was placed on")
	require.Len(t, order.Timeline, 1)
	assert.Equal(t, string(domainorder.StatusPending), order.Timeline[0].Status, "placing the order starts its history")
}

// nextPickupSlot is tomorrow at 10:00 bakery time: inside the seeded opening
// hours, past the seeded lead time and within the seeded booking window.
func nextPickupSlot(now time.Time) time.Time {
	local := now.In(domainstore.Location).AddDate(0, 0, 1)
	return time.Date(local.Year(), local.Month(), local.Day(), 10, 0, 0, 0, domainstore.Location)
}
