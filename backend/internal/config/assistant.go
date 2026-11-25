package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// AssistantConfig is the storefront's automated helper: the language model it
// asks over an OpenAI-compatible API — the bakery's own, so questions reach no
// other company — how long one answer may take, and how long its
// conversations are kept.
type AssistantConfig struct {
	// BaseURL is the API root the chat completions path is added to, such as
	// a model served on this machine at http://127.0.0.1:11434/v1.
	BaseURL string
	Model   string
	// ReasoningEffort is passed on as reasoning_effort when set: "none" stops a
	// reasoning model from thinking before it answers, which would take it far
	// past ReplyTimeout.
	ReasoningEffort string
	ReplyTimeout    time.Duration
	Retention       time.Duration
	// Concurrency is how many answers one API process has the model write at
	// once, one for each caller; a question past them is refused at once
	// rather than left waiting.
	Concurrency int
}

// maxAssistantConcurrency bounds the answers one process asks the model for at once.
const maxAssistantConcurrency = 32

// validate checks what the API needs to ask the model. A deployed site
// sends customers' questions and their orders to a model on another machine
// over https only.
func (c AssistantConfig) validate(env string, writeTimeout time.Duration) error {
	base, err := url.Parse(c.BaseURL)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" ||
		base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return errors.New("assistant.base_url must be an http:// or https:// URL without credentials, query or fragment")
	}
	if (env == "production" || env == "staging") && base.Scheme != "https" && !isLoopbackHost(base.Hostname()) {
		return errors.New("assistant.base_url must use https:// in staging/production unless the model runs on the same machine")
	}
	if strings.TrimSpace(c.Model) == "" {
		return errors.New("assistant.model is required")
	}
	if c.ReplyTimeout <= 0 || c.ReplyTimeout >= writeTimeout {
		return errors.New("assistant.reply_timeout must be positive and below http.write_timeout: an answer streams within one response")
	}
	if c.Concurrency < 1 || c.Concurrency > maxAssistantConcurrency {
		return fmt.Errorf("assistant.concurrency must be between 1 and %d", maxAssistantConcurrency)
	}
	return nil
}

// validateRetention checks what the worker needs to delete old conversations.
func (c AssistantConfig) validateRetention() error {
	if c.Retention <= 0 {
		return errors.New("assistant.retention must be positive")
	}
	return nil
}
