// Package paypal takes payments through PayPal's REST API: Orders v2 with an
// immediate capture, and signed webhook deliveries.
package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boms/backend/internal/config"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	"github.com/boms/backend/internal/port"
)

const (
	sandboxBaseURL = "https://api-m.sandbox.paypal.com"
	liveBaseURL    = "https://api-m.paypal.com"
	// requestTimeout bounds one call to PayPal; the request making it has its
	// own deadline as well.
	requestTimeout = 20 * time.Second
	// tokenMargin renews the access token this long before PayPal says it ends.
	tokenMargin = time.Minute
	// maxResponseBytes caps what is read of one PayPal answer.
	maxResponseBytes = 1 << 20
)

// Client calls PayPal as one REST app.
type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	webhookID    string
	http         *http.Client

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func New(cfg config.PayPalConfig) *Client {
	baseURL := sandboxBaseURL
	if cfg.Mode == config.PayPalModeLive {
		baseURL = liveBaseURL
	}
	return &Client{
		baseURL:      baseURL,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		webhookID:    cfg.WebhookID,
		http:         &http.Client{Timeout: requestTimeout},
	}
}

type amount struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

type captureResource struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	Amount            amount `json:"amount"`
	SupplementaryData struct {
		RelatedIDs struct {
			OrderID string `json:"order_id"`
		} `json:"related_ids"`
	} `json:"supplementary_data"`
}

