"use client";

import { useMe } from "../hooks";

/** "Signed in as" line under the account header — reads the cached /me. */
export function CustomerAccountSignedInAs() {
  const me = useMe();

  // The paragraph keeps its line while /me answers (CSS `min-block-size`), so the
  // email arriving does not push the page down a row.
  return (
    <p aria-busy={!me.data || undefined} className="storefront-account-signed-in">
      {me.data ? (
        <>
          Signed in as <strong>{me.data.email}</strong>
        </>
      ) : null}
    </p>
  );
}
