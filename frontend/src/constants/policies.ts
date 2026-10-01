/**
 * The version of the customer policies — terms of sale, privacy policy and
 * refund policy — that registration and checkout accept (backend
 * `domain/policy.TermsVersion`, fixture `contracts/terms-version.json`). The
 * API refuses any other version, so change it only together with the pages.
 */
export const TERMS_VERSION = "2026-10-01.2";

/** The day the current version took effect, as the policy pages state it. */
export const POLICIES_EFFECTIVE = "1 October 2026";
