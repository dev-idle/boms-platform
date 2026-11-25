// Package policy holds the bakery's customer policies — the terms of sale, the
// privacy policy and the refund policy — and a person's acceptance of them.
package policy

import "time"

// TermsVersion names the policies customers accept today. It changes together
// with the published pages whenever what a customer agrees to changes; every
// acceptance records the version it was given, so a later change never
// rewrites what someone agreed to.
const TermsVersion = "2026-11-25.3"

// PreviousTermsVersion is the version TermsVersion replaced; when TermsVersion
// changes, the version it had moves here.
const PreviousTermsVersion = "2026-11-25.2"

// PreviousTermsOverlap is how long after TermsVersion is first served the
// version before it is still accepted: a change reaches an open page only when
// it loads again, and one opened just before it should not be refused.
const PreviousTermsOverlap = time.Hour

// Acceptance is someone's agreement to one version of the policies.
type Acceptance struct {
	Version    string
	AcceptedAt time.Time
}

// Terms checks the version someone accepted, knowing since when TermsVersion
// has been served.
type Terms struct {
	servedSince time.Time
}

// NewTerms checks acceptances against TermsVersion, served since servedSince.
func NewTerms(servedSince time.Time) Terms {
	return Terms{servedSince: servedSince}
}

// Check reports whether version may be accepted at now: TermsVersion, or
// PreviousTermsVersion within PreviousTermsOverlap of the change. The client
// sends the version it showed, which is the one recorded, so a page left open
// longer is refused rather than taken as agreement to text never seen.
func (t Terms) Check(version string, now time.Time) error {
	if version == TermsVersion {
		return nil
	}
	if version == PreviousTermsVersion && now.Before(t.servedSince.Add(PreviousTermsOverlap)) {
		return nil
	}
	return ErrTermsNotAccepted
}
