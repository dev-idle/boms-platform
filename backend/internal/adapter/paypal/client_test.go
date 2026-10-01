package paypal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainpayment "github.com/boms/backend/internal/domain/payment"
	"github.com/boms/backend/internal/port"
)

// fakePayPal answers the token endpoint and hands every other request to api.
func fakePayPal(t *testing.T, api http.HandlerFunc) (*Client, *atomic.Int32) {
	t.Helper()
	var tokens atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/oauth2/token" {
			user, pass, ok := r.BasicAuth()
			assert.True(t, ok && user == "client-id" && pass == "client-secret", "the app authenticates with its credentials")
			tokens.Add(1)
			_, _ = io.WriteString(w, `{"access_token":"token-fixture","expires_in":32400}`)
			return
		}
		assert.Equal(t, "Bearer token-fixture", r.Header.Get("Authorization"))
		api(w, r)
	}))
	t.Cleanup(server.Close)
	return &Client{
		baseURL: server.URL, clientID: "client-id", clientSecret: "client-secret", webhookID: "webhook-id",
		http: server.Client(),
	}, &tokens
}

const completedOrder = `{"id":"PAYPAL-ORDER","status":"COMPLETED","purchase_units":[{"payments":{"captures":[
	{"id":"CAPTURE-1","status":"COMPLETED","amount":{"currency_code":"USD","value":"12.50"}}]}}]}`

