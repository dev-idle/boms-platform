/**
 * The version of the customer policies — terms of sale, privacy policy and
 * refund policy — that registration and checkout accept (backend
 * `domain/policy.TermsVersion`, fixture `contracts/terms-version.json`). The
 * API refuses any other version but the one before it (`PreviousTermsVersion`)
 * for an hour after it first serves this one, so change it only together with
 * the pages, moving the old value into `PreviousTermsVersion`.
 */
export const TERMS_VERSION = "2026-11-28.1";

/** The day the current version took effect, as the policy pages state it. */
export const POLICIES_EFFECTIVE = "November 28, 2026";
