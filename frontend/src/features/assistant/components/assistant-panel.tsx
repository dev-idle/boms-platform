"use client";

import { useEffect, useId, useRef, type KeyboardEvent } from "react";

import { CloseIcon } from "@/components/icons/storefront-icons";
import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { BRAND } from "@/constants/brand";

import type { AssistantChatState } from "../hooks";
import { AssistantChat } from "./assistant-chat";
import { AssistantStaff } from "./assistant-staff";

export type AssistantView = "assistant" | "staff";

type AssistantPanelProps = {
  id: string;
  view: AssistantView;
  chat: AssistantChatState;
  signedIn: boolean;
  onViewChange: (view: AssistantView) => void;
  onClose: () => void;
};

const VIEWS: readonly { value: AssistantView; label: string }[] = [
  { value: "assistant", label: "Assistant" },
  { value: "staff", label: "Staff" },
];

/**
 * The assistant and the conversation with staff, side by side. It floats over
 * the page without locking it, so the visitor can keep shopping; Escape closes it.
 */
export function AssistantPanel({ id, view, chat, signedIn, onViewChange, onClose }: AssistantPanelProps) {
  const titleId = useId();
  const panelRef = useRef<HTMLElement>(null);

  // Opened, it takes the focus, so the keyboard reaches it at once; when the
  // view changes under the focus, the panel takes it back, so Escape still works.
  useEffect(() => {
    if (!panelRef.current?.contains(document.activeElement)) {
      panelRef.current?.focus();
    }
  }, [view]);

  function closeOnEscape(event: KeyboardEvent<HTMLElement>): void {
    if (event.key === "Escape") {
      event.stopPropagation();
      onClose();
    }
  }

  return (
    <section
      aria-labelledby={titleId}
      className="assistant-panel"
      id={id}
      ref={panelRef}
      role="dialog"
      tabIndex={-1}
      onKeyDown={closeOnEscape}
    >
      <header className="assistant-panel__header">
        <h2 className="assistant-panel__title" id={titleId}>
          Ask {BRAND.name}
        </h2>
        <DashboardFilterGroup aria-label="Who answers" options={VIEWS} value={view} onChange={onViewChange} />
        <button aria-label="Close" className="assistant-panel__close" type="button" onClick={onClose}>
          <CloseIcon />
        </button>
      </header>
      <div className="assistant-panel__body">
        {view === "assistant" ? (
          <AssistantChat chat={chat} onTalkToStaff={() => onViewChange("staff")} />
        ) : (
          <AssistantStaff signedIn={signedIn} />
        )}
      </div>
    </section>
  );
}
