package config

import (
	"errors"
	"fmt"
	"strings"
)

// PayPal environments, as PAYPAL_MODE names them.
const (
	PayPalModeSandbox = "sandbox"
	PayPalModeLive    = "live"
)

// PayPalConfig is the PayPal REST app the API takes payments through.
type PayPalConfig struct {
	Mode         string
	ClientID     string
	ClientSecret string
	// WebhookID is the webhook PayPal signs its deliveries for; a delivery is
	// checked against it, so without it none is accepted.
	WebhookID string
}

func (c PayPalConfig) validate(env string) error {
	if c.Mode != PayPalModeSandbox && c.Mode != PayPalModeLive {
		return fmt.Errorf("paypal.mode must be %q or %q", PayPalModeSandbox, PayPalModeLive)
	}
	if strings.TrimSpace(c.ClientID) == "" || strings.TrimSpace(c.ClientSecret) == "" {
		return errors.New("paypal.client_id and paypal.client_secret are required: the API takes every order's payment through PayPal")
	}
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "production":
		if c.Mode != PayPalModeLive {
			return errors.New("paypal.mode must be live in production: a sandbox takes no real money")
		}
		fallthrough
	case "staging":
		if strings.TrimSpace(c.WebhookID) == "" {
			return errors.New("paypal.webhook_id is required in staging/production")
		}
	}
	return nil
}
