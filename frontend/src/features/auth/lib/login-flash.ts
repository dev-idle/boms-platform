import { toast } from "sonner";

/** Stable sonner ids — dedupe React Strict Mode double effects and rapid remounts. */
const LOGIN_FLASH_TOAST_IDS = {
  passwordChanged: "login-flash-password-changed",
  accountRegistered: "login-flash-account-registered",
} as const;

const LOGIN_FLASH_MESSAGES = {
  passwordChanged: "Password changed. Sign in again.",
  accountRegistered: "Account created. Sign in to continue.",
} as const;

export type LoginFlashKind = keyof typeof LOGIN_FLASH_MESSAGES;

export function showLoginFlashToast(kind: LoginFlashKind): void {
  toast.success(LOGIN_FLASH_MESSAGES[kind], { id: LOGIN_FLASH_TOAST_IDS[kind] });
}
