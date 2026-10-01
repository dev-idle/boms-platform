package email

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
)

func newComposer(t *testing.T) *OrderComposer {
	t.Helper()
	composer, err := NewOrderComposer("https://shop.example")
	require.NoError(t, err)
	return composer
}

func orderEmail(notice domainorder.Notice) port.OrderEmail {
	// 04:30 UTC is 11:30 at the bakery (UTC+7).
	pickup := time.Date(2026, time.September, 30, 4, 30, 0, 0, time.UTC)
	code := "WELCOME10"
	return port.OrderEmail{
		Notice:       notice,
		To:           "mai@example.com",
		CustomerName: "Mai",
		Order: domainorder.Order{
			ID: uuid.MustParse("5b0c7a4e-3c1d-4b8e-9f2a-6d7e8f901234"), Code: "CH-260930-007",
			SubtotalCents: 2450, DiscountCents: 245, TotalCents: 2205,
			DiscountCodeSnapshot: &code, PickupAt: &pickup,
		},
		Items: []domainorder.Item{
			{Name: "Almond Croissant", Quantity: 2, LineTotalCents: 900},
			{Name: "Matcha Roll Cake", Quantity: 1, LineTotalCents: 1550},
		},
	}
}

func TestMoney(t *testing.T) {
	t.Parallel()
	for cents, want := range map[int64]string{
		0:         "$0.00",
		900:       "$9.00",
		123456:    "$1,234.56",
		100000000: "$1,000,000.00",
	} {
		assert.Equal(t, want, money(cents))
	}
}

