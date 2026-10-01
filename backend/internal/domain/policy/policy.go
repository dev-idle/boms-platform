// Package policy holds the bakery's customer policies — the terms of sale, the
// privacy policy and the refund policy — and a person's acceptance of them.
package policy

import "time"

// TermsVersion names the policies customers accept today. It changes together
// with the published pages whenever what a customer agrees to changes; every
// acceptance records the version it was given, so a later change never
// rewrites what someone agreed to.
const TermsVersion = "2026-10-01.3"

// Acceptance is someone's agreement to one version of the policies.
type Acceptance struct {
	Version    string
	AcceptedAt time.Time
}

// RequireAccepted checks that version is the one customers accept today. The
// client sends the version it showed, so a page left open across a policy
// change is refused rather than taken as agreement to text never seen.
func RequireAccepted(version string) error {
	if version != TermsVersion {
		return ErrTermsNotAccepted
	}
	return nil
}
