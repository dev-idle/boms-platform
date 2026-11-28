import { ApiErrorCode, isApiError } from "@/lib/errors";

/** What an admin reads when an account action is refused: the same words from the list and from the user's page. */
export function accountActionErrorMessage(error: unknown, fallback: string): string {
  if (!isApiError(error)) {
    return fallback;
  }
  if (error.isCannotModifySelf()) {
    return "You cannot perform this action on your own account.";
  }
  if (error.isCannotModifyAdmin()) {
    return "Admin accounts cannot be disabled or have sessions revoked.";
  }
  if (error.code === ApiErrorCode.AccountHasOpenOrders) {
    return "This customer has orders the bakery has not handed over or cancelled yet. Settle those first.";
  }
  return error.message;
}
