import { PAGE_TITLES } from "@/lib/metadata/page-title";

/** Heading copy shared by each auth form and its streaming skeleton. */
export const AUTH_FORM_COPY = {
  signIn: {
    description: "Your orders and pickup times, kept in one place.",
    title: PAGE_TITLES.signIn,
  },
  createAccount: {
    description: "Create your account to order for pickup.",
    title: PAGE_TITLES.createAccount,
  },
  forgotPassword: {
    description: "Enter the email you sign in with and we will send you a link to choose a new password.",
    title: PAGE_TITLES.forgotPassword,
  },
  resetPassword: {
    description: "Choose a new password. Saving it signs you out on every device.",
    title: PAGE_TITLES.resetPassword,
  },
  verifyEmail: {
    description: "Confirming your address lets you order and get updates about your orders.",
    title: PAGE_TITLES.verifyEmail,
  },
  unsubscribe: {
    description: "Stopping promotion emails to your address.",
    title: PAGE_TITLES.unsubscribe,
  },
} as const;
