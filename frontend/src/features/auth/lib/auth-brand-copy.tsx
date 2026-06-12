/** Shared left-panel copy — stable across login, register, and forgot password. */
export const AUTH_BRAND_DESCRIPTION =
  "Handcrafted pastries and celebration cakes — order online, pick up at your chosen time.";

export function AuthBrandTitle() {
  return (
    <>
      Sweetly baked,{" "}
      <span className="italic text-matcha-500">ready</span> when you are.
    </>
  );
}