func TestClient_CreateOrder(t *testing.T) {
	t.Parallel()
	orderID := uuid.New()
	client, tokens := fakePayPal(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/checkout/orders", r.URL.Path)
		var body struct {
			Intent        string `json:"intent"`
			PurchaseUnits []struct {
				CustomID string `json:"custom_id"`
				Amount   amount `json:"amount"`
			} `json:"purchase_units"`
			PaymentSource struct {
				PayPal struct {
					ExperienceContext map[string]string `json:"experience_context"`
				} `json:"paypal"`
			} `json:"payment_source"`
		}
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "CAPTURE", body.Intent)
		assert.Equal(t, orderID.String(), body.PurchaseUnits[0].CustomID)
		assert.Equal(t, amount{CurrencyCode: "USD", Value: "12.50"}, body.PurchaseUnits[0].Amount)
		assert.Equal(t, "https://shop.example/orders/1?paypal=approved", body.PaymentSource.PayPal.ExperienceContext["return_url"])
		assert.Equal(t, "IMMEDIATE_PAYMENT_REQUIRED", body.PaymentSource.PayPal.ExperienceContext["payment_method_preference"])
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"id":"PAYPAL-ORDER","status":"PAYER_ACTION_REQUIRED","links":[
			{"href":"https://api-m.sandbox.paypal.com/v2/checkout/orders/PAYPAL-ORDER","rel":"self"},
			{"href":"https://www.sandbox.paypal.com/checkoutnow?token=PAYPAL-ORDER","rel":"payer-action"}]}`)
	})
	req := port.PaymentOrderRequest{
		OrderID: orderID, AmountCents: 1250, Currency: "USD",
		ReturnURL: "https://shop.example/orders/1?paypal=approved", CancelURL: "https://shop.example/orders/1",
	}

	id, approveURL, err := client.CreateOrder(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "PAYPAL-ORDER", id)
	assert.Equal(t, "https://www.sandbox.paypal.com/checkoutnow?token=PAYPAL-ORDER", approveURL)

	_, _, err = client.CreateOrder(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int32(1), tokens.Load(), "the token is reused until it nearly ends")
}

func TestClient_Capture(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("takes_the_approved_money_once", func(t *testing.T) {
		t.Parallel()
		client, _ := fakePayPal(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v2/checkout/orders/PAYPAL-ORDER/capture", r.URL.Path)
			assert.Equal(t, "capture-PAYPAL-ORDER", r.Header.Get("PayPal-Request-Id"), "a retry returns the first answer")
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, completedOrder)
		})

		capture, err := client.Capture(ctx, "PAYPAL-ORDER")

		require.NoError(t, err)
		assert.Equal(t, domainpayment.Capture{ID: "CAPTURE-1", Status: domainpayment.StatusCaptured, AmountCents: 1250, Currency: "USD"}, capture)
	})

	t.Run("a_capture_held_for_review_is_pending", func(t *testing.T) {
		t.Parallel()
		client, _ := fakePayPal(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"purchase_units":[{"payments":{"captures":[
				{"id":"CAPTURE-1","status":"PENDING","amount":{"currency_code":"USD","value":"12.50"}}]}}]}`)
		})

		capture, err := client.Capture(ctx, "PAYPAL-ORDER")

		require.NoError(t, err)
		assert.Equal(t, domainpayment.StatusPending, capture.Status)
	})

	for _, issue := range []string{"INSTRUMENT_DECLINED", "ORDER_NOT_APPROVED"} {
		t.Run("the_buyer_can_retry_after_"+issue, func(t *testing.T) {
			t.Parallel()
			client, _ := fakePayPal(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = io.WriteString(w, `{"name":"UNPROCESSABLE_ENTITY","debug_id":"d1","details":[{"issue":"`+issue+`"}]}`)
			})

			_, err := client.Capture(ctx, "PAYPAL-ORDER")

			require.ErrorIs(t, err, domainpayment.ErrNotCompleted)
		})
	}

	t.Run("any_other_refusal_is_reported_with_its_debug_id", func(t *testing.T) {
		t.Parallel()
		client, _ := fakePayPal(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{"name":"INTERNAL_SERVER_ERROR","debug_id":"dbg-7","payer":{"email_address":"buyer@example.com"}}`)
		})

		_, err := client.Capture(ctx, "PAYPAL-ORDER")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "dbg-7")
		assert.NotContains(t, err.Error(), "buyer@example.com", "a refusal never carries the body")
	})
}

func TestClient_Lookup(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	answerWith := func(status int, body string) *Client {
		client, _ := fakePayPal(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v2/checkout/orders/PAYPAL-ORDER", r.URL.Path)
			w.WriteHeader(status)
			_, _ = io.WriteString(w, body)
		})
		return client
	}
	answer := func(body string) *Client { return answerWith(http.StatusOK, body) }

	t.Run("reads_the_money_taken", func(t *testing.T) {
		t.Parallel()
		capture, err := answer(completedOrder).Lookup(ctx, "PAYPAL-ORDER")
		require.NoError(t, err)
		require.NotNil(t, capture)
		assert.Equal(t, domainpayment.StatusCaptured, capture.Status)
	})

	t.Run("an_order_never_captured_has_nothing", func(t *testing.T) {
		t.Parallel()
		capture, err := answer(`{"id":"PAYPAL-ORDER","status":"APPROVED","purchase_units":[{}]}`).Lookup(ctx, "PAYPAL-ORDER")
		require.NoError(t, err)
		assert.Nil(t, capture)
	})

	t.Run("an_order_paypal_no_longer_has_took_nothing", func(t *testing.T) {
		t.Parallel()
		capture, err := answerWith(http.StatusNotFound, `{"name":"RESOURCE_NOT_FOUND","debug_id":"d2"}`).Lookup(ctx, "PAYPAL-ORDER")
		require.NoError(t, err)
		assert.Nil(t, capture)
	})

	t.Run("a_declined_capture_is_denied", func(t *testing.T) {
		t.Parallel()
		capture, err := answer(`{"purchase_units":[{"payments":{"captures":[
			{"id":"CAPTURE-1","status":"DECLINED","amount":{"currency_code":"USD","value":"12.50"}}]}}]}`).Lookup(ctx, "PAYPAL-ORDER")
		require.NoError(t, err)
		require.NotNil(t, capture)
		assert.Equal(t, domainpayment.StatusDenied, capture.Status)
	})
}

func TestClient_Refund(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	refunding := func(t *testing.T, status int, body string) *Client {
		client, _ := fakePayPal(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v2/payments/captures/CAPTURE-1/refund", r.URL.Path)
			assert.Equal(t, "refund-CAPTURE-1", r.Header.Get("PayPal-Request-Id"), "a retry returns the first refund")
			w.WriteHeader(status)
			_, _ = io.WriteString(w, body)
		})
		return client
	}

	t.Run("returns_the_whole_capture", func(t *testing.T) {
		t.Parallel()
		id, err := refunding(t, http.StatusCreated, `{"id":"REFUND-1","status":"COMPLETED"}`).Refund(ctx, "CAPTURE-1")
		require.NoError(t, err)
		assert.Equal(t, "REFUND-1", id)
	})

	t.Run("a_refund_still_clearing_is_made", func(t *testing.T) {
		t.Parallel()
		id, err := refunding(t, http.StatusCreated, `{"id":"REFUND-1","status":"PENDING"}`).Refund(ctx, "CAPTURE-1")
		require.NoError(t, err)
		assert.Equal(t, "REFUND-1", id)
	})

	t.Run("a_capture_refunded_from_the_dashboard_is_reported_as_such", func(t *testing.T) {
		t.Parallel()
		_, err := refunding(t, http.StatusUnprocessableEntity,
			`{"name":"UNPROCESSABLE_ENTITY","debug_id":"d3","details":[{"issue":"CAPTURE_FULLY_REFUNDED"}]}`).Refund(ctx, "CAPTURE-1")
		require.ErrorIs(t, err, domainpayment.ErrAlreadyRefunded)
	})

	t.Run("a_failed_refund_is_an_error", func(t *testing.T) {
		t.Parallel()
		_, err := refunding(t, http.StatusCreated, `{"id":"REFUND-1","status":"FAILED"}`).Refund(ctx, "CAPTURE-1")
		require.Error(t, err)
		assert.NotErrorIs(t, err, domainpayment.ErrAlreadyRefunded)
	})
}

func TestClient_VerifyWebhook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	body := []byte(`{"id":"WH-1","event_type":"PAYMENT.CAPTURE.COMPLETED","resource":{"id":"CAPTURE-1","status":"COMPLETED",
		"amount":{"currency_code":"USD","value":"12.50"},"supplementary_data":{"related_ids":{"order_id":"PAYPAL-ORDER"}}}}`)
	signed := func(name string) string { return "header:" + name }

	verifier := func(t *testing.T, status string) (*Client, *atomic.Int32) {
		var calls atomic.Int32
		client, _ := fakePayPal(t, func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			assert.Equal(t, "/v1/notifications/verify-webhook-signature", r.URL.Path)
			var check struct {
				WebhookID    string          `json:"webhook_id"`
				AuthAlgo     string          `json:"auth_algo"`
				WebhookEvent json.RawMessage `json:"webhook_event"`
			}
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&check))
			assert.Equal(t, "webhook-id", check.WebhookID)
			assert.Equal(t, "header:PAYPAL-AUTH-ALGO", check.AuthAlgo)
			assert.JSONEq(t, string(body), string(check.WebhookEvent), "the event goes back as it came")
			_, _ = io.WriteString(w, `{"verification_status":"`+status+`"}`)
		})
		return client, &calls
	}

	t.Run("reads_a_signed_capture", func(t *testing.T) {
		t.Parallel()
		client, _ := verifier(t, "SUCCESS")

		event, err := client.VerifyWebhook(ctx, signed, body)

		require.NoError(t, err)
		assert.Equal(t, "WH-1", event.ID)
		assert.Equal(t, "PAYPAL-ORDER", event.ProviderOrderID)
		require.NotNil(t, event.Capture)
		assert.Equal(t, domainpayment.StatusCaptured, event.Capture.Status)
	})

	t.Run("reads_a_capture_paypal_refused_after_review", func(t *testing.T) {
		t.Parallel()
		refused := []byte(`{"id":"WH-2","event_type":"PAYMENT.CAPTURE.DENIED","resource":{"id":"CAPTURE-1","status":"DECLINED",
			"amount":{"currency_code":"USD","value":"12.50"},"supplementary_data":{"related_ids":{"order_id":"PAYPAL-ORDER"}}}}`)
		client, _ := fakePayPal(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"verification_status":"SUCCESS"}`)
		})

		event, err := client.VerifyWebhook(ctx, signed, refused)

		require.NoError(t, err)
		require.NotNil(t, event.Capture)
		assert.Equal(t, domainpayment.StatusDenied, event.Capture.Status)
	})

	t.Run("refuses_a_delivery_paypal_did_not_sign", func(t *testing.T) {
		t.Parallel()
		client, _ := verifier(t, "FAILURE")

		_, err := client.VerifyWebhook(ctx, signed, body)

		require.ErrorIs(t, err, domainpayment.ErrWebhookInvalid)
	})

	t.Run("refuses_an_unsigned_delivery_without_asking_paypal", func(t *testing.T) {
		t.Parallel()
		client, calls := verifier(t, "SUCCESS")

		_, err := client.VerifyWebhook(ctx, func(string) string { return "" }, body)

		require.ErrorIs(t, err, domainpayment.ErrWebhookInvalid)
		assert.Zero(t, calls.Load())
	})
}

func TestAmounts(t *testing.T) {
	t.Parallel()
	for cents, text := range map[int64]string{1250: "12.50", 5: "0.05", 100: "1.00", 0: "0.00"} {
		assert.Equal(t, text, formatAmount(cents))
		parsed, err := parseAmount(text)
		require.NoError(t, err)
		assert.Equal(t, cents, parsed)
	}
	for _, bad := range []string{"12.5", "12", "-1.00", "1e3.00", "12.500"} {
		_, err := parseAmount(bad)
		assert.Error(t, err, bad)
	}
}
