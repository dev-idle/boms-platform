"use client";

import {
  useEffect,
  useId,
  useRef,
  useSyncExternalStore,
  type ReactNode,
} from "react";
import { createPortal } from "react-dom";

import { cn } from "@/lib/utils";

export type AppDialogSize = "sm" | "md" | "lg";

type AppDialogProps = {
  children?: ReactNode;
  className?: string;
  closeOnBackdrop?: boolean;
  closeOnEscape?: boolean;
  description?: string;
  footer?: ReactNode;
  isPending?: boolean;
  onClose?: () => void;
  open: boolean;
  panelClassName?: string;
  size?: AppDialogSize;
  title: string;
};

const PANEL_SIZE_CLASS: Record<AppDialogSize, string> = {
  sm: "app-dialog-panel--sm",
  md: "app-dialog-panel--md",
  lg: "app-dialog-panel--lg",
};

const FOCUSABLE_SELECTOR =
  'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

function getFocusableElements(container: HTMLElement): HTMLElement[] {
  return Array.from(
    container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR),
  ).filter((element) => element.tabIndex !== -1);
}

function subscribeNoop(): () => void {
  return () => {};
}

function getClientSnapshot(): boolean {
  return true;
}

function getServerSnapshot(): boolean {
  return false;
}

export function AppDialog({
  children,
  className,
  closeOnBackdrop = true,
  closeOnEscape = true,
  description,
  footer,
  isPending = false,
  onClose,
  open,
  panelClassName,
  size = "md",
  title,
}: AppDialogProps) {
  const titleId = useId();
  const descriptionId = useId();
  const panelRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);
  const mounted = useSyncExternalStore(
    subscribeNoop,
    getClientSnapshot,
    getServerSnapshot,
  );

  useEffect(() => {
    if (!open) {
      return;
    }

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [open]);

  useEffect(() => {
    if (!open || !closeOnEscape || onClose === undefined) {
      return;
    }

    const closeDialog = onClose;

    function onEscape(event: KeyboardEvent) {
      if (event.key === "Escape" && !isPending) {
        event.preventDefault();
        closeDialog();
      }
    }

    window.addEventListener("keydown", onEscape);
    return () => window.removeEventListener("keydown", onEscape);
  }, [closeOnEscape, isPending, onClose, open]);

  useEffect(() => {
    if (!open) {
      const previous = previousFocusRef.current;
      if (previous?.isConnected) {
        previous.focus();
      }
      previousFocusRef.current = null;
      return;
    }

    previousFocusRef.current = document.activeElement as HTMLElement | null;

    const frame = window.requestAnimationFrame(() => {
      const panel = panelRef.current;
      if (!panel) {
        return;
      }

      getFocusableElements(panel)[0]?.focus();
    });

    return () => {
      window.cancelAnimationFrame(frame);
    };
  }, [open]);

  useEffect(() => {
    if (!open) {
      return;
    }

    function onTab(event: KeyboardEvent) {
      if (event.key !== "Tab") {
        return;
      }

      const panel = panelRef.current;
      if (!panel) {
        return;
      }

      const focusable = getFocusableElements(panel);
      if (focusable.length === 0) {
        return;
      }

      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      const active = document.activeElement;

      if (event.shiftKey && active === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && active === last) {
        event.preventDefault();
        first.focus();
      }
    }

    document.addEventListener("keydown", onTab);
    return () => document.removeEventListener("keydown", onTab);
  }, [open]);

  if (!open || !mounted) {
    return null;
  }

  function handleBackdropClick(): void {
    if (!closeOnBackdrop || isPending || !onClose) {
      return;
    }

    onClose();
  }

  return createPortal(
    <div
      className={cn("app-dialog-overlay", className)}
      onClick={handleBackdropClick}
    >
      <div
        aria-labelledby={titleId}
        aria-modal="true"
        className={cn(
          "app-dialog-panel",
          PANEL_SIZE_CLASS[size],
          panelClassName,
        )}
        onClick={(event) => event.stopPropagation()}
        ref={panelRef}
        role="dialog"
        {...(description ? { "aria-describedby": descriptionId } : {})}
      >
        <header className="app-dialog-header">
          <h2 className="app-dialog-title" id={titleId}>
            {title}
          </h2>
          {description ? (
            <p className="app-dialog-description" id={descriptionId}>
              {description}
            </p>
          ) : null}
        </header>

        {children ? <div className="app-dialog-body">{children}</div> : null}

        {footer ? (
          <footer className="app-dialog-footer">
            <div aria-hidden className="app-dialog-separator" />
            {footer}
          </footer>
        ) : null}
      </div>
    </div>,
    document.body,
  );
}

type AppDialogFooterActionsProps = {
  children: ReactNode;
  className?: string;
};

export function AppDialogFooterActions({
  children,
  className,
}: AppDialogFooterActionsProps) {
  return (
    <div className={cn("app-dialog-footer-actions", className)}>{children}</div>
  );
}
