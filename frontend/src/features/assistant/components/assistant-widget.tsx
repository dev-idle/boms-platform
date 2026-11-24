"use client";

import { useId, useRef, useState } from "react";

import { ChatIcon } from "@/components/icons/storefront-icons";
import { Button } from "@/components/ui/button";
import { BRAND } from "@/constants/brand";
import { USER_ROLE } from "@/constants/roles";
import { useAuthStore } from "@/stores/auth-store";

import { useAssistantChat } from "../hooks";
import { AssistantPanel, type AssistantView } from "./assistant-panel";

/**
 * The assistant in the corner of every storefront page, for guests and
 * customers alike; the bakery's staff, who work elsewhere, are not offered it.
 * Its conversation lasts while the visitor moves within the shop's pages, or
 * within their account's: the two are separate layouts, so crossing between
 * them starts it afresh.
 */
export function AssistantWidget() {
  const role = useAuthStore((state) => (state.status === "authenticated" ? (state.user?.role ?? null) : null));
  const [open, setOpen] = useState(false);
  const [view, setView] = useState<AssistantView>("assistant");
  const chat = useAssistantChat(() => setView("staff"));
  const launcherRef = useRef<HTMLButtonElement>(null);
  const panelId = useId();

  if (role !== null && role !== USER_ROLE.customer) {
    return null;
  }

  function close(): void {
    setOpen(false);
    launcherRef.current?.focus();
  }

  return (
    <div className="assistant">
      {open ? (
        <AssistantPanel
          chat={chat}
          id={panelId}
          signedIn={role === USER_ROLE.customer}
          view={view}
          onClose={close}
          onViewChange={setView}
        />
      ) : null}
      <Button
        aria-controls={open ? panelId : undefined}
        aria-expanded={open}
        ref={launcherRef}
        size="sm"
        type="button"
        onClick={() => (open ? close() : setOpen(true))}
      >
        <ChatIcon />
        Ask {BRAND.name}
      </Button>
    </div>
  );
}
