"use client";

import { useId, useRef, useState } from "react";

import { ChatIcon } from "@/components/icons/storefront-icons";
import { Button } from "@/components/ui/button";
import { BRAND } from "@/constants/brand";
import { USER_ROLE } from "@/constants/roles";
import { useGeneralUnread } from "@/features/customer";
import { plural } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

import { useAssistantChat } from "../hooks";
import { AssistantPanel, type AssistantView } from "./assistant-panel";

/**
 * The assistant in the corner of every storefront page, for guests and
 * customers alike; the bakery's staff, who work elsewhere, are not offered it.
 * Its conversation lasts while the visitor moves within the shop's pages, or
 * within their account's: the two are separate layouts, so crossing between
 * them starts it afresh. Replies from our staff a customer has not read are
 * counted on the launcher, which then opens on them.
 */
export function AssistantWidget() {
  const role = useAuthStore((state) => (state.status === "authenticated" ? (state.user?.role ?? null) : null));
  const [open, setOpen] = useState(false);
  const [view, setView] = useState<AssistantView>("assistant");
  const chat = useAssistantChat(() => setView("staff"));
  const launcherRef = useRef<HTMLButtonElement>(null);
  const panelId = useId();
  const signedIn = role === USER_ROLE.customer;
  const unread = useGeneralUnread(signedIn);

  if (role !== null && role !== USER_ROLE.customer) {
    return null;
  }

  function close(): void {
    setOpen(false);
    launcherRef.current?.focus();
  }

  function launch(): void {
    if (unread > 0) {
      setView("staff");
    }
    setOpen(true);
  }

  return (
    <div className="assistant">
      {open ? (
        <AssistantPanel
          chat={chat}
          id={panelId}
          signedIn={signedIn}
          view={view}
          onClose={close}
          onViewChange={setView}
        />
      ) : null}
      <Button
        aria-controls={open ? panelId : undefined}
        aria-expanded={open}
        aria-label={unread > 0 ? `Ask ${BRAND.name}, ${plural(unread, "new message")} from our staff` : undefined}
        ref={launcherRef}
        size="sm"
        type="button"
        onClick={() => (open ? close() : launch())}
      >
        <ChatIcon />
        Ask {BRAND.name}
        {unread > 0 ? <span className="assistant__unread">{unread}</span> : null}
      </Button>
    </div>
  );
}
