-- name: CountActiveUsersByRole :many
-- The accounts in use, by role.
SELECT role, count(*)::bigint AS count
FROM users
WHERE deleted_at IS NULL
GROUP BY role;

-- name: AdminDashboardCounts :one
-- The figures the admin home shows besides the roles: accounts disabled (an
-- erased one is gone, not disabled), customers who signed up since week_start,
-- customers who have not confirmed their email, events the outbox has held
-- past undelivered_after_seconds, and refunds asked for past
-- stuck_after_seconds and still not made. Ages run on the database clock.
SELECT
  (SELECT count(*) FROM users WHERE deleted_at IS NOT NULL AND erased_at IS NULL)::bigint AS disabled,
  (SELECT count(*) FROM users
   WHERE deleted_at IS NULL AND role = 'customer'::user_role AND created_at >= sqlc.arg('week_start')::timestamptz)::bigint AS new_customers,
  (SELECT count(*) FROM users
   WHERE deleted_at IS NULL AND role = 'customer'::user_role AND email_verified_at IS NULL)::bigint AS unconfirmed_emails,
  (SELECT count(*) FROM outbox_events
   WHERE published_at IS NULL
     AND created_at < now() - make_interval(secs => sqlc.arg('undelivered_after_seconds')::double precision))::bigint AS undelivered_events,
  (SELECT count(*) FROM payments
   WHERE status = 'captured'::payment_status
     AND refund_requested_at < now() - make_interval(secs => sqlc.arg('stuck_after_seconds')::double precision))::bigint AS stuck_refunds;