type orderResource struct {
	ID            string `json:"id"`
	Links         []link `json:"links"`
	PurchaseUnits []struct {
		Payments struct {
			Captures []captureResource `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
}

// CreateOrder implements port.PaymentGateway.
func (c *Client) CreateOrder(ctx context.Context, req port.PaymentOrderRequest) (string, string, error) {
	body := map[string]any{
		"intent": "CAPTURE",
		"purchase_units": []map[string]any{{
			// Ties the PayPal order to ours in PayPal's dashboard and reports.
			"custom_id": req.OrderID.String(),
			"amount":    amount{CurrencyCode: req.Currency, Value: formatAmount(req.AmountCents)},
		}},
		"payment_source": map[string]any{"paypal": map[string]any{"experience_context": map[string]any{
			"return_url": req.ReturnURL,
			"cancel_url": req.CancelURL,
			// The buyer pays on PayPal's page; the order needs no shipping.
			"user_action":         "PAY_NOW",
			"shipping_preference": "NO_SHIPPING",
			// A pickup is paid for now: no eCheck that clears days later.
			"payment_method_preference": "IMMEDIATE_PAYMENT_REQUIRED",
		}}},
	}
	var order orderResource
	if err := c.call(ctx, http.MethodPost, "/v2/checkout/orders", "", body, &order); err != nil {
		return "", "", fmt.Errorf("create paypal order: %w", err)
	}
	for _, l := range order.Links {
		if l.Rel == "payer-action" && order.ID != "" {
			return order.ID, l.Href, nil
		}
	}
	return "", "", errors.New("create paypal order: no page to approve it on")
}

// Capture implements port.PaymentGateway.
func (c *Client) Capture(ctx context.Context, providerOrderID string) (domainpayment.Capture, error) {
	var order orderResource
	// The request id makes a retry return the first answer instead of a second
	// capture; it names the PayPal order, as a restarted payment has a new one.
	err := c.call(ctx, http.MethodPost, "/v2/checkout/orders/"+url.PathEscape(providerOrderID)+"/capture",
		"capture-"+providerOrderID, struct{}{}, &order)
	var apiErr *apiError
	if errors.As(err, &apiErr) && apiErr.status == http.StatusUnprocessableEntity && buyerCanRetry(apiErr.issue) {
		return domainpayment.Capture{}, domainpayment.ErrNotCompleted
	}
	if err != nil {
		return domainpayment.Capture{}, fmt.Errorf("capture paypal order: %w", err)
	}
	if len(order.PurchaseUnits) == 0 || len(order.PurchaseUnits[0].Payments.Captures) == 0 {
		return domainpayment.Capture{}, errors.New("capture paypal order: no capture in the answer")
	}
	return toCapture(order.PurchaseUnits[0].Payments.Captures[0])
}

// Lookup implements port.PaymentGateway.
func (c *Client) Lookup(ctx context.Context, providerOrderID string) (*domainpayment.Capture, error) {
	var order orderResource
	err := c.call(ctx, http.MethodGet, "/v2/checkout/orders/"+url.PathEscape(providerOrderID), "", nil, &order)
	var apiErr *apiError
	if errors.As(err, &apiErr) && apiErr.status == http.StatusNotFound {
		// PayPal drops an order nobody paid after a while: no money moved.
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("look up paypal order: %w", err)
	}
	if len(order.PurchaseUnits) == 0 || len(order.PurchaseUnits[0].Payments.Captures) == 0 {
		return nil, nil
	}
	capture, err := toCapture(order.PurchaseUnits[0].Payments.Captures[0])
	if errors.Is(err, domainpayment.ErrNotCompleted) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &capture, nil
}

// Refund implements port.PaymentGateway. An empty body returns the whole
// capture; PayPal reports a refund still clearing (PENDING) as taken.
func (c *Client) Refund(ctx context.Context, captureID string) (string, error) {
	var refund struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	// The request id makes a retry return the first answer instead of a second
	// refund.
	err := c.call(ctx, http.MethodPost, "/v2/payments/captures/"+url.PathEscape(captureID)+"/refund",
		"refund-"+captureID, struct{}{}, &refund)
	var apiErr *apiError
	if errors.As(err, &apiErr) && apiErr.status == http.StatusUnprocessableEntity && apiErr.issue == "CAPTURE_FULLY_REFUNDED" {
		return "", domainpayment.ErrAlreadyRefunded
	}
	if err != nil {
		return "", fmt.Errorf("refund paypal capture: %w", err)
	}
	if refund.ID == "" || (refund.Status != "COMPLETED" && refund.Status != "PENDING") {
		return "", fmt.Errorf("refund paypal capture: refund %q answered %q", refund.ID, refund.Status)
	}
	return refund.ID, nil
}

// buyerCanRetry reports the capture refusals that approving again fixes: the
// buyer has not approved, or the funding they chose was declined.
func buyerCanRetry(issue string) bool {
	switch issue {
	case "ORDER_NOT_APPROVED", "PAYER_ACTION_REQUIRED", "INSTRUMENT_DECLINED":
		return true
	default:
		return false
	}
}

func toCapture(r captureResource) (domainpayment.Capture, error) {
	var status domainpayment.Status
	switch r.Status {
	case "COMPLETED":
		status = domainpayment.StatusCaptured
	case "PENDING":
		status = domainpayment.StatusPending
	case "DECLINED", "FAILED":
		// No money moved: the buyer may pay again with a new PayPal order.
		status = domainpayment.StatusDenied
	default:
		// Refunded after the fact: nothing to record as a payment taken now.
		return domainpayment.Capture{}, domainpayment.ErrNotCompleted
	}
	cents, err := parseAmount(r.Amount.Value)
	if err != nil {
		return domainpayment.Capture{}, err
	}
	return domainpayment.Capture{ID: r.ID, Status: status, AmountCents: cents, Currency: r.Amount.CurrencyCode}, nil
}

// VerifyWebhook implements port.PaymentGateway: PayPal checks the signature
// against the webhook the delivery is for.
func (c *Client) VerifyWebhook(ctx context.Context, header func(name string) string, body []byte) (port.PaymentWebhookEvent, error) {
	if c.webhookID == "" {
		return port.PaymentWebhookEvent{}, errors.New("verify paypal webhook: paypal.webhook_id is not configured")
	}
	check := map[string]any{
		"auth_algo":         header("PAYPAL-AUTH-ALGO"),
		"cert_url":          header("PAYPAL-CERT-URL"),
		"transmission_id":   header("PAYPAL-TRANSMISSION-ID"),
		"transmission_sig":  header("PAYPAL-TRANSMISSION-SIG"),
		"transmission_time": header("PAYPAL-TRANSMISSION-TIME"),
		"webhook_id":        c.webhookID,
	}
	for _, v := range check {
		if v == "" {
			return port.PaymentWebhookEvent{}, domainpayment.ErrWebhookInvalid
		}
	}
	if !json.Valid(body) {
		return port.PaymentWebhookEvent{}, domainpayment.ErrWebhookInvalid
	}
	// The event as PayPal sent it, byte for byte: the signature covers it.
	check["webhook_event"] = json.RawMessage(body)
	var verdict struct {
		VerificationStatus string `json:"verification_status"`
	}
	if err := c.call(ctx, http.MethodPost, "/v1/notifications/verify-webhook-signature", "", check, &verdict); err != nil {
		return port.PaymentWebhookEvent{}, fmt.Errorf("verify paypal webhook: %w", err)
	}
	if verdict.VerificationStatus != "SUCCESS" {
		return port.PaymentWebhookEvent{}, domainpayment.ErrWebhookInvalid
	}

	var event struct {
		ID        string          `json:"id"`
		EventType string          `json:"event_type"`
		Resource  json.RawMessage `json:"resource"`
	}
	if err := json.Unmarshal(body, &event); err != nil || event.ID == "" {
		return port.PaymentWebhookEvent{}, domainpayment.ErrWebhookInvalid
	}
	out := port.PaymentWebhookEvent{ID: event.ID}
	switch event.EventType {
	case "PAYMENT.CAPTURE.COMPLETED", "PAYMENT.CAPTURE.DENIED":
		var resource captureResource
		if err := json.Unmarshal(event.Resource, &resource); err != nil {
			return port.PaymentWebhookEvent{}, fmt.Errorf("read paypal webhook %s: %w", event.ID, err)
		}
		capture, err := toCapture(resource)
		if err != nil {
			return port.PaymentWebhookEvent{}, fmt.Errorf("read paypal webhook %s: %w", event.ID, err)
		}
		out.ProviderOrderID = resource.SupplementaryData.RelatedIDs.OrderID
		out.Capture = &capture
	}
	return out, nil
}

// call sends one authenticated request, with in as its JSON body unless nil,
// and decodes the answer into out.
func (c *Client) call(ctx context.Context, method, path, requestID string, in, out any) error {
	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}
	var body io.Reader
	if in != nil {
		payload, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if requestID != "" {
		req.Header.Set("PayPal-Request-Id", requestID)
	}
	return c.send(req, out)
}

// accessToken returns the app's OAuth token, fetching a new one when the last
// is about to end. Callers wait for one fetch rather than each making one.
func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.expiresAt) {
		return c.token, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/oauth2/token",
		strings.NewReader(url.Values{"grant_type": {"client_credentials"}}.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var grant struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := c.send(req, &grant); err != nil {
		return "", fmt.Errorf("paypal access token: %w", err)
	}
	if grant.AccessToken == "" {
		return "", errors.New("paypal access token: empty grant")
	}
	c.token = grant.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(grant.ExpiresIn)*time.Second - tokenMargin)
	return c.token, nil
}

// apiError is PayPal refusing a request: its status, error name, the first
// detail's issue and the debug id PayPal support asks for. It never carries
// the body, which can hold the payer's details.
type apiError struct {
	status  int
	name    string
	issue   string
	debugID string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("paypal answered %d %s %s (debug id %s)", e.status, e.name, e.issue, e.debugID)
}

func (c *Client) send(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("reach paypal: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("read paypal answer: %w", err)
	}
	if resp.StatusCode >= http.StatusMultipleChoices {
		var refusal struct {
			Name    string `json:"name"`
			DebugID string `json:"debug_id"`
			Details []struct {
				Issue string `json:"issue"`
			} `json:"details"`
		}
		_ = json.Unmarshal(data, &refusal)
		apiErr := &apiError{status: resp.StatusCode, name: refusal.Name, debugID: refusal.DebugID}
		if len(refusal.Details) > 0 {
			apiErr.issue = refusal.Details[0].Issue
		}
		return apiErr
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode paypal answer: %w", err)
	}
	return nil
}

// formatAmount writes cents as PayPal's decimal string: 1250 is "12.50".
func formatAmount(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

// parseAmount reads PayPal's decimal string back into cents, refusing anything
// but whole digits with two decimals: money is never a float.
func parseAmount(value string) (int64, error) {
	whole, frac, ok := strings.Cut(value, ".")
	if !ok || len(frac) != 2 {
		return 0, fmt.Errorf("paypal amount %q is not a two-decimal value", value)
	}
	units, err := strconv.ParseUint(whole, 10, 62)
	if err != nil {
		return 0, fmt.Errorf("paypal amount %q: %w", value, err)
	}
	cents, err := strconv.ParseUint(frac, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("paypal amount %q: %w", value, err)
	}
	return int64(units)*100 + int64(cents), nil
}

var _ port.PaymentGateway = (*Client)(nil)
