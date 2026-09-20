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
} as const;
