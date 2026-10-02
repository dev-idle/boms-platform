-- name: CreateCustomerProfile :one
-- A customer who agrees to promotions as they sign up agrees now.
INSERT INTO customer_profiles (user_id, display_name, phone, marketing_consent_at)
VALUES (
  sqlc.arg('user_id'),
  sqlc.narg('display_name'),
  sqlc.narg('phone'),
  CASE WHEN sqlc.arg('marketing_opt_in')::boolean THEN now() END
)
RETURNING user_id, display_name, phone, marketing_consent_at, created_at, updated_at;

-- name: GetCustomerProfileByUserID :one
SELECT user_id, display_name, phone, marketing_consent_at, created_at, updated_at
FROM customer_profiles
WHERE user_id = $1;

-- name: UpdateCustomerProfileByUserID :one
-- Agreement to promotions changes only when asked, judged on the row as it is
-- now: given, it keeps the moment it was first given; withdrawn, it is cleared.
UPDATE customer_profiles
SET display_name = sqlc.narg('display_name'),
    phone = sqlc.narg('phone'),
    marketing_consent_at = CASE
      WHEN sqlc.narg('marketing_opt_in')::boolean IS NULL THEN marketing_consent_at
      WHEN sqlc.narg('marketing_opt_in')::boolean THEN COALESCE(marketing_consent_at, now())
    END,
    updated_at = now()
WHERE user_id = sqlc.arg('user_id')
RETURNING user_id, display_name, phone, marketing_consent_at, created_at, updated_at;

-- name: WithdrawMarketingConsent :execrows
UPDATE customer_profiles
SET marketing_consent_at = NULL,
    updated_at = now()
WHERE user_id = $1
  AND marketing_consent_at IS NOT NULL;

-- name: DeleteCustomerProfileByUserID :execrows
DELETE FROM customer_profiles
WHERE user_id = $1;

-- name: EraseCustomerProfile :exec
-- Clears what a customer told us about themselves and the promotions they
-- agreed to; the row stays, empty.
UPDATE customer_profiles
SET display_name = NULL,
    phone = NULL,
    marketing_consent_at = NULL,
    updated_at = now()
WHERE user_id = $1;
