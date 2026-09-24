"use client";

import { useMe } from "../hooks";

/** "Signed in as" line under the account header — reads the cached /me. */
export function CustomerAccountSignedInAs() {
  const me = useMe();

  // The line holds its place while /me answers: arriving late, it would push the
  // whole account page down a row.
  return (
    <p aria-busy={!me.data || undefined} className="storefront-account-signed-in">
      {me.data ? (
        <>
          Signed in as <strong>{me.data.email}</strong>
        </>
      ) : (
        " "
      )}
    </p>
  );
}