func TestOrderComposer(t *testing.T) {
	t.Parallel()

	t.Run("a_received_order_carries_its_receipt", func(t *testing.T) {
		t.Parallel()
		email, err := newComposer(t).ComposeOrderEmail(orderEmail(domainorder.NoticePlaced))
		require.NoError(t, err)

		assert.Equal(t, "We received your order CH-260930-007", email.Subject)
		assert.Equal(t, "mai@example.com", email.To)
		assert.Equal(t, "Mai", email.ToName)
		for _, part := range []string{email.HTML, email.Text} {
			assert.Contains(t, part, "Hi Mai,")
			assert.Contains(t, part, "CH-260930-007")
			assert.Contains(t, part, "Wednesday, September 30 at 11:30 AM", "pickup in the bakery's time")
			assert.Contains(t, part, "Almond Croissant")
			assert.Contains(t, part, "$9.00")
			assert.Contains(t, part, "WELCOME10")
			assert.Contains(t, part, "−$2.45")
			assert.Contains(t, part, "$22.05")
			assert.Contains(t, part, "https://shop.example/orders/5b0c7a4e-3c1d-4b8e-9f2a-6d7e8f901234")
			assert.Contains(t, part, "123 Greige Lane")
		}
	})

	t.Run("the_code_to_collect_the_order_is_given_once_the_bakery_accepts_it", func(t *testing.T) {
		t.Parallel()
		msg := orderEmail(domainorder.NoticeReady)
		msg.PickupCode = "4821"
		ready, err := newComposer(t).ComposeOrderEmail(msg)
		require.NoError(t, err)
		assert.Contains(t, ready.Text, "Pickup code: 4821 (give it at the counter")
		assert.Contains(t, ready.HTML, "4821")
		requested, err := newComposer(t).ComposeOrderEmail(orderEmail(domainorder.NoticeRequested))
		require.NoError(t, err)
		assert.NotContains(t, requested.Text, "Pickup code", "not while staff review the order")
	})

	t.Run("ready_and_cancelled_say_so_without_the_receipt", func(t *testing.T) {
		t.Parallel()
		ready, err := newComposer(t).ComposeOrderEmail(orderEmail(domainorder.NoticeReady))
		require.NoError(t, err)
		assert.Equal(t, "Your order CH-260930-007 is ready to collect", ready.Subject)
		assert.Contains(t, ready.Text, "give your pickup code at the counter")
		assert.NotContains(t, ready.Text, "Subtotal")

		cancelled, err := newComposer(t).ComposeOrderEmail(orderEmail(domainorder.NoticeCancelled))
		require.NoError(t, err)
		assert.Equal(t, "Your order CH-260930-007 was cancelled", cancelled.Subject)
		assert.Contains(t, cancelled.HTML, "will not be made")
		assert.NotContains(t, cancelled.HTML, "Subtotal")

		expired, err := newComposer(t).ComposeOrderEmail(orderEmail(domainorder.NoticeExpired))
		require.NoError(t, err)
		assert.Equal(t, "Your order CH-260930-007 was not paid in time", expired.Subject)
		assert.Contains(t, expired.Text, "Nothing was charged")
		assert.NotContains(t, expired.Text, "Subtotal")
	})

	t.Run("a_cancellation_says_who_cancelled_why_and_what_goes_back", func(t *testing.T) {
		t.Parallel()
		msg := orderEmail(domainorder.NoticeCancelled)
		mine, err := newComposer(t).ComposeOrderEmail(msg)
		require.NoError(t, err)
		assert.Contains(t, mine.Text, "cancelled as you asked")
		assert.NotContains(t, mine.Text, "PayPal", "an order never paid has nothing to return")

		msg.Reason, msg.RefundCents = "Out of <b>matcha</b>", 3940
		bakery, err := newComposer(t).ComposeOrderEmail(msg)
		require.NoError(t, err)
		assert.Contains(t, bakery.Text, "Our reason: Out of <b>matcha</b>")
		assert.Contains(t, bakery.HTML, "Out of &lt;b&gt;matcha&lt;/b&gt;", "the reason is text, not markup")
		assert.Contains(t, bakery.Text, "We are returning $39.40 to your PayPal account")
	})

	t.Run("a_custom_order_is_received_for_review_then_accepted", func(t *testing.T) {
		t.Parallel()
		requested, err := newComposer(t).ComposeOrderEmail(orderEmail(domainorder.NoticeRequested))
		require.NoError(t, err)
		assert.Equal(t, "We received your request CH-260930-007", requested.Subject)
		assert.Contains(t, requested.Text, "refund you in full")
		accepted, err := newComposer(t).ComposeOrderEmail(orderEmail(domainorder.NoticeAccepted))
		require.NoError(t, err)
		assert.Equal(t, "We accepted your order CH-260930-007", accepted.Subject)
	})

	t.Run("a_missed_pickup_is_not_refunded", func(t *testing.T) {
		t.Parallel()
		missed, err := newComposer(t).ComposeOrderEmail(orderEmail(domainorder.NoticeNoShow))
		require.NoError(t, err)
		assert.Equal(t, "Your order CH-260930-007 was not collected", missed.Subject)
		assert.Contains(t, missed.Text, "is not refunded")
	})

	t.Run("a_customer_without_a_name_is_greeted_plainly", func(t *testing.T) {
		t.Parallel()
		msg := orderEmail(domainorder.NoticeReady)
		msg.CustomerName = ""
		email, err := newComposer(t).ComposeOrderEmail(msg)
		require.NoError(t, err)
		assert.Contains(t, email.Text, "Hello,")
	})

	t.Run("a_name_is_text_not_markup", func(t *testing.T) {
		t.Parallel()
		msg := orderEmail(domainorder.NoticeReady)
		msg.CustomerName = `<a href="https://evil.example">Mai</a>`
		email, err := newComposer(t).ComposeOrderEmail(msg)
		require.NoError(t, err)
		assert.NotContains(t, email.HTML, `<a href="https://evil.example">`)
		assert.Contains(t, email.HTML, "&lt;a href=")
	})

	t.Run("an_unknown_notice_is_refused", func(t *testing.T) {
		t.Parallel()
		_, err := newComposer(t).ComposeOrderEmail(orderEmail(domainorder.Notice("shipped")))
		require.Error(t, err)
	})
}
