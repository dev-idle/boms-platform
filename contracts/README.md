# contracts/

Fixtures for rules that **both** sides implement. A rule written twice drifts; a
fixture read by both test suites makes the drift fail CI instead of reaching a
customer.

| File | Rule | Read by |
|---|---|---|
| `vietnam-phone-cases.json` | Vietnam mobile number: accepted spellings, stored form, rejections, display form | `backend/internal/shared/utils/phone_test.go` · `frontend/src/lib/validation/phone.test.ts` |

## What belongs here

A fixture belongs here when the same rule is enforced in Go **and** in
TypeScript, and the two must agree exactly — number formats, code shapes, derived
values. Nothing else: no API samples used by one side only, no test data for a
single suite, no generated files.

Everything else stays where it is enforced. Error codes, for instance, are
defined once in `backend/internal/shared/errors/errors.go` and mirrored in
`frontend/src/lib/errors/api-error.ts`; a contract test on the backend fails when
the two disagree, so no fixture is needed.

## Adding a case

1. Add it to the JSON file, never to one of the two test files.
2. Run both suites (`go test ./...` in `backend/`, `pnpm test` in `frontend/`).
   A case the other side does not satisfy is the point: fix the rule, not the
   fixture.

## Adding a fixture

Name it after the rule, in kebab-case, with the cases grouped by the behaviour
they cover. Add a `$comment` naming both implementations, list it in the table
above, and wire the reader on **both** sides in the same change — a fixture only
one suite reads is worse than none, because it looks enforced and is not.

## CI

`backend-ci.yml` and `frontend-ci.yml` both watch `contracts/**`, so changing a
fixture re-runs both pipelines. Neither pipeline can go green alone.
