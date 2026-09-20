"use client";

import { useMe } from "../hooks";

/** "Signed in as" line under the account header — reads the cached /me. */
export function CustomerAccountSignedInAs() {
  const me = useMe();

  if (!me.data) {
    return null;
  }

  return (
    <p className="storefront-account-signed-in">
      Signed in as <strong>{me.data.email}</strong>
    </p>
  );
}
